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

	baseModelDir = "base-models"
)

// ListModels lists models.
func (s *S) ListModels(
	ctx context.Context,
	req *v1.ListModelsRequest,
) (*v1.ListModelsResponse, error) {
	ms, err := s.store.ListModelsByTenantID(fakeTenantID, true)
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
	}, true)
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

// RegisterModel registers a model.
func (s *IS) RegisterModel(
	ctx context.Context,
	req *v1.RegisterModelRequest,
) (*v1.RegisterModelResponse, error) {
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

	path := fmt.Sprintf("%s/%s/%s", s.pathPrefix, req.TenantId, id)
	_, err = s.store.CreateModel(store.ModelSpec{
		Key: store.ModelKey{
			ModelID:  id,
			TenantID: req.TenantId,
		},
		IsPublished: false,
		Path:        path,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}

	return &v1.RegisterModelResponse{
		Id:   id,
		Path: path,
	}, nil
}

// PublishModel publishes a model.
func (s *IS) PublishModel(
	ctx context.Context,
	req *v1.PublishModelRequest,
) (*v1.PublishModelResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	err := s.store.UpdateModel(
		store.ModelKey{
			ModelID:  req.Id,
			TenantID: req.TenantId,
		},
		true,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}
	return &v1.PublishModelResponse{}, nil
}

// GetModelPath gets a model path.
func (s *IS) GetModelPath(
	ctx context.Context,
	req *v1.GetModelPathRequest,
) (*v1.GetModelPathResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id is required")
	}

	m, err := s.store.GetModel(
		store.ModelKey{
			ModelID:  req.Id,
			TenantID: req.TenantId,
		},
		true,
	)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}
	return &v1.GetModelPathResponse{
		Path: m.Path,
	}, nil
}

// GetBaseModelPath gets a model path.
func (s *IS) GetBaseModelPath(
	ctx context.Context,
	req *v1.GetBaseModelPathRequest,
) (*v1.GetBaseModelPathResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	return &v1.GetBaseModelPathResponse{
		Path: fmt.Sprintf("%s/%s/%s", s.pathPrefix, baseModelDir, req.Id),
	}, nil
}

func (s *IS) genenerateModelID(baseModel, suffix string) (string, error) {
	const randomLength = 10
	// OpenAI uses ':" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	base := fmt.Sprintf("ft:%s:%s-", baseModel, suffix)

	// Randomly create an ID and retry if it already exists.
	for {

		id := fmt.Sprintf("%s%s", base, rand.SafeEncodeString(rand.String(randomLength)))
		if _, err := s.store.GetModel(store.ModelKey{
			ModelID:  id,
			TenantID: fakeTenantID,
		}, false); errors.Is(err, gorm.ErrRecordNotFound) {
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
