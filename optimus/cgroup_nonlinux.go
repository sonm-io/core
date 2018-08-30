// +build !linux

package optimus

type nilDeleter struct{}

func (nilDeleter) Delete() error {
	return nil
}

func RestrictUsage(cfg *RestrictionsConfig) (Deleter, error) {
	return nilDeleter{}, nil
}
