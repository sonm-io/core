package main

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/sonm-io/insonmnia/proto/hub"
	"github.com/urfave/cli"
)

func main() {
	var (
		hubendpoint string
		timeout     time.Duration

		gctx = context.Background()
	)

	app := cli.NewApp()
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
				for _, name := range lr.Name {
					fmt.Println(name)
				}
				return nil
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
	}

	app.Run(os.Args)
}
