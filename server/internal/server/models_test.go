package server

import (
	"context"
	"testing"

	"github.com/go-logr/logr/testr"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
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

	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: modelID, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "m1", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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
		err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: id, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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
			LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		})
		assert.NoError(t, err)
		err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: id, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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

func TestModels_Pagination_ProjectScoped(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	keys := []store.ModelKey{
		{
			ModelID:  "bm0",
			TenantID: defaultTenantID,
		},
		{
			ModelID:   "bm0",
			ProjectID: defaultProjectID,
			TenantID:  defaultTenantID,
		},
		{
			ModelID:   "bm1",
			ProjectID: defaultProjectID,
			TenantID:  defaultTenantID,
		},
		{
			ModelID:   "bm2",
			ProjectID: "different project",
			TenantID:  defaultTenantID,
		},
	}
	for _, k := range keys {
		_, err := st.CreateBaseModel(
			k,
			"path",
			[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
			"gguf-path",
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		)
		assert.NoError(t, err)
		err = st.CreateModelActivationStatus(&store.ModelActivationStatus{
			ModelID:   k.ModelID,
			ProjectID: k.ProjectID,
			TenantID:  k.TenantID,
			Status:    v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE,
		})
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
			LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED,
		})
		assert.NoError(t, err)
		err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: id, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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

func TestListModels_ActivationOrder(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	// Create base models.
	k0 := store.ModelKey{
		ModelID:  "bm0",
		TenantID: defaultTenantID,
	}
	_, err := st.CreateBaseModel(
		k0,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "bm0", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE})
	assert.NoError(t, err)

	k1 := store.ModelKey{
		ModelID:  "bm1",
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(
		k1,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	// Create fine tuned models.
	for _, id := range []string{"m0", "m1"} {
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

	// Activation statuses
	k := store.ModelKey{
		ModelID:  "bm0",
		TenantID: defaultTenantID,
	}
	err = st.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE)
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "m0", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE})
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "bm1", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "m1", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
	assert.NoError(t, err)

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	resp, err := srv.ListModels(ctx, &v1.ListModelsRequest{})
	assert.NoError(t, err)
	var ids []string
	for _, m := range resp.Data {
		ids = append(ids, m.Id)
	}
	assert.Equal(t, []string{"bm0", "m0", "bm1", "m1"}, ids)
}

func TestDeleteModel_BaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	k := store.ModelKey{
		ModelID:  "m0",
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

	k := store.ModelKey{
		ModelID:  modelID,
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

	r := &store.HFModelRepo{
		Name:     hfRepoName,
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	err = st.CreateHFModelRepo(r)
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

	k := store.ModelKey{
		ModelID:  "m0",
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
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: modelID, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
	assert.NoError(t, err)

	k := store.ModelKey{
		ModelID:  baseModelID,
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf-path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: baseModelID, TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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

	k0 := store.ModelKey{
		ModelID:  "bm0",
		TenantID: defaultTenantID,
	}
	_, err := st.CreateBaseModel(
		k0,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf-path",
		v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
	)
	assert.NoError(t, err)

	k1 := store.ModelKey{
		ModelID:  "bm1",
		TenantID: defaultTenantID,
	}
	_, err = st.CreateBaseModelWithLoadingRequested(k1, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)

	assert.NoError(t, err)
	err = st.CreateModelActivationStatus(&store.ModelActivationStatus{ModelID: "bm1", TenantID: defaultTenantID, Status: v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE})
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

	k := store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	}
	as, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)

	// Update the loading status to succeeded.
	err = st.UpdateBaseModelToLoadingStatus(k)
	assert.NoError(t, err)
	err = st.UpdateBaseModelToSucceededStatus(k, "", nil, "")
	assert.NoError(t, err)

	_, err = srv.ActivateModel(ctx, &v1.ActivateModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	as, err = st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, as.Status)

	_, err = srv.DeactivateModel(ctx, &v1.DeactivateModelRequest{
		Id: modelID,
	})
	assert.NoError(t, err)

	as, err = st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, as.Status)
}

func TestModelConfig_BaseModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	ctx := fakeAuthInto(context.Background())

	const modelID = "r/m0"

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               modelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
		Config: &v1.ModelConfig{
			RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
				Replicas: 2,
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), m.Config.RuntimeConfig.Replicas)

	_, err = st.GetModelConfig(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.NoError(t, err)

	resp, err := srv.ListModels(ctx, &v1.ListModelsRequest{
		Limit:                1,
		IncludeLoadingModels: true,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 1)
	got := resp.Data[0]
	assert.Equal(t, modelID, got.Id)
	assert.NotNil(t, got.Config)
	assert.NotNil(t, got.Config.RuntimeConfig)
	assert.Equal(t, int32(2), got.Config.RuntimeConfig.Replicas)
}

func TestModelConfig_FineTunedModel(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	srv := New(st, testr.New(t))
	wsrv := NewWorkerServiceServer(st, testr.New(t))

	ctx := fakeAuthInto(context.Background())

	_, err := wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "models",
	})
	assert.NoError(t, err)

	const baseModelID = "r/m0"

	bm, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		Id:               baseModelID,
		SourceRepository: v1.SourceRepository_SOURCE_REPOSITORY_HUGGING_FACE,
	})
	assert.NoError(t, err)

	m, err := srv.CreateModel(ctx, &v1.CreateModelRequest{
		SourceRepository:  v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
		IsFineTunedModel:  true,
		BaseModelId:       bm.Id,
		Suffix:            "test",
		ModelFileLocation: "s3://test",
		Config: &v1.ModelConfig{
			RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
				Replicas: 2,
			},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), m.Config.RuntimeConfig.Replicas)

	_, err = st.GetModelConfig(store.ModelKey{
		ModelID:  m.Id,
		TenantID: defaultTenantID,
	})
	assert.NoError(t, err)

	resp, err := srv.ListModels(ctx, &v1.ListModelsRequest{
		Limit:                2,
		IncludeLoadingModels: true,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Data, 2)
	got := resp.Data[1]
	assert.Equal(t, m.Id, got.Id)
	assert.NotNil(t, got.Config)
	assert.NotNil(t, got.Config.RuntimeConfig)
	assert.Equal(t, int32(2), got.Config.RuntimeConfig.Replicas)
}

func TestUpdateModel(t *testing.T) {
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

	_, err = st.GetModelConfig(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	m, err := srv.UpdateModel(ctx, &v1.UpdateModelRequest{
		Model: &v1.Model{
			Id: modelID,
			Config: &v1.ModelConfig{
				RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
					Replicas: 2,
				},
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"config"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(2), m.Config.RuntimeConfig.Replicas)

	c, err := st.GetModelConfig(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.NoError(t, err)
	var conf v1.ModelConfig
	err = proto.Unmarshal(c.EncodedConfig, &conf)
	assert.NoError(t, err)
	assert.Equal(t, int32(2), conf.RuntimeConfig.Replicas)

	m, err = srv.UpdateModel(ctx, &v1.UpdateModelRequest{
		Model: &v1.Model{
			Id: modelID,
			Config: &v1.ModelConfig{
				RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
					Replicas: 3,
				},
			},
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"config"},
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, int32(3), m.Config.RuntimeConfig.Replicas)

	c, err = st.GetModelConfig(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.NoError(t, err)
	err = proto.Unmarshal(c.EncodedConfig, &conf)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), conf.RuntimeConfig.Replicas)

	m, err = srv.UpdateModel(ctx, &v1.UpdateModelRequest{
		Model: &v1.Model{
			Id:     modelID,
			Config: nil,
		},
		UpdateMask: &fieldmaskpb.FieldMask{
			Paths: []string{"config"},
		},
	})
	assert.NoError(t, err)
	// Get the default value.
	assert.Equal(t, int32(1), m.Config.RuntimeConfig.Replicas)

	_, err = st.GetModelConfig(store.ModelKey{
		ModelID:  modelID,
		TenantID: defaultTenantID,
	})
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
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

func TestValidateModelConfig(t *testing.T) {
	tcs := []struct {
		name    string
		c       *v1.ModelConfig
		wantErr bool
	}{
		{
			name: "nil",
			c:    nil,
		},
		{
			name: "valid",
			c: &v1.ModelConfig{
				RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
					Resources: &v1.ModelConfig_RuntimeConfig_Resources{
						Gpu: 1,
					},
					Replicas: 1,
				},
			},
		},
		{
			name: "negative gpu",
			c: &v1.ModelConfig{
				RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
					Resources: &v1.ModelConfig_RuntimeConfig_Resources{
						Gpu: -1,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "zero replica",
			c: &v1.ModelConfig{
				RuntimeConfig: &v1.ModelConfig_RuntimeConfig{
					Replicas: 0,
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := validateModelConfig(tc.c)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}

}
