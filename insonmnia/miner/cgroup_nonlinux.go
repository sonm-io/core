// +build !linux

package miner

// Resources is a placeholder for resources
type Resources interface{}

const (
	platformSupportCGroups = false
	parentCgroup           = ""
)

type nilDeleter struct{}

func (*nilDeleter) Delete() error { return nil }

func initializeControlGroup(*Resources) (cGroupDeleter, error) {
	return &nilDeleter{}, nil
}
