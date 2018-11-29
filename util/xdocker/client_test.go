package xdocker

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestXXX(t *testing.T) {
	cl, err := NewClient()
	require.NoError(t, err)

	_, err = cl.Info(context.Background())
	require.NoError(t, err)

	cl.Close()

	_, err = cl.Info(context.Background())
	require.NoError(t, err)

	cl.Close()

	_, err = cl.Info(context.Background())
	require.NoError(t, err)

	wg := sync.WaitGroup{}
	wg.Add(50)

	for i := 0; i < 50; i++ {
		go func() {
			defer wg.Done()

			cl, err := NewClient()
			require.NoError(t, err)

			i, err := cl.Info(context.Background())
			require.NoError(t, err)
			assert.Equal(t, "/var/lib/docker", i.DockerRootDir)

			cl.Close()

		}()
	}

	wg.Wait()
}
