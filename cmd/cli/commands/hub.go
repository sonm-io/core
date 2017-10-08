package commands

import (
	"time"

	"encoding/json"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

func init() {
	hubRootCmd.AddCommand(hubPingCmd, hubStatusCmd, hubShowSlotsCmd)
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
		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return
		}
		hubPingCmdRunner(cmd, itr)
	},
}

var hubStatusCmd = &cobra.Command{
	Use:     "status",
	Short:   "Show hub status",
	PreRunE: checkHubAddressIsSet,
	Run: func(cmd *cobra.Command, args []string) {
		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return
		}

		hubStatusCmdRunner(cmd, itr)
	},
}

var hubShowSlotsCmd = &cobra.Command{
	Use:     "slots",
	Short:   "Show hub's virtual slots",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		itr, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return err
		}

		return hubShowSlotsCmdRunner(cmd, itr, args)
	},
}

func hubPingCmdRunner(cmd *cobra.Command, interactor CliInteractor) {
	_, err := interactor.HubPing(context.Background())
	if err != nil {
		showError(cmd, "Ping failed", err)
		return
	}

	showOk(cmd)
}

func hubStatusCmdRunner(cmd *cobra.Command, interactor CliInteractor) {
	// todo: implement this on hub
	stat, err := interactor.HubStatus(context.Background())
	if err != nil {
		showError(cmd, "Cannot get status", err)
		return
	}

	printHubStatus(cmd, stat)
}

func hubShowSlotsCmdRunner(cmd *cobra.Command, interactor CliInteractor, IDs []string) error {
	ctx := context.Background()
	if len(IDs) == 0 {
		miners, err := interactor.MinerList(ctx)
		if err != nil {
			return err
		}
		for id := range miners.Info {
			IDs = append(IDs, id)
		}
	}

	result := map[string][]*pb.Slot{}
	for _, id := range IDs {
		slots, err := interactor.HubSlotsShow(context.Background(), id)
		if err != nil {
			return err
		}

		result[id] = slots.Slot
	}

	b, _ := json.Marshal(result)
	cmd.Println(string(b))
	return nil
}

func printHubStatus(cmd *cobra.Command, stat *pb.HubStatusReply) {
	if isSimpleFormat() {
		cmd.Printf("Connected miners: %d\r\n", stat.MinerCount)
		cmd.Printf("Uptime:           %s\r\n", (time.Second * time.Duration(stat.Uptime)).String())
		cmd.Printf("Version:          %s %s\r\n", stat.Version, stat.Platform)
		cmd.Printf("Eth address:      %s\r\n", stat.EthAddr)
	} else {
		b, _ := json.Marshal(stat)
		cmd.Println(string(b))
	}
}
