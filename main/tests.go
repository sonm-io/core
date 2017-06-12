package main

import (

)
import (
	"fmt"
	"github.com/sonm-io/Fusrodah"
	"github.com/sonm-io/go-ethereum/crypto"
)

/**
 /--------TEST--------/
 THIS FUNCTION FOR TEST
 /--------------------/
*/

func main() {

	/**
	HUB example
	 */
	hubPrv, _ := crypto.GenerateKey()
	//hubFrd := Fusrodah{prv: hubPrv}
	////hubFrd.start()
	//hub1 := hub.Server{}
	//hub1.DiscoveryHandling(hubFrd)

	/**
	Server example
	 */
	//mainer_1Prv, _ := crypto.GenerateKey()
	//mainer_1Frd := Fusrodah{prv: mainer_1Prv}
	//mainer_1Frd.start()
	//mainer_1 := mainer.Server{}
	//mainer_1.StartDiscovery(hubFrd)

	//fmt.Println("MAIN MAINER 1", mainer_1.Hubs)

	/**
	any Server example
	 */
	//mainer_2Prv, _ := crypto.GenerateKey()
	//mainer_2Frd := Fusrodah{prv: mainer_2Prv}
	//mainer_2 := mainer.Server{}
	//mainer_2.StartDiscovery(hubFrd)

	frd := Fusrodah.Fusrodah{Prv: hubPrv}
	//
	//frd.Start()
	//
	//pk := Fusrodah.ToPubKey(hubPrv)
	pk:= hubPrv.PublicKey
	fmt.Println(pk.X.String())
	fmt.Println(pk.X.String())
	pkString := string(crypto.FromECDSAPub(&pk))
	pkdString := crypto.ToECDSAPub([]byte(pkString))
	fmt.Println(pkString)
	fmt.Println(pkString)
	fmt.Println(pkdString)
	fmt.Println(pkdString)
	fmt.Println(string(pk.Y.String()))

	frd.Send("123", &pk)

	select {}
}
