package main

import (
	"fmt"
	"os"

	_ "github.com/sonm-io/core/cmd/autocli/proto"
	"github.com/sonm-io/core/util/xcode"
)

func main() {
	if err := xcode.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
