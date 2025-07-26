package store

import (
	"errors"
	"testing"

	gerrors "github.com/llmariner/common/pkg/gormlib/errors"
	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	k := ModelKey{
		ModelID:  modelID,
		TenantID: tenantID,
	}
	_, err := st.GetBaseModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	gotM, err := st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, "path", gotM.Path)

	pk := ModelKey{
		ModelID:   modelID,
		TenantID:  tenantID,
		ProjectID: "p0",
	}
	_, err = st.GetBaseModel(pk)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	formats, err := UnmarshalModelFormats(gotM.Formats)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, formats)

	dk := ModelKey{
		ModelID:  modelID,
		TenantID: "different_tenant",
	}
	_, err = st.GetBaseModel(dk)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotMs, err := st.ListBaseModelsByTenantID(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.Equal(t, modelID, gotMs[0].ModelID)
	assert.Equal(t, "path", gotMs[0].Path)
	assert.Equal(t, "gguf_model_path", gotMs[0].GGUFModelPath)

	c, err := st.CountBaseModels("", tenantID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), c)

	gotMs, err = st.ListBaseModelsByTenantID("different_tenant")
	assert.NoError(t, err)
	assert.Empty(t, gotMs)
}

func TestBaseModel_UniqueConstraint(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	k0 := ModelKey{
		ModelID:  "m0",
		TenantID: "t0",
	}
	_, err := st.CreateBaseModel(k0, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	_, err = st.CreateBaseModel(k0, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	k1 := ModelKey{
		ModelID:  "m1",
		TenantID: "t0",
	}
	_, err = st.CreateBaseModel(k1, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)
}

func TestDeleteBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	k := ModelKey{
		ModelID:  modelID,
		TenantID: tenantID,
	}
	_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	err = st.DeleteBaseModel(k)
	assert.NoError(t, err)

	_, err = st.GetBaseModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	err = st.DeleteBaseModel(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestListUnloadedBaseModels(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	k0 := ModelKey{
		ModelID:  "m0",
		TenantID: "t0",
	}
	_, err := st.CreateBaseModelWithLoadingRequested(k0, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	k1 := ModelKey{
		ModelID:  "m1",
		TenantID: "t1",
	}
	_, err = st.CreateBaseModelWithLoadingRequested(k1, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	k2 := ModelKey{
		ModelID:  "m2",
		TenantID: "t0",
	}
	_, err = st.CreateBaseModel(k2, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	ms, err := st.ListUnloadedBaseModels("t0")
	assert.NoError(t, err)
	assert.Len(t, ms, 1)
	assert.Equal(t, "m0", ms[0].ModelID)

	ms, err = st.ListUnloadedBaseModels("t1")
	assert.NoError(t, err)
	assert.Len(t, ms, 1)
	assert.Equal(t, "m1", ms[0].ModelID)
}

func TestListLoadingBaseModels(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	k0 := ModelKey{
		ModelID:  "m0",
		TenantID: "t0",
	}
	_, err := st.CreateBaseModelWithLoadingRequested(k0, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	k1 := ModelKey{
		ModelID:  "m1",
		TenantID: "t0",
	}
	_, err = st.CreateBaseModelWithLoadingRequested(k1, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	k2 := ModelKey{
		ModelID:  "m2",
		TenantID: "t1",
	}
	_, err = st.CreateBaseModelWithLoadingRequested(k2, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	err = st.UpdateBaseModelToLoadingStatus(k0)
	assert.NoError(t, err)
	err = st.UpdateBaseModelToLoadingStatus(k2)
	assert.NoError(t, err)

	ms, err := st.ListLoadingBaseModels("t0")
	assert.NoError(t, err)
	assert.Len(t, ms, 1)
	assert.Equal(t, "m0", ms[0].ModelID)

	ms, err = st.ListLoadingBaseModels("t1")
	assert.NoError(t, err)
	assert.Len(t, ms, 1)
	assert.Equal(t, "m2", ms[0].ModelID)
}

func TestUpdateBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	k := ModelKey{
		ModelID:  modelID,
		TenantID: tenantID,
	}
	_, err := st.CreateBaseModelWithLoadingRequested(k, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
	assert.NoError(t, err)

	m, err := st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	err = st.UpdateBaseModelToLoadingStatus(k)
	assert.NoError(t, err)

	m, err = st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, m.LoadingStatus)

	// Failed to update as the current state does not match.
	err = st.UpdateBaseModelToLoadingStatus(k)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConcurrentUpdate))

	err = st.UpdateBaseModelToSucceededStatus(
		k,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf_model_path",
	)
	assert.NoError(t, err)

	m, err = st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, m.LoadingStatus)
	assert.Equal(t, "path", m.Path)
	assert.Equal(t, "gguf_model_path", m.GGUFModelPath)

	// Set the stateus back to Loading.
	m.LoadingStatus = v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING
	err = st.db.Save(m).Error
	assert.NoError(t, err)

	err = st.UpdateBaseModelToFailedStatus(k, "error")
	assert.NoError(t, err)

	m, err = st.GetBaseModel(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, m.LoadingStatus)
	assert.Equal(t, "error", m.LoadingFailureReason)
}

func TestListBaseModelsByActivationStatusWithPagination(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const tenantID = "t0"

	ids := []string{"bm0", "bm1", "bm2"}
	for i, id := range ids {
		k := ModelKey{
			ModelID:  id,
			TenantID: tenantID,
		}
		_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
		assert.NoError(t, err)
		status := v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE
		if i == 2 {
			status = v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE
		}
		err = st.CreateModelActivationStatus(&ModelActivationStatus{ModelID: id, TenantID: tenantID, Status: status})
		assert.NoError(t, err)
	}

	got, hasMore, err := ListBaseModelsByActivationStatusWithPaginationInTransaction(st.db, "", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.True(t, hasMore)
	assert.Equal(t, "bm0", got[0].ModelID)

	got, hasMore, err = ListBaseModelsByActivationStatusWithPaginationInTransaction(st.db, "", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "bm0", 2, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.False(t, hasMore)
}

func TestListBaseModelsByActivationStatusWithPagination_ProjectScoped(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const tenantID = "t0"

	keys := []ModelKey{
		{
			ModelID:  "bm0",
			TenantID: tenantID,
		},
		{
			ModelID:  "bm1",
			TenantID: tenantID,
		},
		{
			ModelID:   "bm0",
			ProjectID: "p0",
			TenantID:  tenantID,
		},
		{
			ModelID:   "bm2",
			ProjectID: "p0",
			TenantID:  tenantID,
		},
		{
			ModelID:  "bm3",
			TenantID: "t1",
		},
	}

	for _, k := range keys {
		_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
		assert.NoError(t, err)

		err = st.CreateModelActivationStatus(&ModelActivationStatus{
			ModelID:   k.ModelID,
			ProjectID: k.ProjectID,
			TenantID:  k.TenantID,
			Status:    v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE,
		})
		assert.NoError(t, err)
	}

	got, hasMore, err := ListBaseModelsByActivationStatusWithPaginationInTransaction(
		st.db, "p0", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.True(t, hasMore)
	assert.Equal(t, "bm0", got[0].ModelID)
	// Get a project-scoped one instead of a global-scoped one.
	assert.Equal(t, "p0", got[0].ProjectID)

	got, hasMore, err = ListBaseModelsByActivationStatusWithPaginationInTransaction(
		st.db, "p0", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "bm0", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.True(t, hasMore)
	assert.Equal(t, "bm1", got[0].ModelID)

	got, hasMore, err = ListBaseModelsByActivationStatusWithPaginationInTransaction(
		st.db, "p0", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "bm1", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.False(t, hasMore)
	assert.Equal(t, "bm2", got[0].ModelID)

	// Query with a different project.
	got, hasMore, err = ListBaseModelsByActivationStatusWithPaginationInTransaction(
		st.db, "p1", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.True(t, hasMore)
	assert.Equal(t, "bm0", got[0].ModelID)
	// Get a global-scoped one.
	assert.Empty(t, got[0].ProjectID)

	got, hasMore, err = ListBaseModelsByActivationStatusWithPaginationInTransaction(
		st.db, "p1", tenantID, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "bm0", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.False(t, hasMore)
	assert.Equal(t, "bm1", got[0].ModelID)
}

func TestListBaseModelsByModelIDAndTenantID(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const tenantID = "t0"

	keys := []ModelKey{
		{
			ModelID:  "bm0",
			TenantID: tenantID,
		},
		{
			ModelID:   "bm0",
			ProjectID: "p0",
			TenantID:  tenantID,
		},
		{
			ModelID:   "bm0",
			ProjectID: "p1",
			TenantID:  tenantID,
		},
		{
			ModelID:  "bm1",
			TenantID: tenantID,
		},
		{
			ModelID:  "bm0",
			TenantID: "t1",
		},
	}

	for _, k := range keys {
		_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
		assert.NoError(t, err)
	}

	tcs := []struct {
		modelID        string
		tenantID       string
		wantProjectIDs []string
	}{
		{"bm0", tenantID, []string{"p1", "p0", ""}},
		{"bm1", tenantID, []string{""}},
		{"bm0", "t1", []string{""}},
	}
	for _, tc := range tcs {
		got, err := st.ListBaseModelsByModelIDAndTenantID(tc.modelID, tc.tenantID)
		assert.NoError(t, err)
		assert.Len(t, got, len(tc.wantProjectIDs))
		for i, projectID := range tc.wantProjectIDs {
			assert.Equal(t, projectID, got[i].ProjectID)
		}
	}
}

func TestCountBaseModels(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const tenantID = "t0"

	keys := []ModelKey{
		{
			ModelID:  "bm0",
			TenantID: tenantID,
		},
		{
			ModelID:  "bm1",
			TenantID: tenantID,
		},
		{
			ModelID:   "bm0",
			ProjectID: "p0",
			TenantID:  tenantID,
		},
		{
			ModelID:   "bm2",
			ProjectID: "p0",
			TenantID:  tenantID,
		},
		{
			ModelID:  "bm3",
			TenantID: "t1",
		},
	}

	for _, k := range keys {
		_, err := st.CreateBaseModel(k, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE)
		assert.NoError(t, err)
	}

	tcs := []struct {
		projectID string
		tenantID  string
		want      int64
	}{
		{
			projectID: "",
			tenantID:  tenantID,
			want:      2, // bm0 and bm1
		},
		{
			projectID: "p0",
			tenantID:  tenantID,
			want:      3, // bm0, bm1, and bm2
		},
		{
			projectID: "p1",
			tenantID:  tenantID,
			want:      2, // bm0 and bm1
		},
		{
			projectID: "",
			tenantID:  "t1",
			want:      1,
		},
	}
	for _, tc := range tcs {
		got, err := st.CountBaseModels(tc.projectID, tc.tenantID)
		assert.NoError(t, err)
		assert.Equal(t, tc.want, got)
	}
}
