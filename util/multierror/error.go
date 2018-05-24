package multierror

import (
	"strings"

	"github.com/hashicorp/go-multierror"
)

func NewMultiError() *multierror.Error {
	return &multierror.Error{ErrorFormat: errorFormat}
}

func Append(err error, errs ...error) *multierror.Error {
	return multierror.Append(err, errs...)
}

func errorFormat(errs []error) string {
	var s []string
	for _, e := range errs {
		s = append(s, e.Error())
	}

	return strings.Join(s, ", ")

}
