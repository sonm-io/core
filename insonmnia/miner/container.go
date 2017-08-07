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

type containerDescriptor struct {
	ctx    context.Context
	cancel context.CancelFunc

	client *client.Client

	ID    string
	stats types.Stats
}

func newContainer(ctx context.Context, dockerClient *client.Client, d Description) (*containerDescriptor, error) {
	log.G(ctx).Info("start container with application")

	ctx, cancel := context.WithCancel(ctx)
	cont := containerDescriptor{
		ctx:    ctx,
		cancel: cancel,
		client: dockerClient,
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
		RestartPolicy:   d.RestartPolicy,
		// NOTE; we don't want to leave garbage
		AutoRemove: true,
		Resources: container.Resources{
			// TODO: accept a name of a cgroup cooked by user
			// NOTE: on non-Linux platform it's empty
			CgroupParent: parentCgroup,
			Memory:       d.Resources.Memory,
			NanoCPUs:     d.Resources.NanoCPUs,
		},
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

func (c *containerDescriptor) startContainer() error {
	var options types.ContainerStartOptions
	if err := c.client.ContainerStart(c.ctx, c.ID, options); err != nil {
		c.cancel()
		return err
	}
	return nil
}

func (c *containerDescriptor) Kill() (err error) {
	// TODO: add atomic flag to prevent duplicated remove
	defer func() {
		// release HTTP connections
		c.cancel()
	}()

	log.G(c.ctx).Info("kill the container", zap.String("id", c.ID))
	if err = c.client.ContainerKill(context.Background(), c.ID, "SIGKILL"); err != nil {
		log.G(c.ctx).Error("failed to send SIGKILL to the container", zap.String("id", c.ID), zap.Error(err))
		return err
	}
	return nil
}

func (c *containerDescriptor) remove() {
	containerRemove(c.ctx, c.client, c.ID)
}

func containerRemove(ctx context.Context, client client.APIClient, id string) {
	removeOpts := types.ContainerRemoveOptions{}
	if err := client.ContainerRemove(ctx, id, removeOpts); err != nil {
		log.G(ctx).Error("failed to remove the container", zap.String("id", id), zap.Error(err))
	}
}
