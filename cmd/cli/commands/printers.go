package commands

import (
	"fmt"
	"strings"
	"time"

	pb "github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/datasize"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func printTaskStatus(cmd *cobra.Command, id string, taskStatus *pb.TaskStatusReply) {
	if isSimpleFormat() {
		cmd.Printf("ID: %s\r\n", id)
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

		if len(taskStatus.GetPortMap()) > 0 {
			cmd.Printf("  Ports:\r\n")
			for containerPort, portBindings := range taskStatus.GetPortMap() {
				for _, portBinding := range portBindings.GetEndpoints() {
					cmd.Printf("    %s: %s:%d\r\n", containerPort, portBinding.GetAddr(), portBinding.GetPort())
				}
			}
		}
	} else {
		v := map[string]interface{}{
			"id":     id,
			"status": taskStatus.Status.String(),
			"image":  taskStatus.GetImageName(),
			"ports":  taskStatus.GetPortMap(),
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
		cmd.Print(string(out))
	}
}

func printNodeTaskStatus(cmd *cobra.Command, tasksMap map[string]*pb.TaskStatusReply) {
	if isSimpleFormat() {
		for id, task := range tasksMap {
			printTaskStatus(cmd, id, task)
		}
	} else {
		showJSON(cmd, tasksMap)
	}
}

func printWorkerStatus(cmd *cobra.Command, stat *pb.StatusReply) {
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
				cmd.Printf("  index=%d: %s\r\n", i, gpu.Device.GetDeviceName())
				printBenchmarkGroup(cmd, gpu.Benchmarks)
			}
		}

		netIn := datasize.NewBitRate(dev.GetNetwork().GetIn()).HumanReadable()
		netOut := datasize.NewBitRate(dev.GetNetwork().GetOut()).HumanReadable()
		cmd.Println("Network:")
		cmd.Printf("  Incoming: %v\r\n", dev.GetNetwork().GetNetFlags().GetIncoming())
		cmd.Printf("  Overlay:  %v\r\n", dev.GetNetwork().GetNetFlags().GetOverlay())
		cmd.Printf("  In:       %s\r\n", netIn)
		cmd.Printf("  Out:      %s\r\n", netOut)

		// merge network benchmarks to prevent printing two benchmarks groups with one item in each
		networkBenchmarks := dev.GetNetwork().GetBenchmarksIn()
		for k, v := range dev.GetNetwork().GetBenchmarksOut() {
			networkBenchmarks[k] = v
		}
		printBenchmarkGroup(cmd, networkBenchmarks)

		storageAvailable := datasize.NewByteSize(dev.GetStorage().GetDevice().GetBytesAvailable()).HumanReadable()
		cmd.Println("Storage:")
		cmd.Printf("  Volume: %s\r\n", storageAvailable)
		printBenchmarkGroup(cmd, dev.GetStorage().GetBenchmarks())
	} else {
		showJSON(cmd, dev)
	}
}

func printOrdersList(cmd *cobra.Command, orders []*pb.Order) {
	if isSimpleFormat() {
		if len(orders) == 0 {
			cmd.Println("No orders found")
			return
		}

		cmd.Println("        ID | type |        total price       |   duration ")
		for _, order := range orders {
			cmd.Printf("%10s | %4s | %20s USD | %10s\r\n",
				order.GetId().Unwrap().String(),
				order.OrderType.String(),
				order.TotalPrice(),
				(time.Second * time.Duration(order.Duration)).String())
		}
	} else {
		showJSON(cmd, map[string]interface{}{"orders": orders})
	}
}

func printOrderDetails(cmd *cobra.Command, order *pb.Order) {
	if isSimpleFormat() {
		cmd.Printf("ID:              %s\r\n", order.Id)
		if !order.GetDealID().IsZero() {
			cmd.Printf("Deal ID:         %s\r\n", order.GetDealID().Unwrap().String())
		}
		cmd.Printf("Type:            %s\r\n", order.OrderType.String())
		cmd.Printf("Status:          %s\r\n", order.OrderStatus.String())
		cmd.Printf("Duration:        %s\r\n", (time.Duration(order.GetDuration()) * time.Second).String())
		cmd.Printf("Total price:     %s USD (%s USD/sec)\r\n", order.TotalPrice(), order.Price.ToPriceString())

		cmd.Printf("Author ID:       %s\r\n", order.GetAuthorID().Unwrap().Hex())
		if order.GetCounterpartyID() != nil {
			cmd.Printf("Counterparty ID: %s\r\n", order.GetCounterpartyID().Unwrap().Hex())
		}

		cmd.Println("Net:")
		cmd.Printf("  Overlay:  %v\r\n", order.GetNetflags().GetOverlay())
		cmd.Printf("  Outbound: %v\r\n", order.GetNetflags().GetOutbound())
		cmd.Printf("  Incoming: %v\r\n", order.GetNetflags().GetIncoming())

		b := order.GetBenchmarks()
		cmd.Println("Benchmarks:")
		cmd.Printf("  CPU Sysbench Multi:  %d\r\n", b.CPUSysbenchMulti())
		cmd.Printf("  CPU Sysbench Single: %d\r\n", b.CPUSysbenchOne())
		cmd.Printf("  CPU Cores:           %d\r\n", b.CPUCores())
		cmd.Printf("  RAM Size             %d\r\n", b.RAMSize())
		cmd.Printf("  Storage Size         %d\r\n", b.StorageSize())
		cmd.Printf("  Net Download         %d\r\n", b.NetTrafficIn())
		cmd.Printf("  Net Upload           %d\r\n", b.NetTrafficOut())
		cmd.Printf("  GPU Count            %d\r\n", b.GPUCount())
		cmd.Printf("  GPU Mem              %d\r\n", b.GPUMem())
		cmd.Printf("  GPU Eth hashrate     %d\r\n", b.GPUEthHashrate())
		cmd.Printf("  GPU Cash hashrate    %d\r\n", b.GPUCashHashrate())
		cmd.Printf("  GPU Redshift         %d\r\n", b.GPURedshift())
	} else {
		showJSON(cmd, order)
	}
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

func printDealsList(cmd *cobra.Command, deals []*pb.Deal) {
	if isSimpleFormat() {
		if len(deals) == 0 {
			cmd.Println("No deals found")
			return
		}

		for _, deal := range deals {
			printDealInfo(cmd, &pb.DealInfoReply{Deal: deal}, nil)
			cmd.Println()
		}
	} else {
		showJSON(cmd, map[string]interface{}{"deals": deals})
	}

}

func printDealInfo(cmd *cobra.Command, info *pb.DealInfoReply, changes *pb.DealChangeRequestsReply) {
	if isSimpleFormat() {
		deal := info.GetDeal()
		isClosed := deal.GetStatus() == pb.DealStatus_DEAL_CLOSED
		start := deal.StartTime.Unix()
		end := deal.EndTime.Unix()
		dealDuration := end.Sub(start)
		lastBill := deal.GetLastBillTS().Unix()

		cmd.Printf("ID:           %s (%s deal)\r\n", deal.GetId(), deal.GetTypeName())
		cmd.Printf("ASK ID:       %s\r\n", deal.GetAskID().Unwrap().String())
		cmd.Printf("BID ID:       %s\r\n", deal.GetBidID().Unwrap().String())
		cmd.Printf("Status:       %s\r\n", deal.GetStatus())
		if deal.IsSpot() {
			// for active spot deal we can show only pricePerSecond
			cmd.Printf("Price:        %s USD/sec\r\n", deal.GetPrice().ToPriceString())
			if isClosed {
				// for closed deal we also can *calculate* total duration
				cmd.Printf("Duration:     %s\r\n", dealDuration.String())
			}
		} else {
			// for non-spot deal we can show duration, total price and pricePerSecond
			cmd.Printf("Price:        %s USD (%s USD/sec)\r\n", deal.TotalPrice(), deal.GetPrice().ToPriceString())
			cmd.Printf("Duration:     %s\r\n", dealDuration.String())
		}

		cmd.Printf("Total payout: %s SNM\r\n", deal.GetTotalPayout().ToPriceString())
		cmd.Printf("Consumer ID:  %s\r\n", deal.GetConsumerID().Unwrap().Hex())
		cmd.Printf("Supplier ID:  %s\r\n", deal.GetSupplierID().Unwrap().Hex())

		cmd.Printf("Start at:     %s\r\n", start.Format(time.RFC3339))
		if isClosed {
			// correct endTime exists for any closed deal.
			cmd.Printf("End at:       %s\r\n", start.Add(dealDuration).Format(time.RFC3339))
		}

		if lastBill.Unix() > 0 {
			cmd.Printf("Last bill:    %s\r\n", lastBill.Format(time.RFC3339))
		}

		if changes != nil && len(changes.GetRequests()) > 0 {
			cmd.Println("Change requests:")
			for _, req := range changes.GetRequests() {
				cmd.Printf("  id: %s,  new duration: %s, new price: %s USD/s\n",
					req.GetId().Unwrap().String(),
					time.Second*time.Duration(req.GetDuration()),
					req.GetPrice().ToPriceString())
			}
		}

		if info.Resources != nil {
			cmd.Println("Resources:")

			cpuCores := float64(info.GetResources().GetCPU().GetCorePercents()) / 100.0
			ram := info.GetResources().GetRAM().GetSize().Unwrap().HumanReadable()

			cmd.Printf("  CPU:     %.2f cores\n", cpuCores)
			cmd.Printf("  RAM:     %s\n", ram)
			if len(info.GetResources().GetGPU().GetHashes()) > 0 {
				cmd.Printf("  GPUs:    %v\n", strings.Join(info.GetResources().GetGPU().GetHashes(), ", "))
			}
			cmd.Printf("  Storage: %v\n", info.GetResources().GetStorage().GetSize().Unwrap().HumanReadable())
			cmd.Println("  Network:")
			cmd.Printf("    Overlay:  %v\n", info.GetResources().GetNetwork().GetNetFlags().GetOverlay())
			cmd.Printf("    Outbound: %v\n", info.GetResources().GetNetwork().GetNetFlags().GetOutbound())
			cmd.Printf("    Incoming: %v\n", info.GetResources().GetNetwork().GetNetFlags().GetIncoming())

		}

		if info.Running != nil && len(info.Running) > 0 {
			cmd.Println("Running tasks:")
			for id, task := range info.Running {
				cmd.Printf("  %s: %s %s\n", id, task.Status.String(), task.ImageName)
			}
		}

		if info.Completed != nil && len(info.Completed) > 0 {
			cmd.Println("Finished tasks:")
			for id, task := range info.Completed {
				cmd.Printf("  %s: %s %s\n", id, task.Status.String(), task.ImageName)
			}
		}
	} else {
		showJSON(cmd, info)
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
		cmd.Printf("Task ID:    %s\r\n", start.Id)

		for containerPort, portBindings := range start.GetPortMap() {
			for _, portBinding := range portBindings.GetEndpoints() {
				cmd.Printf("  Endpoint: %s: %s:%d\r\n", containerPort, portBinding.GetAddr(), portBinding.GetPort())
			}
		}

		for _, end := range start.GetNetworkIDs() {
			cmd.Printf("  Network:  %s\r\n", end)
		}
	} else {
		showJSON(cmd, start)
	}
}

func printBalanceInfo(cmd *cobra.Command, reply *pb.BalanceReply) {
	side := reply.GetSideBalance().ToPriceString()
	live := reply.GetLiveBalance().ToPriceString()

	if isSimpleFormat() {
		cmd.Printf("On Ethereum: %s SNM\n", live)
		cmd.Printf("On SONM:     %s SNM\n", side)
	} else {
		showJSON(cmd, map[string]map[string]string{"balance": {
			"ethereum": live,
			"sonm":     side,
		}})
	}
}

func printBlacklist(cmd *cobra.Command, list *pb.BlacklistReply) {
	if isSimpleFormat() {
		if len(list.GetAddresses()) == 0 {
			cmd.Println("Blacklist is empty")
			return
		}

		cmd.Println("Blacklisted addresses:")
		for _, addr := range list.GetAddresses() {
			cmd.Printf("  %s\n", addr)
		}
	} else {
		showJSON(cmd, list)
	}
}

func printWorkersList(cmd *cobra.Command, list *pb.WorkerListReply) {
	if isSimpleFormat() {
		if len(list.GetWorkers()) == 0 {
			cmd.Println("No workers for current master")
			return
		}

		var confirmed, notConfirmed []*pb.DWHWorker
		for _, w := range list.GetWorkers() {
			if w.GetConfirmed() {
				confirmed = append(confirmed, w)
			} else {
				notConfirmed = append(notConfirmed, w)
			}
		}

		if len(confirmed) > 0 {
			cmd.Println("Confirmed workers:")
			for _, w := range confirmed {
				cmd.Printf("  %s\n", w.SlaveID.Unwrap().Hex())
			}
		}

		if len(notConfirmed) > 0 {
			cmd.Println("Waiting for confirmation:")
			for _, w := range notConfirmed {
				cmd.Printf("  %s\n", w.SlaveID.Unwrap().Hex())
			}
		}
	} else {
		showJSON(cmd, list)
	}
}
