package worker

import (
	"context"
	"fmt"
	"io"

	"github.com/sonm-io/core/insonmnia/worker/plugin"
	"github.com/sonm-io/core/util/multierror"
	"go.uber.org/zap"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/gliderlabs/ssh"
	log "github.com/noxiouz/zapctx/ctxlog"
)

type containerDescriptor struct {
	ctx    context.Context
	cancel context.CancelFunc

	client *client.Client

	ID              string
	CommitedImageID string
	description     Description
	stats           types.StatsJSON

	cleanup plugin.Cleanup
}

func newContainer(ctx context.Context, dockerClient *client.Client, d Description, tuners *plugin.Repository) (*containerDescriptor, error) {
	log.S(ctx).Infof("start container with application, reference %s", d.Reference.String())

	ctx, cancel := context.WithCancel(ctx)
	cont := containerDescriptor{
		ctx:         ctx,
		cancel:      cancel,
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
		Labels:  map[string]string{overseerTag: ""},
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
		RestartPolicy:   d.RestartPolicy,
		AutoRemove:      d.autoremove,
		Resources:       d.Resources.ToHostConfigResources(d.CGroupParent),
	}

	networkingConfig := network.NetworkingConfig{}

	cleanup, err := tuners.Tune(ctx, &d, &hostConfig, &networkingConfig)
	if err != nil {
		log.G(ctx).Error("failed to tune container", zap.Error(err))
		return nil, err
	}

	// create new container
	// assign resulted containerid
	// log all warnings
	resp, err := cont.client.ContainerCreate(ctx, &config, &hostConfig, &networkingConfig, "")
	if err != nil {
		return nil, err
	}
	cont.ID = resp.ID
	cont.ctx = log.WithLogger(cont.ctx, log.G(ctx).With(zap.String("id", cont.ID)))
	cont.cleanup = cleanup
	if len(resp.Warnings) > 0 {
		log.G(ctx).Warn("ContainerCreate finished with warnings", zap.Strings("warnings", resp.Warnings))
	}

	return &cont, nil
}

func (c *containerDescriptor) startContainer() error {
	var options types.ContainerStartOptions
	if err := c.client.ContainerStart(c.ctx, c.ID, options); err != nil {
		log.G(c.ctx).Warn("ContainerStart finished with error", zap.Error(err))
		c.cancel()
		return err
	}
	return nil
}

func (c *containerDescriptor) execCommand(cmd []string, env []string, isTty bool, wCh <-chan ssh.Window) (conn types.HijackedResponse, err error) {
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

	log.G(c.ctx).Info("attaching command", zap.Any("config", cfg))

	execId, err := c.client.ContainerExecCreate(c.ctx, c.ID, cfg)
	if err != nil {
		log.G(c.ctx).Warn("ContainerExecCreate finished with error", zap.Error(err))
		return
	}

	conn, err = c.client.ContainerExecAttach(c.ctx, execId.ID, cfg)
	if err != nil {
		log.G(c.ctx).Warn("ContainerExecAttach finished with error", zap.Error(err))
	}

	err = c.client.ContainerExecStart(c.ctx, execId.ID, types.ExecStartCheck{Detach: false, Tty: true})
	if err != nil {
		log.G(c.ctx).Warn("ContainerExecStart finished with error", zap.Error(err))
		return
	}
	go func() {
		for {
			select {
			case w, ok := <-wCh:
				if !ok {
					return
				}
				log.G(c.ctx).Info("resizing tty", zap.Int("height", w.Height), zap.Int("width", w.Width))
				err = c.client.ContainerExecResize(c.ctx, execId.ID, types.ResizeOptions{Height: uint(w.Height), Width: uint(w.Width)})
				if err != nil {
					log.G(c.ctx).Warn("ContainerExecResize finished with error", zap.Error(err))
				}
			case <-c.ctx.Done():
				return
			}
		}
	}()

	log.G(c.ctx).Info("attached command to container")
	return
}

func (c *containerDescriptor) Kill() (err error) {
	log.G(c.ctx).Info("kill the container", zap.String("id", c.ID))
	if err = c.client.ContainerKill(context.Background(), c.ID, "SIGKILL"); err != nil {
		log.G(c.ctx).Error("failed to send SIGKILL to the container", zap.String("id", c.ID), zap.Error(err))
		return err
	}
	return nil
}

func (c *containerDescriptor) Remove() error {
	log.G(c.ctx).Info("remove the container", zap.String("id", c.ID))
	result := multierror.NewMultiError()
	if err := containerRemove(c.ctx, c.client, c.ID); err != nil {
		result = multierror.Append(result, err)
	}
	if len(c.CommitedImageID) != 0 {
		if err := imageRemove(c.ctx, c.client, c.CommitedImageID); err != nil {
			result = multierror.Append(result, err)
		}
	}
	return result.ErrorOrNil()
}

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

func (c *containerDescriptor) upload() error {
	opts := types.ContainerCommitOptions{}
	resp, err := c.client.ContainerCommit(c.ctx, c.ID, opts)
	if err != nil {
		return err
	}
	c.CommitedImageID = resp.ID
	log.G(c.ctx).Info("committed container", zap.String("id", c.ID), zap.String("newId", resp.ID))

	tag := fmt.Sprintf("%s_%s", c.description.DealId, c.description.TaskId)

	newImg, err := reference.WithTag(c.description.Reference, tag)
	if err != nil {
		log.G(c.ctx).Error("failed to add tag", zap.String("id", resp.ID), zap.Error(err))
		return err
	}

	log.G(c.ctx).Info("tagging image", zap.String("from", resp.ID), zap.Stringer("to", newImg))
	err = c.client.ImageTag(c.ctx, resp.ID, newImg.String())
	if err != nil {
		log.G(c.ctx).Error("failed to tag image", zap.String("id", resp.ID), zap.Any("name", newImg), zap.Error(err))
		return err
	}

	options := types.ImagePushOptions{
		RegistryAuth: c.description.Auth,
	}

	log.G(c.ctx).Info("pushing image", zap.Any("name", newImg))
	reader, err := c.client.ImagePush(c.ctx, newImg.String(), options)
	if err != nil {
		log.G(c.ctx).Error("failed to push image", zap.Any("name", newImg), zap.Error(err))
		return err
	}
	defer reader.Close()
	buffer := make([]byte, 100*1024)
	for {
		readCnt, err := reader.Read(buffer)
		if readCnt != 0 {
			log.G(c.ctx).Info(string(buffer[:readCnt]))
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
