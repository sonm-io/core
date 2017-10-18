package commands

import "github.com/spf13/cobra"

func init() {
	nodeACLRootCmd.AddCommand(
		nodeACLListCmd,
		nodeACLRegisterCmd,
		nodeACLUnregisterCmd,
	)
}

var nodeACLRootCmd = &cobra.Command{
	Use:     "acl",
	Short:   "Operations with Access Control Lists",
	PreRunE: checkNodeAddressIsSet,
}

var nodeACLListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show current ACLs",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeACLRegisterCmd = &cobra.Command{
	Use:     "register <worker_id>",
	Short:   "Register new Worker",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeACLUnregisterCmd = &cobra.Command{
	Use:     "unregister <worker_id>",
	Short:   "Unregister known worker",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
