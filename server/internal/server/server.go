package server

import (
	"context"
	"fmt"
	"net"

	"github.com/go-logr/logr"
	"github.com/llmariner/api-usage/pkg/sender"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/config"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/llmariner/rbac-manager/pkg/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
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

	store *store.S
	log   logr.Logger
}

// Run starts the gRPC server.
func (s *S) Run(ctx context.Context, port int, authConfig config.AuthConfig, usage sender.UsageSetter) error {
	s.log.Info("Starting gRPC server...", "port", port)

	var opt grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewInterceptor(ctx, auth.Config{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
			AccessResource: "api.model",
		})
		if err != nil {
			return err
		}
		opt = grpc.ChainUnaryInterceptor(ai.Unary("/grpc.health.v1.Health/Check"), sender.Unary(usage))
	} else {
		fakeAuth := func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
			return handler(fakeAuthInto(ctx), req)
		}
		opt = grpc.ChainUnaryInterceptor(fakeAuth, sender.Unary(usage))
	}

	grpcServer := grpc.NewServer(opt)
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

// fakeAuthInto sets dummy user info and token into the context.
func fakeAuthInto(ctx context.Context) context.Context {
	return auth.AppendUserInfoToContext(ctx, auth.UserInfo{
		OrganizationID: "default",
		ProjectID:      defaultProjectID,
		AssignedKubernetesEnvs: []auth.AssignedKubernetesEnv{
			{
				ClusterID: defaultClusterID,
				Namespace: "default",
			},
		},
		TenantID: defaultTenantID,
	})
}
