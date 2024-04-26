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

	const modelID = "m0"

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: fakeTenantID,
	}
	_, err := st.CreateModel(store.ModelSpec{
		Key:         k,
		IsPublished: true,
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

func TestGetAndListModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID     = "m0"
		baseModelID = "bm0"
	)

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: fakeTenantID,
	}
	_, err := st.CreateModel(store.ModelSpec{
		Key:         k,
		IsPublished: true,
	})
	assert.NoError(t, err)

	_, err = st.CreateBaseModel(baseModelID, "path", "gguf-path")
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

func TestRegisterAndPublishModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	isrv := NewInternal(st, "models")
	ctx := context.Background()
	resp, err := isrv.RegisterModel(ctx, &v1.RegisterModelRequest{
		BaseModel: "my-model",
		Suffix:    "fine-tuning",
		TenantId:  fakeTenantID,
	})
	assert.NoError(t, err)
	modelID := resp.Id
	assert.True(t, strings.HasPrefix(modelID, "ft:my-model:fine-tuning-"))

	m, err := st.GetModel(store.ModelKey{
		ModelID:  modelID,
		TenantID: fakeTenantID,
	}, false)
	assert.NoError(t, err)
	assert.False(t, m.IsPublished)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = isrv.PublishModel(ctx, &v1.PublishModelRequest{
		Id:       modelID,
		TenantId: fakeTenantID,
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

	const modelID = "m0"

	isrv := NewInternal(st, "models")
	ctx := context.Background()
	_, err := isrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id:       modelID,
		TenantId: fakeTenantID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: fakeTenantID,
	}
	_, err = st.CreateModel(store.ModelSpec{
		Key:         k,
		Path:        "model-path",
		IsPublished: false,
	})
	assert.NoError(t, err)

	_, err = isrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id:       modelID,
		TenantId: fakeTenantID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = isrv.PublishModel(ctx, &v1.PublishModelRequest{
		Id:       modelID,
		TenantId: fakeTenantID,
	})
	assert.NoError(t, err)

	resp, err := isrv.GetModelPath(ctx, &v1.GetModelPathRequest{
		Id:       modelID,
		TenantId: fakeTenantID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "model-path", resp.Path)
}

func TestBaseModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st)
	ctx := context.Background()

	isrv := NewInternal(st, "models")

	const modelID = "m0"

	_, err := isrv.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listResp, err := srv.ListBaseModels(ctx, &v1.ListBaseModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 0)

	// Create a base model.
	_, err = isrv.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
		Id:            modelID,
		Path:          "path",
		GgufModelPath: "gguf-path",
	})
	assert.NoError(t, err)

	getResp, err := isrv.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "path", getResp.Path)

	listResp, err = srv.ListBaseModels(ctx, &v1.ListBaseModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 1)
	assert.Equal(t, modelID, listResp.Data[0].Id)
}
