package server

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID = "m0"
		orgID   = "o0"
	)

	_, err := st.CreateModel(store.ModelSpec{
		ModelID:        modelID,
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	srv := New(st)
	ctx := context.Background()
	getResp, err := srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, modelID, getResp.Id)

	listResp, err := srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 1)
	assert.Equal(t, modelID, listResp.Data[0].Id)

	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listResp, err = srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 0)
}

func TestDeleteModel_BaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("m0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", defaultTenantID)
	assert.NoError(t, err)

	srv := New(st)
	ctx := context.Background()
	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: "m0",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetAndListModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID     = "m0"
		baseModelID = "bm0"
		orgID       = "o0"
	)

	_, err := st.CreateModel(store.ModelSpec{
		ModelID:        modelID,
		TenantID:       defaultTenantID,
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	_, err = st.CreateBaseModel(baseModelID, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", defaultTenantID)
	assert.NoError(t, err)

	srv := New(st)
	ctx := context.Background()
	getResp, err := srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, modelID, getResp.Id)

	getResp, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: baseModelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, baseModelID, getResp.Id)

	listResp, err := srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	var ids []string
	for _, m := range listResp.Data {
		ids = append(ids, m.Id)
	}
	assert.ElementsMatch(t, []string{modelID, baseModelID}, ids)
}

func TestInternalGetModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("model0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", defaultTenantID)
	assert.NoError(t, err)

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        "model1",
		TenantID:       defaultTenantID,
		OrganizationID: "o0",
		ProjectID:      defaultProjectID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	wsrv := NewWorkerServiceServer(st)

	ctx := context.Background()
	got, err := wsrv.GetModel(ctx, &v1.GetModelRequest{
		Id: "model0",
	})
	assert.NoError(t, err)
	assert.Equal(t, "system", got.OwnedBy)

	got, err = wsrv.GetModel(ctx, &v1.GetModelRequest{
		Id: "model1",
	})
	assert.NoError(t, err)
	assert.Equal(t, "user", got.OwnedBy)

	_, err = wsrv.GetModel(ctx, &v1.GetModelRequest{
		Id: "model2",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestRegisterAndPublishModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	wsrv := NewWorkerServiceServer(st)
	ctx := context.Background()

	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	resp, err := wsrv.RegisterModel(ctx, &v1.RegisterModelRequest{
		BaseModel:      "my-model",
		Suffix:         "fine-tuning",
		OrganizationId: "o0",
		ProjectId:      defaultProjectID,
	})
	assert.NoError(t, err)
	modelID := resp.Id
	assert.True(t, strings.HasPrefix(modelID, "ft:my-model:fine-tuning-"))

	m, err := st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.False(t, m.IsPublished)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = wsrv.PublishModel(ctx, &v1.PublishModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
}

func TestGetModelPath(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID = "m0"
		orgID   = "o0"
	)

	wsrv := NewWorkerServiceServer(st)
	ctx := context.Background()
	_, err := wsrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        modelID,
		TenantID:       defaultTenantID,
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		Path:           "model-path",
		IsPublished:    false,
	})
	assert.NoError(t, err)

	_, err = wsrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = wsrv.PublishModel(ctx, &v1.PublishModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	resp, err := wsrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "model-path", resp.Path)
}

func TestBaseModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := context.Background()

	wsrv := NewWorkerServiceServer(st)

	const modelID = "m0"

	_, err := wsrv.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listResp, err := srv.ListBaseModels(ctx, &v1.ListBaseModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 0)

	// Create a base model.
	_, err = wsrv.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
		Id:            modelID,
		Path:          "path",
		GgufModelPath: "gguf-path",
	})
	assert.NoError(t, err)

	getResp, err := wsrv.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "path", getResp.Path)
	assert.ElementsMatch(t, []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, getResp.Formats)
}
