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
	v1 "github.com/llmariner/model-manager/api/v1"
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

func (d *S3Downloader) download(ctx context.Context, modelName, filename, destDir string) error {
	d.log.Info("Downloading the model", "name", modelName)

	var keys []string
	prefix := filepath.Join(d.pathPrefix, modelName)
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
	filePath := filepath.Join(destDir, strings.TrimPrefix(key, prefix))
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return err
	}
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	d.log.Info("Downloading S3 object", "key", key, "filePath", filePath)
	if err := d.s3Client.Download(ctx, f, key); err != nil {
		return err
	}

	return nil
}

func (d *S3Downloader) sourceRepository() v1.SourceRepository {
	return v1.SourceRepository_SOURCE_REPOSITORY_OBJECT_STORE
}
