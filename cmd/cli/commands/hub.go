package commands

import (
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/sonm-io/core/proto/hub"
)

func init() {
	hubRootCmd.AddCommand(hubPingCmd, hubStatusCmd)
}

// --- hub commands
var hubRootCmd = &cobra.Command{
	Use:     "hub",
	Short:   "Operations with hub",
	PreRunE: checkHubAddressIsSet,
}

var hubPingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Ping the hub",
	PreRunE: checkHubAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError("Cannot create connection", err)
			return
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		_, err = pb.NewHubClient(cc).Ping(ctx, &pb.PingRequest{})
		if err != nil {
			showError("Ping failed", err)
			return
		}

		showOk()
		// fmt.Printf("Ping hub %s... OK\r\n", hubAddress)
	},
}

var hubStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show hub status",
	PreRunE: checkHubAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		// todo: implement this on hub
		showOk()
		// fmt.Printf("Hub %s status: OK\r\n", hubAddress)
	},
}
