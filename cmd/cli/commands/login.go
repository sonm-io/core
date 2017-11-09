package commands

import (
	"os"
	"path"

	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var defaultKeystorePath = ".sonm/keystore/"

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Open or generate Etherum keys",
	Run: func(cmd *cobra.Command, _ []string) {
		var err error
		keyStorageDir := cfg.KeyStore()

		// use default key store dir if not specified in config
		if keyStorageDir == "" {
			keyStorageDir, err = getDefaultKeyStorePath()
			if err != nil {
				showError(cmd, "Cannot get default keystore", err)
				os.Exit(1)
			}
		}

		cmd.Printf("Using %s as KeyStore directory\r\n", keyStorageDir)

		if !util.DirectoryExists(keyStorageDir) {
			cmd.Printf("KeyStore directory does not exists, try to create it...\r\n")
			err = os.MkdirAll(keyStorageDir, 0700)
			if err != nil {
				showError(cmd, "Cannot create KeyStore directory", err)
				os.Exit(1)
			}
		}

		// ask for pass-phrase if not specified in config
		var pf accounts.PassPhraser
		if cfg.PassPhrase() == "" {
			pf = accounts.NewInteractivePassPhraser()
		} else {
			pf = accounts.NewStaticPassPhraser(cfg.PassPhrase())
		}

		ko := accounts.NewKeyOpener(keyStorageDir, pf)
		created, err := ko.OpenKeystore()
		if err != nil {
			showError(cmd, "Cannot open KeyStore", err)
			os.Exit(1)
		}

		if created {
			cmd.Printf("Keystore successfully created\r\n")
		} else {
			cmd.Printf("Keystore successfully opened\r\n")
		}
	},
}

func getDefaultKeyStorePath() (string, error) {
	home, err := util.GetUserHomeDir()
	if err != nil {
		return "", err
	}

	keyDir := path.Join(home, defaultKeystorePath)
	return keyDir, nil
}
