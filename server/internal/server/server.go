package server

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/config"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultProjectID = "default"
	defaultTenantID  = "default-tenant-id"
)

// New creates a server.
func New(store *store.S, log logr.Logger) *S {
	return &S{
		store: store,
		log:   log.WithName("grpc"),
	}
}

// S is a server.
type S struct {
	v1.UnimplementedModelsServiceServer

	srv *grpc.Server

	store      *store.S
	log        logr.Logger
	enableAuth bool
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	s.log.Info("Starting gRPC server...", "port", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.model",
		})
		if err != nil {
			return err
		}
		authFn := ai.Unary()
		healthSkip := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			if info.FullMethod == "/grpc.health.v1.Health/Check" {
				// Skip authentication for health check
				return handler(ctx, req)
			}
			return authFn(ctx, req, info, handler)
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(healthSkip))
		s.enableAuth = true
	}

	grpcServer := grpc.NewServer(opts...)
	v1.RegisterModelsServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

	healthCheck := health.NewServer()
	healthCheck.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthCheck)

	s.srv = grpcServer

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %s", err)
	}
	if err := grpcServer.Serve(l); err != nil {
		return fmt.Errorf("serve: %s", err)
	}
	return nil
}

// Stop stops the gRPC server.
func (s *S) Stop() {
	s.srv.Stop()
}

func (s *S) extractUserInfoFromContext(ctx context.Context) (*auth.UserInfo, error) {
	if !s.enableAuth {
		return &auth.UserInfo{
			OrganizationID: "default",
			ProjectID:      defaultProjectID,
			AssignedKubernetesEnvs: []auth.AssignedKubernetesEnv{
				{
					ClusterID: defaultClusterID,
					Namespace: "default",
				},
			},
			TenantID: defaultTenantID,
		}, nil
	}

	userInfo, ok := auth.ExtractUserInfoFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user info not found")
	}
	return userInfo, nil
}
