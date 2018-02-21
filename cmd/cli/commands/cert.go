package commands

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
)

var (
	certCertOutput string
	certKeyOutput  string
)

func init() {
	makeCertCmd.PersistentFlags().StringVar(&certCertOutput, "cert", "sonm.crt", "file to save certificate")
	makeCertCmd.PersistentFlags().StringVar(&certKeyOutput, "key", "sonm.key", "file to save key")
}

var makeCertCmd = &cobra.Command{
	Use:    "make-cert",
	Short:  "Make TLS certificate for gRPC based on Eth key",
	PreRun: loadKeyStoreWrapper,
	Run: func(cmd *cobra.Command, args []string) {
		cert, key, err := util.GenerateCert(sessionKey, time.Duration(time.Hour))
		if err != nil {
			showError(cmd, "cannot generate cert", err)
			os.Exit(1)
		}

		keyFile := cmd.Flag("key").Value.String()
		certFile := cmd.Flag("cert").Value.String()

		err = ioutil.WriteFile(certFile, cert, 0600)
		if err != nil {
			showError(cmd, "cannot write cert file", err)
			os.Exit(1)
		}

		err = ioutil.WriteFile(keyFile, key, 0600)
		if err != nil {
			showError(cmd, "cannot write key file", err)
			os.Exit(1)
		}
	},
}
