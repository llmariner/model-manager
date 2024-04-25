package s3

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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
func NewClient(o NewOptions) *Client {
	var sopts session.Options
	if o.UseAnonymousCredentials {
		sopts = session.Options{
			Config: aws.Config{
				Credentials: credentials.AnonymousCredentials,
			},
		}
	} else {
		sopts = session.Options{
			SharedConfigState: session.SharedConfigEnable,
		}
	}

	sess := session.Must(session.NewSessionWithOptions(sopts))

	conf := &aws.Config{
		Endpoint: aws.String(o.EndpointURL),
		Region:   aws.String(o.Region),
		// This is needed as the minio server does not support the virtual host style.
		S3ForcePathStyle: aws.Bool(true),
	}
	return &Client{
		svc:    s3.New(sess, conf),
		bucket: o.Bucket,
	}
}

// Client is a client for S3.
type Client struct {
	svc    *s3.S3
	bucket string
}

// Upload uploads the data that buf contains to a S3 object.
func (c *Client) Upload(r io.Reader, key string) error {
	uploader := s3manager.NewUploaderWithClient(c.svc, func(u *s3manager.Uploader) {
		u.PartSize = partMiBs * 1024 * 1024
	})
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
		Body:   r,
	})
	if err != nil {
		return err
	}
	return nil
}

// Download downloads the data from a S3 object.
func (c *Client) Download(w io.WriterAt, key string) error {
	downloader := s3manager.NewDownloaderWithClient(c.svc, func(d *s3manager.Downloader) {
		d.PartSize = partMiBs * 1024 * 1024
	})
	_, err := downloader.Download(w, &s3.GetObjectInput{
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
	prefix string,
	f func(page *s3.ListObjectsOutput, lastPage bool) bool,
) error {
	return c.svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(c.bucket),
		Prefix: aws.String(prefix),
	}, f)
}
