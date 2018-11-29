package xdocker

import (
	"sync"

	"github.com/docker/docker/client"
)

var (
	mu             sync.Mutex
	clientInstance *client.Client
)

func NewClient() (*client.Client, error) {
	mu.Lock()
	defer mu.Unlock()

	if clientInstance == nil {
		cl, err := client.NewEnvClient()
		if err != nil {
			return nil, err
		}

		clientInstance = cl
	}

	return clientInstance, nil
}
