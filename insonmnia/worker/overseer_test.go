package worker

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOvsSpool(t *testing.T) {
	ctx := context.Background()
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	defer ovs.Close()
	require.NoError(t, err, "failed to create Overseer")

	ref, err := reference.ParseNormalizedNamed("docker.io/alpine")
	require.NoError(t, err, "failed to create Overseer")
	err = ovs.Spool(ctx, Description{Reference: ref})
	require.NoError(t, err, "failed to pull an image")

	ref, err = reference.ParseNormalizedNamed("docker2.io/alpine")
	err = ovs.Spool(ctx, Description{Reference: ref})
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
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	require.NoError(t, err)
	ref, err := reference.ParseNormalizedNamed("worker")
	require.NoError(t, err)
	ch, info, err := ovs.Start(ctx, Description{Reference: ref})
	require.NoError(t, err)
	cjson, err := cl.ContainerInspect(ctx, info.ID)
	require.NoError(t, err)
	//assrt.True(cjson.HostConfig.AutoRemove)
	assrt.True(cjson.HostConfig.PublishAllPorts)
	t.Logf("spawned %s %v", info.ID, info.Ports)
	_, ok := cjson.NetworkSettings.Ports["20000/tcp"]
	assrt.True(ok)
	_, ok = cjson.NetworkSettings.Ports["20001/udp"]
	assrt.True(ok)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tk := time.NewTicker(time.Second * 10)
		defer tk.Stop()
		defer wg.Done()
		select {
		case <-ch:
		case <-tk.C:
			t.Error("waiting for stop status timed out")
		}
	}()

	err = ovs.Stop(ctx, info.ID)
	require.NoError(t, err)
	wg.Wait()
}

func TestOvsAttach(t *testing.T) {
	assrt := assert.New(t)
	buildTestImage(t)
	cl, err := client.NewEnvClient()
	assrt.NoError(err)
	defer cl.Close()
	ctx := context.Background()
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	require.NoError(t, err)
	ref, err := reference.ParseNormalizedNamed("worker")
	require.NoError(t, err)
	_, info, err := ovs.Start(ctx, Description{Reference: ref})
	require.NoError(t, err)
	cjson, err := cl.ContainerInspect(ctx, info.ID)
	require.NoError(t, err)
	t.Logf("spawned %s %v", info.ID, info.Ports)
	_, ok := cjson.NetworkSettings.Ports["20000/tcp"]
	assrt.True(ok)
	_, ok = cjson.NetworkSettings.Ports["20001/udp"]
	assrt.True(ok)
	ovs.Close()

	ovs2, err := NewOverseer(ctx, plugin.EmptyRepository())
	require.NoError(t, err)
	descr := Description{Reference: ref}
	ch, err := ovs2.Attach(ctx, info.ID, descr)
	t.Logf("attached to container %s", info.ID)
	require.NoError(t, err)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		tk := time.NewTicker(time.Second * 10)
		defer tk.Stop()
		defer wg.Done()
		select {
		case <-ch:
		case <-tk.C:
			t.Error("waiting for stop status timed out")
		}
	}()
	metrics, err := ovs.Info(ctx)
	require.NoError(t, err)
	_, ok = metrics[info.ID]
	if !ok {
		t.Error("failed to find container with id ", info.ID)
	}
	err = ovs2.Stop(ctx, info.ID)
	require.NoError(t, err)
	t.Logf("successfully stopped container %s", info.ID)
	ovs2.Close()
	wg.Wait()
}

func TestExpose(t *testing.T) {
	portMap, portBinding, err := nat.ParsePortSpecs([]string{"81:80", "443:443", "8.8.8.8:53:10053", "22"})

	require.NoError(t, err)

	assert.Equal(t, map[nat.Port]struct{}{"80/tcp": {}, "443/tcp": {}, "10053/tcp": {}, "22/tcp": {}}, portMap)
	assert.Equal(t, map[nat.Port][]nat.PortBinding{
		"80/tcp":    {{HostIP: "", HostPort: "81"}},
		"443/tcp":   {{HostIP: "", HostPort: "443"}},
		"10053/tcp": {{HostIP: "8.8.8.8", HostPort: "53"}},
		"22/tcp":    {{HostIP: "", HostPort: ""}},
	}, portBinding)
}
