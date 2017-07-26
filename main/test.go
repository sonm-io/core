package main

import (
	"github.com/sonm-io/metricsStructs"
	"fmt"
)

type jsonobject struct {
	Object metricsStructs.MetricsHub
}

func main() {
	myFile := "./metricsStructs/config.json"

	m := &metricsStructs.MetricsHub{}
	err := m.LoadFromFile(myFile)
	if err != nil {
		fmt.Printf("Cannot load from file: %s\r\n", err)
		return
	}

	m.HubStack = "hooe"
	err = m.SaveToFile(myFile)
	if err != nil {
		fmt.Printf("Cannot write to file: %s\r\n", err)
		return
	}

	//json.NewEncoder(os.Stdout).Encode()
	//
	//file, err := ioutil.ReadFile("./metricsStructs/config.json")
	//if err != nil {
	//	fmt.Print("File error: %v\n", err)
	//	os.Exit(1)
	//}
	//fmt.Printf("Read json file: %s\n", string(file))
	//
	//metricsHubVar1 := &metricsStructs.MetricsHub{
	//	HubAddress:   "12345",
	//	HubPing:      "1234556",
	//	HubService:   "sercive",
	//	HubStack:     "dfdws",
	//	CreationDate: "ddsv",
	//}
	//
	//to := metricsHubVar1.ToJSON()
	//fmt.Printf("Result: %s", string(to))
	//
	//// by := []byte(file)
	//err = metricsHubVar1.FromJSON(file)
	//fmt.Printf("Result from JSON: %s", err)

	//m := map[string]*metricsStructs.MetricsHub
	//m["c"] = metricsHubVar1

	//metricsHubVar2, _ := json.Marshal(metricsHubVar1)
	//fmt.Printf("Standart marshal structure: %s\n",string(metricsHubVar2))

}
