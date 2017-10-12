package commands

import (
	"encoding/json"
	"io/ioutil"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"

	ds "github.com/c2h5oh/datasize"
	"github.com/sonm-io/core/cmd/cli/task_config"
	pb "github.com/sonm-io/core/proto"
)

func init() {
	minerRootCmd.AddCommand(minersListCmd, minerStatusCmd, minerPropertiesCmd, minerSlotCmd)
	minerPropertiesCmd.AddCommand(minerSetPropertiesCmd, minerGetPropertiesCmd)
	minerSlotCmd.AddCommand(minerShowSlotsCmd, minerAddSlotCmd)
}

func printMinerList(cmd *cobra.Command, lr *pb.ListReply) {
	if isSimpleFormat() {
		if len(lr.Info) == 0 {
			cmd.Printf("No miners connected\r\n")
			return
		}

		for addr, meta := range lr.Info {
			cmd.Printf("Miner: %s", addr)

			taskCount := len(meta.Values)
			if taskCount == 0 {
				cmd.Printf("\t\tIdle\r\n")
			} else {
				cmd.Printf("\t\t%d active task(s)\r\n", taskCount)
			}
		}
	} else {
		b, _ := json.Marshal(lr)
		cmd.Println(string(b))
	}
}

func printCpuInfo(cmd *cobra.Command, cap *pb.Capabilities) {
	for i, cpu := range cap.Cpu {
		cmd.Printf("    CPU%d: %d x %s\r\n", i, cpu.GetCores(), cpu.GetName())
	}
}

func printGpuInfo(cmd *cobra.Command, cap *pb.Capabilities) {
	if len(cap.Gpu) > 0 {
		for i, gpu := range cap.Gpu {
			cmd.Printf("    GPU%d: %s %s\r\n", i, gpu.VendorName, gpu.Name)
		}
	} else {
		cmd.Println("    GPU: None")
	}
}

func printMemInfo(cmd *cobra.Command, cap *pb.Capabilities) {
	cmd.Println("    RAM:")
	cmd.Printf("      Total: %s\r\n", ds.ByteSize(cap.Mem.GetTotal()).HR())
	cmd.Printf("      Used:  %s\r\n", ds.ByteSize(cap.Mem.GetUsed()).HR())
}

func printMinerStatus(cmd *cobra.Command, minerID string, metrics *pb.InfoReply) {
	if isSimpleFormat() {
		if metrics.Name == "" {
			cmd.Printf("Miner: \"%s\":\r\n", minerID)
		} else {
			cmd.Printf("Miner: \"%s\" (%s):\r\n", minerID, metrics.Name)
		}

		if metrics.Capabilities != nil {
			cmd.Println("  Hardware:")
			printCpuInfo(cmd, metrics.Capabilities)
			printGpuInfo(cmd, metrics.Capabilities)
			printMemInfo(cmd, metrics.Capabilities)
		}

		if len(metrics.GetUsage()) == 0 {
			cmd.Println("  No active tasks")
		} else {
			cmd.Println("  Tasks:")
			i := 1
			for task := range metrics.Usage {
				cmd.Printf("    %d) %s\r\n", i, task)
				i++
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

var minerPropertiesCmd = &cobra.Command{
	Use:     "properties",
	Short:   "Miner properties",
	PreRunE: checkHubAddressIsSet,
}

var minerGetPropertiesCmd = &cobra.Command{
	Use:     "get MINER_ID...",
	Short:   "Get miner(s) properties",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, IDs []string) error {
		ctx := context.Background()
		grpc, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return nil
		}
		if len(IDs) == 0 {
			miners, err := grpc.MinerList(ctx)
			if err != nil {
				return err
			}
			for id := range miners.Info {
				IDs = append(IDs, id)
			}
		}

		properties := map[string]map[string]string{}
		for _, id := range IDs {
			property, err := grpc.MinerGetProperties(ctx, id)
			if err != nil {
				return err
			}
			properties[id] = property.Properties
		}

		dump, err := json.Marshal(properties)
		if err != nil {
			return err
		}
		cmd.Println(string(dump))
		return nil
	},
}

var minerSetPropertiesCmd = &cobra.Command{
	Use:     "set MINER_ID PATH",
	Short:   "Set miner properties",
	PreRunE: checkHubAddressIsSet,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ID := args[0]
		path := args[1]

		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		cfg := map[string]string{}
		err = yaml.Unmarshal(buf, &cfg)
		if err != nil {
			return err
		}

		grpc, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return nil
		}
		_, err = grpc.MinerSetProperties(context.Background(), ID, cfg)
		if err != nil {
			return err
		}

		cmd.Println("OK")
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

	printMinerStatus(cmd, minerID, metrics)
}

var minerSlotCmd = &cobra.Command{
	Use:     "slot",
	Short:   "Show hub's virtual slots",
	PreRunE: checkHubAddressIsSet,
}

var minerShowSlotsCmd = &cobra.Command{
	Use:     "show",
	Short:   "Show hub's virtual slots",
	PreRunE: checkHubAddressIsSet,
	RunE: func(cmd *cobra.Command, IDs []string) error {
		grpc, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return err
		}

		ctx := context.Background()
		if len(IDs) == 0 {
			miners, err := grpc.MinerList(ctx)
			if err != nil {
				return err
			}
			for id := range miners.Info {
				IDs = append(IDs, id)
			}
		}

		result := map[string][]*pb.Slot{}
		for _, id := range IDs {
			slots, err := grpc.MinerShowSlots(context.Background(), id)
			if err != nil {
				return err
			}

			result[id] = slots.Slot
		}

		b, _ := json.Marshal(result)
		cmd.Println(string(b))
		return nil
	},
}

var minerAddSlotCmd = &cobra.Command{
	Use:     "add ID PATH",
	Short:   "Add a virtual slot",
	PreRunE: checkHubAddressIsSet,
	Args:    cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ID := args[0]
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

		grpc, err := NewGrpcInteractor(hubAddress, timeout)
		if err != nil {
			return err
		}
		slot, err := cfg.IntoSlot()
		if err != nil {
			return err
		}

		_, err = grpc.MinerAddSlot(context.Background(), ID, slot)
		if err != nil {
			return err
		}

		cmd.Println("OK")
		return nil
	},
}
