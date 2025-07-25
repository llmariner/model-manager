package store

import (
	"testing"

	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestModelActivationStatus(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	err := st.CreateModelActivationStatus(&ModelActivationStatus{
		ModelID:  "m0",
		TenantID: "t0",
		Status:   v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE,
	})
	assert.NoError(t, err)

	k := ModelKey{
		ModelID:  "m0",
		TenantID: "t0",
	}
	status, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, status.Status)

	projectKey := ModelKey{
		ModelID:   "m0",
		TenantID:  "t0",
		ProjectID: "p0",
	}
	_, err = st.GetModelActivationStatus(projectKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = st.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE)
	assert.NoError(t, err)

	status, err = st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, status.Status)

	err = st.DeleteModelActivationStatus(k)
	assert.NoError(t, err)

	_, err = st.GetModelActivationStatus(k)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}

func TestModelActivationStatus_WithProject(t *testing.T) {
	st, tearDown := NewTest(t)
	defer tearDown()

	err := st.CreateModelActivationStatus(&ModelActivationStatus{
		ModelID:   "m0",
		TenantID:  "t0",
		ProjectID: "p0",
		Status:    v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE,
	})
	assert.NoError(t, err)

	k := ModelKey{
		ModelID:   "m0",
		TenantID:  "t0",
		ProjectID: "p0",
	}
	status, err := st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, status.Status)

	noProjectKey := ModelKey{
		ModelID:  "m0",
		TenantID: "t0",
		// Empty ProjectID
	}
	_, err = st.GetModelActivationStatus(noProjectKey)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	err = st.UpdateModelActivationStatus(k, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE)
	assert.NoError(t, err)

	status, err = st.GetModelActivationStatus(k)
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, status.Status)

	err = st.DeleteModelActivationStatus(k)
	assert.NoError(t, err)

	_, err = st.GetModelActivationStatus(k)
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
