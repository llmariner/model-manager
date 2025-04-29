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
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const (
	defaultClusterID = "default"
)

// NewWorkerServiceServer creates a new worker service server.
func NewWorkerServiceServer(s *store.S, log logr.Logger) *WS {
	return &WS{
		store: s,
		log:   log.WithName("worker"),
	}
}

// WS is a server for worker services.
type WS struct {
	v1.UnimplementedModelsWorkerServiceServer

	srv   *grpc.Server
	store *store.S
	log   logr.Logger

	enableAuth bool
}

// Run runs the worker service server.
func (ws *WS) Run(ctx context.Context, port int, authConfig config.AuthConfig) error {
	ws.log.Info("Starting worker service server...", "port", port)

	var opts []grpc.ServerOption
	if authConfig.Enable {
		ai, err := auth.NewWorkerInterceptor(ctx, auth.WorkerConfig{
			RBACServerAddr: authConfig.RBACInternalServerAddr,
		})
		if err != nil {
			return err
		}
		opts = append(opts, grpc.ChainUnaryInterceptor(ai.Unary()))
		ws.enableAuth = true
	}

	srv := grpc.NewServer(opts...)
	v1.RegisterModelsWorkerServiceServer(srv, ws)
	reflection.Register(srv)

	ws.srv = srv

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	if err := srv.Serve(l); err != nil {
		return fmt.Errorf("serve: %w", err)
	}
	return nil
}

// Stop stops the worker service server.
func (ws *WS) Stop() {
	ws.srv.Stop()
}

// GracefulStop gracefully stops the worker service server.
func (ws *WS) GracefulStop() {
	ws.srv.GracefulStop()
}

func (ws *WS) extractClusterInfoFromContext(ctx context.Context) (*auth.ClusterInfo, error) {
	if !ws.enableAuth {
		return &auth.ClusterInfo{
			ClusterID: defaultClusterID,
			TenantID:  defaultTenantID,
		}, nil
	}
	clusterInfo, ok := auth.ExtractClusterInfoFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "user info not found")
	}
	return clusterInfo, nil
}
