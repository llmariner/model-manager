package loader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/go-logr/logr"
)

type s3Client interface {
	Download(ctx context.Context, w io.WriterAt, key string) error
	ListObjectsPages(ctx context.Context, prefix string, f func(page *s3.ListObjectsV2Output, lastPage bool) bool) error
}

// NewS3Downloader returns a new S3Downloader.
func NewS3Downloader(s3Client s3Client, pathPrefix string, log logr.Logger) *S3Downloader {
	return &S3Downloader{
		s3Client:   s3Client,
		log:        log.WithName("s3"),
		pathPrefix: pathPrefix,
	}
}

// S3Downloader downloads models from S3.
type S3Downloader struct {
	s3Client   s3Client
	log        logr.Logger
	pathPrefix string
}

func (d *S3Downloader) download(ctx context.Context, modelPath, filename, destDir string) error {
	d.log.Info("Downloading the model", "modelPath", modelPath)

	var keys []string
	// Use pathPrefix if the model name does not start with "s3://".
	prefix := modelPath
	if !strings.HasPrefix(modelPath, "s3://") {
		prefix = filepath.Join(d.pathPrefix, modelPath)
	}
	f := func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			if filename != "" && *obj.Key != filepath.Join(prefix, filename) {
				// Exclude objects that do not match the specified filename.
				continue
			}

			keys = append(keys, *obj.Key)
		}
		return lastPage
	}

	// We need to append "/". Otherwise, we will download all objects with the same prefix
	// (e.g., "google/gemma-2b" will download "google/gemma-2b" and "google/gemma-2b-it").
	if err := d.s3Client.ListObjectsPages(ctx, prefix+"/", f); err != nil {
		return err
	}
	if len(keys) == 0 {
		return fmt.Errorf("no objects found under %s", prefix)
	}

	for _, key := range keys {
		if err := d.downloadOneObject(ctx, key, prefix, destDir); err != nil {
			return err
		}
	}

	return nil
}

func (d *S3Downloader) downloadOneObject(ctx context.Context, key, prefix, destDir string) error {
	p := strings.TrimPrefix(key, prefix)
	if p == "" || p == "/" {
		// Do nothing if there is an object whose key exactly matches with the model path. We don't need to copy that one.
		// For example, if the model path is 'google/gemma-2b', we would like to copy object under 'google/gemme-2b', but
		// not the object at 'google/gemma-2b'.
		return nil
	}

	filePath := filepath.Join(destDir, p)
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("mkdirall for key %q: %s", key, err)
	}
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("create for key %q: %s", key, err)
	}
	defer func() {
		_ = f.Close()
	}()

	d.log.Info("Downloading S3 object", "key", key, "filePath", filePath)
	if err := d.s3Client.Download(ctx, f, key); err != nil {
		return fmt.Errorf("download key %q: %s", key, err)
	}

	return nil
}
