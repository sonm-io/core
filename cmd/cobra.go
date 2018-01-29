package cmd

import (
	"fmt"
	"os"

	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

// Runner func starts service wrapper by Cobra command
type Runner func()

// NewCmd returns new cobra.Command with --config and --version flags attached and output set to os.Stdout
func NewCmd(name, appVersion string, conf *string, version *bool, run Runner) *cobra.Command {
	c := &cobra.Command{
		Use: name,
		Run: func(_ *cobra.Command, _ []string) {
			if *version {
				fmt.Printf("%s %s (%s)\r\n", name, appVersion, util.GetPlatformName())
				return
			}

			run()
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(conf, "config", "node.yaml", "Path to the Node config file")
	c.PersistentFlags().BoolVar(version, "version", false, "Show Node version and exit")

	return c
}
