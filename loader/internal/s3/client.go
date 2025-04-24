package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	laws "github.com/llmariner/common/pkg/aws"
)

const (
	partMiBs int64 = 128
)

// NewClient returns a new S3 client.
func NewClient(ctx context.Context, o laws.NewS3ClientOptions) (*Client, error) {
	svc, err := laws.NewS3Client(ctx, o)
	if err != nil {
		return nil, err
	}
	return &Client{
		svc: svc,
	}, nil
}

// Client is a client for S3.
type Client struct {
	svc    *s3.Client
	bucket string
}

// Upload uses an upload manager to upload a file to an object in a bucket.
// The upload manager breaks large file into parts and uploads the parts concurrently.
func (c *Client) Upload(ctx context.Context, r io.Reader, bucket, key string) error {
	uploader := manager.NewUploader(c.svc, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return err
	}
	return nil
}

// Download uses a download manager to download an object from a bucket.
// The download manager gets the data in parts and writes them to a buffer until all of
// the data has been downloaded.
func (c *Client) Download(ctx context.Context, w io.WriterAt, bucket, key string) error {
	downloader := manager.NewDownloader(c.svc, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	_, err := downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

// ListObjectsPages returns S3 objects with pagination.
func (c *Client) ListObjectsPages(
	ctx context.Context,
	bucket,
	prefix string,
	f func(page *s3.ListObjectsV2Output, lastPage bool) bool,
) error {
	const maxKeys = int32(1000)
	p := s3.NewListObjectsV2Paginator(
		c.svc,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(bucket),
			Prefix: aws.String(prefix),
		},
		func(o *s3.ListObjectsV2PaginatorOptions) {
			o.Limit = maxKeys
		},
	)
	for p.HasMorePages() {
		page, err := p.NextPage(ctx)
		if err != nil {
			return err
		}
		f(page, !p.HasMorePages())
	}
	return nil
}
