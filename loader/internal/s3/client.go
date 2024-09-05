package s3

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	partMiBs int64 = 128
)

// NewOptions is the options for NewClient.
type NewOptions struct {
	EndpointURL string
	Region      string
	Bucket      string

	UseAnonymousCredentials bool
}

// NewClient returns a new S3 client.
func NewClient(ctx context.Context, o NewOptions) (*Client, error) {
	var err error
	var conf aws.Config
	if o.UseAnonymousCredentials {
		conf, err = config.LoadDefaultConfig(ctx,
			config.WithCredentialsProvider(aws.AnonymousCredentials{}),
		)
	} else {
		conf, err = config.LoadDefaultConfig(ctx)
	}
	if err != nil {
		return nil, err
	}
	s3Client := s3.NewFromConfig(conf, func(opt *s3.Options) {
		opt.BaseEndpoint = aws.String(o.EndpointURL)
		opt.Region = o.Region
		// This is needed as the minio server does not support the virtual host style.
		opt.UsePathStyle = true
	})
	return &Client{
		svc:    s3Client,
		bucket: o.Bucket,
	}, nil
}

// Client is a client for S3.
type Client struct {
	svc    *s3.Client
	bucket string
}

// Upload uses an upload manager to upload a file to an object in a bucket.
// The upload manager breaks large file into parts and uploads the parts concurrently.
func (c *Client) Upload(ctx context.Context, r io.Reader, key string) error {
	uploader := manager.NewUploader(c.svc, func(u *manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucket),
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
func (c *Client) Download(ctx context.Context, w io.WriterAt, key string) error {
	downloader := manager.NewDownloader(c.svc, func(d *manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	_, err := downloader.Download(ctx, w, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket),
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
	prefix string,
	f func(page *s3.ListObjectsV2Output, lastPage bool) bool,
) error {
	const maxKeys = int32(1000)
	p := s3.NewListObjectsV2Paginator(
		c.svc,
		&s3.ListObjectsV2Input{
			Bucket: aws.String(c.bucket),
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
