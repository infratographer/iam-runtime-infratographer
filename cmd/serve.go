package cmd

import (
	"context"
	"os"
	"os/signal"

	"go.infratographer.com/iam-runtime-infratographer/internal/accesstoken"
	"go.infratographer.com/iam-runtime-infratographer/internal/config"
	"go.infratographer.com/iam-runtime-infratographer/internal/eventsx"
	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"go.infratographer.com/iam-runtime-infratographer/internal/otelx"
	"go.infratographer.com/iam-runtime-infratographer/internal/permissions"
	"go.infratographer.com/iam-runtime-infratographer/internal/server"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "starts the IAM runtime server",
	RunE: func(cmd *cobra.Command, _ []string) error {
		return serve(cmd.Context(), viper.GetViper(), appConfig)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	cmdFlags := serveCmd.Flags()

	otelx.AddFlags(cmdFlags)
	jwt.AddFlags(cmdFlags)
	permissions.AddFlags(cmdFlags)
	eventsx.AddFlags(cmdFlags)
	server.AddFlags(cmdFlags)
	accesstoken.AddFlags(cmdFlags)

	if err := viper.BindPFlags(cmdFlags); err != nil {
		panic(err)
	}
}

func serve(ctx context.Context, _ *viper.Viper, cfg config.Config) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	err := otelx.Initialize(cfg.Tracing, appName)
	if err != nil {
		logger.Fatalw("unable to initialize tracing system", "error", err)
	}

	validator, err := jwt.NewValidator(cfg.JWT)
	if err != nil {
		logger.Fatalw("failed to create validator", "error", err)
	}

	tokenSource, err := accesstoken.NewTokenSource(ctx, cfg.AccessToken)
	if err != nil {
		logger.Fatalw("failed to configure token source", "error", err)
	}

	permClient, err := permissions.NewClient(cfg.Permissions, logger)
	if err != nil {
		logger.Fatalw("failed to create permissions-api client", "error", err)
	}

	publisher, err := eventsx.NewPublisher(cfg.Events)
	if err != nil {
		logger.Fatalw("failed to create events publisher", "error", err)
	}

	iamSrv, err := server.NewServer(cfg.Server, validator, permClient, publisher, tokenSource, logger)
	if err != nil {
		logger.Fatalw("failed to create server", "error", err)
	}

	go func() {
		if err := iamSrv.Listen(); err != nil {
			logger.Fatalw("failed starting server", "error", err)
		}
	}()

	<-c

	logger.Info("signal received, stopping server")

	iamSrv.Stop()

	return nil
}
