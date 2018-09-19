package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Runner func starts service wrapper by Cobra command.
type Runner func() error

func NewCmd(name, appVersion string, cfg *string, version *bool, run Runner) *cobra.Command {
	wrapped := func() error {
		expandedCfg, err := homedir.Expand(*cfg)
		if err != nil {
			return fmt.Errorf("failed to parse config path: %v", err)
		}
		*cfg = expandedCfg
		return run()
	}

	c := &cobra.Command{
		Use: appName(name),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if version != nil && *version {
				fmt.Println(versionString(name, appVersion))
				os.Exit(0)
			}
			return checkRequiredFlags(cmd.PersistentFlags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := wrapped(); err != nil {
				fmt.Println(capitalize(err.Error()))
				os.Exit(1)
			}
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(cfg, "config", "", configFlagHelp(name))
	c.PersistentFlags().BoolVar(version, "version", false, versionFlagHelp(name))

	c.MarkPersistentFlagRequired("config")

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

func configFlagHelp(name string) string {
	return fmt.Sprintf("Path to the %s config file (required)", name)
}

func versionFlagHelp(name string) string {
	return fmt.Sprintf("Show %s version and exit", name)
}

func versionString(name, appVersion string) string {
	return fmt.Sprintf("%s %s (%s)", name, appVersion, util.GetPlatformName())
}

func checkRequiredFlags(flags *pflag.FlagSet) error {
	requiredError := false
	flagName := ""

	flags.VisitAll(func(flag *pflag.Flag) {
		requiredAnnotation := flag.Annotations[cobra.BashCompOneRequiredFlag]
		if len(requiredAnnotation) == 0 {
			return
		}

		flagRequired := requiredAnnotation[0] == "true"

		if flagRequired && !flag.Changed {
			requiredError = true
			flagName = flag.Name
		}
	})

	if requiredError {
		return fmt.Errorf("required flag `%s` has not been set", flagName)
	}

	return nil
}
