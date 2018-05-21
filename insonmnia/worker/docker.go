package worker

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

type imageLoadStatus struct {
	Status string `json:"status"`
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

	return status, nil
}
