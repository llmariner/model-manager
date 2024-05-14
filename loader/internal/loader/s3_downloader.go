package loader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Client interface {
	Download(w io.WriterAt, key string) error
	ListObjectsPages(prefix string, f func(page *s3.ListObjectsOutput, lastPage bool) bool) error
}

// NewS3Downloader returns a new S3Downloader.
func NewS3Downloader(s3Client s3Client, pathPrefix string) *S3Downloader {
	return &S3Downloader{
		s3Client:   s3Client,
		pathPrefix: pathPrefix,
	}
}

// S3Downloader downloads models from S3.
type S3Downloader struct {
	s3Client   s3Client
	pathPrefix string
}

func (d *S3Downloader) download(modelName, destDir string) error {
	log.Printf("Downloading the model %q\n", modelName)

	var keys []string
	f := func(page *s3.ListObjectsOutput, lastPage bool) bool {
		for _, obj := range page.Contents {
			keys = append(keys, *obj.Key)
		}
		return lastPage
	}
	prefix := filepath.Join(d.pathPrefix, modelName)
	// We need to append "/". Otherwise, we will download all objects with the same prefix
	// (e.g., "google/gemma-2b" will download "google/gemma-2b" and "google/gemma-2b-it").
	if err := d.s3Client.ListObjectsPages(prefix+"/", f); err != nil {
		return err
	}
	if len(keys) == 0 {
		return fmt.Errorf("no objects found under %s", prefix)
	}

	for _, key := range keys {
		if err := d.downloadOneObject(key, prefix, destDir); err != nil {
			return err
		}
	}

	return nil
}

func (d *S3Downloader) downloadOneObject(key, prefix, destDir string) error {
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

	log.Printf("Downloading S3 object %q and writing to %q\n", key, filePath)
	if err := d.s3Client.Download(f, key); err != nil {
		return err
	}

	return nil
}
