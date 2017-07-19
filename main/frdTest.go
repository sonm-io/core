package main


import (
	"fmt"
	"github.com/sonm-io/Fusrodah"
)

func main(){
	done := make(chan struct{})
	frd := Fusrodah.NewFusrodah(nil)

	//frd.Init()
	frd.Start()

	frd.Send("Hello world", "Main")

	fmt.Println(frd)


	close(done)
	select {
	case <-done:
	}
}