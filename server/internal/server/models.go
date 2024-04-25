package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/server/internal/store"
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
	var modelProtos []*v1.Model
	// First include base models.
	bms, err := s.store.ListBaseModels()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}
	for _, m := range bms {
		modelProtos = append(modelProtos, baseToModelProto(m))
	}

	// Then add generated models owned by the tenant.
	ms, err := s.store.ListModelsByTenantID(fakeTenantID, true)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list models: %s", err)
	}

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

	// First check if it's a base model.
	bm, err := s.store.GetBaseModel(req.Id)
	if err == nil {
		return baseToModelProto(bm), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, status.Errorf(codes.Internal, "get model: %s", err)
	}

	// Then check if it's a generated model.
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

// ListBaseModels lists base models.
func (s *S) ListBaseModels(
	ctx context.Context,
	req *v1.ListBaseModelsRequest,
) (*v1.ListBaseModelsResponse, error) {
	ms, err := s.store.ListBaseModels()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list base models: %s", err)
	}
	var modelProtos []*v1.BaseModel
	for _, m := range ms {
		modelProtos = append(modelProtos, toBaseModelProto(m))
	}
	return &v1.ListBaseModelsResponse{
		Object: "list",
		Data:   modelProtos,
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

// CreateBaseModel creates a base model.
func (s *IS) CreateBaseModel(
	ctx context.Context,
	req *v1.CreateBaseModelRequest,
) (*v1.BaseModel, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if req.Path == "" {
		return nil, status.Error(codes.InvalidArgument, "path is required")
	}

	m, err := s.store.CreateBaseModel(req.Id, req.Path)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "create base model: %s", err)
	}

	return toBaseModelProto(m), nil
}

// GetBaseModelPath gets a model path.
func (s *IS) GetBaseModelPath(
	ctx context.Context,
	req *v1.GetBaseModelPathRequest,
) (*v1.GetBaseModelPathResponse, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	m, err := s.store.GetBaseModel(req.Id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "model %q not found", req.Id)
		}
		return nil, status.Errorf(codes.Internal, "create model: %s", err)
	}
	return &v1.GetBaseModelPathResponse{
		Path: m.Path,
	}, nil
}

func (s *IS) genenerateModelID(baseModel, suffix string) (string, error) {
	const randomLength = 10
	// OpenAI uses ':" as a separator, but Ollama does not accept. Use '-' instead for now.
	// TODO(kenji): Revisit this.
	// Replace "/" with "-'. HuggingFace model contains "/", but that doesn't work for Ollama.
	base := fmt.Sprintf("ft:%s:%s-", strings.ReplaceAll(baseModel, "/", "-"), suffix)

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
		OwnedBy: fakeTenantID,
	}
}

func baseToModelProto(m *store.BaseModel) *v1.Model {
	return &v1.Model{
		Id:      m.ModelID,
		Object:  "model",
		Created: m.CreatedAt.UTC().Unix(),
		OwnedBy: "system",
	}
}

func toBaseModelProto(m *store.BaseModel) *v1.BaseModel {
	return &v1.BaseModel{
		Id:      m.ModelID,
		Object:  "basemodel",
		Created: m.CreatedAt.UTC().Unix(),
	}
}
