// How to use it?
//
// Direct usage: `pandora_OS_ARCH etc/pandora.yaml`

package main

import (
	"github.com/spf13/afero"
	"github.com/yandex/pandora/cli"
	importer "github.com/yandex/pandora/core/import"
	"github.com/yandex/pandora/core/register"
)

func main() {
	fs := afero.NewOsFs()
	importer.Import(fs)

	AmmoRegistry.Register("sonm.marketplace.GetOrderInfo", newOrderInfoAmmoFactory)
	AmmoRegistry.Register("sonm.marketplace.PlaceOrder", newOrderPlaceAmmoFactory)
	AmmoRegistry.Register("sonm.DWH.Orders", newDWHOrdersAmmoFactory)

	register.Gun("sonm.marketplace", NewMarketplaceGun)
	register.Gun("sonm.DWH", NewDWHGun)

	register.Provider("sonm", NewProvider)

	cli.Run()
}
