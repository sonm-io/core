package worker

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"
)

type imageLoadStatus struct {
	Id string `json:"stream"`
}

func decodeImageLoad(rd io.Reader) (imageLoadStatus, error) {
	result, err := ioutil.ReadAll(rd)
	if err != nil {
		return imageLoadStatus{}, err
	}
	var status imageLoadStatus
	if err := json.Unmarshal([]byte(result), &status); err != nil {
		return imageLoadStatus{}, err
	}
	status.Id = strings.Replace(status.Id, "Loaded image ID: ", "", -1)
	status.Id = strings.Trim(status.Id, "\n")

	return status, nil
}
