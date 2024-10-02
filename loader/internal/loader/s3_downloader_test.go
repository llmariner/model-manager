package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/go-logr/logr/testr"
	"github.com/stretchr/testify/assert"
)

func TestS3Download(t *testing.T) {
	destDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(destDir)
		assert.NoError(t, err)
	}()

	client := &fakeS3Client{
		objs: map[string][]byte{
			"v1/base-models/google/gemma-2b/key0":      []byte("object0"),
			"v1/base-models/google/gemma-2b/key1":      []byte("object1"),
			"v1/base-models/google/gemma-2b/key2/key3": []byte("object2"),
		},
	}
	ctx := context.Background()
	d := NewS3Downloader(client, "v1/base-models", testr.New(t))
	err = d.download(ctx, "google/gemma-2b", destDir)
	assert.NoError(t, err)

	want := map[string]string{
		"key0":      "object0",
		"key1":      "object1",
		"key2/key3": "object2",
	}
	for key, object := range want {
		b, err := os.ReadFile(filepath.Join(destDir, key))
		assert.NoError(t, err)
		assert.Equal(t, object, string(b))
	}
}

type fakeS3Client struct {
	objs map[string][]byte
}

func (c *fakeS3Client) Download(ctx context.Context, w io.WriterAt, key string) error {
	b, ok := c.objs[key]
	if !ok {
		return fmt.Errorf("unknown key: %s", key)
	}
	_, err := w.WriteAt(b, 0)
	return err

}

func (c *fakeS3Client) ListObjectsPages(ctx context.Context, prefix string, f func(page *s3.ListObjectsV2Output, lastPage bool) bool) error {
	var objs []types.Object
	for key := range c.objs {
		objs = append(objs, types.Object{Key: aws.String(key)})
	}
	f(&s3.ListObjectsV2Output{Contents: objs}, true)
	return nil
}
