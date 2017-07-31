package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/pkg/errors"
	pb "github.com/sonm-io/core/proto/hub"
	"github.com/urfave/cli"
)

func main() {
	var (
		hubendpoint string
		timeout     time.Duration

		gctx = context.Background()
	)

	app := cli.NewApp()
	app.Name = "command line application for SONM"
	app.Usage = ""
	app.Description = "CLI for SONM"
	app.Author = "noxiouz"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "hub",
			Value:       "",
			Usage:       "Hub to communicate",
			Destination: &hubendpoint,
		},
		cli.DurationFlag{
			Name:        "timeout",
			Value:       time.Second * 60,
			Usage:       "timeout for communication with the Hub",
			Destination: &timeout,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:    "list",
			Aliases: []string{"l"},
			Usage:   "get listing of connected miners to the Hub",
			Action: func(c *cli.Context) error {
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				lr, err := pb.NewHubClient(cc).List(ctx, &pb.ListRequest{})
				if err != nil {
					return err
				}
				fmt.Println("Connected Miners")
				buff := new(bytes.Buffer)
				enc := json.NewEncoder(buff)
				enc.SetIndent("", "\t")
				enc.Encode(lr.Info)
				fmt.Printf("%s\n", buff.Bytes())
				return err
			},
		},
		{
			Name:    "info",
			Aliases: []string{"i"},
			Usage:   "get runtime metrics from the specified Hub",
			Action: func(c *cli.Context) error {
				conn, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer conn.Close()

				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				var req = pb.InfoRequest{
					Miner: c.String("miner"),
				}
				metrics, err := pb.NewHubClient(conn).Info(ctx, &req)
				if err != nil {
					return err
				}

				js, err := json.Marshal(metrics)
				if err != nil {
					return err
				}

				fmt.Printf("%s", js)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "miner",
					Value: "",
					Usage: "miner endpoint",
				},
			},
		},
		{
			Name:    "ping",
			Aliases: []string{"p"},
			Usage:   "ping the Hub",
			Action: func(c *cli.Context) error {
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				_, err = pb.NewHubClient(cc).Ping(ctx, &pb.PingRequest{})
				if err != nil {
					return err
				}
				fmt.Println("OK")
				return nil
			},
		},
		{
			Name:  "start",
			Usage: "start a container using Hub",
			Action: func(c *cli.Context) error {
				fmt.Println(hubendpoint)
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				var req = pb.StartTaskRequest{
					Miner:    c.String("miner"),
					Image:    c.String("image"),
					Registry: c.String("registry"),
				}
				rep, err := pb.NewHubClient(cc).StartTask(ctx, &req)
				if err != nil {
					return err
				}
				fmt.Printf("ID %s, Endpoint %s", rep.Id, rep.Endpoint)
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "miner",
					Value: "",
					Usage: "miner to launch the task",
				},
				cli.StringFlag{
					Name:  "registry",
					Value: "",
					Usage: "registry for image",
				},
				cli.StringFlag{
					Name:  "image",
					Value: "",
					Usage: "image name in registry",
				},
			},
		},
		{
			Name:  "stop",
			Usage: "stop a container using Hub",
			Action: func(c *cli.Context) error {
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				var req = pb.StopTaskRequest{
					Id: c.String("id"),
				}
				_, err = pb.NewHubClient(cc).StopTask(ctx, &req)
				if err != nil {
					fmt.Println(err)
					return err
				}
				fmt.Println("OK")
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "id",
					Value: "",
					Usage: "job id to stop",
				},
			},
		},
		{
			Name:  "miner-status",
			Usage: "status of tasks on a specific miner",
			Action: func(c *cli.Context) error {
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				miner := c.String("miner")
				if len(miner) == 0 {
					return errors.New("required parameter\"miner\"")
				}
				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				var req = pb.MinerStatusRequest{
					Miner: miner,
				}
				minerStatus, err := pb.NewHubClient(cc).MinerStatus(ctx, &req)
				if err != nil {
					fmt.Println(err)
					return err
				}
				buff := new(bytes.Buffer)
				enc := json.NewEncoder(buff)
				enc.SetIndent("", "\t")
				//TODO: human readable statuses
				enc.Encode(minerStatus.Statuses)
				fmt.Printf("%s\n", buff.Bytes())
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "miner",
					Value: "",
					Usage: "miner to request statuses from",
				},
			},
		},
		{
			Name:  "task-status",
			Usage: "status of specific task",
			Action: func(c *cli.Context) error {
				cc, err := grpc.Dial(hubendpoint, grpc.WithInsecure())
				if err != nil {
					return err
				}
				defer cc.Close()

				task := c.String("task")
				if len(task) == 0 {
					return errors.New("required parameter\"miner\"")
				}
				ctx, cancel := context.WithTimeout(gctx, timeout)
				defer cancel()
				var req = pb.TaskStatusRequest{
					Id: task,
				}
				taskStatus, err := pb.NewHubClient(cc).TaskStatus(ctx, &req)
				if err != nil {
					fmt.Println(err)
					return err
				}
				fmt.Printf("Task %s status is %s\n", req.Id, taskStatus.Status.Status.String())
				return nil
			},
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "task",
					Value: "",
					Usage: "task to fetch status from",
				},
			},
		},
	}

	app.Run(os.Args)
}
