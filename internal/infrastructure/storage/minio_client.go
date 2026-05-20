package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client *minio.Client
	bucket string
}

func NewMinioClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinioClient, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	return &MinioClient{
		client: client,
		bucket: bucket,
	}, nil
}

func (m *MinioClient) EnsureBucket(ctx context.Context) error {
	exists, err := m.client.BucketExists(ctx, m.bucket)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.client.MakeBucket(ctx, m.bucket, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket %s: %w", m.bucket, err)
		}
	}

	return nil
}

func (m *MinioClient) Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload object to MinIO: %w", err)
	}

	return nil
}

func (m *MinioClient) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object from MinIO: %w", err)
	}

	return object, nil
}

func (m *MinioClient) Delete(ctx context.Context, objectKey string) error {
	err := m.client.RemoveObject(ctx, m.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object from MinIO: %w", err)
	}

	return nil
}

func (m *MinioClient) GetPresignedURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

func (m *MinioClient) GetThumbnailURL(ctx context.Context, objectKey string, expiry time.Duration) (string, error) {
	// Generate a presigned URL; thumbnail generation can be handled by
	// MinIO's built-in image processing if configured, or by a separate service.
	// For now, return the presigned URL with a query parameter indicating thumbnail.
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectKey, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate thumbnail URL: %w", err)
	}

	return presignedURL.String(), nil
}
