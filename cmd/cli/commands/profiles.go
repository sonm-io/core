package commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/proto"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func init() {
	profileRootCmd.AddCommand(
		profileStatusCmd,
		profileRemoveAttrCmd,
	)
}

var profileRootCmd = &cobra.Command{
	Use:               "profile",
	Short:             "Manage profiles",
	PersistentPreRunE: loadKeyStoreWrapper,
}

var profileStatusCmd = &cobra.Command{
	Use:   "status [addr]",
	Short: "Show profile details",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		key, err := getDefaultKey()
		if err != nil {
			return err
		}

		addr := crypto.PubkeyToAddress(key.PublicKey)
		if len(args) > 0 {
			addr, err = util.HexToAddress(args[0])
			if err != nil {
				return err
			}
		}

		client, err := newProfilesClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		profile, err := client.Status(ctx, &sonm.EthID{Id: sonm.NewEthAddress(addr)})
		if err != nil {
			if status.Code(err) == codes.NotFound {
				return fmt.Errorf("cannot find profile for address `%s`", addr.Hex())
			}

			return fmt.Errorf("cannot get profile info: %v", err)
		}

		printProfileInfo(cmd, profile)
		return nil
	},
}

var profileRemoveAttrCmd = &cobra.Command{
	Use:   "remove-attr <id>",
	Short: "Remove attribute form your profile",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := newTimeoutContext()
		defer cancel()

		id, err := sonm.NewBigIntFromString(args[0])
		if err != nil {
			return err
		}

		client, err := newProfilesClient(ctx)
		if err != nil {
			return fmt.Errorf("cannot create client connection: %v", err)
		}

		if _, err := client.RemoveAttribute(ctx, id); err != nil {
			return fmt.Errorf("cannot remove profile attribue: %v", err)
		}

		showOk(cmd)
		return nil
	},
}
