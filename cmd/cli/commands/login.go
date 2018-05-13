package commands

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

func init() {
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
	Use:   "list",
	Short: "Show Ethereum accounts list",
	Run: func(cmd *cobra.Command, _ []string) {
		defaultAddr := getDefaultKey("/Users/alex/.sonm")

		// TODO(sshaman1101): use `cfg.KeyStore()`
		dir := "/Users/alex/.sonm/test_keystore"
		ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

		// TODO(sshaman1101): make if JSON-friendly
		if len(ks.Accounts()) == 0 {
			cmd.Println("keystore is empty")
			return
		}

		for idx, acc := range ks.Accounts() {
			prefix := "  "
			if acc.Address.Big().Cmp(defaultAddr.Big()) == 0 {
				prefix = "* "
			}
			cmd.Printf("%s%d: %s\n", prefix, idx+1, acc.Address.Hex())
		}
	},
}

var accountsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create new Ethereum account",
	Run: func(cmd *cobra.Command, _ []string) {
		pf := accounts.NewInteractivePassPhraser()
		pass, err := pf.GetPassPhrase()
		if err != nil {
			showError(cmd, "Cannot read pass phrase", err)
			os.Exit(1)
		}

		// TODO(sshaman1101): unhardcode path
		dir := "/Users/alex/.sonm/test_keystore"
		ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

		// set key as default key it is first key in storage
		setDefault := len(ks.Accounts()) == 0
		acc, err := ks.NewAccount(pass)
		if err != nil {
			showError(cmd, "Cannot create account", err)
			os.Exit(1)
		}

		if setDefault {
			setDefaultKey("/Users/alex/.sonm", acc.Address)
		}

		// todo: JSON-friendly
		cmd.Printf("New account address = %s\r\n", acc.Address.Hex())
	},
}

var accountsImportCmd = &cobra.Command{
	Use:   "import <key.json>",
	Short: "Import exiting Ethereum account",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyPath := args[0]

		if !util.FileExists(keyPath) {
			showError(cmd, "File not exists", nil)
			os.Exit(1)
		}

		keyData, err := ioutil.ReadFile(keyPath)
		if err != nil {
			showError(cmd, "Cannot read key file", err)
			os.Exit(1)
		}

		pf := accounts.NewInteractivePassPhraser()
		pass, err := pf.GetPassPhrase()
		if err != nil {
			showError(cmd, "Cannot read pass phrase", err)
			os.Exit(1)
		}

		dir := "/Users/alex/.sonm/test_keystore"
		ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)

		acc, err := ks.Import(keyData, pass, pass)
		if err != nil {
			showError(cmd, "Cannot import account", err)
			os.Exit(1)
		}

		// TODO(sshaman1101): json-friendly
		cmd.Printf("Successfully imported account \"%s\"\n", acc.Address.Hex())
	},
}

var accountsSetDefaultCmd = &cobra.Command{
	Use:   "set-default <addr>",
	Short: "Set default account for keystore",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if !common.IsHexAddress(args[0]) {
			showError(cmd, "Given parameter is not an Ethereum address", nil)
			os.Exit(1)
		}

		addr := common.HexToAddress(args[0])

		// TODO(sshaman1101): unhardcode path
		dir := "/Users/alex/.sonm/test_keystore"
		ks := keystore.NewKeyStore(dir, keystore.LightScryptN, keystore.LightScryptP)
		for _, acc := range ks.Accounts() {
			if acc.Address.Big().Cmp(addr.Big()) == 0 {
				setDefaultKey("/Users/alex/.sonm", acc.Address)
				cmd.Printf("Using \"%s\" as default account\n", acc.Address.Hex())
				return
			}
		}

		showError(cmd, "Given address does not exists in keystore", nil)
		os.Exit(1)
	},
}

func getDefaultKey(p string) common.Address {
	stateFile := path.Join(p, "context")
	if !util.FileExists(stateFile) {
		fmt.Printf(" >>> context file not exists\n")
		return common.Address{}
	}

	data, err := ioutil.ReadFile(stateFile)
	if err != nil {
		fmt.Printf(" >>> cannot read file: %v\n", err)
		return common.Address{}
	}

	if !common.IsHexAddress(string(data)) {
		fmt.Printf(" >>> value is not CommonAddress \n")
		return common.Address{}
	}

	return common.HexToAddress(string(data))
}

func setDefaultKey(p string, addr common.Address) {
	stateFile := path.Join(p, "context")
	fmt.Printf(" >>> writing state to %s\n", stateFile)
	if err := ioutil.WriteFile(stateFile, []byte(addr.Hex()), 0600); err != nil {
		fmt.Printf(" >>> cannot write state %v\n", err)
	}
}
