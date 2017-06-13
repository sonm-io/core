package miner

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"golang.org/x/net/context"
)

func TestOvsSpool(t *testing.T) {
	ctx := context.Background()
	ovs, err := NewOverseer(ctx)
	defer ovs.Close()
	require.NoError(t, err, "failed to create Overseer")
	err = ovs.Spool(ctx, Description{Registry: "docker.io", Image: "alpine"})
	require.NoError(t, err, "failed to pull an image")
	err = ovs.Spool(ctx, Description{Registry: "docker2.io", Image: "alpine"})
	require.NotNil(t, err)
}

const scriptWorkerSh = `#!/bin/sh
# we need this to give an isolation system the gap to attach
sleep 300
echo $@
printenv
`

func buildTestImage(t *testing.T) {
	assert := assert.New(t)
	const dockerFile = `
FROM ubuntu:trusty
COPY worker.sh /usr/bin/worker.sh
EXPOSE 20000
EXPOSE 20001/udp
ENTRYPOINT /usr/bin/worker.sh
	`
	cl, err := client.NewEnvClient()
	assert.NoError(err)
	defer cl.Close()

	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)

	files := []struct {
		Name, Body string
		Mode       int64
	}{
		{"worker.sh", scriptWorkerSh, 0777},
		{"Dockerfile", dockerFile, 0666},
	}

	for _, file := range files {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: file.Mode,
			Size: int64(len(file.Body)),
		}
		assert.Nil(tw.WriteHeader(hdr))
		_, err = tw.Write([]byte(file.Body))
		assert.Nil(err)
	}
	assert.Nil(tw.Close())

	opts := types.ImageBuildOptions{
		Tags: []string{"worker"},
	}

	_, err = cl.ImageRemove(context.Background(), "worker", types.ImageRemoveOptions{PruneChildren: true, Force: true})
	if err != nil {
		t.Logf("ImageRemove returns error: %v", err)
	}

	resp, err := cl.ImageBuild(context.Background(), buf, opts)
	assert.Nil(err)
	defer resp.Body.Close()

	var p = make([]byte, 1024)
	for {
		_, err = resp.Body.Read(p)
		if err != nil {
			assert.EqualError(err, io.EOF.Error())
			break
		}
	}
}

func TestOvsSpawn(t *testing.T) {
	assrt := assert.New(t)
	buildTestImage(t)
	cl, err := client.NewEnvClient()
	assrt.NoError(err)
	defer cl.Close()
	ctx := context.Background()
	ovs, err := NewOverseer(ctx)
	require.NoError(t, err)
	info, err := ovs.Spawn(ctx, Description{Registry: "", Image: "worker"})
	require.NoError(t, err)
	cjson, err := cl.ContainerInspect(ctx, info.ID)
	require.NoError(t, err)
	assrt.True(cjson.HostConfig.AutoRemove)
	assrt.True(cjson.HostConfig.PublishAllPorts)
	t.Logf("spawned %s %v", info.ID, info.Ports)
	_, ok := cjson.NetworkSettings.Ports["20000/tcp"]
	assrt.True(ok)
	_, ok = cjson.NetworkSettings.Ports["20001/udp"]
	assrt.True(ok)
	err = ovs.Stop(ctx, info.ID)
	require.NoError(t, err)
}
