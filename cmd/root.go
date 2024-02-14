// Package cmd is our cobra/viper cli implementation
package cmd

import (
	"strings"

	"go.infratographer.com/iam-runtime-infratographer/internal/config"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const appName = "iam-runtime-equinix"

var (
	cfgFile   string
	appConfig config.Config
	logger    *zap.SugaredLogger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   appName,
	Short: "Infratographer IAM runtime",
	Long:  "iam-runtime-infratographer is an IAM runtime implementation that uses the Infratographer IAM stack for authentication and authorization.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is /etc/"+appName+"/config.yaml)")

	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")
	viperBindFlag("logging.debug", rootCmd.PersistentFlags().Lookup("debug"))

	rootCmd.PersistentFlags().Bool("pretty", false, "enable pretty (human readable) logging output")
	viperBindFlag("logging.pretty", rootCmd.PersistentFlags().Lookup("pretty"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc/" + appName + "/")
		viper.SetConfigName("config")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	viper.SetEnvPrefix("iamruntime")

	viper.AutomaticEnv() // read in environment variables that match

	setupLogging()

	// If a config file is found, read it in.
	err := viper.ReadInConfig()
	if err == nil {
		logger.Infow("using config file",
			"file", viper.ConfigFileUsed(),
		)
	}

	if err := viper.Unmarshal(&appConfig); err != nil {
		logger.Fatalw("unable to process app config", "error", err.Error())
	}
}

func setupLogging() {
	cfg := zap.NewProductionConfig()
	if viper.GetBool("logging.pretty") {
		cfg = zap.NewDevelopmentConfig()
	}

	if viper.GetBool("logging.debug") {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	logger = l.Sugar().With("app", appName)
	defer logger.Sync() //nolint:errcheck
}

// viperBindFlag provides a wrapper around the viper bindings that handles error checks
func viperBindFlag(name string, flag *pflag.Flag) {
	if err := viper.BindPFlag(name, flag); err != nil {
		panic(err)
	}
}
