package loader

import (
	"context"

	v1 "github.com/llmariner/model-manager/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// NewFakeModelClient creates a fake model client.
func NewFakeModelClient() *FakeModelClient {
	return &FakeModelClient{
		pathsByID:   map[string]string{},
		ggufsByID:   map[string]string{},
		formatsByID: map[string][]v1.ModelFormat{},

		hfModelRepos: map[string]bool{},
	}
}

// FakeModelClient is a fake model client.
type FakeModelClient struct {
	pathsByID   map[string]string
	ggufsByID   map[string]string
	formatsByID map[string][]v1.ModelFormat

	hfModelRepos map[string]bool
}

// CreateBaseModel creates a base model.
func (c *FakeModelClient) CreateBaseModel(ctx context.Context, in *v1.CreateBaseModelRequest, opts ...grpc.CallOption) (*v1.BaseModel, error) {
	if _, ok := c.pathsByID[in.Id]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", in.Id)
	}
	c.pathsByID[in.Id] = in.Path
	c.ggufsByID[in.Id] = in.GgufModelPath
	c.formatsByID[in.Id] = in.Formats

	return &v1.BaseModel{
		Id: in.Id,
	}, nil
}

// GetBaseModelPath gets the path of a base model.
func (c *FakeModelClient) GetBaseModelPath(ctx context.Context, in *v1.GetBaseModelPathRequest, opts ...grpc.CallOption) (*v1.GetBaseModelPathResponse, error) {
	path, ok := c.pathsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "model %q not found", in.Id)
	}
	ggufPath, ok := c.ggufsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "GGUF model %q not found", in.Id)
	}
	formats, ok := c.formatsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "formats for model %q not found", in.Id)
	}

	return &v1.GetBaseModelPathResponse{
		Path:          path,
		Formats:       formats,
		GgufModelPath: ggufPath,
	}, nil
}

// GetModelPath gets the path of a model.
func (c *FakeModelClient) GetModelPath(ctx context.Context, in *v1.GetModelPathRequest, opts ...grpc.CallOption) (*v1.GetModelPathResponse, error) {
	path, ok := c.pathsByID[in.Id]
	if !ok {
		return nil, status.Errorf(codes.NotFound, "model %q not found", in.Id)
	}
	return &v1.GetModelPathResponse{
		Path: path,
	}, nil
}

// RegisterModel register a model.
func (c *FakeModelClient) RegisterModel(ctx context.Context, in *v1.RegisterModelRequest, opts ...grpc.CallOption) (*v1.RegisterModelResponse, error) {
	if _, ok := c.pathsByID[in.Id]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "model %q already exists", in.Id)
	}
	// path := "models/default-tenant-id/" + in.Id
	c.pathsByID[in.Id] = in.Path
	return &v1.RegisterModelResponse{
		Id:   in.Id,
		Path: in.Path,
	}, nil
}

// PublishModel publishes a model.
func (c *FakeModelClient) PublishModel(ctx context.Context, in *v1.PublishModelRequest, opts ...grpc.CallOption) (*v1.PublishModelResponse, error) {
	return &v1.PublishModelResponse{}, nil
}

// CreateHFModelRepo creates a new HuggingFace model repo.
func (c *FakeModelClient) CreateHFModelRepo(ctx context.Context, in *v1.CreateHFModelRepoRequest, opts ...grpc.CallOption) (*v1.HFModelRepo, error) {
	if _, ok := c.hfModelRepos[in.Name]; ok {
		return nil, status.Errorf(codes.AlreadyExists, "hugging-face model repo %q already exists", in.Name)
	}
	c.hfModelRepos[in.Name] = true
	return &v1.HFModelRepo{Name: in.Name}, nil

}

// GetHFModelRepo returns a HuggingFace model repo.
func (c *FakeModelClient) GetHFModelRepo(ctx context.Context, in *v1.GetHFModelRepoRequest, opts ...grpc.CallOption) (*v1.HFModelRepo, error) {
	if _, ok := c.hfModelRepos[in.Name]; !ok {
		return nil, status.Errorf(codes.NotFound, "hugging-face model repo %q not found", in.Name)
	}
	return &v1.HFModelRepo{Name: in.Name}, nil
}
