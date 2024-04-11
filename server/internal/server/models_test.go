package server

import (
	"context"
	"strings"
	"testing"

	v1 "github.com/llm-operator/model-manager/api/v1"
	"github.com/llm-operator/model-manager/common/pkg/store"
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
	assert.True(t, strings.HasPrefix(modelID, "ft:my-model:fine-tuning:"))

	m, err := st.GetModel(store.ModelKey{
		ModelID:  modelID,
		TenantID: fakeTenantID,
	}, false)
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
