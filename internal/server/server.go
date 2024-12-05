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
	"golang.org/x/oauth2"

	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authentication"
	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/authorization"
	"github.com/metal-toolbox/iam-runtime/pkg/iam/runtime/identity"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/attribute"
	tcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
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
	validator   jwt.Validator
	permClient  permissions.Client
	publisher   eventsx.Publisher
	logger      *zap.SugaredLogger
	socketPath  string
	tokenSource oauth2.TokenSource

	grpcSrv *grpc.Server

	authentication.UnimplementedAuthenticationServer
	authorization.UnimplementedAuthorizationServer
	identity.UnimplementedIdentityServer
}

// NewServer creates a new runtime server.
func NewServer(cfg Config, validator jwt.Validator, permClient permissions.Client, publisher eventsx.Publisher, tokenSource oauth2.TokenSource, logger *zap.SugaredLogger) (Server, error) {
	out := &server{
		validator:   validator,
		permClient:  permClient,
		publisher:   publisher,
		logger:      logger,
		socketPath:  cfg.SocketPath,
		tokenSource: tokenSource,
	}

	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	authorization.RegisterAuthorizationServer(grpcSrv, out)
	authentication.RegisterAuthenticationServer(grpcSrv, out)
	identity.RegisterIdentityServer(grpcSrv, out)

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
func (s *server) ValidateCredential(ctx context.Context, req *authentication.ValidateCredentialRequest) (*authentication.ValidateCredentialResponse, error) {
	span := trace.SpanFromContext(ctx)

	s.logger.Info("received ValidateCredential request")

	sub, claims, err := s.validator.ValidateToken(req.Credential)
	if err != nil {
		span.RecordError(err)

		s.logger.Errorw("invalid token", "error", err)

		resp := &authentication.ValidateCredentialResponse{
			Result: authentication.ValidateCredentialResponse_RESULT_INVALID,
		}

		return resp, nil
	}

	claimsStruct, err := structpb.NewStruct(claims)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(tcodes.Error, err.Error())

		return nil, status.Error(codes.FailedPrecondition, err.Error())
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

// GetAccessToken returns a token from the configured token source.
func (s *server) GetAccessToken(ctx context.Context, req *identity.GetAccessTokenRequest) (*identity.GetAccessTokenResponse, error) {
	span := trace.SpanFromContext(ctx)

	s.logger.Infow("received GetAccessToken request")

	token, err := s.tokenSource.Token()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(tcodes.Error, "failed to fetch token from token source: "+err.Error())

		s.logger.Errorw("failed to fetch token from token source", "error", err)

		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := &identity.GetAccessTokenResponse{
		Token: token.AccessToken,
	}

	return resp, nil
}

// CheckAccess takes the given request and sends it to permissions-api, using the given credential
// as a bearer token.
func (s *server) CheckAccess(ctx context.Context, req *authorization.CheckAccessRequest) (*authorization.CheckAccessResponse, error) {
	span := trace.SpanFromContext(ctx)

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
		span.AddEvent("allowed")

		out := &authorization.CheckAccessResponse{
			Result: authorization.CheckAccessResponse_RESULT_ALLOWED,
		}

		return out, nil
	case errors.Is(err, permissions.ErrUnauthenticated):
		span.RecordError(err)

		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, permissions.ErrPermissionDenied):
		span.AddEvent("denied")

		out := &authorization.CheckAccessResponse{
			Result: authorization.CheckAccessResponse_RESULT_DENIED,
		}

		return out, nil
	default:
		span.RecordError(err)
		span.SetStatus(tcodes.Error, "unexpected error: "+err.Error())

		return nil, status.Error(codes.Unavailable, err.Error())
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
	span := trace.SpanFromContext(ctx)

	span.SetAttributes(
		attribute.String("resource.id", resourceIDStr),
		attribute.String("resource.action", string(action)),
		attribute.Int("resource.relationships", len(relationships)),
	)

	resourceID, err := gidx.Parse(resourceIDStr)
	if err != nil {
		span.RecordError(err)

		return err
	}

	relations, err := buildAuthRelations(relationships)
	if err != nil {
		span.RecordError(err)

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
		span.RecordError(err)

		return err
	}

	if err := authResp.Error(); err != nil {
		span.RecordError(err)

		return err
	}

	if errs := authResp.Message().Errors; len(errs) != 0 {
		span.RecordError(errs)

		return errs
	}

	span.AddEvent("relationships published")

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
