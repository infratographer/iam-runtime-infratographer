package server

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"

	"go.infratographer.com/iam-runtime-infratographer/internal/eventsx"
	"go.infratographer.com/iam-runtime-infratographer/internal/jwt"
	"go.infratographer.com/iam-runtime-infratographer/internal/permissions"
	"go.infratographer.com/x/events"
	"go.infratographer.com/x/gidx"

	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authentication"
	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authorization"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

// Server represents an IAM runtime server.
type Server interface {
	Listen() error
	Stop()
}

type server struct {
	validator  jwt.Validator
	permClient permissions.Client
	publisher  eventsx.Publisher
	logger     *zap.SugaredLogger
	socketPath string

	grpcSrv *grpc.Server

	authentication.UnimplementedAuthenticationServer
	authorization.UnimplementedAuthorizationServer
}

// NewServer creates a new runtime server.
func NewServer(cfg Config, validator jwt.Validator, permClient permissions.Client, publisher eventsx.Publisher, logger *zap.SugaredLogger) (Server, error) {
	out := &server{
		validator:  validator,
		permClient: permClient,
		publisher:  publisher,
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

// ValidateCredential ensures that the given credential is a valid JWT issued by the OIDC issuer
// the runtime was configured with.
func (s *server) ValidateCredential(_ context.Context, req *authentication.ValidateCredentialRequest) (*authentication.ValidateCredentialResponse, error) {
	s.logger.Info("received CheckAccess request")

	sub, claims, err := s.validator.ValidateToken(req.Credential)
	if err != nil {
		resp := &authentication.ValidateCredentialResponse{
			Result: authentication.ValidateCredentialResponse_RESULT_INVALID,
		}

		return resp, nil
	}

	claimsStruct, err := structpb.NewStruct(claims)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	resp := &authentication.ValidateCredentialResponse{
		Result: authentication.ValidateCredentialResponse_RESULT_VALID,
		Subject: &authentication.Subject{
			SubjectId: sub,
			Claims:    claimsStruct,
		},
	}

	return resp, nil
}

// CheckAccess takes the given request and sends it to permissions-api, using the given credential
// as a bearer token.
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

	// Per the IAM runtime spec, a 401 from permissions-api should result in an InvalidArgument
	// status. Otherwise, we return a denial if the result was an explicit denial.
	switch {
	case err == nil:
		out := &authorization.CheckAccessResponse{
			Result: authorization.CheckAccessResponse_RESULT_ALLOWED,
		}

		return out, nil
	case errors.Is(err, permissions.ErrUnauthenticated):
		return nil, status.Errorf(codes.InvalidArgument, err.Error())
	case errors.Is(err, permissions.ErrPermissionDenied):
		out := &authorization.CheckAccessResponse{
			Result: authorization.CheckAccessResponse_RESULT_DENIED,
		}

		return out, nil
	default:
		return nil, status.Errorf(codes.Unavailable, err.Error())
	}
}

func buildAuthRelations(rels []*authorization.Relationship) ([]events.AuthRelationshipRelation, error) {
	out := make([]events.AuthRelationshipRelation, len(rels))

	for i, rel := range rels {

		subjID, err := gidx.Parse(rel.SubjectId)
		if err != nil {
			return nil, err
		}

		out[i] = events.AuthRelationshipRelation{
			Relation:  rel.Relation,
			SubjectID: subjID,
		}
	}

	return out, nil
}

func (s *server) publishRelationships(ctx context.Context, action events.AuthRelationshipAction, resourceIDStr string, relationships []*authorization.Relationship) error {
	resourceID, err := gidx.Parse(resourceIDStr)
	if err != nil {
		return err
	}

	relations, err := buildAuthRelations(relationships)
	if err != nil {
		return err
	}

	authReq := events.AuthRelationshipRequest{
		Action:    action,
		ObjectID:  resourceID,
		Relations: relations,
	}

	s.logger.Infow("request", "req", authReq)

	authResp, err := s.publisher.PublishAuthRelationshipRequest(ctx, authReq)
	if err != nil {
		return err
	}

	if authResp.Error() != nil {
		return err
	}

	return nil
}

// CreateRelationships publishes the relationships provided to permissions-api with a write operation
// via NATS and waits for a reply.
func (s *server) CreateRelationships(ctx context.Context, req *authorization.CreateRelationshipsRequest) (*authorization.CreateRelationshipsResponse, error) {
	s.logger.Info("received CreateRelationships request")

	err := s.publishRelationships(ctx, events.WriteAuthRelationshipAction, req.ResourceId, req.Relationships)
	if err != nil {
		return nil, err
	}

	out := &authorization.CreateRelationshipsResponse{}

	return out, nil
}

// CreateRelationships publishes the relationships provided to permissions-api with a delete operation
// via NATS and waits for a reply.
func (s *server) DeleteRelationships(ctx context.Context, req *authorization.DeleteRelationshipsRequest) (*authorization.DeleteRelationshipsResponse, error) {
	s.logger.Info("received DeleteRelationships request")

	err := s.publishRelationships(ctx, events.DeleteAuthRelationshipAction, req.ResourceId, req.Relationships)
	if err != nil {
		return nil, err
	}

	out := &authorization.DeleteRelationshipsResponse{}

	return out, nil
}
