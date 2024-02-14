package server

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"

	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"go.infratographer.com/iam-runtime-infratographer/internal/permissions"

	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authentication"
	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authorization"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server represents an IAM runtime server.
type Server interface {
	Listen() error
	Stop()
}

type server struct {
	validator  jwt.Validator
	permClient permissions.Client
	logger     *zap.SugaredLogger
	socketPath string

	grpcSrv *grpc.Server

	authentication.UnimplementedAuthenticationServer
	authorization.UnimplementedAuthorizationServer
}

// NewServer creates a new runtime server.
func NewServer(cfg Config, validator jwt.Validator, permClient permissions.Client, logger *zap.SugaredLogger) (Server, error) {
	out := &server{
		validator:  validator,
		permClient: permClient,
		logger:     logger,
		socketPath: cfg.SocketPath,
	}

	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	authorization.RegisterAuthorizationServer(grpcSrv, out)
	authentication.RegisterAuthenticationServer(grpcSrv, out)

	out.grpcSrv = grpcSrv

	return out, nil
}

func (s *server) Listen() error {
	if _, err := os.Stat(s.socketPath); err == nil {
		s.logger.Warnw("socket found, unlinking", "socket_path", s.socketPath)

		if err := syscall.Unlink(s.socketPath); err != nil {
			s.logger.Errorw("error unlinking socket", "error", err)
			return err
		}
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		s.logger.Errorw("failed to listen", "error", err)
		return err
	}

	s.logger.Infow("starting server",
		"address", s.socketPath,
	)

	return s.grpcSrv.Serve(listener)
}

func (s *server) Stop() {
	if s.grpcSrv == nil {
		return
	}

	s.grpcSrv.GracefulStop()
}

func (s *server) AuthenticateSubject(_ context.Context, req *authentication.AuthenticateSubjectRequest) (*authentication.AuthenticateSubjectResponse, error) {
	subjClaims, err := s.validator.ValidateToken(req.Credential)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	}

	resp := &authentication.AuthenticateSubjectResponse{
		SubjectClaims: subjClaims,
	}

	return resp, nil
}

func (s *server) CheckAccess(ctx context.Context, req *authorization.CheckAccessRequest) (*authorization.CheckAccessResponse, error) {
	s.logger.Info("received CheckAccess request")

	actions := make([]permissions.RequestAction, 0, len(req.Actions))

	for _, a := range req.Actions {
		action := permissions.RequestAction{
			Action:     a.Action,
			ResourceID: a.ResourceId,
		}
		actions = append(actions, action)
	}

	err := s.permClient.CheckAccess(ctx, req.Credential, actions)

	switch {
	case err == nil:
		return &authorization.CheckAccessResponse{}, nil
	case errors.Is(err, permissions.ErrUnauthenticated):
		return nil, status.Errorf(codes.Unauthenticated, err.Error())
	case errors.Is(err, permissions.ErrPermissionDenied):
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	default:
		return nil, status.Errorf(codes.Unavailable, err.Error())
	}
}
