package grpc

import (
	"context"
	"net"
	"time"

	"github.com/hiamthach108/dreon-auth/config"
	"github.com/hiamthach108/dreon-auth/internal/aggregate"
	"github.com/hiamthach108/dreon-auth/internal/errorx"
	"github.com/hiamthach108/dreon-auth/internal/service"
	"github.com/hiamthach108/dreon-auth/pkg/logger"
	authinternal "github.com/hiamthach108/dreon-auth/presentation/grpc/gen/proto"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AuthInternalServer implements AuthInternalServiceServer by delegating to RelationSvc and RoleSvc.
type AuthInternalServer struct {
	authinternal.UnimplementedAuthInternalServiceServer
	relationSvc service.IRelationSvc
	roleSvc     service.IRoleSvc
	logger      logger.ILogger
}

// NewAuthInternalServer creates a new gRPC server implementation.
func NewAuthInternalServer(
	relationSvc service.IRelationSvc,
	roleSvc service.IRoleSvc,
	logger logger.ILogger,
) *AuthInternalServer {
	return &AuthInternalServer{
		relationSvc: relationSvc,
		roleSvc:     roleSvc,
		logger:      logger,
	}
}

// GrantRelationTuple creates a Zanzibar-style relation tuple.
func (s *AuthInternalServer) GrantRelationTuple(ctx context.Context, req *authinternal.GrantRelationTupleRequest) (*authinternal.GrantRelationTupleResponse, error) {
	agg := aggregate.GrantRelationReq{
		Namespace:        req.GetNamespace(),
		ObjectID:         req.GetObjectId(),
		Relation:         req.GetRelation(),
		SubjectNamespace: req.GetSubjectNamespace(),
		SubjectObjectID:  req.GetSubjectObjectId(),
		SubjectRelation:  req.GetSubjectRelation(),
	}
	if req.ExpiresAt != nil {
		t := req.ExpiresAt.AsTime()
		agg.ExpiresAt = &t
	}
	r, err := s.relationSvc.GrantRelation(ctx, agg)
	if err != nil {
		return nil, errToStatus(err)
	}
	resp := &authinternal.GrantRelationTupleResponse{
		Id:               r.ID,
		Namespace:        r.Namespace,
		ObjectId:         r.ObjectID,
		Relation:         r.Relation,
		SubjectNamespace: r.SubjectNamespace,
		SubjectObjectId:  r.SubjectObjectID,
		SubjectRelation:  r.SubjectRelation,
		IsActive:         r.IsActive,
		CreatedAt:        timestamppb.New(r.CreatedAt),
		UpdatedAt:        timestamppb.New(r.UpdatedAt),
	}
	return resp, nil
}

// CheckPermission checks whether a subject has a specific relation on an object.
func (s *AuthInternalServer) CheckPermission(ctx context.Context, req *authinternal.CheckPermissionRequest) (*authinternal.CheckPermissionResponse, error) {
	r, err := s.relationSvc.CheckRelation(ctx, aggregate.CheckRelationReq{
		Namespace:        req.GetNamespace(),
		ObjectID:         req.GetObjectId(),
		Relation:         req.GetRelation(),
		SubjectNamespace: req.GetSubjectNamespace(),
		SubjectObjectID:  req.GetSubjectObjectId(),
	})
	if err != nil {
		return nil, errToStatus(err)
	}
	return &authinternal.CheckPermissionResponse{Allowed: r.Allowed, Reason: &r.Reason}, nil
}

// GetUserPermissions returns the permissions for a user.
func (s *AuthInternalServer) GetUserPermissions(ctx context.Context, req *authinternal.GetUserPermissionsRequest) (*authinternal.GetUserPermissionsResponse, error) {
	permissions, err := s.roleSvc.GetUserPermissions(ctx, req.GetUserId())
	if err != nil {
		return nil, errToStatus(err)
	}
	return &authinternal.GetUserPermissionsResponse{Permissions: permissions}, nil
}

func errToStatus(err error) error {
	if err == nil {
		return nil
	}
	code := errorx.GetCode(err)
	msg := err.Error()
	switch code {
	case errorx.ErrBadRequest, errorx.ErrInvalidPermission, errorx.ErrInvalidTupleFormat, errorx.ErrInvalidRole, errorx.ErrUnprocessable:
		return status.Error(codes.InvalidArgument, msg)
	case errorx.ErrNotFound, errorx.ErrPermissionNotFound, errorx.ErrRoleNotFound, errorx.ErrUserNotFound, errorx.ErrProjectNotFound:
		return status.Error(codes.NotFound, msg)
	case errorx.ErrConflict, errorx.ErrPermissionConflict, errorx.ErrRoleConflict, errorx.ErrUserConflict, errorx.ErrProjectConflict:
		return status.Error(codes.AlreadyExists, msg)
	case errorx.ErrUnauthorized:
		return status.Error(codes.Unauthenticated, msg)
	case errorx.ErrForbidden, errorx.ErrPermissionDenied, errorx.ErrSystemRoleProtected:
		return status.Error(codes.PermissionDenied, msg)
	default:
		return status.Error(codes.Internal, msg)
	}
}

// GRPCServer holds the grpc.Server and config for lifecycle management.
type GRPCServer struct {
	server *grpc.Server
	config *config.AppConfig
	logger logger.ILogger
}

// NewGRPCServer creates and configures the gRPC server with AuthInternalService registered.
func NewGRPCServer(
	cfg *config.AppConfig,
	authInternal *AuthInternalServer,
	logger logger.ILogger,
) *GRPCServer {
	s := grpc.NewServer()
	authinternal.RegisterAuthInternalServiceServer(s, authInternal)
	reflection.Register(s)
	return &GRPCServer{
		server: s,
		config: cfg,
		logger: logger,
	}
}

// RegisterHooks registers the gRPC server with fx lifecycle (start listening on GRPC_PORT).
func RegisterHooks(lc fx.Lifecycle, srv *GRPCServer) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			port := srv.config.Server.GRPCPort
			if port == "" {
				port = "9090"
			}
			addr := net.JoinHostPort(srv.config.Server.Host, port)
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}
			srv.logger.Info("Starting gRPC server", "addr", addr)
			go func() {
				if err := srv.server.Serve(lis); err != nil {
					srv.logger.Fatal("gRPC server failed", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			srv.logger.Info("Shutting down gRPC server...")
			stopped := make(chan struct{})
			go func() {
				srv.server.GracefulStop()
				close(stopped)
			}()
			select {
			case <-ctx.Done():
				srv.server.Stop()
			case <-stopped:
			case <-time.After(5 * time.Second):
				srv.server.Stop()
			}
			return nil
		},
	})
}
