package miner

import (
	"context"
	"io"
	"path/filepath"

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

	ID          string
	description Description
	stats       types.StatsJSON
}

func newContainer(ctx context.Context, dockerClient *client.Client, d Description, tuner nvidiaGPUTuner) (*containerDescriptor, error) {
	log.G(ctx).Info("start container with application")

	ctx, cancel := context.WithCancel(ctx)
	cont := containerDescriptor{
		ctx:         ctx,
		cancel:      cancel,
		client:      dockerClient,
		description: d,
	}

	// NOTE: command to launch must be specified via ENTRYPOINT and CMD in Dockerfile
	var config = container.Config{
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,

		Image: filepath.Join(d.Registry, d.Image),
		// TODO: set actual name
		Labels:  map[string]string{overseerTag: ""},
		Env:     d.Env,
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
		RestartPolicy:   d.RestartPolicy,
		// NOTE: we perform cleanup after commit manually
		AutoRemove: false,
		Resources: container.Resources{
			// TODO: accept a name of a cgroup cooked by user
			// NOTE: on non-Linux platform it's empty
			CgroupParent: parentCgroup,
			Memory:       d.Resources.Memory,
			NanoCPUs:     d.Resources.NanoCPUs,
		},
	}

	var networkingConfig network.NetworkingConfig

	if err := tuner.Tune(&hostConfig); err != nil {
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
				log.G(c.ctx).Info("resising tty", zap.Int("height", w.Height), zap.Int("width", w.Width))
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

func (c *containerDescriptor) remove() {
	containerRemove(c.ctx, c.client, c.ID)
}

func containerRemove(ctx context.Context, client client.APIClient, id string) {
	removeOpts := types.ContainerRemoveOptions{}
	if err := client.ContainerRemove(ctx, id, removeOpts); err != nil {
		log.G(ctx).Error("failed to remove the container", zap.String("id", id), zap.Error(err))
	}
}

func (c *containerDescriptor) upload() error {
	opts := types.ContainerCommitOptions{}
	resp, err := c.client.ContainerCommit(c.ctx, c.ID, opts)
	if err != nil {
		return err
	}
	log.G(c.ctx).Info("commited container", zap.String("id", c.ID), zap.String("newId", resp.ID))

	image := filepath.Join(c.description.Registry, c.description.Image)
	named, err := reference.ParseNormalizedNamed(image)
	if err != nil {
		log.G(c.ctx).Error("failed to parse", zap.String("image", image), zap.Error(err))
		return err
	}

	newImg, err := reference.WithTag(named, c.description.TaskId)
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
