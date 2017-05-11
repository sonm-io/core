package insonmnia

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImagePullFromMock(t *testing.T) {
	assert := assert.New(t)

	fixtures := []struct {
		name string
		body []byte
		err  error
	}{
		{"NoEOF", []byte("{\"Status\": \"OK\"}\n"), nil},
		{"LinesCase", []byte("{\"Status\": \"OK\"}\n{\"Status\": \"OK\"}\n"), nil},
		{"LinesCaseNoEnd", []byte("{\"Status\": \"OK\"}\n{\"Status\": \"OK\"}"), nil},
		{"LinesCaseError", []byte("{\"Status\": \"OK\"}\n{\"Status\": \"OK\"}{\"Error\": \"blabla\"}"), fmt.Errorf("blabla")},
		{"FlatCase", []byte("{\"Status\": \"OK\"}{\"Status\": \"OK\"}"), nil},
		{"FlatCaseError", []byte(`{"Status": "OK"}{"Status": "OK"}{"Error": "blabla"}`), fmt.Errorf("blabla")},
		{"MixedCase", []byte("{\"Status\": \"OK\"}\n{\"Status\": \"OK\"}{\"Status\": \"OK\"}"), nil},
		{"MixedCaseError", []byte("{\"Status\": \"OK\"}\n{\"Status\": \"OK\"}{\"Error\": \"blabla\"}"), fmt.Errorf("blabla")},
	}

	for _, fixt := range fixtures {
		err := decodeImagePull(bytes.NewReader(fixt.body))
		assert.Equal(fixt.err, err, "invalid error for %v", fixt.name)
	}
}
