package miner

import (
	"fmt"

	"github.com/sonm-io/fusrodah/hub"
)

func (srv *Server) firstFilter(neededBalance float64) []hub.HubsType {

	//use filter: balance more then neededBalance
	var someList []hub.HubsType
	for _, h := range srv.Hubs {
		if h.Balance >= neededBalance {
			someList = append(someList, h)
		}
	}

	srv.Hubs = someList
	fmt.Println("WHITELIST", srv.Hubs)
	return someList
}
func (srv *Server) secondFilter(neededBalance float64) []hub.HubsType {
	//use filter: balance less then neededBalance
	var someList []hub.HubsType
	for _, h := range srv.Hubs {
		if h.Balance <= neededBalance {
			someList = append(someList, h)
		}
	}
	srv.Hubs = someList
	fmt.Println("WhiteList2", srv.Hubs)
	return someList
}
func (srv *Server) AccountingPeriodFilter(neededPeriod int) []hub.HubsType {
	//use filter: accountingPeriod > neededPeriod
	var someList []hub.HubsType
	for _, h := range srv.Hubs {
		if h.AccountingPeriod > neededPeriod {
			someList = append(someList, h)
		}
	}
	srv.Hubs = someList
	fmt.Println("FilterPeriodList", srv.Hubs)
	return someList
}
