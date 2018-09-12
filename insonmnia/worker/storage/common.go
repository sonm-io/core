package storage

import (
	"context"
	"fmt"
)

type ErrDriverNotSupported struct {
	driver string
}

func (e ErrDriverNotSupported) Error() string {
	return fmt.Sprintf("driver %s not supported", e.driver)
}

type QuotaDescription struct {
	Bytes uint64
}

type Cleanup interface {
	Close() error
}

type StorageQuotaTuner interface {
	SetQuota(ctx context.Context, ID string, quotaID string, bytes uint64) (Cleanup, error)
}
