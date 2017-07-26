package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

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
	}

	app.Run(os.Args)
}
