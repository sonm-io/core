package commands

import (
	"time"

	"encoding/json"
	"io/ioutil"

	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

func init() {
	hubRootCmd.AddCommand(hubPingCmd, hubStatusCmd, hubSlotCmd)
	hubSlotCmd.AddCommand(minerShowSlotsCmd, hubAddSlotCmd)
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
		itr, err := NewGrpcInteractor(hubAddressFlag, timeoutFlag)
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
		itr, err := NewGrpcInteractor(hubAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to hub", err)
			return
		}

		hubStatusCmdRunner(cmd, itr)
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
	stat, err := interactor.HubStatus(context.Background())
	if err != nil {
		showError(cmd, "Cannot get status", err)
		return
	}

	printHubStatus(cmd, stat)
}

var hubSlotCmd = &cobra.Command{
	Use:     "slot",
	Short:   "Show hub's virtual slots",
	PreRunE: checkHubAddressIsSet,
}

var minerShowSlotsCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show hub's virtual slots",
	Args:    cobra.MaximumNArgs(0),
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, args []string) error {
		grpc, err := NewGrpcInteractor(hubAddressFlag, timeoutFlag)
		if err != nil {
			return err
		}

		slots, err := grpc.HubShowSlots(context.Background())
		if err != nil {
			return err
		}

		dump, err := json.Marshal(slots.Slot)
		if err != nil {
			return err
		}
		cmd.Println(string(dump))
		return nil
	},
}

var hubAddSlotCmd = &cobra.Command{
	Use:     "add PRICE PATH",
	Short:   "Add a virtual slot",
	PreRunE: checkHubAddressIsSet,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		price := args[0]
		path := args[1]

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		cfg := task_config.SlotConfig{}
		err = yaml.Unmarshal(buf, &cfg)
		if err != nil {
			return err
		}

		grpc, err := NewGrpcInteractor(hubAddressFlag, timeoutFlag)
		if err != nil {
			return err
		}
		slot, err := cfg.IntoSlot()
		if err != nil {
			return err
		}

		id, err := grpc.HubInsertSlot(context.Background(), slot, price)
		if err != nil {
			return err
		}

		cmd.Printf("id = %s\r\n", id.Id)
		return nil
	},
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
