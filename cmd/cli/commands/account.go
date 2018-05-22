package commands

import (
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

var setImportAccountAsDefault bool

func init() {
	accountsImportCmd.PersistentFlags().BoolVar(&setImportAccountAsDefault, "as-default",
		false, "Set imported account as default")
	accountsRootCmd.AddCommand(
		accountsListCmd,
		accountsCreateCmd,
		accountsImportCmd,
		accountsSetDefaultCmd,
	)
}

var accountsRootCmd = &cobra.Command{
	Use: "accounts",
}

var accountsListCmd = &cobra.Command{
	Use:    "list",
	Short:  "Show Ethereum accounts list",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, _ []string) {
		accounts, err := keystore.List()
		if err != nil {
			showError(cmd, "cannot obtain accounts list", err)
			os.Exit(1)
		}

		if len(accounts) == 0 {
			cmd.Println("keystore is empty")
			return
		}

		// todo: JSON-friendly
		defaultKey, _ := keystore.GetDefault()
		defaultAddr := crypto.PubkeyToAddress(defaultKey.PublicKey)

		for idx, acc := range accounts {
			prefix := "  "
			if acc.Address.Big().Cmp(defaultAddr.Big()) == 0 {
				prefix = "* "
			}
			cmd.Printf("%s%d: %s\n", prefix, idx+1, acc.Address.Hex())
		}
	},
}

var accountsCreateCmd = &cobra.Command{
	Use:    "create",
	Short:  "Create new Ethereum account",
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, _ []string) {
		key, err := keystore.Generate()
		if err != nil {
			showError(cmd, "Cannot create account", err)
			os.Exit(1)
		}

		// todo: JSON-friendly
		cmd.Printf("New account address = %s\r\n", crypto.PubkeyToAddress(key.PublicKey).Hex())
	},
}

var accountsImportCmd = &cobra.Command{
	Use:    "import <key.json>",
	Short:  "Import exiting Ethereum account",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := keystore.Import(args[0])
		if err != nil {
			showError(cmd, "Cannot import account", err)
			os.Exit(1)
		}

		if setImportAccountAsDefault {
			keystore.SetDefault(addr)
		}

		// TODO(sshaman1101): json-friendly
		cmd.Printf("Successfully imported account \"%s\"\n", addr.Hex())
	},
}

var accountsSetDefaultCmd = &cobra.Command{
	Use:    "set-default <addr>",
	Short:  "Set default account for keystore",
	Args:   cobra.MinimumNArgs(1),
	PreRun: loadKeyStoreIfRequired,
	Run: func(cmd *cobra.Command, args []string) {
		if !common.IsHexAddress(args[0]) {
			showError(cmd, "Given parameter is not an Ethereum address", nil)
			os.Exit(1)
		}

		addr := common.HexToAddress(args[0])
		if err := keystore.SetDefault(addr); err != nil {
			showError(cmd, "Cannot set default address", nil)
			os.Exit(1)
		}

		// todo: JSON?
		cmd.Printf("Using \"%s\" as default account\n", addr.Hex())
	},
}
