package server

import (
	"context"
	"errors"

	gerrors "github.com/llm-operator/common/pkg/gormlib/errors"
	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/server/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// CreateStorageConfig creates a storage config.
func (s *WS) CreateStorageConfig(
	ctx context.Context,
	req *v1.CreateStorageConfigRequest,
) (*v1.StorageConfig, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.PathPrefix == "" {
		return nil, status.Error(codes.InvalidArgument, "path prefix is required")
	}

	c, err := s.store.CreateStorageConfig(clusterInfo.TenantID, req.PathPrefix)
	if err != nil {
		if gerrors.IsUniqueConstraintViolation(err) {
			return nil, status.Errorf(codes.AlreadyExists, "storage config %q already exists", req.PathPrefix)
		}
		return nil, status.Errorf(codes.Internal, "create storage config: %s", err)
	}

	return toStorageConfigProto(c), nil
}

// GetStorageConfig gets a storage config.
func (s *WS) GetStorageConfig(
	ctx context.Context,
	req *v1.GetStorageConfigRequest,
) (*v1.StorageConfig, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}
	c, err := s.store.GetStorageConfig(clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "storage config not found")
		}
		return nil, status.Errorf(codes.Internal, "get storage config: %s", err)
	}
	return toStorageConfigProto(c), nil
}

func toStorageConfigProto(c *store.StorageConfig) *v1.StorageConfig {
	return &v1.StorageConfig{
		PathPrefix: c.PathPrefix,
	}
}
