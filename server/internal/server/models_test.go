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
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
	})
	assert.NoError(t, err)

	_, err = st.CreateModel(store.ModelSpec{
		ModelID:        "m1",
		OrganizationID: orgID,
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
		IsPublished:    true,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
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
	// The total items will now include base models if any are present.
	// Let's add a base model to make this test more robust with the new ListModels logic.
	_, err = st.CreateBaseModel("bm-del-test", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "bm-del-test", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
	assert.NoError(t, err)

	listResp, err = srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	assert.Len(t, listResp.Data, 2) // m1 and bm-del-test
}

func TestListModels_SortingAndPagination(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()
	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	modelsToCreate := []struct {
		id               string
		isBase           bool
		status           v1.ActivationStatus
		sourceRepository v1.SourceRepository // Only for base models
		modelFileLoc     string              // Only for fine-tuned
		baseModelID      string              // Only for fine-tuned
	}{
		// Active Base Models
		{id: "abm1", isBase: true, status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE},
		{id: "abm2", isBase: true, status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE},
		// Active Fine-Tuned Models
		{id: "aftm1", isBase: false, status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, baseModelID: "abm1", modelFileLoc: "s3://bucket/aftm1"},
		{id: "aftm2", isBase: false, status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, baseModelID: "abm2", modelFileLoc: "s3://bucket/aftm2"},
		// Inactive Base Models
		{id: "ibm1", isBase: true, status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OLLAMA},
		{id: "ibm2", isBase: true, status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE},
		// Inactive Fine-Tuned Models
		{id: "iftm1", isBase: false, status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, baseModelID: "ibm1", modelFileLoc: "s3://bucket/iftm1"},
		{id: "iftm2", isBase: false, status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, baseModelID: "ibm2", modelFileLoc: "s3://bucket/iftm2"},
		// Unspecified status models (should be treated as inactive or lowest priority)
		{id: "ubm1", isBase: true, status: v1.ActivationStatus_ACTIVATION_STATUS_UNSPECIFIED, sourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE},
		{id: "uftm1", isBase: false, status: v1.ActivationStatus_ACTIVATION_STATUS_UNSPECIFIED, baseModelID: "ubm1", modelFileLoc: "s3://bucket/uftm1"},
	}

	for _, m := range modelsToCreate {
		if m.isBase {
			_, err := st.CreateBaseModel(m.id, "path/"+m.id, []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf/"+m.id, m.sourceRepository, defaultTenantID)
			assert.NoError(t, err)
		} else {
			// Ensure base model for fine-tuned model exists or CreateModel will fail.
			// For simplicity, we assume base models like "abm1" are created if used.
			// Let's ensure the base models for fine-tuned ones are created first or available.
			// The test data has baseModelID pointing to other models in the list.
			if _, errLookup := st.GetBaseModel(m.baseModelID, defaultTenantID); errors.Is(errLookup, gorm.ErrRecordNotFound) {
				// If base model doesn't exist, create a dummy one for test purposes
				_, errCreateBase := st.CreateBaseModel(m.baseModelID, "path/"+m.baseModelID, []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf/"+m.baseModelID, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, defaultTenantID)
				assert.NoError(t, errCreateBase)
				// Also set its activation status, default to inactive for simplicity if not specified
				errStatus := st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: m.baseModelID, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
				assert.NoError(t, errStatus)
			}

			_, err := st.CreateModel(store.ModelSpec{
				ModelID:           m.id,
				TenantID:          defaultTenantID,
				OrganizationID:    "orgTest",
				ProjectID:         defaultProjectID,
				IsPublished:       true,
				LoadingStatus:     v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
				BaseModelID:       m.baseModelID,
				ModelFileLocation: m.modelFileLoc,
				SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, // Fine-tuned models have specific source repo
			})
			assert.NoError(t, err)
		}
		err := st.CreateModelActivationStatus(&store.ModelActivationStatus{
			ModelID:  m.id,
			TenantID: defaultTenantID,
			Status:   m.status,
		})
		assert.NoError(t, err)
	}

	// Expected order: Active Base (abm1, abm2), Active FT (aftm1, aftm2),
	// Inactive Base (ibm1, ibm2), Inactive FT (iftm1, iftm2),
	// Unspecified Base (ubm1), Unspecified FT (uftm1)
	// IDs are sorted alphabetically within each group by the sort implementation.
	expectedFullOrder := []string{
		"abm1", "abm2", "aftm1", "aftm2", "ibm1", "ibm2", "iftm1", "iftm2", "ubm1", "uftm1",
	}

	tcs := []struct {
		name         string
		req          *v1.ListModelsRequest
		wantModelIDs []string
		wantHasMore  bool
		wantTotal    int32
	}{
		{
			name:         "list all models",
			req:          &v1.ListModelsRequest{Limit: 20}, // Limit greater than total models
			wantModelIDs: expectedFullOrder,
			wantHasMore:  false,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name:         "page 1 limit 3",
			req:          &v1.ListModelsRequest{Limit: 3},
			wantModelIDs: []string{"abm1", "abm2", "aftm1"},
			wantHasMore:  true,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name:         "page 2 limit 3 after aftm1",
			req:          &v1.ListModelsRequest{Limit: 3, After: "aftm1"},
			wantModelIDs: []string{"aftm2", "ibm1", "ibm2"},
			wantHasMore:  true,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name:         "page 3 limit 3 after ibm2",
			req:          &v1.ListModelsRequest{Limit: 3, After: "ibm2"},
			wantModelIDs: []string{"iftm1", "iftm2", "ubm1"},
			wantHasMore:  true,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name:         "page 4 limit 3 after ubm1",
			req:          &v1.ListModelsRequest{Limit: 3, After: "ubm1"},
			wantModelIDs: []string{"uftm1"},
			wantHasMore:  false,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name:         "limit 0 (default page size)",
			req:          &v1.ListModelsRequest{Limit: 0}, // default is 50
			wantModelIDs: expectedFullOrder,               // Assuming defaultPageSize > len(expectedFullOrder)
			wantHasMore:  false,
			wantTotal:    int32(len(expectedFullOrder)),
		},
		{
			name: "invalid after ID",
			req:  &v1.ListModelsRequest{Limit: 5, After: "nonexistent"},
			// wantModelIDs: nil, // Expect error, not specific model IDs
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			got, err := srv.ListModels(ctx, tc.req)
			if tc.name == "invalid after ID" {
				assert.Error(t, err)
				assert.Equal(t, codes.InvalidArgument, status.Code(err))
				return
			}

			assert.NoError(t, err)
			var gotIDs []string
			for _, m := range got.Data {
				gotIDs = append(gotIDs, m.Id)
			}
			assert.Equal(t, tc.wantModelIDs, gotIDs, "Model IDs mismatch")
			assert.Equal(t, tc.wantHasMore, got.HasMore, "HasMore mismatch")
			assert.Equal(t, tc.wantTotal, got.TotalItems, "TotalItems mismatch")
		})
	}
}

func TestDeleteModel_BaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel(
		"m0",
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
	)
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

	const (
		modelID    = "meta-llama-Llama-3.2-1B-Instruct"
		hfRepoName = "meta-llama/Llama-3.2-1B-Instruct"
	)

	_, err := st.CreateBaseModel(
		modelID,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
	)
	assert.NoError(t, err)

	_, err = st.CreateHFModelRepo(hfRepoName, modelID, defaultTenantID)
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{Id: modelID})
	assert.NoError(t, err)

	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{Id: modelID})
	assert.NoError(t, err)

	_, err = srv.GetModel(ctx, &v1.GetModelRequest{Id: modelID})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	hfs, err := st.ListHFModelRepos(defaultTenantID)
	assert.NoError(t, err)
	assert.Empty(t, hfs)
}

func TestDeleteModel_ActiveBaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel(
		"m0",
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
	)
	assert.NoError(t, err)

	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{
		ModelID:  "m0",
		TenantID: defaultTenantID,
		Status:   v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE,
	})
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: "m0",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
}

func TestDeleteModel_ActiveFineTunedModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	_, err := st.CreateModel(store.ModelSpec{
		ModelID:        "m0",
		OrganizationID: "o0",
		ProjectID:      defaultProjectID,
		TenantID:       defaultTenantID,
		IsPublished:    true,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
	})
	assert.NoError(t, err)

	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{
		ModelID:  "m0",
		TenantID: defaultTenantID,
		Status:   v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE,
	})
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())
	_, err = srv.DeleteModel(ctx, &v1.DeleteModelRequest{
		Id: "m0",
	})
	assert.Error(t, err)
	assert.Equal(t, codes.FailedPrecondition, status.Code(err))
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
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
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

	as, err := st.GetModelActivationStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)
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

	as, err := st.GetModelActivationStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

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

	_, err = st.CreateBaseModel(
		baseModelID,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
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

	as, err := st.GetModelActivationStatus(m.Id, defaultTenantID)
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

	_, err = st.CreateBaseModel(
		baseModelID,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
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

	_, err = st.CreateBaseModel(
		baseModelID,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_HUGGING_FACE},
		"",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		defaultTenantID,
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

func TestActivateModelAndDeactivateModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	const modelID = "r/m0"

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	as, err := st.GetModelActivationStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

	// Update the loading status to succeeded.
	err = st.UpdateBaseModelToLoadingStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	err = st.UpdateBaseModelToSucceededStatus(modelID, defaultTenantID, "", nil, "")
	assert.NoError(t, err)

	_, err = srv.ActivateModel(ctx, &v1.ActivateModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	as, err = st.GetModelActivationStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, as.Status)

	_, err = srv.DeactivateModel(ctx, &v1.DeactivateModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	as, err = st.GetModelActivationStatus(modelID, defaultTenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)
}

func TestListModels(t *testing.T) {
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
