package main

import (
	"github.com/sonm-io/metricsStructs"
	"fmt"
	"io/ioutil"
	"os"
	"encoding/json"
)

type jsonobject struct {
	Object metricsStructs.MetricsHub
}

func main() {
	json.NewEncoder(os.Stdout).Encode("./metricsStructs/config.json")

	file, err := ioutil.ReadFile("./metricsStructs/config.json")
	if err != nil {
		fmt.Print("File error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Read json file: %s\n", string(file))

	metricsHubVar1 := &metricsStructs.MetricsHub{
		HubAddress:   "12345",
		HubPing:      "1234556",
		HubService:   "sercive",
		HubStack:     "dfdws",
		CreationDate: "ddsv",
	}

	to := metricsHubVar1.ToJSON()
	fmt.Printf("Result: %s", string(to))

	b := []byte(file)

	from := metricsHubVar1.FromJSON(b)
	fmt.Printf("Result from JSON: %s", string(from))

	//m := map[string]*metricsStructs.MetricsHub
	//m["c"] = metricsHubVar1

	//metricsHubVar2, _ := json.Marshal(metricsHubVar1)
	//fmt.Printf("Standart marshal structure: %s\n",string(metricsHubVar2))

}
