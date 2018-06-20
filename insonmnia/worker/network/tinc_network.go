package network

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

// A comparable type for v4 address
type IP4 struct {
	a, b, c, d byte
}

func newIP4(ip net.IP) IP4 {
	v4 := ip.To4()
	return IP4{v4[0], v4[1], v4[2], v4[3]}
}

func (i *IP4) ToCommon() net.IP {
	return net.IP{i.a, i.b, i.c, i.d}
}

func (t *TincNetwork) Init(ctx context.Context) error {
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.NodeID, "-c", t.ConfigPath, "init", "initial_node_"+t.NodeID)
	if err != nil {
		t.logger.Errorf("failed to init network: %s", err)
	} else {
		t.logger.Info("succesefully initialized tinc network")
	}
	return err
}

func (t *TincNetwork) Join(ctx context.Context) error {
	if len(t.Invitation) == 0 {
		return errors.New("can not join to network without invitation")
	}
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.NodeID, "-c", t.ConfigPath, "join", t.Invitation)
	if err != nil {
		t.logger.Errorf("failed to join network: %s", err)
	} else {
		t.logger.Info("succesefully joined tinc network")
	}
	return err
}

func (t *TincNetwork) Start(ctx context.Context, addr string) error {
	iface := t.NodeID[:15]

	err := t.runCommand(ctx, "tinc", "-n", t.NodeID, "-c", t.ConfigPath, "start",
		"-o", "Interface="+iface, "-o", "Subnet="+t.Pool.String(), "-o", "Subnet="+addr+"/32", "-o", "LogLevel=0")
	if err != nil {
		t.logger.Errorf("failed to start tinc: %s", err)
	} else {
		t.logger.Info("started tinc")
	}
	return err
}

func (t *TincNetwork) Shutdown(ctx context.Context) error {
	t.cli.ContainerKill(ctx, t.TincContainerID, "SIGKILL")
	return nil
}

func (t *TincNetwork) Stop(ctx context.Context) error {
	err := t.runCommand(ctx, "tinc", "--batch", "-n", t.NodeID, "-c", t.ConfigPath, "stop")
	if err != nil {
		t.logger.Errorf("failed to stop tinc: %s", err)
		return err
	} else {
		t.logger.Info("successfully stoppped tinc")
	}
	return err
}

func (t *TincNetwork) Invite(ctx context.Context, inviteeID string) (string, error) {
	out, _, err := t.runCommandWithOutput(ctx, "tinc", "--batch", "-n", t.NodeID, "-c", t.ConfigPath, "invite", inviteeID)
	out = strings.Trim(out, "\n")
	return out, err
}

func (t *TincNetwork) OccupiedIPs(ctx context.Context) (map[IP4]struct{}, error) {
	err := t.runCommand(ctx, "tinc", "-n", t.NodeID, "-c", t.ConfigPath, "start")
	if err != nil {
		return nil, err
	}
	stdout, _, err := t.runCommandWithOutput(ctx, "tinc", "-n", t.NodeID, "-c", t.ConfigPath, "dump", "subnets")
	if err != nil {
		return nil, err
	}
	err = t.runCommand(ctx, "tinc", "-n", t.NodeID, "-c", t.ConfigPath, "stop")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(stdout, "\n")
	ips := map[IP4]struct{}{}
	for _, line := range lines {
		addr := strings.Split(line, " ")[0]
		if strings.Contains(addr, "/") {
			continue
		}
		ip := net.ParseIP(addr)
		if ip == nil {
			continue
		}
		if ip.To4() == nil {
			continue
		}
		if !ip.IsGlobalUnicast() {
			continue
		}
		ips[newIP4(ip)] = struct{}{}
	}

	return ips, nil
}

func (t *TincNetwork) runCommand(ctx context.Context, name string, arg ...string) error {
	_, _, err := t.runCommandWithOutput(ctx, name, arg...)
	return err
}
func (t *TincNetwork) runCommandWithOutput(ctx context.Context, name string, arg ...string) (string, string, error) {
	cmd := append([]string{name}, arg...)
	cfg := types.ExecConfig{
		User:         "root",
		Detach:       false,
		Cmd:          cmd,
		AttachStderr: true,
		AttachStdout: true,
	}

	execId, err := t.cli.ContainerExecCreate(ctx, t.TincContainerID, cfg)
	if err != nil {
		t.logger.Warnf("ContainerExecCreate finished with error: %s", err)
		return "", "", err
	}

	conn, err := t.cli.ContainerExecAttach(ctx, execId.ID, cfg)
	if err != nil {
		t.logger.Warnf("ContainerExecAttach finished with error: %s", err)
	}
	stdoutBuf := bytes.Buffer{}
	stderrBuf := bytes.Buffer{}
	stdcopy.StdCopy(&stdoutBuf, &stderrBuf, conn.Reader)
	stdout := stdoutBuf.String()
	stderr := stderrBuf.String()

	if err != nil {
		t.logger.Warnf("failed to execute command - %s %s, stdout - %s, stderr - %s", name, arg, stdout, stderr)
		return stdout, stderr, err
	}

	inspect, err := t.cli.ContainerExecInspect(ctx, execId.ID)
	if err != nil {
		t.logger.Warnf("failed to inspect command: %s", err)
		return stdout, stderr, err
	}

	if inspect.ExitCode != 0 {
		return stdout, stderr, fmt.Errorf("failed to execute command %s %s, exit code %d, stdout - %s, stderr - %s", name, arg, inspect.ExitCode, stdout, stderr)
	} else {
		t.logger.Debugf("finished command - %s %s, stdout - %s, stderr - %s", name, arg, stdout, stderr)
		return stdout, stderr, err
	}
}
