package storage

import "fmt"

type ErrDriverNotSupported struct {
	driver string
}

func (e ErrDriverNotSupported) Error() string {
	return fmt.Sprintf("driver %s not supported", e.driver)
}
