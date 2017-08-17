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
		if len(metrics.GetUsage()) == 0 {
			cmd.Println("Miner is idle")
		} else {
			cmd.Println("Miner tasks:")
			for task, usage := range metrics.Usage {
				cmd.Printf("  ID: %s\r\n", task)
				cmd.Printf("      CPU: %d\r\n", usage.Cpu.Total)
				cmd.Printf("      RAM: %s\r\n")
				cmd.Printf("      CPU: %d\r\n", usage.Cpu.Total)
				cmd.Printf("      RAM: %s\r\n", ds.ByteSize(usage.Memory.MaxUsage).String())
				cmd.Printf("      NET:\r\n")

				for i, net := range usage.Network {
					cmd.Printf("          %s:\r\n", i)
					cmd.Printf("            Tx/Rx bytes: %d/%d\r\n", net.TxBytes, net.RxBytes)
					cmd.Printf("            Tx/Rx packets: %d/%d\r\n", net.TxPackets, net.RxPackets)
					cmd.Printf("            Tx/Rx errors: %d/%d\r\n", net.TxErrors, net.RxErrors)
					cmd.Printf("            Tx/Rx dropped: %d/%d\r\n", net.TxDropped, net.RxDropped)
				}
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
