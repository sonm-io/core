package main

import (
	"fmt"
	"github.com/sonm-io/Fusrodah"
	"time"
)

func main() {
	done := make(chan struct{})

	cfg := &Fusrodah.DefaultConfig
	cfg.Verbosity = 5
	cfg.P2pPort = ":30347"
	cfg.Enode = "enode://8614cbcc79eaede3f26731c0002e768f15adbcdb5f7dab2961536959d36e580fd9183075dc89a3732805f4ce83a9bfb0612f5bc9ad61c01beebad0dea52dd4f9@192.168.10.51:30348"
	frd := Fusrodah.NewFusrodah(cfg)

	frd.Start()

	if err := frd.Send("Hello world", "main"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Message sended.")
	}

	filter, err := frd.AddHandler("main")
	if err != nil {
		fmt.Println(err)
	}

	ticker := time.NewTicker(time.Millisecond * 50)

	for {

		select {
		case <-ticker.C:
			//fmt.Println("Tik")
			if err := frd.Send("Hello world", "main"); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Message sended.")
			}
			fmt.Println(frd.WH.Envelopes())
			envs := frd.WH.Envelopes()
			for _, env := range envs {
				mess, err := env.OpenAsymmetric(frd.Conf.AsymKey)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println(string(mess.Payload))
				}
			}
			messages := filter.Retrieve()

			for _, msg := range messages {
				fmt.Println(string(msg.Payload))
			}
		}

	}

	if err := frd.Send("Hello world", "main"); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println("Message sended.")
	}

	close(done)
	select {
	case <-done:
	}
}
