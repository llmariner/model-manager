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

// NewInternal creates an internal server.
func NewInternal(store *store.S, pathPrefix string) *IS {
	return &IS{
		store:      store,
		pathPrefix: pathPrefix,
	}
}

// IS is an internal server.
type IS struct {
	v1.UnimplementedModelsInternalServiceServer

	srv *grpc.Server

	store *store.S

	pathPrefix string
}

// Run starts the gRPC server.
func (s *IS) Run(port int) error {
	log.Printf("Starting server on port %d\n", port)

	grpcServer := grpc.NewServer()
	v1.RegisterModelsInternalServiceServer(grpcServer, s)
	reflection.Register(grpcServer)

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
