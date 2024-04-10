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
	_, err := st.CreateModel(k)
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

func TestCreatModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := NewInternal(st)
	ctx := context.Background()
	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		BaseModel: "my-model",
		Suffix:    "fine-tuning",
		TenantId:  "tid",
	})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(m.Id, "ft:my-model:fine-tuning:"))
}
