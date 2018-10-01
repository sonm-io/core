package commands

import (
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/olekukonko/tablewriter"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/sonm-io/core/util/datasize"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

const (
	printEverything = iota
	suppressWarnings
)

type printerFlags int

type Printer interface {
	Printf(string, ...interface{})
	Println(i ...interface{})
}

type IndentPrinter struct {
	Subprinter Printer
	IdentCount uint64
	Ident      rune
}

func (m *IndentPrinter) Printf(format string, args ...interface{}) {
	ident := strings.Repeat(string(m.Ident), int(m.IdentCount))
	args = append([]interface{}{ident}, args...)
	m.Subprinter.Printf("%s"+format, args...)
}

func (m *IndentPrinter) Println(args ...interface{}) {
	ident := strings.Repeat(string(m.Ident), int(m.IdentCount))
	m.Subprinter.Printf("%s", ident)
	m.Subprinter.Println(args...)
}

func (p printerFlags) WarningSuppressed() bool {
	return int(p)&suppressWarnings == 1
}

func printTaskStatus(cmd *cobra.Command, id string, taskStatus *sonm.TaskStatusReply) {
	if isSimpleFormat() {
		cmd.Printf("ID: %s\r\n", id)
		cmd.Printf("  Image:  %s\r\n", taskStatus.GetImageName())
		cmd.Printf("  Status: %s\r\n", taskStatus.GetStatus().String())
		if tag := taskStatus.GetTag(); len(tag.GetData()) != 0 {
			tagData, err := yaml.Marshal(tag)
			if err != nil {
				cmd.Printf("  Tag:    failed to marshal tag: %v\r\n", err)
			} else {
				cmd.Printf("  Tag:    %s\r\n", strings.TrimRight(string(tagData), "\n"))
			}
		}
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

func printNetworkSpec(cmd *cobra.Command, spec *sonm.NetworkSpec) {
	out, err := yaml.Marshal(spec)
	if err != nil {
		cmd.Printf("%s", err)
	} else {
		cmd.Print(string(out))
	}
}

func printNodeTaskStatus(cmd *cobra.Command, tasksMap map[string]*sonm.TaskStatusReply) {
	if isSimpleFormat() {
		if len(tasksMap) == 0 {
			cmd.Printf("No active tasks\r\n")
			return
		}
		for id, task := range tasksMap {
			printTaskStatus(cmd, id, task)
		}
	} else {
		showJSON(cmd, tasksMap)
	}
}

func printWorkerStatus(cmd *cobra.Command, stat *sonm.StatusReply) {
	if isSimpleFormat() {
		cmd.Printf("Uptime:            %s\r\n", (time.Second * time.Duration(stat.GetUptime())).String())
		cmd.Printf("Version:           %s %s\r\n", stat.GetVersion(), stat.GetPlatform())
		cmd.Printf("Eth address:       %s\r\n", stat.GetEthAddr())
		cmd.Printf("Master address:    %s (confirmed: %v)\r\n",
			stat.GetMaster().Unwrap().Hex(), stat.GetIsMasterConfirmed())
		if !stat.GetAdmin().IsZero() {
			cmd.Printf("Admin address:     %s\r\n", stat.GetAdmin().Unwrap().Hex())
		}
		cmd.Printf("Task count:        %d\r\n", stat.GetTaskCount())
		cmd.Printf("DWH status:        %s\r\n", stat.GetDWHStatus())
		cmd.Printf("Rendezvous status: %s\r\n", stat.GetRendezvousStatus())
		if !stat.GetIsBenchmarkFinished() {
			cmd.Printf("[WARN] Worker is benchmarking now\r\n")
		}
	} else {
		showJSON(cmd, stat)
	}
}

func printBenchmarkGroup(cmd *cobra.Command, benchmarks map[uint64]*sonm.Benchmark) {
	cmd.Println("  Benchmarks:")
	for _, bn := range benchmarks {
		cmd.Printf("    %s: %v\r\n", bn.Code, bn.Result)
	}
	cmd.Println()
}

func printDeviceList(cmd *cobra.Command, dev *sonm.DevicesReply) {
	if isSimpleFormat() {
		cpu := dev.GetCPU().GetDevice()
		cmd.Printf("CPU: %d cores at %d sockets\r\n", cpu.GetCores(), cpu.GetSockets())
		cmd.Printf("  %s\r\n", cpu.GetModelName())
		printBenchmarkGroup(cmd, dev.GetCPU().GetBenchmarks())

		ram := datasize.NewByteSize(dev.GetRAM().GetDevice().GetAvailable()).HumanReadable()
		cmd.Printf("RAM: %s\r\n", ram)
		printBenchmarkGroup(cmd, dev.GetRAM().GetBenchmarks())

		GPUs := dev.GetGPUs()
		if len(GPUs) > 0 {
			cmd.Printf("GPUs:\r\n")
			for i, gpu := range GPUs {
				cmd.Printf("  %s (index=%d: hash=%s)\r\n", gpu.Device.GetDeviceName(), i, gpu.GetDevice().GetHash())
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

func printOrdersList(cmd *cobra.Command, orders []*sonm.Order) {
	if isSimpleFormat() {
		if len(orders) == 0 {
			cmd.Println("No orders found")
			return
		}

		w := tablewriter.NewWriter(cmd.OutOrStdout())
		w.SetHeader([]string{"ID", "type", "tag", "price", "duration"})
		w.SetBorder(false)

		for _, order := range orders {
			var duration string
			if order.GetDuration() == 0 {
				duration = "spot"
			} else {
				duration = (time.Second * time.Duration(order.Duration)).String()
			}

			w.Append([]string{
				order.GetId().Unwrap().String(),
				order.OrderType.String(),
				string(order.GetTag()),
				order.PricePerHour(),
				duration,
			})
		}

		w.SetCaption(true, fmt.Sprintf("count: %d	", len(orders)))
		w.Render()
	} else {
		showJSON(cmd, map[string]interface{}{"orders": orders})
	}
}

func printOrderDetails(cmd Printer, order *sonm.Order) {
	if isSimpleFormat() {
		cmd.Printf("ID:              %s\r\n", order.Id)
		if !order.GetDealID().IsZero() {
			cmd.Printf("Deal ID:         %s\r\n", order.GetDealID().Unwrap().String())
		}
		cmd.Printf("Type:            %s\r\n", order.OrderType.String())
		cmd.Printf("Status:          %s\r\n", order.OrderStatus.String())
		if len(order.GetTag()) > 0 {
			cmd.Printf("Tag:             %s\r\n", string(order.GetTag()))
		}
		cmd.Printf("Identity:        %s\r\n", order.GetIdentityLevel().String())

		cmd.Printf("Duration:        %s\r\n", (time.Duration(order.GetDuration()) * time.Second).String())
		cmd.Printf("Total price:     %s USD (%s USD/hour)\r\n", order.TotalPrice(), order.PricePerHour())

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
		cmd.Printf("  GPU Ethash           %d\r\n", b.GPUEthHashrate())
		cmd.Printf("  GPU Equihash         %d\r\n", b.GPUCashHashrate())
		cmd.Printf("  GPU Redshift         %d\r\n", b.GPURedshift())
		cmd.Printf("  CPU Cryptonight      %d\r\n", b.CPUCryptonight())
	} else {
		showJSON(cmd, order)
	}
}

// TODO: Breaking issue #1225.
func typeEraseWithFieldMap(plan *sonm.AskPlan, mapping map[string]func(v interface{}) interface{}) yaml.MapSlice {
	v := reflect.Indirect(reflect.ValueOf(plan))
	ty := v.Type()

	values := yaml.MapSlice{}
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// Mimic previous behaviour.
		fieldName := strings.ToLower(ty.Field(i).Name)
		fieldValue := field.Interface()

		if fn, ok := mapping[fieldName]; ok {
			fieldValue = reflect.ValueOf(fn(fieldValue)).Interface()
		}

		if !(field.Kind() == reflect.Ptr && field.IsNil()) {
			values = append(values, yaml.MapItem{
				Key:   fieldName,
				Value: fieldValue,
			})
		}
	}

	return values
}

func printAskList(cmd *cobra.Command, slots *sonm.AskPlansReply) {
	if isSimpleFormat() {
		plans := map[string]yaml.MapSlice{}
		for k, v := range slots.GetAskPlans() {
			plans[k] = typeEraseWithFieldMap(v, map[string]func(v interface{}) interface{}{
				"tag": func(v interface{}) interface{} {
					return string(v.([]byte))
				},
			})
		}

		if len(plans) == 0 {
			cmd.Printf("No Ask Order configured\r\n")
			return
		}
		out, err := yaml.Marshal(plans)
		if err != nil {
			ShowError(cmd, "could not marshall ask plans", err)
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

func getDealCounterpartyString(d *sonm.Deal) string {
	addr, _ := keystore.GetDefaultAddress()
	if d.GetConsumerID().Unwrap() == addr {
		return d.GetSupplierID().Unwrap().Hex()
	} else {
		return d.GetConsumerID().Unwrap().Hex()
	}
}

func printDealsList(cmd *cobra.Command, deals []*sonm.Deal) {
	if isSimpleFormat() {
		if len(deals) == 0 {
			cmd.Println("No deals found")
			return
		}

		w := tablewriter.NewWriter(cmd.OutOrStdout())
		w.SetHeader([]string{"ID", "price", "started at", "duration", "counterparty"})
		w.SetCaption(true, fmt.Sprintf("count: %d", len(deals)))
		w.SetBorder(false)
		for _, deal := range deals {
			var duration string
			if deal.GetDuration() == 0 {
				// deal have no duration, show as "spot"
				duration = "spot"
			} else {
				// otherwise, print duration of forward deal
				duration = (time.Second * time.Duration(deal.GetDuration())).String()
			}

			w.Append([]string{
				deal.GetId().Unwrap().String(),
				deal.PricePerHour(),
				deal.StartTime.Unix().Format(time.RFC3339),
				duration,
				getDealCounterpartyString(deal),
			})
		}

		w.Render()
	} else {
		showJSON(cmd, map[string]interface{}{"deals": deals})
	}

}

type ExtendedDealInfo struct {
	*sonm.DealInfoReply
	ChangeRequests *sonm.DealChangeRequestsReply `json:"changeRequests"`
	Ask            *sonm.Order                   `json:"ask"`
	Bid            *sonm.Order                   `json:"bid"`
}

func printDealInfo(cmd *cobra.Command, info *ExtendedDealInfo, flags printerFlags) {
	if isSimpleFormat() {
		deal := info.GetDeal()
		isClosed := deal.GetStatus() == sonm.DealStatus_DEAL_CLOSED
		start := deal.StartTime.Unix()
		end := deal.EndTime.Unix()
		dealDuration := end.Sub(start)
		lastBill := deal.GetLastBillTS().Unix()

		cmd.Printf("ID:           %s (%s deal)\r\n", deal.GetId(), deal.GetTypeName())

		if info.Ask != nil {
			cmd.Printf("ASK:\r\n")
			printer := &IndentPrinter{cmd, 4, ' '}
			printOrderDetails(printer, info.Ask)
		} else {
			cmd.Printf("ASK ID:       %s\r\n", deal.GetAskID().Unwrap().String())
		}
		if info.Bid != nil {
			cmd.Printf("BID:\r\n")
			printer := &IndentPrinter{cmd, 4, ' '}
			printOrderDetails(printer, info.Bid)
		} else {
			cmd.Printf("BID ID:       %s\r\n", deal.GetBidID().Unwrap().String())
		}
		cmd.Printf("Status:       %s\r\n", deal.GetStatus())
		if deal.IsSpot() {
			// for active spot deal we can show only pricePerHour
			cmd.Printf("Price:        %s USD/hour\r\n", deal.PricePerHour())
			if isClosed {
				// for closed deal we also can *calculate* total duration
				cmd.Printf("Duration:     %s\r\n", dealDuration.String())
			}
		} else {
			// for non-spot deal we can show duration, total price and pricePerSecond
			cmd.Printf("Price:        %s USD (%s USD/hour)\r\n", deal.TotalPrice(), deal.PricePerHour())
			cmd.Printf("Duration:     %s\r\n", dealDuration.String())
		}

		if info.GetResources().GetNetwork().GetNetFlags().GetIncoming() {
			cmd.Printf("Public IPs:   %s\r\n", strings.Join(info.PublicIPs, ", "))
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

		if info.ChangeRequests != nil && len(info.ChangeRequests.GetRequests()) > 0 {
			cmd.Println("Change requests:")
			for _, req := range info.ChangeRequests.GetRequests() {
				cmd.Printf("  id: %s,  new duration: %s, new price: %s USD/h\n",
					req.GetId().Unwrap().String(),
					time.Second*time.Duration(req.GetDuration()),
					req.GetPrice().PricePerHour())
			}
		}

		key, err := getDefaultKey()
		if err != nil {
			cmd.Printf("cannot get default key: %v\r\n", err)
			return
		}

		noWorkerRespond := info.GetResources() == nil && info.GetRunning() == nil && info.GetCompleted() == nil
		iamConsumer := crypto.PubkeyToAddress(key.PublicKey).Big().Cmp(deal.GetConsumerID().Unwrap().Big()) == 0

		if noWorkerRespond && iamConsumer && !flags.WarningSuppressed() {
			// seems like worker is offline, notify user about it
			cmd.Println("WARN: Seems like worker is offline: no respond for the resources and tasks request.")
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

func printErrorById(cmd *cobra.Command, errors *sonm.ErrorByID) {
	if isSimpleFormat() {
		for _, err := range errors.GetResponse() {
			status := "OK"
			if len(err.Error) != 0 {
				status = "FAIL: " + err.Error
			}
			cmd.Printf("ID %s: %s\n", err.Id.Unwrap().String(), status)
		}
	} else {
		showJSON(cmd, errors.GetResponse())
	}
}

func printID(cmd *cobra.Command, id string) {
	if isSimpleFormat() {
		cmd.Printf("ID = %s\r\n", id)
	} else {
		showJSON(cmd, map[string]string{"id": id})
	}
}

func printTaskStart(cmd *cobra.Command, start *sonm.StartTaskReply) {
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

func printBalanceInfo(cmd *cobra.Command, reply *sonm.BalanceReply) {
	sideSNM := reply.GetSideBalance().ToPriceString()
	liveSNM := reply.GetLiveBalance().ToPriceString()
	liveEth := reply.GetLiveEthBalance().ToPriceString()

	if isSimpleFormat() {
		cmd.Printf("On Ethereum: %s SNM | %s Eth\n", liveSNM, liveEth)
		cmd.Printf("On SONM:     %s SNM\n", sideSNM)
	} else {
		showJSON(cmd, map[string]map[string]string{"balance": {
			"eth_live": liveEth,
			"snm_live": liveSNM,
			"snm_side": sideSNM,
			// TODO: Will be removed in the next
			// minor version update #1499
			"ethereum": liveSNM,
			"sonm":     sideSNM,
		}})
	}
}

func printMarketAllowance(cmd *cobra.Command, reply *sonm.BigInt) {
	if isSimpleFormat() {
		allowance := big.NewFloat(0.0).SetPrec(256).SetInt(reply.Unwrap())
		allowance = allowance.Quo(allowance, big.NewFloat(1e18))
		cmd.Printf("%s SNM\n", allowance.Text('g', 4))
	} else {
		showJSON(cmd, map[string]string{"allowance": reply.Unwrap().String()})
	}
}

func printBlacklist(cmd *cobra.Command, list *sonm.BlacklistReply) {
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

func printWorkersList(cmd *cobra.Command, list *sonm.WorkerListReply) {
	if isSimpleFormat() {
		if len(list.GetWorkers()) == 0 {
			cmd.Println("No workers for current master")
			return
		}

		var confirmed, notConfirmed []*sonm.DWHWorker
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

func printProfileInfo(cmd *cobra.Command, p *sonm.Profile) {
	if isSimpleFormat() {
		level := sonm.IdentityLevel(int32(p.GetIdentityLevel()))
		cmd.Printf("User ID: %s (%s)\r\n", p.GetUserID().Unwrap().Hex(), level.String())
		if len(p.GetName()) > 0 {
			cmd.Printf("Name: %s\r\n", p.GetName())
		}
		if len(p.GetCountry()) > 0 {
			cmd.Printf("Country:  %s\r\n", p.GetCountry())
		}
		cmd.Printf("Active orders: %d Bids, %d Asks\r\n", p.GetActiveBids(), p.GetActiveAsks())

		if len(p.GetCertificates()) > 0 {
			var certs []*sonm.Certificate
			if err := json.Unmarshal([]byte(p.GetCertificates()), &certs); err == nil {
				cmd.Println("  Certificates:")
				for _, cert := range certs {
					cmd.Printf("    %s) %s: %s\r\n", cert.GetId().Unwrap().String(), cert.GetAttributeName(), string(cert.Value))
				}
			}
		}

	} else {
		showJSON(cmd, p)
	}
}
