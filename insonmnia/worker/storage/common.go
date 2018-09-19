package storage

import (
	"context"
)

type QuotaDescription struct {
	Bytes uint64
}

type Cleanup interface {
	Close() error
}

type StorageQuotaTuner interface {
	SetQuota(ctx context.Context, ID string, quotaID string, bytes uint64) (Cleanup, error)
}
