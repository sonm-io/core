package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

// Runner func starts service wrapper by Cobra command.
type Runner func() error

func NewCmd(name, appVersion string, cfg *string, version *bool, run Runner) *cobra.Command {
	c := &cobra.Command{
		Use: appName(name),
		Run: func(_ *cobra.Command, _ []string) {
			if version != nil && *version {
				fmt.Println(versionString(name, appVersion))
				return
			}

			if err := run(); err != nil {
				fmt.Println(capitalize(err.Error()))
				os.Exit(1)
			}
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(cfg, "config", configName(name), configFlagHelp(name))
	c.PersistentFlags().BoolVar(version, "version", false, versionFlagHelp(name))

	return c
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func appName(name string) string {
	return "sonm" + name
}

func configName(name string) string {
	return name + ".yaml"
}

func configFlagHelp(name string) string {
	return fmt.Sprintf("Path to the %s config file", name)
}

func versionFlagHelp(name string) string {
	return fmt.Sprintf("Show %s version and exit", name)
}

func versionString(name, appVersion string) string {
	return fmt.Sprintf("%s %s (%s)", name, appVersion, util.GetPlatformName())
}
