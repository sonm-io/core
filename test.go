package main

import "fmt"

func main() {
	wHashrate := 49000000
	dealHashrate := 50000000
	changeHashrate := float64(wHashrate) / float64(dealHashrate)
	fmt.Printf("res %v", changeHashrate)
}
