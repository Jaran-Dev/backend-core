package storage

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	Put(ctx context.Context, key string, r io.Reader, contentType string) (url string, err error)
	Delete(ctx context.Context, key string) error
	SignUploadURL(ctx context.Context, key string, ttl time.Duration) (string, error)
}
