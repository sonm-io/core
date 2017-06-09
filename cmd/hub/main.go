package main

import (
	"log"
	"net"
	"sync"
	"time"

	pb "github.com/sonm-io/insonmnia/proto/hub"
	pbminer "github.com/sonm-io/insonmnia/proto/miner"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	// TODO: make it configurable
	externalGRPCEndpoint         = ":10001"
	minerHubInterconnectEndpoint = ":10002"
)

type daemon struct {
	externalGrpc  *grpc.Server
	minerListener net.Listener

	mu     sync.Mutex
	miners map[string]pbminer.MinerClient

	wg sync.WaitGroup
}

func createDaemon() (*daemon, error) {
	// TODO: add secure mechanism
	grpcServer := grpc.NewServer()
	// TODO: assign protocol

	il, err := net.Listen("tcp", minerHubInterconnectEndpoint)
	if err != nil {
		return nil, err
	}

	grpcL, err := net.Listen("tcp", externalGRPCEndpoint)
	if err != nil {
		il.Close()
		return nil, err
	}

	pb.RegisterHubServer(grpcServer, nil)

	d := &daemon{
		externalGrpc:  grpcServer,
		minerListener: il,

		miners: make(map[string]pbminer.MinerClient),
	}

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		d.externalGrpc.Serve(grpcL)
	}()

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()
		for {
			conn, err := d.minerListener.Accept()
			if err != nil {
				return
			}
			go d.handlerInterconnect(conn)
		}
	}()

	return d, nil
}

func (d *daemon) handlerInterconnect(conn net.Conn) {
	defer conn.Close()
	log.Printf("miner connected from %v", conn.RemoteAddr())

	cc, err := grpc.Dial("miner", grpc.WithDialer(func(_ string, _ time.Duration) (net.Conn, error) {
		return conn, nil
	}))
	if err != nil {
		log.Printf("failed to connect to Miner's grpc server: %v", err)
		return
	}
	defer cc.Close()
	minerClient := pbminer.NewMinerClient(cc)

	d.mu.Lock()
	d.miners[conn.RemoteAddr().String()] = minerClient
	d.mu.Unlock()

	defer func() {
		d.mu.Lock()
		delete(d.miners, conn.RemoteAddr().String())
		d.mu.Unlock()
	}()

	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for range t.C {
		// TODO: identify miner via Authorization mechanism
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		_, err = minerClient.Ping(ctx, &pbminer.PingRequest{})
		cancel()
		if err != nil {
			log.Printf("failed to ping miner %v", err)
			return
		}
	}
}

func (d *daemon) Stop() {
	d.externalGrpc.Stop()
	d.minerListener.Close()
	d.wg.Wait()
}

func main() {
	d, err := createDaemon()
	if err != nil {
		log.Fatalf("failed to create hub daemon: %v", err)
	}

	// TODO: call it from SIGNAL handler
	d.Stop()
}
