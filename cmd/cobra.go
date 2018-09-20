package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	appVersion string
	app        = AppContext{
		Name:    path.Base(os.Args[0]),
		Version: appVersion,
	}
	showVersion bool
)

// Runner func starts service wrapper by Cobra command.
type Runner func(appContext AppContext) error

// AppContext represents the application context.
type AppContext struct {
	// Name holds the current application name.
	Name string
	// ConfigPath holds the current application's config path, if any.
	ConfigPath string
	// Version holds string representation of a current application's version.
	Version string
}

// RunWithHomeExpand runs the given runner, expanding the config path field
// by taking in action HOME directory symbol (~) if it exists.
func (m *AppContext) RunWithHomeExpand(runner Runner) error {
	expandedCfg, err := homedir.Expand(app.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse config path: %v", err)
	}

	app.ConfigPath = expandedCfg
	return runner(app)
}

func NewCmd(runner Runner) *cobra.Command {
	c := &cobra.Command{
		Use: app.Name,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Println(versionString(app.Name, app.Version))
				os.Exit(0)
			}
			return checkRequiredFlags(cmd.PersistentFlags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := app.RunWithHomeExpand(runner); err != nil {
				fmt.Println(capitalize(err.Error()))
				os.Exit(1)
			}
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(&app.ConfigPath, "config", "", configFlagHelp())
	c.PersistentFlags().BoolVar(&showVersion, "version", false, versionFlagHelp())

	c.MarkPersistentFlagRequired("config")

	return c
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	return strings.ToUpper(s[:1]) + s[1:]
}

func configFlagHelp() string {
	return "Path to the config file (required)"
}

func versionFlagHelp() string {
	return "Show version and exit"
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
