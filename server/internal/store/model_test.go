package store

import (
	"errors"
	"testing"

	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModel(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID   = "m0"
		tenantID  = "tid0"
		orgID     = "org0"
		projectID = "project0"
	)

	_, err := st.GetPublishedModelByModelIDAndTenantID(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	m0, err := st.CreateModel(ModelSpec{
		ModelID:        modelID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    true,
	})
	assert.NoError(t, err)

	gotM, err := st.GetPublishedModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, modelID, gotM.ModelID)
	assert.Equal(t, tenantID, gotM.TenantID)

	gotMs, err := st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	_, err = st.CreateModel(ModelSpec{
		ModelID:        "m1",
		TenantID:       "tid1",
		OrganizationID: "oid1",
		ProjectID:      "pid1",
		Path:           "path",
		IsPublished:    true,
	})
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	a1, err := st.CreateModel(ModelSpec{
		ModelID:        "a1",
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    true,
	})
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 2)
	assert.Equal(t, a1.ID, gotMs[0].ID)
	assert.Equal(t, m0.ID, gotMs[1].ID)

	c, err := st.CountModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), c)

	err = st.DeleteModel(modelID, tenantID)
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	gotMs, err = st.ListModelsByTenantID(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.Equal(t, a1.ID, gotMs[0].ID)

	gotMs, err = st.ListModelsByTenantID("tid1")
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.Equal(t, "m1", gotMs[0].ModelID)

	err = st.DeleteModel(modelID, tenantID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestModel_Unpublished(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID         = "m0"
		tenantID        = "tid0"
		orgID           = "org0"
		projectID       = "project0"
		defaultPageSize = 10
	)

	_, err := st.CreateModel(ModelSpec{
		ModelID:        modelID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    false,
	})
	assert.NoError(t, err)

	_, err = st.GetPublishedModelByModelIDAndTenantID(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	gotMs, err := st.ListModelsByProjectID(projectID, false)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)

	gotMs, err = st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)
}

func TestListModels_LoadingStatus(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID         = "m0"
		tenantID        = "tid0"
		orgID           = "org0"
		projectID       = "project0"
		defaultPageSize = 10
	)

	_, err := st.CreateModel(ModelSpec{
		ModelID:        modelID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    true,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
	})
	assert.NoError(t, err)

	gotMs, err := st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
}

func TestUpdateModelPublishingStatus(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID   = "m0"
		tenantID  = "tid0"
		orgID     = "org0"
		projectID = "project0"
	)

	_, err := st.CreateModel(ModelSpec{
		ModelID:        modelID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    false,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING,
	})
	assert.NoError(t, err)

	got, err := st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)

	err = st.UpdateModelPublishingStatus(modelID, tenantID, true, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)
	assert.NoError(t, err)
	got, err = st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.True(t, got.IsPublished)
	assert.Equal(t, got.LoadingStatus, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)

	err = st.UpdateModelPublishingStatus(modelID, tenantID, false, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED)
	assert.NoError(t, err)
	got, err = st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)
}

func TestLoadingStatus(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		modelID   = "m0"
		tenantID  = "tid0"
		orgID     = "org0"
		projectID = "project0"
	)

	m, err := st.CreateModel(ModelSpec{
		ModelID:        modelID,
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    false,
		LoadingStatus:  v1.ModelLoadingStatus_MODEL_LOADING_STATUS_REQUESTED,
	})
	assert.NoError(t, err)

	gots, err := st.ListUnloadedModels(tenantID)
	assert.NoError(t, err)
	assert.Len(t, gots, 1)
	assert.Equal(t, modelID, gots[0].ModelID)

	err = st.UpdateModelToLoadingStatus(modelID, tenantID)
	assert.NoError(t, err)

	got, err := st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING, got.LoadingStatus)

	// Calling again.
	err = st.UpdateModelToLoadingStatus(modelID, tenantID)
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrConcurrentUpdate)

	err = st.UpdateModelToSucceededStatus(modelID, tenantID)
	assert.NoError(t, err)

	got, err = st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_SUCCEEDED, got.LoadingStatus)

	// Set the loading status back to loading.
	m.LoadingStatus = v1.ModelLoadingStatus_MODEL_LOADING_STATUS_LOADING
	err = st.db.Save(m).Error
	assert.NoError(t, err)

	err = st.UpdateModelToFailedStatus(modelID, tenantID, "fake-error")
	assert.NoError(t, err)

	got, err = st.GetModelByModelIDAndTenantID(modelID, tenantID)
	assert.NoError(t, err)
	assert.Equal(t, v1.ModelLoadingStatus_MODEL_LOADING_STATUS_FAILED, got.LoadingStatus)
	assert.Equal(t, "fake-error", got.LoadingFailureReason)
}

func TestListModelsByActivationStatusWithPagination(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	const (
		tenantID  = "tid0"
		orgID     = "org0"
		projectID = "project0"
	)

	ids := []string{"m0", "m1", "m2"}
	for i, id := range ids {
		_, err := st.CreateModel(ModelSpec{
			ModelID:        id,
			TenantID:       tenantID,
			OrganizationID: orgID,
			ProjectID:      projectID,
			Path:           "path",
			IsPublished:    true,
		})
		assert.NoError(t, err)
		status := v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE
		if i == 2 {
			status = v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE
		}
		err = st.CreateModelActivationStatus(&ModelActivationStatus{ModelID: id, TenantID: tenantID, Status: status})
		assert.NoError(t, err)
	}

	got, hasMore, err := ListModelsByActivationStatusWithPaginationInTransaction(st.db, projectID, true, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, "", 1, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.True(t, hasMore)
	assert.Equal(t, []string{ids[0]}, []string{got[0].ModelID})

	got, hasMore, err = ListModelsByActivationStatusWithPaginationInTransaction(st.db, projectID, true, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, ids[0], 2, true)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.False(t, hasMore)
	assert.Equal(t, []string{ids[1]}, []string{got[0].ModelID})
}
