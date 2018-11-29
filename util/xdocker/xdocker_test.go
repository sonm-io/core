package xdocker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/assert"
)

func TestImagePullFromMock(t *testing.T) {
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
		err := DecodeImagePull(bytes.NewReader(fixt.body))
		assert.Equal(t, fixt.err, err, "invalid error for %v", fixt.name)
	}
}

func TestImagePull(t *testing.T) {
	dockclient, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}

	var opt = types.ImagePullOptions{}
	rd, err := dockclient.ImagePull(context.Background(), "alpine:latest", opt)
	if err != nil {
		t.Fatal(err)
	}
	defer rd.Close()

	err = DecodeImagePull(rd)
	if err != nil {
		t.Fatal(err)
	}
}
