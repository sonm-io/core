package main


import (
	"fmt"
	"github.com/sonm-io/Fusrodah"
)

func main(){
	done := make(chan struct{})

	cfg := &Fusrodah.DefaultConfig
	cfg.Verbosity = 2

	frd := Fusrodah.NewFusrodah(cfg)

	frd.Start()

	if err:=frd.Send("Hello world", "Main"); err!=nil{
		fmt.Println(err)
	}

	fmt.Println(frd)


	close(done)
	select {
	case <-done:
	}
}