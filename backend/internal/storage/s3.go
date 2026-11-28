package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage stores files in an S3-compatible bucket.
type S3Storage struct {
	client *s3.Client
	bucket string
}

// NewS3Storage creates an S3-compatible storer from environment variables.
// Required: S3_BUCKET. Optional: S3_REGION, S3_ENDPOINT, S3_ACCESS_KEY_ID, S3_SECRET_ACCESS_KEY.
func NewS3Storage(ctx context.Context) (*S3Storage, error) {
	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		return nil, fmt.Errorf("S3_BUCKET is required")
	}
	region := os.Getenv("S3_REGION")
	if region == "" {
		region = "us-east-1"
	}

	var opts []func(*config.LoadOptions) error
	opts = append(opts, config.WithRegion(region))
	if key, secret := os.Getenv("S3_ACCESS_KEY_ID"), os.Getenv("S3_SECRET_ACCESS_KEY"); key != "" && secret != "" {
		opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(key, secret, "")))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	endpoint := os.Getenv("S3_ENDPOINT")
	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		}
	})

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) Save(ctx context.Context, key string, r io.Reader, contentType string) (*Object, error) {
	// Buffer to count size; for large files a multipart upload would be better.
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(string(body)),
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	if _, err := s.client.PutObject(ctx, input); err != nil {
		return nil, fmt.Errorf("s3 put: %w", err)
	}
	return &Object{Key: key, Size: int64(len(body)), ContentType: contentType}, nil
}

func (s *S3Storage) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("s3 get: %w", err)
	}
	return out.Body, nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("s3 delete: %w", err)
	}
	return nil
}
