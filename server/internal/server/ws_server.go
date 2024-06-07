package server

import (
	"fmt"
	"log"
	"net"

	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/server/internal/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// NewWorkerServiceServer creates a new worker service server.
func NewWorkerServiceServer(s *store.S, pathPrefix string) *WS {
	return &WS{
		store:      s,
		pathPrefix: pathPrefix,
	}
}

// WS is a server for worker services.
type WS struct {
	v1.UnimplementedModelsWorkerServiceServer

	srv *grpc.Server

	store *store.S

	pathPrefix string
}

// Run runs the worker service server.
func (ws *WS) Run(port int) error {
	log.Printf("Starting worker service server on port %d", port)

	// TODO(kenji): configure request authN/Z

	srv := grpc.NewServer()
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
