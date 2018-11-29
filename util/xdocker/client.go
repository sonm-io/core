package xdocker

import (
	"github.com/docker/docker/client"
)

var clientInstance *client.Client

func NewClient() (*client.Client, error) {
	if clientInstance == nil {
		cl, err := client.NewEnvClient()
		if err != nil {
			return nil, err
		}

		clientInstance = cl
	}

	return clientInstance, nil
}
