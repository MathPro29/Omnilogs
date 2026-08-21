package objectstore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type R2 struct {
	client *s3.Client
	bucket string
	prefix string
	err    error
}

func NewR2FromEnv() *R2 {
	endpoint := strings.TrimSpace(os.Getenv("R2_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
	secretKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
	bucket := strings.TrimSpace(os.Getenv("R2_BUCKET"))
	region := strings.TrimSpace(os.Getenv("R2_REGION"))
	if region == "" {
		region = "auto"
	}
	value := &R2{bucket: bucket, prefix: strings.Trim(strings.TrimSpace(os.Getenv("R2_OBJECT_PREFIX")), "/")}
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		value.err = errors.New("R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY and R2_BUCKET are required")
		return value
	}
	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		value.err = fmt.Errorf("configure R2: %w", err)
		return value
	}
	value.client = s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})
	return value
}

func (r *R2) Available() bool { return r != nil && r.err == nil && r.client != nil }

func (r *R2) ConfigurationError() error {
	if r == nil {
		return errors.New("R2 is not configured")
	}
	return r.err
}

func (r *R2) Bucket() string {
	if r == nil {
		return ""
	}
	return r.bucket
}

func (r *R2) Key(relative string) string {
	relative = strings.TrimLeft(filepath.ToSlash(relative), "/")
	if r == nil || r.prefix == "" {
		return relative
	}
	return r.prefix + "/" + relative
}

func (r *R2) PutFile(ctx context.Context, key, path, tagging string, metadata map[string]string) error {
	if !r.Available() {
		return r.ConfigurationError()
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(r.bucket), Key: aws.String(key), Body: file,
		ContentType: aws.String("application/gzip"), Tagging: aws.String(tagging), Metadata: metadata,
	})
	if err != nil {
		return fmt.Errorf("upload archive to R2: %w", err)
	}
	return nil
}

// Materialize downloads an object to a private temporary file. The caller
// must invoke the returned cleanup function.
func (r *R2) Materialize(ctx context.Context, key string) (string, func(), error) {
	if !r.Available() {
		return "", func() {}, r.ConfigurationError()
	}
	response, err := r.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	if err != nil {
		return "", func() {}, fmt.Errorf("download archive from R2: %w", err)
	}
	defer response.Body.Close()
	file, err := os.CreateTemp("", "omnilogs-r2-*.ndjson.gz")
	if err != nil {
		return "", func() {}, err
	}
	path := file.Name()
	cleanup := func() { _ = os.Remove(path) }
	if _, err := io.Copy(file, response.Body); err != nil {
		_ = file.Close()
		cleanup()
		return "", func() {}, err
	}
	if err := file.Close(); err != nil {
		cleanup()
		return "", func() {}, err
	}
	return path, cleanup, nil
}

func (r *R2) Delete(ctx context.Context, key string) error {
	if !r.Available() {
		return r.ConfigurationError()
	}
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(r.bucket), Key: aws.String(key)})
	return err
}
