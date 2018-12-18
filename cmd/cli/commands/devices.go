package commands

import (
	"fmt"

	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
	devicesRootCmd.AddCommand(
		devicesLoadCmd,
		devicesLoadRawCmd,
	)
}

var devicesRootCmd = &cobra.Command{
	Use:               "devices",
	Short:             "load stored worker devices from blockchain",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var devicesLoadRawCmd = &cobra.Command{
	Use:   "loadRaw [addr]",
	Short: "Load worker devices from blockchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dev, err := newDevicesClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		ownerAddr, err := keystore.GetDefaultAddress()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			ownerAddr, err = util.HexToAddress(args[0])
			if err != nil {
				return err
			}
		}

		list, err := dev.RawDevices(ctx, sonm.NewEthAddress(ownerAddr))
		if err != nil {
			return fmt.Errorf("cannot get blacklist: %v", err)
		}
		printRawDevices(cmd, list)
		return nil
	},
}

var devicesLoadCmd = &cobra.Command{
	Use:   "load [addr]",
	Short: "Load worker devices from blockchain",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		dev, err := newDevicesClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		ownerAddr, err := keystore.GetDefaultAddress()
		if err != nil {
			return err
		}

		if len(args) > 0 {
			ownerAddr, err = util.HexToAddress(args[0])
			if err != nil {
				return err
			}
		}

		list, err := dev.Devices(ctx, sonm.NewEthAddress(ownerAddr))
		if err != nil {
			return fmt.Errorf("cannot get blacklist: %v", err)
		}

		printStoredDevices(cmd, list)
		return nil
	},
}
