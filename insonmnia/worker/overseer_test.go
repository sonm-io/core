package worker

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/go-connections/nat"
	"github.com/gliderlabs/ssh"
	"github.com/sonm-io/core/insonmnia/structs"
	"github.com/sonm-io/core/insonmnia/worker/network"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util/xdocker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOvsSpool(t *testing.T) {
	ctx := context.Background()
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	defer ovs.Close()
	require.NoError(t, err, "failed to create Overseer")

	ref, err := xdocker.NewReference("docker.io/alpine")
	require.NoError(t, err, "failed to create Overseer")
	err = ovs.Spool(ctx, Description{Reference: ref})
	require.NoError(t, err, "failed to pull an image")

	ref, err = xdocker.NewReference("docker2.io/alpine")
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
	cl, err := xdocker.NewClient()
	assert.NoError(err)

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
	cl, err := xdocker.NewClient()
	assrt.NoError(err)
	ctx := context.Background()
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	require.NoError(t, err)
	ref, err := xdocker.NewReference("worker")
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
	cl, err := xdocker.NewClient()
	assrt.NoError(err)
	ctx := context.Background()
	ovs, err := NewOverseer(ctx, plugin.EmptyRepository())
	require.NoError(t, err)
	ref, err := xdocker.NewReference("worker")
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

func TestMarshalDescription(t *testing.T) {
	ref, err := xdocker.NewReference("docker.io/sonm-io/tests:latest")
	require.NoError(t, err)

	d := Description{
		Reference: ref,
		NetworkOptions: &network.Network{
			ID:               "1234",
			Name:             "net_name",
			Alias:            "net_alias",
			RateLimitEgress:  100,
			RateLimitIngress: 200,
		},
		NetworkSpecs: []*structs.NetworkSpec{
			{
				NetworkSpec: &sonm.NetworkSpec{
					Type:    "type1",
					Options: map[string]string{"opt1": "val1"},
					Subnet:  "/12",
					Addr:    "1.1.1.1",
				},
				NetID: "1111",
			},
			{
				NetworkSpec: &sonm.NetworkSpec{
					Type:    "type2",
					Options: map[string]string{"opt2": "val2"},
					Subnet:  "/28",
					Addr:    "2.2.2.2",
				},
				NetID: "2222",
			},
		},
	}

	data, err := json.Marshal(&d)
	require.NoError(t, err)

	n := Description{}
	err = json.Unmarshal(data, &n)
	require.NoError(t, err)

	assert.NotEmpty(t, n.Reference.String())
	assert.Equal(t, d.Reference.String(), n.Reference.String())
	assert.Equal(t, ref.String(), n.Reference.String())
	assert.Equal(t, len(d.NetworkSpecs), len(n.NetworkSpecs))

	assert.Equal(t, d.NetworkSpecs[0].Type, n.NetworkSpecs[0].Type)
	assert.Equal(t, d.NetworkSpecs[0].Options, n.NetworkSpecs[0].Options)
	assert.Equal(t, d.NetworkSpecs[0].Subnet, n.NetworkSpecs[0].Subnet)
	assert.Equal(t, d.NetworkSpecs[0].Addr, n.NetworkSpecs[0].Addr)

	assert.Equal(t, d.NetworkOptions.ID, n.NetworkOptions.ID)
	assert.Equal(t, d.NetworkOptions.Name, n.NetworkOptions.Name)
	assert.Equal(t, d.NetworkOptions.Alias, n.NetworkOptions.Alias)
	assert.Equal(t, d.NetworkOptions.RateLimitEgress, n.NetworkOptions.RateLimitEgress)
	assert.Equal(t, d.NetworkOptions.RateLimitIngress, n.NetworkOptions.RateLimitIngress)
}

func TestMarshalContainerInfo(t *testing.T) {
	keyRaw := []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCh+u6UN26+nIc42aRhnuDeralPivXZDi3ETSugsNlOfMww5YdqSJc9otSGooPRbXhOVguoEZfBvLNNd4xTkYtaCsWmFGbq3JXCjtH22V3VeqDc1zd3iJGtQU2BInC0HHvR4M5U4ayN4Ur3bEwgBViv7J+2lABmOArVwOlxacI/m2FtmUPrXKLh98eZgvAxd7DLwTjL8DKLJVqk2hqPRbqvX+CVHVZ4EeS63k0ji2mHDDlZrCsm2n6CnOau4sIND4Xiibdtt6dHnXKXxyC1SLQlH1W+6fxdiQSWXK4/Q4ryA0L/t89CoSp+/uRy4xnP3z5ntI7vE+I3Y1kFeTpOy1v9 alex@Dikobrazzers.local")
	pkey := PublicKey{}
	err := pkey.UnmarshalText(keyRaw)
	require.NoError(t, err)

	c := ContainerInfo{
		PublicKey: pkey,
		DealID:    sonm.NewBigIntFromInt(1234),
	}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	n := ContainerInfo{}
	err = json.Unmarshal(data, &n)
	require.NoError(t, err)

	assert.True(t, ssh.KeysEqual(c.PublicKey, n.PublicKey))
	assert.Equal(t, n.DealID, sonm.NewBigIntFromInt(1234))
}
