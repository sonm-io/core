package commands

import (
	"os"

	ds "github.com/c2h5oh/datasize"
	"github.com/sonm-io/core/cmd/cli/task_config"
	"github.com/sonm-io/core/insonmnia/structs"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	nodeOrderRootCmd.AddCommand(
		nodeOrderListCmd,
		nodeOrderCreateCmd,
		nodeOrderRemoveCmd,
	)
}

var nodeOrderRootCmd = &cobra.Command{
	Use:   "ask-plan",
	Short: "Operations with ask order plan",
}

func printAskList(cmd *cobra.Command, slots *pb.SlotsReply) {
	if isSimpleFormat() {
		slots := slots.GetSlots()
		if len(slots) == 0 {
			cmd.Printf("No Ask Order configured\r\n")
			return
		}

		for id, slot := range slots {
			cmd.Printf(" ID:  %s", id)
			cmd.Printf(" CPU: %d Cores\r\n", slot.Resources.CpuCores)
			cmd.Printf(" GPU: %d Devices\r\n", slot.Resources.GpuCount)
			cmd.Printf(" RAM: %s\r\n", ds.ByteSize(slot.Resources.RamBytes).HR())
			cmd.Printf(" Net: %s\r\n", slot.Resources.NetworkType.String())
			cmd.Printf("     %s IN\r\n", ds.ByteSize(slot.Resources.NetTrafficIn).HR())
			cmd.Printf("     %s OUT\r\n", ds.ByteSize(slot.Resources.NetTrafficOut).HR())

			if slot.Geo != nil && slot.Geo.City != "" && slot.Geo.Country != "" {
				cmd.Printf(" Geo: %s, %s\r\n", slot.Geo.City, slot.Geo.Country)
			}
			cmd.Println("")
		}
	} else {
		showJSON(cmd, slots)
	}
}

var nodeOrderListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show current ask plans",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		asks, err := hub.GetAskPlans()
		if err != nil {
			showError(cmd, "Cannot get Ask Orders from Worker", err)
			os.Exit(1)
		}

		printAskList(cmd, asks)
	},
}

var nodeOrderCreateCmd = &cobra.Command{
	Use:    "create <price> <plan.yaml>",
	Short:  "Create new plan",
	Args:   cobra.MinimumNArgs(2),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		price := args[0]
		planPath := args[1]

		_, err = util.ParseBigInt(price)
		if err != nil {
			showError(cmd, "Cannot parse price", err)
			os.Exit(1)
		}

		slot, err := loadSlotFile(planPath)
		if err != nil {
			showError(cmd, "Cannot load AskOrder definition", err)
			os.Exit(1)
		}

		id, err := hub.CreateAskPlan(slot, price)
		if err != nil {
			showError(cmd, "Cannot create new AskOrder", err)
			os.Exit(1)
		}

		showJSON(cmd, id)
	},
}

var nodeOrderRemoveCmd = &cobra.Command{
	Use:    "remove <order_id>",
	Short:  "Remove plan by",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		hub, err := NewHubInteractor(nodeAddressFlag, timeoutFlag)
		if err != nil {
			showError(cmd, "Cannot connect to Node", err)
			os.Exit(1)
		}

		ID := args[0]
		_, err = hub.RemoveAskPlan(ID)
		if err != nil {
			showError(cmd, "Cannot remove AskOrder", err)
			os.Exit(1)
		}

		showOk(cmd)
	},
}

func loadOrderFile(path string) (*structs.Order, error) {
	cfg := task_config.OrderConfig{}
	err := util.LoadYamlFile(path, &cfg)
	if err != nil {
		return nil, err
	}

	order, err := cfg.IntoOrder()
	if err != nil {
		return nil, err
	}

	return order, nil
}

func loadSlotFile(path string) (*structs.Slot, error) {
	cfg := task_config.SlotConfig{}
	err := util.LoadYamlFile(path, &cfg)
	if err != nil {
		return nil, err
	}

	slot, err := cfg.IntoSlot()
	if err != nil {
		return nil, err
	}

	return slot, nil
}

func loadPropsFile(path string) (map[string]float64, error) {
	props := map[string]float64{}
	err := util.LoadYamlFile(path, &props)
	if err != nil {
		return nil, err
	}

	return props, nil
}
