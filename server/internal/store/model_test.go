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
	)

	_, err := st.GetPublishedModelByModelIDAndTenantID(modelID, tenantID)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))

	_, err = st.CreateModel(ModelSpec{
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

	err = st.DeleteModel(modelID, projectID)
	assert.NoError(t, err)

	gotMs, err = st.ListModelsByProjectID(projectID, true)
	assert.NoError(t, err)
	assert.Len(t, gotMs, 0)

	err = st.DeleteModel(modelID, projectID)
	assert.True(t, errors.Is(err, gorm.ErrRecordNotFound))
}

func TestModel_Unpublished(t *testing.T) {
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
