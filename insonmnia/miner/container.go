package miner

import (
	"context"
	"path/filepath"

	"go.uber.org/zap"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	log "github.com/noxiouz/zapctx/ctxlog"
)

type dcontainer struct {
	ctx    context.Context
	cancel context.CancelFunc

	client *client.Client

	ID    string
	stats types.Stats
}

func newContainer(ctx context.Context, dclient *client.Client, d Description) (*dcontainer, error) {
	log.G(ctx).Info("start container with application")

	ctx, cancel := context.WithCancel(ctx)
	cont := dcontainer{
		ctx:    ctx,
		cancel: cancel,
		client: dclient,
	}

	// NOTE: command to launch must be specified via ENTRYPOINT and CMD in Dockerfile
	var config = container.Config{
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,

		Image: filepath.Join(d.Registry, d.Image),
		// TODO: set actual name
		Labels: map[string]string{overseerTag: ""},
	}
	// NOTE: all ports are EXPOSE as PublishAll
	// TODO: detect network network mode and interface
	var hostConfig = container.HostConfig{
		PublishAllPorts: true,
		// NOTE; we don't want to leave garbage
		AutoRemove: true,
	}
	var networkingConfig network.NetworkingConfig

	// create new container
	// assign resulted containerid
	// log all warnings
	resp, err := cont.client.ContainerCreate(ctx, &config, &hostConfig, &networkingConfig, "")
	if err != nil {
		return nil, err
	}
	cont.ID = resp.ID
	cont.ctx = log.WithLogger(cont.ctx, log.G(ctx).With(zap.String("id", cont.ID)))
	if len(resp.Warnings) > 0 {
		log.G(ctx).Warn("ContainerCreate finished with warnings", zap.Strings("warnings", resp.Warnings))
	}

	return &cont, nil
}

func (c *dcontainer) startContainer() error {
	var options types.ContainerStartOptions
	if err := c.client.ContainerStart(c.ctx, c.ID, options); err != nil {
		c.cancel()
		return err
	}
	return nil
}

func (c *dcontainer) Kill() (err error) {
	// TODO: add atomic flag to prevent duplicated remove
	defer func() {
		c.remove()
		// release HTTP connections
		c.cancel()
	}()
	log.G(c.ctx).Info("kill the container")

	if err = c.client.ContainerKill(c.ctx, c.ID, "SIGKILL"); err != nil {
		log.G(c.ctx).Error("failed to send SIGKILL to the container", zap.Error(err))
		return err
	}
	return nil
}

func (c *dcontainer) remove() {
	containerRemove(c.ctx, c.client, c.ID)
}

func containerRemove(ctx context.Context, client client.APIClient, id string) {
	removeOpts := types.ContainerRemoveOptions{}
	if err := client.ContainerRemove(ctx, id, removeOpts); err != nil {
		log.G(ctx).Error("failed to remove the container", zap.Error(err))
	}
}
