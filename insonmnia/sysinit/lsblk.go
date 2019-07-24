package sysinit

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
)

type BlockDevice struct {
	Name     string         `json:"name"`
	FsType   string         `json:"fstype"`
	Children []*BlockDevice `json:"children"`
}

func ListBlockDevices(ctx context.Context) ([]*BlockDevice, error) {
	cmd := exec.CommandContext(ctx, "lsblk", "--output", "NAME,FSTYPE", "--json")

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute %v: %v", cmd.Args, err)
	}

	type container struct {
		BlockDevices []*BlockDevice `json:"blockdevices"`
	}

	result := &container{}
	if err := json.Unmarshal(out, &result); err != nil {
		return nil, fmt.Errorf("failed to decode %v output: %v", cmd.Args, err)
	}

	return result.BlockDevices, nil
}
