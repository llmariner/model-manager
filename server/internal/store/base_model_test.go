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

	_, err := st.GetBaseModel(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateBaseModel(modelID, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, tenantID)
	assert.NoError(t, err)

	gotM, err := st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, "path", gotM.Path)

	formats, err := UnmarshalModelFormats(gotM.Formats)
	assert.NoError(t, err)
	assert.ElementsMatch(t, []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, formats)

	_, err = st.GetBaseModel(modelID, "different_tenant")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotMs, err := st.ListBaseModels(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.Equal(t, modelID, gotMs[0].ModelID)
	assert.Equal(t, "path", gotMs[0].Path)
	assert.Equal(t, "gguf_model_path", gotMs[0].GGUFModelPath)

	c, err := st.CountBaseModels(tenantID)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), c)

	gotMs, err = st.ListBaseModels("different_tenant")
	assert.NoError(t, err)
	assert.Empty(t, gotMs)
}

func TestListBaseModelsWithPagination(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const tenantID = "t0"

	modelIDs := []string{"m0", "m1", "m2"}

	for _, modelID := range modelIDs {
		_, err := st.CreateBaseModel(
			modelID,
			"path",
			[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
			"gguf_model_path",
			v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE,
			tenantID,
		)
		assert.NoError(t, err)
	}

	// Test without pagination.
	gotMs, err := st.ListBaseModels(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, len(modelIDs))
	for i, got := range gotMs {
		assert.Equal(t, modelIDs[2-i], got.ModelID)
	}

	// Test with pagination.
	tcs := []struct {
		name         string
		afterID      uint
		limit        int
		wantModelIDs []string
		wantHaMore   bool
	}{
		{
			name:         "page 1",
			afterID:      0,
			limit:        2,
			wantModelIDs: []string{"m2", "m1"},
			wantHaMore:   true,
		},
		{
			name:         "page 2",
			afterID:      2,
			limit:        2,
			wantModelIDs: []string{"m0"},
			wantHaMore:   false,
		},
		{
			name:         "page 1 with limit 1",
			afterID:      0,
			limit:        1,
			wantModelIDs: []string{"m2"},
			wantHaMore:   true,
		},
		{
			name:         "page 1 with limit 10",
			afterID:      0,
			limit:        10,
			wantModelIDs: []string{"m2", "m1", "m0"},
			wantHaMore:   false,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			gotMs, gotHasMore, err := st.ListBaseModelsWithPagination(tenantID, tc.afterID, tc.limit)
			assert.NoError(t, err)
			assert.Len(t, gotMs, len(tc.wantModelIDs))
			for i, got := range gotMs {
				assert.Equal(t, tc.wantModelIDs[i], got.ModelID)
			}
			assert.Equal(t, tc.wantHaMore, gotHasMore)
		})
	}
}

func TestBaseModel_UniqueConstraint(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModel("m0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t0")
	assert.NoError(t, err)

	_, err = st.CreateBaseModel("m0", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t0")
	assert.Error(t, err)
	assert.True(t, gerrors.IsUniqueConstraintViolation(err))

	_, err = st.CreateBaseModel("m1", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t0")
	assert.NoError(t, err)
}

func TestDeleteBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	_, err := st.CreateBaseModel(modelID, "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, tenantID)
	assert.NoError(t, err)

	err = st.DeleteBaseModel(modelID, tenantID)
	assert.NoError(t, err)

	_, err = st.GetBaseModel(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	err = st.DeleteBaseModel(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestListUnloadedBaseModels(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	_, err := st.CreateBaseModelWithLoadingRequested("m0", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t0")
	assert.NoError(t, err)

	_, err = st.CreateBaseModelWithLoadingRequested("m1", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t1")
	assert.NoError(t, err)

	_, err = st.CreateBaseModel("m2", "path", []v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF}, "gguf_model_path", v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, "t0")
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

func TestUpdateBaseModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID  = "m0"
		tenantID = "t0"
	)

	_, err := st.CreateBaseModelWithLoadingRequested(modelID, v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE, tenantID)
	assert.NoError(t, err)

	m, err := st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED, m.LoadingStatus)

	err = st.UpdateBaseModelToLoadingStatus(modelID, tenantID)
	assert.NoError(t, err)

	m, err = st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, m.LoadingStatus)

	// Failed to update as the current state does not match.
	err = st.UpdateBaseModelToLoadingStatus(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrConcurrentUpdate))

	err = st.UpdateBaseModelToSucceededStatus(
		modelID,
		tenantID,
		"path",
		[]v1.ModelFormat{v1.ModelFormat_MODEL_FORMAT_GGUF},
		"gguf_model_path",
	)
	assert.NoError(t, err)

	m, err = st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, m.LoadingStatus)
	assert.Equal(t, "path", m.Path)
	assert.Equal(t, "gguf_model_path", m.GGUFModelPath)

	// Set the stateus back to Loading.
	m.LoadingStatus = v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING
	err = st.db.Save(m).Error
	assert.NoError(t, err)

	err = st.UpdateBaseModelToFailedStatus(
		modelID,
		tenantID,
		"error",
	)
	assert.NoError(t, err)

	m, err = st.GetBaseModel(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, m.LoadingStatus)
	assert.Equal(t, "error", m.LoadingFailureReason)
}
