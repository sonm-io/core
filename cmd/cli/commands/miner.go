package commands

import (
	"encoding/json"

	ds "github.com/c2h5oh/datasize"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func init() {
	minerRootCmd.AddCommand(minersListCmd, minerStatusCmd)
}

func printMinerList(cmd *cobra.Command, lr *pb.ListReply) {
	if isSimpleFormat() {
		if len(lr.Info) == 0 {
			cmd.Printf("No miners connected\r\n")
			return
		}

		for addr, meta := range lr.Info {
			cmd.Printf("Miner: %s\r\n", addr)

			if len(meta.Values) == 0 {
				cmd.Println("Miner is idle")
			} else {
				cmd.Println("Tasks:")
				for i, task := range meta.Values {
					cmd.Printf("  %d) %s\r\n", i+1, task)
				}
			}
		}
	} else {
		b, _ := json.Marshal(lr)
		cmd.Println(string(b))
	}
}

func printMinerStatus(cmd *cobra.Command, metrics *pb.InfoReply) {
	if isSimpleFormat() {
		if len(metrics.Stats) == 0 {
			cmd.Println("Miner is idle")
		} else {
			cmd.Println("Miner tasks:")
			for task, stat := range metrics.Stats {
				cmd.Printf("  ID: %s\r\n", task)
				cmd.Printf("      CPU: %d\r\n", stat.CPU.TotalUsage)
				cmd.Printf("      RAM: %s\r\n", ds.ByteSize(stat.Memory.MaxUsage).String())
			}
		}
	} else {
		b, _ := json.Marshal(metrics)
		cmd.Println(string(b))
	}
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
		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return
		}

		minerListCmdRunner(cmd, itr)
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

		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return nil
		}

		minerStatusCmdRunner(cmd, minerID, itr)
		return nil
	},
}

func minerListCmdRunner(cmd *cobra.Command, interactor CliInteractor) {
	list, err := interactor.MinerList(context.Background())
	if err != nil {
		showError(cmd, "Cannot get miners list", err)
		return
	}

	printMinerList(cmd, list)
}

func minerStatusCmdRunner(cmd *cobra.Command, minerID string, interactor CliInteractor) {
	metrics, err := interactor.MinerStatus(minerID, context.Background())
	if err != nil {
		showError(cmd, "Cannot get miner status", err)
		return
	}

	printMinerStatus(cmd, metrics)
}
