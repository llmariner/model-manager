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

func TestInternalGetModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	k := store.ModelKey{
		ModelID:  "model0",
		TenantID: defaultTenantID,
	}
	_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        "model1",
		TenantID:       defaultTenantID,
		OrganizationID: "o0",
		ProjectID:      defaultProjectID,
		IsPublished:    true,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
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

	m, err := st.GetModelByModelIDAndTenantID(modelID, defaultTenantID)
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
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
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
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
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

	ctx := fakeAuthInto(context.Background())
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	const modelID = "m0"

	_, err := wsrv.GetBaseModelPath(ctx, &v1.GetBaseModelPathRequest{
		Id: modelID,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

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

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	as, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)
}

func TestGetBaseModelPath(t *testing.T) {
	const modelID = "m0"

	tcs := []struct {
		name       string
		createReq  *v1.CreateBaseModelRequest
		getReq     *v1.GetBaseModelPathRequest
		wantPath   string
		wantFormat v1.ModelFormat
		wantErr    bool
	}{
		{
			name: "global-scope model, get without project id",
			createReq: &v1.CreateBaseModelRequest{
				Id:            modelID,
				Path:          "path",
				GgufModelPath: "gguf-path",
			},
			getReq: &v1.GetBaseModelPathRequest{
				Id: modelID,
			},
			wantPath:   "path",
			wantFormat: v1.ModelFormat_MODEL_FORMAT_GGUF,
		},
		{
			name: "global-scope model, get with project id",
			createReq: &v1.CreateBaseModelRequest{
				Id:            modelID,
				Path:          "path",
				GgufModelPath: "gguf-path",
			},
			getReq: &v1.GetBaseModelPathRequest{
				Id:        modelID,
				ProjectId: "p0",
			},
			wantErr: true,
		},
		{
			name: "project-scope model, get without project id",
			createReq: &v1.CreateBaseModelRequest{
				Id:            modelID,
				Path:          "path",
				GgufModelPath: "gguf-path",
				ProjectId:     "p0",
			},
			getReq: &v1.GetBaseModelPathRequest{
				Id: modelID,
			},
			wantPath:   "path",
			wantFormat: v1.ModelFormat_MODEL_FORMAT_GGUF,
		},
		{
			name: "project-scope model, get with project id",
			createReq: &v1.CreateBaseModelRequest{
				Id:            modelID,
				Path:          "path",
				GgufModelPath: "gguf-path",
				ProjectId:     "p0",
			},
			getReq: &v1.GetBaseModelPathRequest{
				Id:        modelID,
				ProjectId: "p0",
			},
			wantPath:   "path",
			wantFormat: v1.ModelFormat_MODEL_FORMAT_GGUF,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			st, tearDown := store.NewTest(t)
			defer tearDown()

			ctx := fakeAuthInto(context.Background())
			wsrv := NewWorkerServiceServer(st, testr.New(t))

			_, err := wsrv.CreateBaseModel(ctx, tc.createReq)
			assert.NoError(t, err)

			getResp, err := wsrv.GetBaseModelPath(ctx, tc.getReq)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantPath, getResp.Path)
			assert.Len(t, getResp.Formats, 1)
			assert.Equal(t, tc.wantFormat, getResp.Formats[0])
		})
	}
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

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	as, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

	resp, err = wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, modelID, resp.BaseModelId)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE, resp.SourceRepository)

	got, err := st.GetBaseModel(k)
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

	got, err = st.GetBaseModel(k)
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

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	got, err := st.GetBaseModel(k)
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
	_, err = st.GetBaseModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestBaseModelCreation_ProjectScoped(t *testing.T) {
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
		IsProjectScoped:  true,
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	k := store.ModelKey{
		ModelID:   modelID,
		ProjectID: defaultProjectID,
		TenantID:  defaultTenantID,
	}
	as, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

	resp, err = wsrv.AcquireUnloadedBaseModel(ctx, &v1.AcquireUnloadedBaseModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, modelID, resp.BaseModelId)
	assert.Equal(t, defaultProjectID, resp.ProjectId)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE, resp.SourceRepository)

	got, err := st.GetBaseModel(k)
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
		ProjectId:     defaultProjectID,
	})
	assert.NoError(t, err)

	_, err = wsrv.UpdateBaseModelLoadingStatus(ctx, &v1.UpdateBaseModelLoadingStatusRequest{
		Id:        modelID,
		ProjectId: defaultProjectID,
		LoadingResult: &v1.UpdateBaseModelLoadingStatusRequest_Success_{
			Success: &v1.UpdateBaseModelLoadingStatusRequest_Success{},
		},
	})
	assert.NoError(t, err)

	got, err = st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, got.LoadingStatus)
	assert.Equal(t, "path", got.Path)
	assert.Equal(t, "gguf-path", got.GGUFModelPath)
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

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	got, err := st.GetBaseModel(k)
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

	got, err = st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, got.LoadingStatus)
	assert.Equal(t, "error", got.LoadingFailureReason)
}

func TestBaseModelCreation_Duplicate(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	const modelID = "r/m0"

	_, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)

	// Send the RPC again.
	_, err = srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.Error(t, err)
	assert.Equal(t, codes.AlreadyExists, status.Code(err))
}

func TestFineTunedModelCreation(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	ctx := fakeAuthInto(context.Background())
	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	// No model to be acquired.
	resp, err := wsrv.AcquireUnloadedModel(ctx, &v1.AcquireUnloadedModelRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.ModelId)

	// Create a base model.
	const baseModelID = "bm0"

	k := store.ModelKey{
		ModelID:  baseModelID,
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(
		k,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		IsFineTunedModel:  true,
		BaseModelId:       baseModelID,
		Suffix:            "suffix0",
		SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		ModelFileLocation: "s3://bucket0/path0",
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)
	assert.Equal(t, "ft:bm0:suffix0", m.Id)

	k = store.ModelKey{
		ModelID:  m.Id,
		TenantID: defaultTenantID,
	}
	as, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

	resp, err = wsrv.AcquireUnloadedModel(ctx, &v1.AcquireUnloadedModelRequest{})
	assert.NoError(t, err)
	assert.Equal(t, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, resp.SourceRepository)
	assert.Equal(t, "s3://bucket0/path0", resp.ModelFileLocation)
	assert.Equal(t, "models/default-tenant-id/ft:bm0:suffix0", resp.DestPath)

	modelID := resp.ModelId

	got, err := st.GetModelByModelIDAndTenantID(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	// No model to be acquired as the model has already been acquired.
	resp, err = wsrv.AcquireUnloadedModel(ctx, &v1.AcquireUnloadedModelRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.ModelId)

	_, err = wsrv.UpdateModelLoadingStatus(ctx, &v1.UpdateModelLoadingStatusRequest{
		Id: modelID,
		LoadingResult: &v1.UpdateModelLoadingStatusRequest_Success_{
			Success: &v1.UpdateModelLoadingStatusRequest_Success{},
		},
	})
	assert.NoError(t, err)

	got, err = st.GetModelByModelIDAndTenantID(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, got.LoadingStatus)
}

func TestFineTunedModelCreation_Failure(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	ctx := fakeAuthInto(context.Background())
	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	// Create a base model.
	const baseModelID = "bm0"

	k := store.ModelKey{
		ModelID:  baseModelID,
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(
		k,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		IsFineTunedModel:  true,
		BaseModelId:       baseModelID,
		Suffix:            "suffix0",
		SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		ModelFileLocation: "s3://bucket0/path0",
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	resp, err := wsrv.AcquireUnloadedModel(ctx, &v1.AcquireUnloadedModelRequest{})
	assert.NoError(t, err)

	modelID := resp.ModelId

	got, err := st.GetModelByModelIDAndTenantID(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	_, err = wsrv.UpdateModelLoadingStatus(ctx, &v1.UpdateModelLoadingStatusRequest{
		Id: modelID,
		LoadingResult: &v1.UpdateModelLoadingStatusRequest_Failure_{
			Failure: &v1.UpdateModelLoadingStatusRequest_Failure{
				Reason: "error",
			},
		},
	})
	assert.NoError(t, err)

	got, err = st.GetModelByModelIDAndTenantID(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, got.LoadingStatus)
	assert.Equal(t, "error", got.LoadingFailureReason)
}

func TestFineTunedModelCreation_CreateModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	ctx := fakeAuthInto(context.Background())
	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	// No model to be acquired.
	resp, err := wsrv.AcquireUnloadedModel(ctx, &v1.AcquireUnloadedModelRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.ModelId)

	// Create a base model.
	const baseModelID = "bm0"

	k := store.ModelKey{
		ModelID:  baseModelID,
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(
		k,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	tcs := []struct {
		name    string
		req     *v1.CreateModelRequest
		wantErr bool
	}{
		{
			name: "success",
			req: &v1.CreateModelRequest{
				IsFineTunedModel:  true,
				BaseModelId:       baseModelID,
				Suffix:            "suffix0",
				SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
				ModelFileLocation: "s3://bucket0/path0",
			},
		},
		{
			name: "no base model",
			req: &v1.CreateModelRequest{
				IsFineTunedModel:  true,
				BaseModelId:       "invalid base model ID",
				Suffix:            "suffix0",
				SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
				ModelFileLocation: "s3://bucket0/path0",
			},
			wantErr: true,
		},
		{
			name: "too long suffix",
			req: &v1.CreateModelRequest{
				IsFineTunedModel:  true,
				BaseModelId:       "invalid base model ID",
				Suffix:            "12345678910234567890",
				SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
				ModelFileLocation: "s3://bucket0/path0",
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			_, err := srv.CreateModel(ctx, tc.req)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestListModels(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	baseModelIDs := []string{"bm0", "bm1"}
	for _, id := range baseModelIDs {
		k := store.ModelKey{
			ModelID:  id,
			TenantID: defaultTenantID,
		}
		_, err := st.CreateBaseModel(
			k,
			"path",
			[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
			"gguf-path",
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		)
		assert.NoError(t, err)
	}

	modelIDs := []string{"m0", "m1", "m2"}
	for _, id := range modelIDs {
		_, err := st.CreateModel(store.ModelSpec{
			ModelID:        id,
			OrganizationID: "o0",
			ProjectID:      defaultProjectID,
			TenantID:       defaultTenantID,
			IsPublished:    true,
			LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		})
		assert.NoError(t, err)
	}

	wsrv := NewWorkerServiceServer(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	got, err := wsrv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)

	wantModelIDs := []string{"bm1", "bm0", "m2", "m1", "m0"}
	assert.Len(t, got.Data, len(wantModelIDs))
	for i, id := range wantModelIDs {
		assert.Equal(t, id, got.Data[i].Id)
	}
}

func TestBaseModelUpdateLoadingStatusMessage(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	wsrv := NewWorkerServiceServer(st, testr.New(t))

	const modelID = "repo/m0"

	_, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)

	_, err = wsrv.UpdateBaseModelLoadingStatus(ctx, &v1.UpdateBaseModelLoadingStatusRequest{
		Id:            modelID,
		StatusMessage: "msg",
	})
	assert.NoError(t, err)

	got, err := st.GetBaseModel(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.NoError(t, err)
	assert.Equal(t, "msg", got.LoadingStatusMessage)
}

func TestFineTunedModelUpdateLoadingStatusMessage(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	// Create a base model.
	const baseModelID = "bm0"

	k := store.ModelKey{
		ModelID:  baseModelID,
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(
		k,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		IsFineTunedModel:  true,
		BaseModelId:       baseModelID,
		Suffix:            "suffix0",
		SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		ModelFileLocation: "s3://bucket0/path0",
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	_, err = wsrv.UpdateModelLoadingStatus(ctx, &v1.UpdateModelLoadingStatusRequest{
		Id:            m.Id,
		StatusMessage: "msg",
	})
	assert.NoError(t, err)

	got, err := st.GetModelByModelIDAndTenantID(m.Id, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, "msg", got.LoadingStatusMessage)
}
