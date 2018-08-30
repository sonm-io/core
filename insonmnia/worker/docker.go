package worker

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

type imageID string

func (m imageID) String() string {
	return string(m)
}

func getImageID(rd io.Reader) (imageID, error) {
	resultBytes, err := ioutil.ReadAll(rd)
	if err != nil {
		return "", err
	}

	var result = struct {
		ID  string `json:"stream"`
		Err string `json:"error"`
	}{}

	if err := json.Unmarshal(resultBytes, &result); err != nil {
		return "", fmt.Errorf("failed to unmarshal Docker response: %v", err)
	}

	if len(result.Err) > 0 {
		return "", errors.New(result.Err)
	}

	result.ID = strings.Replace(result.ID, "Loaded image ID: ", "", -1)
	result.ID = strings.Trim(result.ID, "\n")

	return imageID(result.ID), nil
}
