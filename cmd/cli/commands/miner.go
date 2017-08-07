package commands

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/sonm-io/core/proto/hub"
)

func init() {
	minerRootCmd.AddCommand(minersListCmd, minerStatusCmd)
}

var minerRootCmd = &cobra.Command{
	Use:     "miner",
	Short:   "Operations with miners",
	PreRunE: checkHubAddressIsSet,
}

var minersListCmd = &cobra.Command{
	Use:     "list",
	Short:   "Show connected miners",
	PreRunE: minerRootCmd.PreRunE,
	Run: func(cmd *cobra.Command, args []string) {
		cc, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError("Cannot create connection", err)
			return
		}
		defer cc.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()
		lr, err := pb.NewHubClient(cc).List(ctx, &pb.ListRequest{})
		if err != nil {
			showError("Cannot get miners list", err)
			return
		}

		for addr, meta := range lr.Info {
			fmt.Printf("Miner: %s\r\n", addr)

			if len(meta.Values) == 0 {
				fmt.Println("Miner is idle")
			} else {
				fmt.Println("Tasks:")
				for i, task := range meta.Values {
					fmt.Printf("  %d) %s\r\n", i+1, task)
				}
			}
		}
	},
}

var minerStatusCmd = &cobra.Command{
	Use:     "status <miner_addr>",
	Short:   "Miner status",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errMinerAddressRequired
		}
		minerID := args[0]

		conn, err := grpc.Dial(hubAddress, grpc.WithInsecure())
		if err != nil {
			showError("Cannot create connection", err)
			return nil
		}
		defer conn.Close()

		ctx, cancel := context.WithTimeout(gctx, timeout)
		defer cancel()

		var req = pb.InfoRequest{Miner: minerID}
		metrics, err := pb.NewHubClient(conn).Info(ctx, &req)
		if err != nil {
			showError("Cannot get miner status", err)
			return nil
		}

		if len(metrics.Stats) == 0 {
			fmt.Println("Miner is idle")
		} else {
			fmt.Println("Miner tasks:")
			for task, stat := range metrics.Stats {
				// fixme: what the hell with this ID?
				fmt.Printf("  ID: %s\r\n", task)
				fmt.Printf("      CPU: %d\r\n", stat.CPU.TotalUsage)
				fmt.Printf("      RAM: %s\r\n", humanize.Bytes(stat.Memory.MaxUsage))
			}
		}

		return nil
	},
}
