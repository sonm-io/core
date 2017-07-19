package mainer
//
//import (
//	"github.com/sonm-io/Fusrodah/hub"
//	"fmt"
//)
//
//func (mainer Server) firstFilter(neededBalance float64) []hub.HubsType {
//
//	//use filter: balance more then neededBalance
//	var someList []hub.HubsType
//	for _, hub := range mainer.Hubs {
//		if hub.Balance >= neededBalance {
//			someList = append(someList, hub)
//		}
//	}
//
//	mainer.Hubs = someList
//	fmt.Println("WHITELIST", mainer.Hubs)
//	return someList
//}
//func (mainer Server) secondFilter(neededBalance float64) []hub.HubsType {
//	//use filter: balance less then neededBalance
//	var someList []hub.HubsType
//	for _, hub := range mainer.Hubs {
//		if hub.Balance <= neededBalance {
//			someList = append(someList, hub)
//		}
//	}
//	mainer.Hubs = someList
//	fmt.Println("WhiteList2", mainer.Hubs)
//	return someList
//}
//func (mainer Server) AccountingPeriodFilter(neededPeriod int) []hub.HubsType {
//	//use filter: accountingPeriod > neededPeriod
//	var someList []hub.HubsType
//	for _, hub := range mainer.Hubs {
//		if hub.AccountingPeriod > neededPeriod {
//			someList = append(someList, hub)
//		}
//	}
//	mainer.Hubs = someList
//	fmt.Println("FilterPeriodList", mainer.Hubs)
//	return someList
//}