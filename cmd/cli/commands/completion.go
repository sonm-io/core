package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var autoCompleteCmd = &cobra.Command{
	Use:   "completion <bash|zsh>",
	Short: "Generate shell-completion script",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		shell := args[0]

		switch shell {
		case "zsh":
			err = rootCmd.GenZshCompletion(os.Stdout)
		case "bash":
			err = rootCmd.GenBashCompletion(os.Stdout)
		default:
			err = fmt.Errorf("unknown shell type `%s`", shell)
		}

		return err
	},
}
