package store

import (
	"testing"

	"github.com/llm-operator/model-manager/server/internal/gormlib/testdb"
	"github.com/stretchr/testify/assert"
)

// NewTest returns a new test store.
func NewTest(t *testing.T) (*S, func()) {
	db, tearDown := testdb.New(t)
	err := autoMigrate(db)
	assert.NoError(t, err)
	return New(db), tearDown
}