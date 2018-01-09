package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		printVersion(cmd, version)
	},
}

var autoCompleteCmd = &cobra.Command{
	Use:   "completion <bash|zsh>",
	Short: "Generate shell-completion script",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		shell := args[0]

		switch shell {
		case "zsh":
			err = rootCmd.GenZshCompletion(os.Stdout)
		case "bash":
			err = rootCmd.GenBashCompletion(os.Stdout)
		default:
			showError(cmd, "Unknown shell type", nil)
			os.Exit(1)
		}

		if err != nil {
			showError(cmd, "Cannot generate completion script", nil)
			os.Exit(1)
		}
	},
}
