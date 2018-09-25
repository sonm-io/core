package commands

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	passwordFlag string
)

func init() {
	loginCmd.Flags().StringVar(&passwordFlag, "password", "", "Explicitly set password")
}

var loginCmd = &cobra.Command{
	Use:   "login [addr]",
	Short: "Open or generate Ethereum keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		var passReader accounts.PassPhraser
		if len(passwordFlag) > 0 {
			passReader = accounts.NewStaticPassPhraser(passwordFlag)
		} else {
			passReader = accounts.NewInteractivePassPhraser()
		}

		ks, err := initKeystore(passReader)
		if err != nil {
			return fmt.Errorf("cannot init keystore: %v", err)
		}

		keydir, err := keystorePath()
		if err != nil {
			return err
		}
		cmd.Printf("Keystore path: %s\n", keydir)

		if len(args) > 0 { // have a key
			if len(ks.List()) == 0 {
				return fmt.Errorf("cannot switch default address: keystore is empty")
			}

			// check if valid
			addr, err := util.HexToAddress(args[0])
			if err != nil {
				return err
			}

			// ask for password for default key
			pass, err := passReader.GetPassPhrase()
			if err != nil {
				return fmt.Errorf("cannot read pass phrase: %v", err)
			}

			// try to decrypt default key with given pass phrase
			if _, err := ks.GetKeyWithPass(addr, pass); err != nil {
				return fmt.Errorf("cannot decrypt default key with given pass: %v", err)
			}

			// mark key as default if we can decrypt it with given pass phrase
			if err := ks.SetDefault(addr); err != nil {
				cmd.Printf("Given address is not present in keystore.\r\nAvailable addresses:\r\n")
				for _, addr := range ks.List() {
					cmd.Println(addr.Address.Hex())
				}
				return nil
			}

			cfg.Eth.Passphrase = pass
			cfg.Eth.Keystore = keydir
			cfg.Save()

			cmd.Printf("Set \"%s\" as default keystore address\r\n", addr.Hex())
		} else { // no keys
			ls := ks.List()
			if len(ls) == 0 {
				// generate new key
				cmd.Println("Keystore is empty, generating new key...")
				// ask for password for default key
				pass, err := passReader.GetPassPhrase()
				if err != nil {
					return fmt.Errorf("cannot read pass phrase: %v", err)
				}

				newKey, err := ks.GenerateWithPassword(pass)
				if err != nil {
					return fmt.Errorf("cannot generate new key: %v", err)
				}

				cmd.Printf("Generated key %s set as default\r\n", crypto.PubkeyToAddress(newKey.PublicKey).Hex())
				cfg.Eth.Passphrase = pass
				cfg.Eth.Keystore = keydir
				cfg.Save()
				return nil
			}

			defaultAddr, err := ks.GetDefaultAddress()
			if err != nil {
				cmd.Printf("No default address for account, select one from list and use `sonmcli login [addr]`\r\n")
			} else {
				cmd.Printf("Default key: %s\r\n", defaultAddr.Hex())
				// try to decrypt default key with pre-defined pass
				if len(cfg.Eth.Passphrase) == 0 {
					pass, err := passReader.GetPassPhrase()
					if err != nil {
						return fmt.Errorf("cannot read pass phrase: %v", err)
					}

					cfg.Eth.Passphrase = pass
				}

				_, err = ks.GetKeyWithPass(defaultAddr, cfg.Eth.Passphrase)
				if err != nil {
					return fmt.Errorf("cannot decrypt default key with given pass: %v", err)
				}

				cfg.Eth.Keystore = keydir
				cfg.Save()
			}

			cmd.Println("Keystore contains following keys:")
			for _, acc := range ls {
				cmd.Printf("  %s\r\n", acc.Address.Hex())
			}
		}
		return nil
	},
}
