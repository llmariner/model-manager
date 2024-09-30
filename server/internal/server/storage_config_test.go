package server

import (
	"context"
	"testing"

	v1 "github.com/llmariner/model-manager/api/v1"
	"github.com/llmariner/model-manager/server/internal/store"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestStorageConfig(t *testing.T) {
	st, tearDown := store.NewTest(t)
	defer tearDown()

	wsrv := NewWorkerServiceServer(st)
	ctx := context.Background()

	_, err := wsrv.GetStorageConfig(ctx, &v1.GetStorageConfigRequest{})
	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, status.Code(err))

	_, err = wsrv.CreateStorageConfig(ctx, &v1.CreateStorageConfigRequest{
		PathPrefix: "path-prefix",
	})
	assert.NoError(t, err)

	c, err := wsrv.GetStorageConfig(ctx, &v1.GetStorageConfigRequest{})
	assert.NoError(t, err)
	assert.Equal(t, "path-prefix", c.PathPrefix)
}
