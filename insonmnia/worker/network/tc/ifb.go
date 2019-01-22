package tc

import (
	"fmt"
	"os/exec"
	"strings"
)

const (
	modprobe      = "modprobe"
	ifbModuleName = "ifb"
)

// IFBInit loads the Intermediate Functional Block device kernel module.
//
// It is used in conjunction with HTB queueing discipline to allow for
// queueing incoming traffic for shaping instead of dropping.
func IFBInit() error {
	return execModProbe(ifbModuleName, "numifbs=0")
}

// IFBClose unloads the Intermediate Functional Block device kernel module.
func IFBClose() error {
	return execModProbe(ifbModuleName, "-r")
}

// IFBFlush completely reloads the Intermediate Functional Block device kernel
// module.
func IFBFlush() error {
	if err := IFBClose(); err != nil {
		return err
	}
	if err := IFBInit(); err != nil {
		return err
	}

	return nil
}

func execModProbe(args ...string) error {
	bin, err := exec.LookPath(modprobe)
	if err != nil {
		return fmt.Errorf("failed to find `modprobe`: %s", err)
	}

	cmd := exec.Command(bin, args...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute `modprobe %s`: %v", strings.Join(args, " "), err)
	}

	return nil
}
