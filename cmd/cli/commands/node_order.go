package commands

import "github.com/spf13/cobra"

func init() {
	nodeOrderRootCmd.AddCommand(
		nodeOrderListCmd,
		nodeOrderCreateCmd,
		nodeOrderGetCmd,
		nodeOrderDeleteCmd,
	)
}

var nodeOrderRootCmd = &cobra.Command{
	Use:     "ask-plan",
	Short:   "Operations with ask order plan",
	PreRunE: checkNodeAddressIsSet,
}

var nodeOrderListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show current ask plans",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeOrderCreateCmd = &cobra.Command{
	Use:     "create",
	Short:   "Create new plan",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeOrderGetCmd = &cobra.Command{
	Use:     "get <plan_id>",
	Short:   "Get plan details",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var nodeOrderDeleteCmd = &cobra.Command{
	Use:     "delete <plan_id>",
	Short:   "Delete plan",
	PreRunE: checkNodeAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {

	},
}
