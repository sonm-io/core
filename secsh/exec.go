package secsh

import (
	"bytes"
	"io"
	"os/exec"
)

func parsePipedCommand(args []string) ([][]string, error) {
	commands := make([][]string, 0)
	currentCommand := make([]string, 0)
	for _, arg := range args {
		if arg == "|" {
			if len(currentCommand) != 0 {
				commands = append(commands, currentCommand)
				currentCommand = nil
			}
			continue
		}

		currentCommand = append(currentCommand, arg)
	}

	if len(currentCommand) != 0 {
		commands = append(commands, currentCommand)
	}

	return commands, nil
}

func execPipedCommand(out io.Writer, cmds ...*exec.Cmd) error {
	if len(cmds) == 0 {
		return nil
	}

	var errBuffer bytes.Buffer
	pipes := make([]*io.PipeWriter, len(cmds)-1)
	i := 0
	for ; i < len(cmds)-1; i++ {
		stdinPipe, stdoutPipe := io.Pipe()

		cmds[i].Stdout = stdoutPipe
		cmds[i].Stderr = &errBuffer
		cmds[i+1].Stdin = stdinPipe
		pipes[i] = stdoutPipe
	}
	cmds[i].Stdout = out
	cmds[i].Stderr = &errBuffer

	if err := execNext(cmds, pipes); err != nil {
		return err
	}

	return nil
}

func execNext(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
	if stack[0].Process == nil {
		if err = stack[0].Start(); err != nil {
			return err
		}
	}

	if len(stack) > 1 {
		if err = stack[1].Start(); err != nil {
			return err
		}
		defer func() {
			if err == nil {
				_ = pipes[0].Close()
				err = execNext(stack[1:], pipes[1:])
			}
		}()
	}

	return stack[0].Wait()
}
