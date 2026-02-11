package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3 wraps MinIO clients and bucket configuration
type S3 struct {
	internal *minio.Client
	presign  *minio.Client

	Bucket        string
	PublicBaseURL string
}

// Normalize endpoint and detect if HTTPS is used
func normalizeEndpoint(raw string) (endpoint string, secure bool, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", false, fmt.Errorf("empty endpoint")
	}

	// Parse full URL and extract host + scheme
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		u, e := url.Parse(raw)
		if e != nil {
			return "", false, fmt.Errorf("invalid endpoint url %q: %w", raw, e)
		}
		if u.Path != "" && u.Path != "/" {
			return "", false, fmt.Errorf("endpoint must not contain a path: %q", raw)
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("endpoint url missing host: %q", raw)
		}
		return u.Host, u.Scheme == "https", nil
	}

	// Assume raw host without scheme
	return raw, false, nil
}

// Create and configure MinIO clients
func NewS3() (*S3, error) {
	internalRaw := os.Getenv("S3_ENDPOINT")
	externalRaw := os.Getenv("S3_EXTERNAL_ENDPOINT")

	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")

	bucket := os.Getenv("S3_BUCKET")
	publicBase := os.Getenv("S3_PUBLIC_BASE_URL")

	internalEndpoint, internalSecure, err := normalizeEndpoint(internalRaw)
	if err != nil {
		return nil, fmt.Errorf("S3_ENDPOINT: %w", err)
	}

	// Fallback to internal endpoint if external is not set
	if strings.TrimSpace(externalRaw) == "" {
		externalRaw = internalRaw
	}
	externalEndpoint, externalSecure, err := normalizeEndpoint(externalRaw)
	if err != nil {
		return nil, fmt.Errorf("S3_EXTERNAL_ENDPOINT: %w", err)
	}

	// Helper to create a MinIO client
	newClient := func(endpoint string, secure bool) (*minio.Client, error) {
		return minio.New(endpoint, &minio.Options{
			Creds:        credentials.NewStaticV4(accessKey, secretKey, ""),
			Secure:       secure,
			Region:       "us-east-1",
			BucketLookup: minio.BucketLookupPath,
		})
	}

	// Internal client (server-to-server)
	internalClient, err := newClient(internalEndpoint, internalSecure)
	if err != nil {
		return nil, err
	}

	// External client (presigned URLs)
	presignClient, err := newClient(externalEndpoint, externalSecure)
	if err != nil {
		return nil, err
	}

	return &S3{
		internal:      internalClient,
		presign:       presignClient,
		Bucket:        bucket,
		PublicBaseURL: publicBase,
	}, nil
}

// Ensure the bucket exists (create if missing)
func (s *S3) EnsureBucket() error {
	ctx := context.Background()

	exists, err := s.internal.BucketExists(ctx, s.Bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return s.internal.MakeBucket(ctx, s.Bucket, minio.MakeBucketOptions{})
}

// Generate presigned PUT URL for uploading a file
func (s *S3) PresignPut(objectKey string, expiry time.Duration) (string, error) {
	u, err := s.presign.PresignedPutObject(
		context.Background(),
		s.Bucket,
		objectKey,
		expiry,
	)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Generate presigned GET URL for downloading a file
func (s *S3) PresignGet(objectKey string, expiry time.Duration) (string, error) {
	u, err := s.presign.PresignedGetObject(
		context.Background(),
		s.Bucket,
		objectKey,
		expiry,
		nil,
	)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Upload a file directly to MinIO using multipart upload
func (s *S3) PutMultipart(
	objectKey string,
	file multipart.File,
	size int64,
	contentType string,
) error {
	_, err := s.internal.PutObject(
		context.Background(),
		s.Bucket,
		objectKey,
		file,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	return err
}
