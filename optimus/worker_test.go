package optimus

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNamedErrorGroup_Set(t *testing.T) {
	errs := newNamedErrorGroup()
	errs.Set("0", errors.New("error"))

	assert.EqualError(t, errs, `{"0":"error"}`)
	assert.NotNil(t, errs.ErrorOrNil())
}

func TestNamedErrorGroup_SetUnique(t *testing.T) {
	errs := newNamedErrorGroup()
	errs.Set("0", errors.New("error"))
	errs.SetUnique([]string{"0", "1", "2"}, errors.New("fail"))

	assert.EqualError(t, errs, `{"0":"error","1":"fail","2":"fail"}`)
	assert.NotNil(t, errs.ErrorOrNil())
}

func TestNamedErrorGroup_ErrorOrNil(t *testing.T) {
	assert.Nil(t, newNamedErrorGroup().ErrorOrNil())
}
