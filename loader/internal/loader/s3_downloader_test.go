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

	tcs := []struct {
		name      string
		modelName string
		filename  string
		want      map[string]string
	}{
		{
			name:      "download all objects",
			modelName: "google/gemma-2b",
			filename:  "",
			want: map[string]string{
				"key0":      "object0",
				"key1":      "object1",
				"key2/key3": "object2",
			},
		},
		{
			name:      "download specific file",
			modelName: "google/gemma-2b",
			filename:  "key0",
			want: map[string]string{
				"key0": "object0",
			},
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			d := NewS3Downloader(client, "bucket", "v1/base-models", testr.New(t))
			err = d.download(ctx, tc.modelName, tc.filename, destDir)
			assert.NoError(t, err)

			for key, object := range tc.want {
				b, err := os.ReadFile(filepath.Join(destDir, key))
				assert.NoError(t, err)
				assert.Equal(t, object, string(b))
			}
		})
	}
}

// TestS3Download_IgnoreExactMatch verifies that an object whose key is
// exactly the same as the model path is not downloaded.
func TestS3Download_IgnoreExactMatch(t *testing.T) {
	destDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer func() {
		err := os.RemoveAll(destDir)
		assert.NoError(t, err)
	}()

	client := &fakeS3Client{
		objs: map[string][]byte{
			"v1/base-models/google/gemma-2b":      []byte("object0"),
			"v1/base-models/google/gemma-2b/":     []byte("object1"),
			"v1/base-models/google/gemma-2b/key0": []byte("object2"),
		},
	}

	ctx := context.Background()
	d := NewS3Downloader(client, "bucket", "v1/base-models", testr.New(t))
	err = d.download(ctx, "google/gemma-2b", "", destDir)
	assert.NoError(t, err)

	b, err := os.ReadFile(filepath.Join(destDir, "key0"))
	assert.NoError(t, err)
	assert.Equal(t, "object2", string(b))
}

func TestSplitS3Path(t *testing.T) {
	tcs := []struct {
		path       string
		wantBucket string
		wantPrefix string
		wantErr    bool
	}{
		{
			path:       "s3://bucket/prefix",
			wantBucket: "bucket",
			wantPrefix: "prefix",
		},
		{
			path:    "s3://bucket",
			wantErr: true,
		},
		{
			path:    "foo",
			wantErr: true,
		},
		{
			path:    "https://",
			wantErr: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.path, func(t *testing.T) {
			bucket, prefix, err := splitS3Path(tc.path)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantBucket, bucket)
			assert.Equal(t, tc.wantPrefix, prefix)
		})
	}

}

type fakeS3Client struct {
	objs map[string][]byte
}

func (c *fakeS3Client) Download(ctx context.Context, w io.WriterAt, bucket, key string) error {
	b, ok := c.objs[key]
	if !ok {
		return fmt.Errorf("unknown key: %s", key)
	}
	_, err := w.WriteAt(b, 0)
	return err

}

func (c *fakeS3Client) ListObjectsPages(ctx context.Context, bucket, prefix string, f func(page *s3.ListObjectsV2Output, lastPage bool) bool) error {
	var objs []types.Object
	for key := range c.objs {
		objs = append(objs, types.Object{Key: aws.String(key)})
	}
	f(&s3.ListObjectsV2Output{Contents: objs}, true)
	return nil
}
