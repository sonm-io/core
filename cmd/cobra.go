package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Runner func starts service wrapper by Cobra command
type Runner func()

// NewCmd returns new cobra.Command with --config and --version flags attached and output set to os.Stdout
func NewCmd(name string, conf *string, version *bool, run Runner) *cobra.Command {
	c := &cobra.Command{
		Use: name,
		Run: func(_ *cobra.Command, _ []string) {
			run()
		},
	}

	c.SetOutput(os.Stdout)
	c.PersistentFlags().StringVar(conf, "config", "node.yaml", "Path to the Node config file")
	c.PersistentFlags().BoolVar(version, "version", false, "Show Node version and exit")

	return c
}
