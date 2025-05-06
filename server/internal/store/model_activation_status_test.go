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

	status, err := st.GetModelActivationStatus("m0", "t0")
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_ACTIVE, status.Status)

	err = st.UpdateModelActivationStatus("m0", "t0", v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE)
	assert.NoError(t, err)

	status, err = st.GetModelActivationStatus("m0", "t0")
	assert.NoError(t, err)
	assert.Equal(t, v1.ActivationStatus_ACTIVATION_STATUS_INACTIVE, status.Status)

	err = st.DeleteModelActivationStatus("m0", "t0")
	assert.NoError(t, err)

	_, err = st.GetModelActivationStatus("m0", "t0")
	assert.Error(t, err)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
}
