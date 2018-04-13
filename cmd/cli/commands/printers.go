package commands

import (
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/ethereum/go-ethereum/core/types"
	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/datasize"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func printTaskStatus(cmd *cobra.Command, id string, taskStatus *pb.TaskStatusReply) {
	if isSimpleFormat() {
		portsParsedOK := false
		ports := nat.PortMap{}
		if len(taskStatus.GetPorts()) > 0 {
			err := json.Unmarshal([]byte(taskStatus.GetPorts()), &ports)
			portsParsedOK = err == nil
		}

		cmd.Printf("Task %s (on %s):\r\n", id, taskStatus.MinerID)
		cmd.Printf("  Image:  %s\r\n", taskStatus.GetImageName())
		cmd.Printf("  Status: %s\r\n", taskStatus.GetStatus().String())
		cmd.Printf("  Uptime: %s\r\n", time.Duration(taskStatus.GetUptime()).String())

		if taskStatus.GetUsage() != nil {
			cmd.Println("  Resources:")
			cmd.Printf("    CPU: %d\r\n", taskStatus.Usage.GetCpu().GetTotal())
			cmd.Printf("    MEM: %s\r\n", datasize.NewByteSize(taskStatus.Usage.GetMemory().GetMaxUsage()).HumanReadable())
			if taskStatus.GetUsage().GetNetwork() != nil {
				cmd.Printf("    NET:\r\n")
				for i, net := range taskStatus.GetUsage().GetNetwork() {
					cmd.Printf("      %s:\r\n", i)
					cmd.Printf("        Tx/Rx bytes: %d/%d\r\n", net.TxBytes, net.RxBytes)
					cmd.Printf("        Tx/Rx packets: %d/%d\r\n", net.TxPackets, net.RxPackets)
					cmd.Printf("        Tx/Rx errors: %d/%d\r\n", net.TxErrors, net.RxErrors)
					cmd.Printf("        Tx/Rx dropped: %d/%d\r\n", net.TxDropped, net.RxDropped)
				}
			}
		}

		if portsParsedOK && len(ports) > 0 {
			cmd.Printf("  Ports:\r\n")
			for containerPort, portBindings := range ports {
				for _, portBinding := range portBindings {
					cmd.Printf("    %s: %s:%s\r\n", containerPort, portBinding.HostIP, portBinding.HostPort)
				}
			}
		}
	} else {
		v := map[string]interface{}{
			"id":     id,
			"worker": taskStatus.MinerID,
			"status": taskStatus.Status.String(),
			"image":  taskStatus.GetImageName(),
			"ports":  taskStatus.GetPorts(),
			"uptime": fmt.Sprintf("%d", time.Duration(taskStatus.GetUptime())),
		}
		if taskStatus.GetUsage() != nil {
			v["cpu"] = fmt.Sprintf("%d", taskStatus.GetUsage().GetCpu().GetTotal())
			v["mem"] = fmt.Sprintf("%d", taskStatus.GetUsage().GetMemory().GetMaxUsage())
			v["net"] = taskStatus.GetUsage().GetNetwork()
		}

		showJSON(cmd, v)
	}
}

func printNetworkSpec(cmd *cobra.Command, spec *pb.NetworkSpec) {
	out, err := yaml.Marshal(spec)
	if err != nil {
		cmd.Printf("%s", err)
	} else {
		cmd.Print(out)
	}
}

func printNodeTaskStatus(cmd *cobra.Command, tasksMap map[string]*pb.TaskStatusReply) {
	if isSimpleFormat() {
		for id, task := range tasksMap {
			i := 1
			up := time.Duration(task.GetUptime())

			cmd.Printf("  %d) %s \r\n     %s  %s (up: %v)\r\n",
				i, id, task.GetImageName(), task.GetStatus().String(), up.String())
			i++
		}
	} else {
		showJSON(cmd, tasksMap)
	}
}

func printHubStatus(cmd *cobra.Command, stat *pb.HubStatusReply) {
	if isSimpleFormat() {
		cmd.Printf("Uptime:             %s\r\n", (time.Second * time.Duration(stat.GetUptime())).String())
		cmd.Printf("Version:            %s %s\r\n", stat.GetVersion(), stat.GetPlatform())
		cmd.Printf("Eth address:        %s\r\n", stat.GetEthAddr())
		cmd.Printf("Task count:         %d\r\n", stat.GetTaskCount())
		cmd.Printf("DWH status:         %s\r\n", stat.GetDWHStatus())
		cmd.Printf("Rendezvous status:  %s\r\n", stat.GetRendezvousStatus())
	} else {
		showJSON(cmd, stat)
	}
}

func printBenchmarkGroup(cmd *cobra.Command, benchmarks map[uint64]*pb.Benchmark) {
	cmd.Println("  Benchmarks:")
	for _, bn := range benchmarks {
		cmd.Printf("    %s: %v\r\n", bn.Code, bn.Result)
	}
	cmd.Println()
}

func printDeviceList(cmd *cobra.Command, dev *pb.DevicesReply) {
	if isSimpleFormat() {
		cpu := dev.GetCPU().GetDevice()
		cmd.Printf("CPU: %d cores at %d sockets\r\n", cpu.GetCores(), cpu.GetSockets())
		printBenchmarkGroup(cmd, dev.GetCPU().GetBenchmarks())

		ram := datasize.NewByteSize(dev.GetRAM().GetDevice().GetAvailable()).HumanReadable()
		cmd.Printf("RAM: %s\r\n", ram)
		printBenchmarkGroup(cmd, dev.GetRAM().GetBenchmarks())

		GPUs := dev.GetGPUs()
		if len(GPUs) > 0 {
			cmd.Printf("GPUs:\r\n")
			for i, gpu := range GPUs {
				cmd.Printf("  id=%d: %s\r\n", i, gpu.Device.GetDeviceName())
				printBenchmarkGroup(cmd, gpu.Benchmarks)
			}
		}

		netIn := datasize.NewBitRate(dev.GetNetwork().GetDevice().GetBandwidthIn()).HumanReadable()
		netOut := datasize.NewBitRate(dev.GetNetwork().GetDevice().GetBandwidthOut()).HumanReadable()
		cmd.Println("Network:")
		cmd.Printf("  In: %s | Out: %s \r\n", netIn, netOut)
		printBenchmarkGroup(cmd, dev.GetNetwork().GetBenchmarks())

		storageAvailable := datasize.NewByteSize(dev.GetStorage().GetDevice().GetBytesAvailable()).HumanReadable()
		cmd.Println("Storage:")
		cmd.Printf("  Volume: %s\r\n", storageAvailable)
		printBenchmarkGroup(cmd, dev.GetStorage().GetBenchmarks())
	} else {
		showJSON(cmd, dev)
	}
}

func printTransactionInfo(cmd *cobra.Command, tx *types.Transaction) {
	if isSimpleFormat() {
		cmd.Printf("Hash:      %s\r\n", tx.Hash().String())
		cmd.Printf("Value:     %s\r\n", tx.Value().String())
		cmd.Printf("To:        %s\r\n", tx.To().String())
		cmd.Printf("Cost:      %s\r\n", tx.Cost().String())
		cmd.Printf("Gas:       %d\r\n", tx.Gas())
		cmd.Printf("Gas price: %s\r\n", tx.GasPrice().String())
	} else {
		showJSON(cmd, convertTransactionInfo(tx))
	}
}

func convertTransactionInfo(tx *types.Transaction) map[string]interface{} {
	return map[string]interface{}{
		"hash":      tx.Hash().String(),
		"value":     tx.Value().String(),
		"to":        tx.To().String(),
		"cost":      tx.Cost().String(),
		"gas":       tx.Gas(),
		"gas_price": tx.GasPrice().String(),
	}
}

func printSearchResults(cmd *cobra.Command, orders []*pb.MarketOrder) {
	if isSimpleFormat() {
		if len(orders) == 0 {
			cmd.Printf("No matching orders found")
			return
		}

		for i, order := range orders {
			cmd.Printf("%d) %s %s | price = %s\r\n", i+1,
				order.OrderType.String(), order.GetId(), order.GetPrice().ToPriceString())
		}
	} else {
		showJSON(cmd, map[string]interface{}{"orders": orders})
	}
}

func printOrderDetails(cmd *cobra.Command, order *pb.MarketOrder) {
	if isSimpleFormat() {
		cmd.Printf("ID:             %s\r\n", order.Id)
		cmd.Printf("Type:           %s\r\n", order.OrderType.String())
		cmd.Printf("Price:          %s\r\n", order.GetPrice().ToPriceString())

		cmd.Printf("AuthorID:     %s\r\n", order.GetAuthor())
		cmd.Printf("CounterpartyID:        %s\r\n", order.GetCounterparty())

		// todo: find a way to print resources as they presented into MarketOrder struct.
	} else {
		showJSON(cmd, order)
	}
}

func printOrderResources(cmd *cobra.Command, rs *pb.Resources) {
	cmd.Printf("Resources:\r\n")
	cmd.Printf("  CPU:     %d\r\n", rs.CpuCores)
	cmd.Printf("  GPU:     %s\r\n", rs.GpuCount.String())
	cmd.Printf("  RAM:     %s\r\n", datasize.NewByteSize(rs.RamBytes).HumanReadable())
	cmd.Printf("  Storage: %s\r\n", datasize.NewByteSize(rs.Storage).HumanReadable())
	cmd.Printf("  Network: %s\r\n", rs.NetworkType.String())
	cmd.Printf("    In:   %s\r\n", datasize.NewByteSize(rs.NetTrafficIn).HumanReadable())
	cmd.Printf("    Out:  %s\r\n", datasize.NewByteSize(rs.NetTrafficOut).HumanReadable())
}

func printAskList(cmd *cobra.Command, slots *pb.AskPlansReply) {
	if isSimpleFormat() {
		plans := slots.GetAskPlans()
		if len(plans) == 0 {
			cmd.Printf("No Ask Order configured\r\n")
			return
		}
		out, err := yaml.Marshal(plans)
		if err != nil {
			showError(cmd, "could not marshall ask plans", err)
		} else {
			cmd.Println(string(out))
		}
	} else {
		showJSON(cmd, slots)
	}
}

func printVersion(cmd *cobra.Command, v string) {
	if isSimpleFormat() {
		cmd.Printf("sonmcli %s (%s)\r\n", v, util.GetPlatformName())
	} else {
		showJSON(cmd, map[string]string{
			"version":  v,
			"platform": util.GetPlatformName(),
		})
	}
}

func printDealsList(cmd *cobra.Command, deals []*pb.MarketDeal) {
	if isSimpleFormat() {
		if len(deals) == 0 {
			cmd.Println("No deals found")
			return
		}

		for _, deal := range deals {
			printDealInfo(cmd, deal)
			cmd.Println()
		}
	} else {
		showJSON(cmd, map[string]interface{}{"deals": deals})
	}

}

func printDealInfo(cmd *cobra.Command, deal *pb.MarketDeal) {
	if isSimpleFormat() {
		start := deal.StartTime.Unix()
		end := deal.EndTime.Unix()

		ppsBig := pb.NewBigInt(nil)
		dealDuration := end.Sub(start)
		if dealDuration > 0 {
			durationBig := big.NewInt(int64(dealDuration.Seconds()))
			pps := big.NewInt(0).Div(deal.GetPrice().Unwrap(), durationBig)
			ppsBig = pb.NewBigInt(pps)
		}

		cmd.Printf("ID:       %s\r\n", deal.GetId())
		cmd.Printf("Status:   %s\r\n", deal.GetStatus())
		cmd.Printf("Duraton:  %s\r\n", dealDuration.String())
		cmd.Printf("Price:    %s (%s SNM/sec)\r\n", deal.GetPrice().ToPriceString(), ppsBig.ToPriceString())
		cmd.Printf("Consumer: %s\r\n", deal.GetConsumerID())
		cmd.Printf("Supplier: %s\r\n", deal.GetSupplierID())
		cmd.Printf("Start at: %s\r\n", start.Format(time.RFC3339))
		cmd.Printf("End at:   %s\r\n", end.Format(time.RFC3339))
	} else {
		showJSON(cmd, deal)
	}
}

func printDealTasksShort(cmd *cobra.Command, tasks map[string]*pb.TaskStatusReply) {
	for id, info := range tasks {
		cmd.Printf("%s ID: %s | image \"%s\"\r\n", info.GetStatus(), id, info.GetImageName())
	}
}

func printID(cmd *cobra.Command, id string) {
	if isSimpleFormat() {
		cmd.Printf("ID = %s\r\n", id)
	} else {
		showJSON(cmd, map[string]string{"id": id})
	}
}

func printTaskStart(cmd *cobra.Command, start *pb.StartTaskReply) {
	if isSimpleFormat() {
		cmd.Printf("Task ID:      %s\r\n", start.Id)
		cmd.Printf("Hub Address:  %s\r\n", start.HubAddr)
		for _, end := range start.GetEndpoint() {
			cmd.Printf("  Endpoint:    %s\r\n", end)
		}
		for _, end := range start.GetNetworkIDs() {
			cmd.Printf("  Network:    %s\r\n", end)
		}
	} else {
		showJSON(cmd, start)
	}
}

func printDealDetails(cmd *cobra.Command, d *pb.DealStatusReply) {
	if !isSimpleFormat() {
		showJSON(cmd, d)
		return
	}

	if d.GetDeal() != nil {
		cmd.Printf("Deal info:\r\n")
		printDealInfo(cmd, d.GetDeal())
	}

	if d.GetInfo().GetOrder() != nil {
		cmd.Printf("\r\n")
		printOrderResources(cmd, d.GetInfo().GetOrder().GetSlot().GetResources())
	}

	if d.GetInfo().GetRunning() != nil && len(d.GetInfo().GetRunning().GetStatuses()) > 0 {
		cmd.Printf("\r\nRunning tasks:\r\n")
		printDealTasksShort(cmd, d.GetInfo().GetRunning().GetStatuses())
	}

	if d.GetInfo().GetCompleted() != nil && len(d.GetInfo().GetCompleted().GetStatuses()) > 0 {
		cmd.Printf("\r\nCompleted tasks:\r\n")
		printDealTasksShort(cmd, d.GetInfo().GetCompleted().GetStatuses())
	}
}
