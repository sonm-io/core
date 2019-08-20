package secsh

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"
	"strings"

	"github.com/sonm-io/core/proto"
	"go.uber.org/zap"
)

type RemotePTYService struct {
	execPath   string
	policyPath string
	log        *zap.SugaredLogger
}

func (m *RemotePTYService) Banner(ctx context.Context, request *sonm.RemotePTYBannerRequest) (*sonm.RemotePTYBannerResponse, error) {
	cmds, err := m.commandsList()
	if err != nil {
		return nil, err
	}

	banner := NewBanner(ctx, m.log)
	banner.AddLine("")
	banner.AddLine(fmt.Sprintf("List of available commands: %s", strings.Join(cmds, ", ")))

	return &sonm.RemotePTYBannerResponse{Banner: banner.String()}, nil
}

func (m *RemotePTYService) commandsList() ([]string, error) {
	fileInfoList, err := ioutil.ReadDir(m.policyPath)
	if err != nil {
		return nil, err
	}

	cmds := make([]string, 0)
	for _, file := range fileInfoList {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		name := strings.Replace(path.Base(file.Name()), ".yaml", "", -1)

		cmds = append(cmds, name)
	}

	return cmds, nil
}

func (m *RemotePTYService) Exec(request *sonm.RemotePTYExecRequest, server sonm.RemotePTY_ExecServer) error {
	return m.execCmd(server.Context(), request.Args, &execStream{Server: server})
}

func (m *RemotePTYService) execCmd(ctx context.Context, args []string, stream *execStream) error {
	if len(args) == 0 {
		return fmt.Errorf("nothing to execute")
	}

	if len(args) == 1 && args[0] == "help" {
		return m.returnHelp(stream)
	}

	commands, err := parsePipedCommand(args)
	if err != nil {
		return err
	}

	cmds := make([]*exec.Cmd, len(commands))
	for id, command := range commands {
		args, err := m.prepareArguments(command)
		if err != nil {
			return err
		}

		cmds[id] = exec.CommandContext(ctx, m.execPath, args...)
		m.log.Debugf("%s", cmds[id].Args)
	}

	return execPipedCommand(stream, cmds...)
}

func (m *RemotePTYService) prepareArguments(args []string) ([]string, error) {
	execFullPath, err := m.resolveExecPath(args[0])
	if err != nil {
		return nil, err
	}

	return append([]string{m.policyPath, execFullPath}, args[1:]...), nil
}

func (m *RemotePTYService) resolveExecPath(execName string) (string, error) {
	p, err := exec.LookPath(execName)
	if err != nil {
		return "", err
	}

	return p, nil
}

func (m *RemotePTYService) returnHelp(stream *execStream) error {
	files, err := ioutil.ReadDir(m.policyPath)
	if err != nil {
		return err
	}

	out := fmt.Sprintf("List of available commands:\n\r")
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if !strings.HasSuffix(file.Name(), ".yaml") {
			continue
		}

		name := strings.Replace(path.Base(file.Name()), ".yaml", "", -1)

		out += fmt.Sprintf("- %s\n\r", name)
	}

	return stream.Server.Send(&sonm.RemotePTYExecResponseChunk{
		Out:  []byte(out),
		Done: true,
	})
}
