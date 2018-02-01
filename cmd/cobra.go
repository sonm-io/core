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
	appName := "sonm" + name
	configName := name + ".yaml"
	configFlagHelp := fmt.Sprintf("Path to the %s config file", name)
	versionFlagHelp := fmt.Sprintf("Show %s version and exit", name)

	c := &cobra.Command{
		Use: appName,
		Run: func(_ *cobra.Command, _ []string) {
			if version != nil && *version {
				fmt.Printf("%s %s (%s)\r\n", name, appVersion, util.GetPlatformName())
				return
			}

			run()
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(conf, "config", configName, configFlagHelp)
	c.PersistentFlags().BoolVar(version, "version", false, versionFlagHelp)

	return c
}
