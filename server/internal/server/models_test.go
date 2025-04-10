package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/go-logr/logr/testr"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func TestModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID = "m0"
		orgID   = "o0"
	)

	m0, err := st.CreateModel(store.ModelSpec{
		ModelID:        modelID,
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        "m1",
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
	getResp, err := srv.GetModel(ctx, &v1.GetModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, modelID, getResp.Id)

	listResp, err := srv.ListModels(ctx, &v1.ListModelsRequest{
		Limit: 1,
	})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 1)
	assert.Equal(t, m0.ModelID, listResp.Data[0].Id)
	assert.Equal(t, int32(2), listResp.TotalItems)
	assert.True(t, listResp.HasMore)

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
	assert.Len(t, listResp.Data, 1)
}

func TestModels_Pagination(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	baseModelIDs := []string{"bm0", "bm1"}
	for _, id := range baseModelIDs {
		_, err := st.CreateBaseModel(
			id,
			"path",
			[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
			"gguf-path",
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			defaultTenantID,
		)
		assert.NoError(t, err)
	}

	const orgID = "o0"
	modelIDs := []string{"m0", "m1", "m2"}
	for _, id := range modelIDs {
		_, err := st.CreateModel(store.ModelSpec{
			ModelID:        id,
			OrganizationID: orgID,
			ProjectID:      defaultProjectID,
			TenantID:       defaultTenantID,
			IsPublished:    true,
		})
		assert.NoError(t, err)
	}

	tcs := []struct {
		name         string
		req          *v1.ListModelsRequest
		wantModelIDs []string
		wantHasMore  bool
	}{
		{
			name: "page 0 with limit 2",
			req: &v1.ListModelsRequest{
				Limit: 2,
			},
			wantModelIDs: []string{
				"bm0", "bm1",
			},
			wantHasMore: true,
		},
		{
			name: "page 1 with limit 2",
			req: &v1.ListModelsRequest{
				Limit: 2,
				After: "bm1",
			},
			wantModelIDs: []string{
				"m0", "m1",
			},
			wantHasMore: true,
		},
		{
			name: "page 2 with limit 2",
			req: &v1.ListModelsRequest{
				Limit: 2,
				After: "m1",
			},
			wantModelIDs: []string{
				"m2",
			},
			wantHasMore: false,
		},
		{
			name: "page 0 with limit 3",
			req: &v1.ListModelsRequest{
				Limit: 3,
			},
			wantModelIDs: []string{
				"bm0", "bm1", "m0",
			},
			wantHasMore: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			srv := New(st, testr.New(t))
			ctx := fakeAuthInto(context.Background())

			got, err := srv.ListModels(ctx, tc.req)
			assert.NoError(t, err)
			assert.Len(t, got.Data, len(tc.wantModelIDs))
			for i, g := range got.Data {
				assert.Equal(t, tc.wantModelIDs[i], g.Id)
			}
			assert.Equal(t, tc.wantHasMore, got.HasMore)
			assert.Equal(t, int32(5), got.TotalItems)
		})
	}
}

func TestDeleteModel_BaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("m0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: "m0",
	})
	assert.NoError(t, err)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: "m0",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	listResp, err := srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 0)
}

func TestDeleteModel_BaseModelAndHFModelRepo(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("repo-m0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
	assert.NoError(t, err)

	_, err = st.CreateHFModelRepo("repo/m0", defaultTenantID)
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: "repo-m0",
	})
	assert.NoError(t, err)

	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: "repo-m0",
	})
	assert.NoError(t, err)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{
		Id: "repo-m0",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	hfs, err := st.ListHFModelRepos(defaultTenantID)
	assert.NoError(t, err)
	assert.Empty(t, hfs)
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

	_, err = st.CreateBaseModel(baseModelID, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
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

func TestIncludeLoadingModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel(
		"bm0",
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
	)
	assert.NoError(t, err)

	_, err = st.CreateBaseModelWithLoadingRequested(
		"bm1",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
	)
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	tcs0 := []struct {
		id      string
		wantErr bool
	}{
		{"bm0", false},
		{"bm1", true},
	}

	for _, tc := range tcs0 {
		_, err := srv.GetModel(ctx, &v1.GetModelRequest{
			Id: tc.id,
		})
		if tc.wantErr {
			assert.Error(t, err)
			assert.Equal(t, codes.NotFound, status.Code(err))
			return
		}
		assert.NoError(t, err)

		// The RPC always succeeds if the loading model is included.
		_, err = srv.GetModel(ctx, &v1.GetModelRequest{
			Id:                  tc.id,
			IncludeLoadingModel: true,
		})
		assert.NoError(t, err)

	}

	tcs1 := []struct {
		includeLoadingModels bool
		wantIDs              []string
	}{
		{includeLoadingModels: false, wantIDs: []string{"bm0"}},
		{includeLoadingModels: true, wantIDs: []string{"bm0", "bm1"}},
	}

	for _, tc := range tcs1 {
		resp, err := srv.ListModels(ctx, &v1.ListModelsRequest{IncludeLoadingModels: tc.includeLoadingModels})
		assert.NoError(t, err)
		var ids []string
		for _, m := range resp.Data {
			ids = append(ids, m.Id)
		}
		assert.ElementsMatch(t, tc.wantIDs, ids)
	}
}

func TestInternalGetModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("model0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
	assert.NoError(t, err)

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        "model1",
		TenantID:       defaultTenantID,
		OrganizationID: "o0",
		ProjectID:      defaultProjectID,
		IsPublished:    true,
	})
	assert.NoError(t, err)

	wsrv := NewWorkerServiceServer(st, testr.New(t))

	ctx := fakeAuthInto(context.Background())
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

	srv := New(st, testr.New(t))
	wsrv := NewWorkerServiceServer(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	resp, err := wsrv.RegisterModel(ctx, &v1.RegisterModelRequest{
		BaseModel:      "my-model",
		Suffix:         "fine-tuning",
		OrganizationId: "o0",
		ProjectId:      defaultProjectID,
		Adapter:        v1.AdapterType_ADAPTER_TYPE_LORA,
		Quantization:   v1.QuantizationType_QUANTIZATION_TYPE_AWQ,
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

	_, err = wsrv.RegisterModel(ctx, &v1.RegisterModelRequest{
		Id:             modelID,
		BaseModel:      "my-model",
		OrganizationId: "o0",
		ProjectId:      defaultProjectID,
		Adapter:        v1.AdapterType_ADAPTER_TYPE_LORA,
		Quantization:   v1.QuantizationType_QUANTIZATION_TYPE_AWQ,
	})
	assert.Error(t, err)
}

func TestGetModelPath(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID = "m0"
		orgID   = "o0"
	)

	wsrv := NewWorkerServiceServer(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
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

func TestGetModelAttributes(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	const (
		modelID = "m0"
		orgID   = "o0"
	)

	wsrv := NewWorkerServiceServer(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
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
		BaseModelID:    "base-model0",
		Adapter:        v1.AdapterType_ADAPTER_TYPE_LORA,
		Quantization:   v1.QuantizationType_QUANTIZATION_TYPE_AWQ,
	})
	assert.NoError(t, err)

	_, err = wsrv.GetModelAttributes(ctx, &v1.GetModelAttributesRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = wsrv.PublishModel(ctx, &v1.PublishModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	resp, err := wsrv.GetModelAttributes(ctx, &v1.GetModelAttributesRequest{
		Id: modelID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "model-path", resp.Path)
	assert.Equal(t, "base-model0", resp.BaseModel)
	assert.Equal(t, v1.AdapterType_ADAPTER_TYPE_LORA, resp.Adapter)
	assert.Equal(t, v1.QuantizationType_QUANTIZATION_TYPE_AWQ, resp.Quantization)
}

func TestBaseModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	wsrv := NewWorkerServiceServer(st, testr.New(t))

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

func TestBaseModelCreation(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	wsrv := NewWorkerServiceServer(st, testr.New(t))

	// No model to be acquired.
	resp, err := wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.BaseModelId)

	const modelID = "r/m0"

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	resp, err = wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, modelID, resp.BaseModelId)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE, resp.SourceRepository)

	got, err := st.GetBaseModel(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	// No model to be acquired as the model has already been acquired.
	resp, err = wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.BaseModelId)

	// Create a base model.
	_, err = wsrv.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
		Id:            modelID,
		Path:          "path",
		GgufModelPath: "gguf-path",
	})
	assert.NoError(t, err)

	_, err = wsrv.UpdateBaseModelLoadingStatus(ctx, &v1.UpdateBaseModelLoadingStatusRequest{
		Id: modelID,
		LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Success_{
			Success: &v1.UpdateBaseModelLoadingStatusRequest_Success{},
		},
	})
	assert.NoError(t, err)

	got, err = st.GetBaseModel(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, got.LoadingStatus)
	assert.Equal(t, "path", got.Path)
	assert.Equal(t, "gguf-path", got.GGUFModelPath)
}

func TestBaseModelCreation_CreateModelOfDifferentID(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	wsrv := NewWorkerServiceServer(st, testr.New(t))

	const modelID = "r/m0"

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	resp, err := wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, modelID, resp.BaseModelId)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE, resp.SourceRepository)

	got, err := st.GetBaseModel(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	// Create base models.
	_, err = wsrv.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
		Id:            modelID + "-model0.gguf",
		Path:          "path0",
		GgufModelPath: "gguf-path0",
	})
	assert.NoError(t, err)

	_, err = wsrv.CreateBaseModel(ctx, &v1.CreateBaseModelRequest{
		Id:            modelID + "-model1.gguf",
		Path:          "path0",
		GgufModelPath: "gguf-path0",
	})
	assert.NoError(t, err)

	_, err = wsrv.UpdateBaseModelLoadingStatus(ctx, &v1.UpdateBaseModelLoadingStatusRequest{
		Id: modelID,
		LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Success_{
			Success: &v1.UpdateBaseModelLoadingStatusRequest_Success{},
		},
	})
	assert.NoError(t, err)

	// The requested model has been deleted.
	_, err = st.GetBaseModel(modelID, defaultTenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestBaseModelCreation_Failure(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	wsrv := NewWorkerServiceServer(st, testr.New(t))

	const modelID = "r/m0"

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	resp, err := wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, modelID, resp.BaseModelId)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE, resp.SourceRepository)

	got, err := st.GetBaseModel(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	_, err = wsrv.UpdateBaseModelLoadingStatus(ctx, &v1.UpdateBaseModelLoadingStatusRequest{
		Id: modelID,
		LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Failure_{
			Failure: &v1.UpdateBaseModelLoadingStatusRequest_Failure{
				Reason: "error",
			},
		},
	})
	assert.NoError(t, err)

	got, err = st.GetBaseModel(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, got.LoadingStatus)
	assert.Equal(t, "error", got.LoadingFailureReason)
}

func TestValidateIdAndSourceRepository(t *testing.T) {
	tcs := []struct {
		name             string
		id               string
		sourceRepository v1.SourceRepository
		wantErr          bool
	}{
		{
			name:             "valid object store",
			id:               "r/m0",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			wantErr:          false,
		},
		{
			name:             "invalid object store",
			id:               "s3://a/b",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			wantErr:          false,
		},
		{
			name:             "valid hugging face model id",
			id:               "r/m0",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			wantErr:          false,
		},
		{
			name:             "valid hugging face model id with file",
			id:               "r/m0/file",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			wantErr:          false,
		},
		{
			name:             "invalid hugging face model id without org",
			id:               "m0",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			wantErr:          true,
		},
		{
			name:             "invalid hugging face model id with empty repo",
			id:               "r/",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
			wantErr:          true,
		},
		{
			name:             "valid ollama id",
			id:               "m0:tag",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA,
			wantErr:          false,
		},
		{
			name:             "invalid ollama id",
			id:               "m0",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA,
			wantErr:          true,
		},
		{
			name:             "invalid ollama id with empty tag",
			id:               "m0:",
			sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA,
			wantErr:          true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := validateIDAndSourceRepository(tc.id, tc.sourceRepository)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
