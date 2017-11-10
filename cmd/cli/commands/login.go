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
		ko, err := cliKeyOpener(cmd, cfg.KeyStore(), cfg.PassPhrase())
		if err != nil {
			showError(cmd, "Cannot init KeyOpener", err)
			os.Exit(1)
		}

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

// cliKeyOpener return KeyOpener configured for using with cli
func cliKeyOpener(p printer, keyDir, passPhrase string) (accounts.KeyOpener, error) {
	var err error
	// use default key store dir if not specified in config
	if keyDir == "" {
		keyDir, err = getDefaultKeyStorePath()
		if err != nil {
			return nil, err
		}
	}

	p.Printf("Using %s as KeyStore directory\r\n", keyDir)

	if !util.DirectoryExists(keyDir) {
		p.Printf("KeyStore directory does not exists, try to create it...\r\n")
		err = os.MkdirAll(keyDir, 0700)
		if err != nil {
			return nil, err
		}
	}

	// ask for pass-phrase if not specified in config
	var pf accounts.PassPhraser
	if passPhrase == "" {
		pf = accounts.NewInteractivePassPhraser()
	} else {
		pf = accounts.NewStaticPassPhraser(passPhrase)
	}

	ko := accounts.NewKeyOpener(keyDir, pf)
	return ko, nil
}

func getDefaultKeyStorePath() (string, error) {
	home, err := util.GetUserHomeDir()
	if err != nil {
		return "", err
	}

	keyDir := path.Join(home, defaultKeystorePath)
	return keyDir, nil
}
