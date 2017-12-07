package commands

import (
	"encoding/json"
	"fmt"
	"time"

	ds "github.com/c2h5oh/datasize"
	"github.com/docker/go-connections/nat"
	pb "github.com/sonm-io/core/proto"
	"github.com/spf13/cobra"
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
			cmd.Printf("    MEM: %s\r\n", ds.ByteSize(taskStatus.Usage.GetMemory().GetMaxUsage()).HR())
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
			for containerPort, host := range ports {
				if len(host) > 0 {
					cmd.Printf("    %s: %s:%s\r\n", containerPort, host[0].HostIP, host[0].HostPort)
				} else {
					cmd.Printf("    %s\r\n", containerPort)
				}
			}
		}
	} else {
		v := map[string]interface{}{
			"id":     id,
			"miner":  taskStatus.MinerID,
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

		b, _ := json.Marshal(v)
		cmd.Println(string(b))
	}
}

func printNodeTaskStatus(cmd *cobra.Command, tasksMap map[string]*pb.TaskListReply_TaskInfo) {
	if isSimpleFormat() {
		for worker, tasks := range tasksMap {
			if len(tasks.GetTasks()) == 0 {
				cmd.Printf("Worker \"%s\" has no tasks\r\n", worker)
				continue
			}

			cmd.Printf("Worker \"%s\":\r\n", worker)
			i := 1
			for ID, status := range tasks.GetTasks() {
				up := time.Duration(status.GetUptime())
				cmd.Printf("  %d) %s \r\n     %s  %s (up: %v)\r\n",
					i, ID, status.Status.String(), status.ImageName, up.String())
				i++
			}
		}
	} else {
		b, _ := json.Marshal(tasksMap)
		fmt.Printf("%s\r\n", string(b))
	}
}

func printWorkerList(cmd *cobra.Command, lr *pb.ListReply) {
	if isSimpleFormat() {
		if len(lr.Info) == 0 {
			cmd.Printf("No workers connected\r\n")
			return
		}

		for addr, meta := range lr.Info {
			cmd.Printf("Worker: %s", addr)

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
		cmd.Printf("    CPU%d: %d x %s\r\n", i, cpu.GetCores(), cpu.GetModelName())
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

func printWorkerStatus(cmd *cobra.Command, workerID string, metrics *pb.InfoReply) {
	if isSimpleFormat() {
		cmd.Printf("Worker \"%s\":\r\n", workerID)

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

func printDeviceList(cmd *cobra.Command, devices *pb.DevicesReply) {
	if isSimpleFormat() {
		CPUs := devices.GetCPUs()
		GPUs := devices.GetGPUs()

		if len(CPUs) == 0 && len(GPUs) == 0 {
			cmd.Printf("No devices detected.\r\n")
			return
		}

		if len(CPUs) > 0 {
			cmd.Printf("CPUs:\r\n")
			for id, cpu := range CPUs {
				cmd.Printf(" %s: %s\r\n", id, cpu.Device.ModelName)
			}
		} else {
			cmd.Printf("No CPUs detected.\r\n")
		}

		if len(GPUs) > 0 {
			cmd.Printf("GPUs:\r\n")
			for id, gpu := range GPUs {
				cmd.Printf(" %s: %s\r\n", id, gpu.Device.Name)
			}
		} else {
			cmd.Printf("No GPUs detected.\r\n")
		}
	} else {
		b, _ := json.Marshal(devices)
		cmd.Println(string(b))
	}
}

func printDevicesProps(cmd *cobra.Command, props map[string]float64) {
	if isSimpleFormat() {
		for k, v := range props {
			cmd.Printf("%s = %f\r\n", k, v)
		}
	} else {
		b, _ := json.Marshal(props)
		cmd.Println(string(b))
	}
}

func printWorkerAclList(cmd *cobra.Command, list *pb.GetRegisteredWorkersReply) {
	if isSimpleFormat() {
		for i, id := range list.GetIds() {
			cmd.Printf("%d) %s\r\n", i+1, id.GetId())
		}

	} else {
		b, _ := json.Marshal(list)
		cmd.Printf("%s\r\n", string(b))
	}
}
