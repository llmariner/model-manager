package store

import (
	"errors"
	"testing"

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

		defaultPageSize = 10
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

	gotMs, hasMore, err := st.ListModelsByProjectIDWithPagination(projectID, true, 0, defaultPageSize)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.False(t, hasMore)

	_, err = st.CreateModel(ModelSpec{
		ModelID:        "m1",
		TenantID:       "tid1",
		OrganizationID: "oid1",
		ProjectID:      "pid1",
		Path:           "path",
		IsPublished:    true,
	})
	assert.NoError(t, err)

	gotMs, hasMore, err = st.ListModelsByProjectIDWithPagination(projectID, true, 0, defaultPageSize)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.False(t, hasMore)

	a1, err := st.CreateModel(ModelSpec{
		ModelID:        "a1",
		TenantID:       tenantID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		Path:           "path",
		IsPublished:    true,
	})
	assert.NoError(t, err)

	gotMs, hasMore, err = st.ListModelsByProjectIDWithPagination(projectID, true, 0, 1)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.True(t, hasMore)
	assert.Equal(t, a1.ID, gotMs[0].ID)

	c, err := st.CountModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), c)

	gotMs, hasMore, err = st.ListModelsByProjectIDWithPagination(projectID, true, a1.ID, 1)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.False(t, hasMore)
	assert.Equal(t, m0.ID, gotMs[0].ID)

	err = st.DeleteModel(modelID, projectID)
	assert.NoError(t, err)

	gotMs, hasMore, err = st.ListModelsByProjectIDWithPagination(projectID, true, 0, defaultPageSize)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.False(t, hasMore)

	err = st.DeleteModel(modelID, projectID)
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

	gotMs, hasMore, err := st.ListModelsByProjectIDWithPagination(projectID, false, 0, defaultPageSize)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 1)
	assert.False(t, hasMore)

	gotMs, hasMore, err = st.ListModelsByProjectIDWithPagination(projectID, true, 0, defaultPageSize)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)
	assert.False(t, hasMore)
}

func TestUpdateModel(t *testing.T) {
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
	})
	assert.NoError(t, err)

	got, err := st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)

	err = st.UpdateModel(modelID, tenantID, true)
	assert.NoError(t, err)
	got, err = st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.True(t, got.IsPublished)

	err = st.UpdateModel(modelID, tenantID, false)
	assert.NoError(t, err)
	got, err = st.GetModelByModelID(modelID)
	assert.NoError(t, err)
	assert.False(t, got.IsPublished)
}
