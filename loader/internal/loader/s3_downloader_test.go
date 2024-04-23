package loader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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
			"v1/base-models/google/gemma-2b/key0": []byte("object0"),
			"v1/base-models/google/gemma-2b/key1": []byte("object1"),
		},
	}
	d := NewS3Downloader(client, "v1/base-models")
	err = d.download("google/gemma-2b", destDir)
	assert.NoError(t, err)

	b, err := os.ReadFile(filepath.Join(destDir, "key0"))
	assert.NoError(t, err)
	assert.Equal(t, "object0", string(b))

	b, err = os.ReadFile(filepath.Join(destDir, "key1"))
	assert.NoError(t, err)
	assert.Equal(t, "object1", string(b))
}

type fakeS3Client struct {
	objs map[string][]byte
}

func (c *fakeS3Client) Download(w io.WriterAt, key string) error {
	b, ok := c.objs[key]
	if !ok {
		return fmt.Errorf("unknown key: %s", key)
	}
	_, err := w.WriteAt(b, 0)
	return err

}

func (c *fakeS3Client) ListObjectsPages(prefix string, f func(page *s3.ListObjectsOutput, lastPage bool) bool) error {
	var objs []*s3.Object
	for key := range c.objs {
		objs = append(objs, &s3.Object{Key: aws.String(key)})
	}
	f(&s3.ListObjectsOutput{Contents: objs}, true)
	return nil
}
