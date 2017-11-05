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
	hubWorkerACLCmd.AddCommand(registerWorkerCmd, deregisterWorkerCmd)
}

func wrapHubCommand(use, short string, command *cobra.Command) *cobra.Command {
	command.Use = use
	command.Short = short

	if command.PreRunE == nil {
		command.PreRunE = checkHubAddressIsSet
	}
	return command
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
		grpc, err := NewGrpcInteractor(hubAddress, timeout)
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
	Use:     "add PATH",
	Short:   "Add a virtual slot",
	PreRunE: checkHubAddressIsSet,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		cfg := task_config.SlotConfig{}
		err = yaml.Unmarshal(buf, &cfg)
		if err != nil {
			return err
		}

		grpc, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return err
		}
		slot, err := cfg.IntoSlot()
		if err != nil {
			return err
		}

		_, err = grpc.HubInsertSlot(context.Background(), slot)
		if err != nil {
			return err
		}

		cmd.Println("OK")
		return nil
	},
}

var hubWorkerACLCmd = wrapHubCommand("acl", "Worker ACL management", &cobra.Command{})

var registerWorkerCmd = wrapHubCommand("register", "Registers a worker credentials", &cobra.Command{
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		hub, err := NewHubInteractor(nodeAddress, timeout)
		if err != nil {
			return err
		}

		reply, err := hub.RegisterWorker(id)
		if err != nil {
			return err
		}

		dump, err := json.Marshal(reply)
		if err != nil {
			return err
		}

		cmd.Println(string(dump))

		return nil
	},
})

var deregisterWorkerCmd = wrapHubCommand("deregister", "Deregisters a worker credentials", &cobra.Command{
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		hub, err := NewHubInteractor(nodeAddress, timeout)
		if err != nil {
			return err
		}

		reply, err := hub.DeregisterWorker(id)
		if err != nil {
			return err
		}

		dump, err := json.Marshal(reply)
		if err != nil {
			return err
		}

		cmd.Println(string(dump))

		return nil
	},
})

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
