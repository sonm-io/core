package multierror

import (
	"strings"
	"sync"

	"github.com/hashicorp/go-multierror"
)

func NewMultiError() *multierror.Error {
	return &multierror.Error{ErrorFormat: errorFormat}
}

func NewTSMultiError() *TSMultiError {
	return &TSMultiError{Error: &multierror.Error{ErrorFormat: errorFormat}}
}

func Append(err error, errs ...error) *multierror.Error {
	return multierror.Append(err, errs...)
}

func AppendUnique(err *multierror.Error, other error) *multierror.Error {
	for _, e := range err.WrappedErrors() {
		if e.Error() == other.Error() {
			return err
		}
	}
	return Append(err, other)
}

func errorFormat(errs []error) string {
	var s []string
	for _, e := range errs {
		s = append(s, e.Error())
	}

	return strings.Join(s, ", ")

}

type TSMultiError struct {
	*multierror.Error
	mu sync.Mutex
}

func (m *TSMultiError) Append(errs ...error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = multierror.Append(m.Error, errs...)
}
