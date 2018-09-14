package worker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-units"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"
)

type containerDescriptor struct {
	client *client.Client
	log    *zap.SugaredLogger

	ID              string
	CommitedImageID string
	description     Description
	stats           types.StatsJSON

	cleanup plugin.Cleanup
}

func attachContainer(ctx context.Context, dockerClient *client.Client, ID string, d Description, tuners *plugin.Repository) (*containerDescriptor, error) {
	log.S(ctx).Infof("attaching to container with ID: %s reference: %s", ID, d.Reference.String())

	cont := containerDescriptor{
		client:      dockerClient,
		description: d,
	}

	cleanup, err := tuners.GetCleanup(ctx, &d)
	if err != nil {
		return nil, err
	}

	cont.cleanup = cleanup
	cont.log = log.S(ctx).With(zap.String("container_id", cont.ID))

	return &cont, nil
}

func newContainer(ctx context.Context, dockerClient *client.Client, d Description, tuners *plugin.Repository) (*containerDescriptor, error) {
	log.S(ctx).Infof("start container with application, reference %s", d.Reference.String())

	cont := containerDescriptor{
		client:      dockerClient,
		description: d,
	}

	exposedPorts, portBindings, err := d.Expose()
	if err != nil {
		log.G(ctx).Error("failed to parse `expose` section", zap.Error(err))
		return nil, err
	}

	log.G(ctx).Debug("exposing ports", zap.Any("portBindings", portBindings))

	// NOTE: command to launch must be specified via ENTRYPOINT and CMD in Dockerfile
	var config = container.Config{
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,

		ExposedPorts: exposedPorts,

		Image: d.Reference.String(),
		// TODO: set actual name
		Labels:  map[string]string{overseerTag: "", dealIDTag: d.DealId},
		Env:     d.FormatEnv(),
		Volumes: make(map[string]struct{}),
	}

	// NOTE: all ports are EXPOSE as PublishAll
	// TODO: detect network network mode and interface
	logOpts := make(map[string]string)
	// TODO: Move to StartTask?
	logOpts["max-size"] = "100m"
	var hostConfig = container.HostConfig{
		LogConfig:       container.LogConfig{Type: "json-file", Config: logOpts},
		PublishAllPorts: true,
		PortBindings:    portBindings,
		RestartPolicy:   d.RestartPolicy.Unwrap(),
		AutoRemove:      d.Autoremove,
		Resources:       d.Resources.ToHostConfigResources(d.CGroupParent),
	}

	networkingConfig := network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	if d.NetworkOptions != nil {
		networkingConfig.EndpointsConfig[d.NetworkOptions.Name] = &network.EndpointSettings{
			NetworkID: d.NetworkOptions.ID,
		}
	}

	cleanup, err := tuners.Tune(ctx, &d, &hostConfig, &networkingConfig)
	if err != nil {
		log.G(ctx).Error("failed to tune container", zap.Error(err))
		return nil, err
	}
	log.G(ctx).Debug("successfully tuned container")

	// create new container
	// assign resulted containerid
	// log all warnings
	resp, err := cont.client.ContainerCreate(ctx, &config, &hostConfig, &networkingConfig, "")
	if err != nil {
		return nil, err
	}
	cont.ID = resp.ID
	cont.log = log.S(ctx).With(zap.String("container_id", cont.ID))
	cont.cleanup = cleanup
	if len(resp.Warnings) > 0 {
		log.G(ctx).Warn("ContainerCreate finished with warnings", zap.Strings("warnings", resp.Warnings))
	}

	return &cont, nil
}

func (c *containerDescriptor) startContainer(ctx context.Context) error {
	var options types.ContainerStartOptions
	if err := c.client.ContainerStart(ctx, c.ID, options); err != nil {
		c.log.Warn("ContainerStart finished with error", zap.Error(err))
		return err
	}
	return nil
}

func (c *containerDescriptor) execCommand(ctx context.Context, cmd []string, env []string, isTty bool, wCh <-chan ssh.Window) (conn types.HijackedResponse, err error) {
	cfg := types.ExecConfig{
		User:         "root",
		Tty:          isTty,
		AttachStderr: true,
		AttachStdout: true,
		AttachStdin:  true,
		Detach:       false,
		Cmd:          cmd,
		Env:          env,
	}

	c.log.With(zap.Any("config", cfg)).Info("attaching command")

	execId, err := c.client.ContainerExecCreate(ctx, c.ID, cfg)
	if err != nil {
		c.log.Warnf("ContainerExecCreate finished with error: %s", err)
		return
	}

	conn, err = c.client.ContainerExecAttach(ctx, execId.ID, cfg)
	if err != nil {
		c.log.Warn("ContainerExecAttach finished with error: %s", err)
	}

	err = c.client.ContainerExecStart(ctx, execId.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		c.log.Warn("ContainerExecStart finished with error: %s", err)
		return
	}
	go func() {
		for {
			select {
			case w, ok := <-wCh:
				if !ok {
					return
				}
				c.log.Infof("resizing tty to %dx%d", w.Height, w.Width)
				err = c.client.ContainerExecResize(ctx, execId.ID, types.ResizeOptions{Height: uint(w.Height), Width: uint(w.Width)})
				if err != nil {
					log.G(ctx).Warn("ContainerExecResize finished with error", zap.Error(err))
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	c.log.Info("attached command to container")
	return
}

func (c *containerDescriptor) Kill(ctx context.Context) error {
	c.log.Info("kill the container")
	if err := c.client.ContainerKill(ctx, c.ID, "SIGKILL"); err != nil {
		c.log.Warnf("failed to send SIGKILL to the container: %s", err)
		return err
	}
	return nil
}

func (c *containerDescriptor) Remove(ctx context.Context) error {
	c.log.Info("remove the container")
	result := multierror.NewMultiError()
	if err := containerRemove(ctx, c.client, c.ID); err != nil {
		result = multierror.Append(result, err)
	}
	if len(c.CommitedImageID) != 0 {
		if err := imageRemove(ctx, c.client, c.CommitedImageID); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

//TODO: pass context
func (c *containerDescriptor) Cleanup() error {
	return c.cleanup.Close()
}

func containerRemove(ctx context.Context, client client.APIClient, id string) error {
	removeOpts := types.ContainerRemoveOptions{}
	if err := client.ContainerRemove(ctx, id, removeOpts); err != nil {
		return fmt.Errorf("failed to remove the container %s: %s", id, err)
	}
	log.S(ctx).Infof("removed container %s", id)
	return nil
}

func imageRemove(ctx context.Context, client client.APIClient, id string) error {
	removeOpts := types.ImageRemoveOptions{}
	if _, err := client.ImageRemove(ctx, id, removeOpts); err != nil {
		return fmt.Errorf("failed to remove the committed image %s: %s", id, err)
	}
	log.S(ctx).Infof("removed image %s", id)
	return nil
}

func (c *containerDescriptor) upload(ctx context.Context) error {
	if len(c.CommitedImageID) == 0 {
		opts := types.ContainerCommitOptions{}
		resp, err := c.client.ContainerCommit(ctx, c.ID, opts)
		if err != nil {
			return err
		}
		c.CommitedImageID = resp.ID
		c.log.Infof("committed container with new id %s", resp.ID)
	}
	tag := fmt.Sprintf("%s_%s", c.description.DealId, c.description.TaskId)

	newImg, err := c.description.Reference.WithTag(tag)
	if err != nil {
		c.log.Errorf("failed to add tag: %s", err)
		return err
	}

	c.log.Infof("tagging image %s from %s", newImg.String(), c.CommitedImageID)
	err = c.client.ImageTag(ctx, c.CommitedImageID, newImg.String())
	if err != nil {
		c.log.Errorf("failed to tag image: %s", err)
		return err
	}

	if c.description.PushOnStop {
		options := types.ImagePushOptions{
			RegistryAuth: c.description.Auth,
		}

		c.log.Infof("pushing image %s", newImg)
		reader, err := c.client.ImagePush(ctx, newImg.String(), options)
		if err != nil {
			c.log.Errorf("failed to push image: %s", err)
			return err
		}
		defer reader.Close()
		buf := bytes.NewBuffer(nil)
		if _, err := io.CopyN(buf, reader, units.MB); err != nil && err != io.EOF {
			return fmt.Errorf("pushing image: failed to read Docker's response: %v", err)
		}
		c.log.Infof("pushed image: %s", buf.String())
	}

	return nil
}
