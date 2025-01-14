package server

import (
	"context"
	"errors"

	v1 "github.com/llmariner/model-manager/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// CreateHFModelRepo creates a HuggingFace model repo.
func (s *WS) CreateHFModelRepo(
	ctx context.Context,
	req *v1.CreateHFModelRepoRequest,
) (*v1.HFModelRepo, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	r, err := s.store.CreateHFModelRepo(req.Name, clusterInfo.TenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get hugging-face model repo: %s", err)
	}

	return &v1.HFModelRepo{
		Name: r.Name,
	}, nil
}

// GetHFModelRepo gets a HuggingFace model repo.
func (s *WS) GetHFModelRepo(
	ctx context.Context,
	req *v1.GetHFModelRepoRequest,
) (*v1.HFModelRepo, error) {
	clusterInfo, err := s.extractClusterInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	r, err := s.store.GetHFModelRepo(req.Name, clusterInfo.TenantID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "hugging-face model repo %q not found", req.Name)
		}
		return nil, status.Errorf(codes.Internal, "get hugging-face model repo: %s", err)
	}

	return &v1.HFModelRepo{
		Name: r.Name,
	}, nil
}
