package main

import (
	"fmt"
	"os"

	_ "github.com/sonm-io/core/proto"
	"github.com/sshaman1101/grpccmd"
)

func main() {
	grpccmd.SetCmdInfo("autocli", "Call SONM services directly")
	if err := grpccmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
