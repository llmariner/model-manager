package server

import (
	"context"
	"errors"
	"fmt"

	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/common/pkg/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"

	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	fakeTenantID = "fake-tenant-id"
)

// ListModels lists models.
func (s *S) ListModels(
	ctx context.Context,
	req *v1.ListModelsRequest,
) (*v1.ListModelsResponse, error) {
	ms, err := s.store.ListModelsByTenantID(fakeTenantID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}

	var modelProtos []*v1.Model
	for _, m := range ms {
		modelProtos = append(modelProtos, toModelProto(m))
	}
	return &v1.ListModelsResponse{
		Object: "list",
		Data:   modelProtos,
	}, nil
}

// GetModel gets a model.
func (s *S) GetModel(
	ctx context.Context,
	req *v1.GetModelRequest,
) (*v1.Model, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	m, err := s.store.GetModel(store.ModelKey{
		ModelID:  req.Id,
		TenantID: fakeTenantID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "get model: %s", err)
	}
	return toModelProto(m), nil
}

// DeleteModel deletes a model.
func (s *S) DeleteModel(
	ctx context.Context,
	req *v1.DeleteModelRequest,
) (*v1.DeleteModelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	if err := s.store.DeleteModel(store.ModelKey{
		ModelID:  req.Id,
		TenantID: fakeTenantID,
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "delete model: %s", err)
	}
	return &v1.DeleteModelResponse{
		Id:      req.Id,
		Object:  "model",
		Deleted: true,
	}, nil
}

func (s *IS) CreateModel(
	ctx context.Context,
	req *v1.CreateModelRequest,
) (*v1.Model, error) {
	if req.BaseModel == "" {
		return nil, status.Error(codes.InvalidArgument, "base_model is required")
	}
	if req.Suffix == "" {
		return nil, status.Error(codes.InvalidArgument, "suffix is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	id, err := s.genenerateModelID(req.BaseModel, req.Suffix)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "generate model ID: %s", err)
	}

	m, err := s.store.CreateModel(store.ModelKey{
		ModelID:  id,
		TenantID: req.TenantId,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}

	return toModelProto(m), nil
}

func (s *IS) genenerateModelID(baseModel, suffix string) (string, error) {
	const randomLength = 5
	base := fmt.Sprintf("ft:%s:%s:", baseModel, suffix)

	// Randomly create an ID and retry if it already exists.
	for {

		id := fmt.Sprintf("%s%s", base, rand.SafeEncodeString(rand.String(randomLength)))
		if _, err := s.store.GetModel(store.ModelKey{
			ModelID:  id,
			TenantID: fakeTenantID,
		}); errors.Is(err, gorm.ErrRecordNotFound) {
			return id, nil
		}
	}
}

func toModelProto(m *store.Model) *v1.Model {
	return &v1.Model{
		Id:      m.ModelID,
		Object:  "model",
		Created: m.CreatedAt.UTC().Unix(),
		// TODO(kenji): Set OwnedBy
	}
}
