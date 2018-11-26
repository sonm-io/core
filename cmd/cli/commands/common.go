package commands

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/mitchellh/go-homedir"
	"github.com/sonm-io/core/accounts"
	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/util"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/credentials"
)

const (
	// log flag names
	logTypeFlag       = "type"
	sinceFlag         = "since"
	addTimestampsFlag = "ts"
	followFlag        = "follow"
	tailFlag          = "tail"
	detailsFlag       = "detailed"
	prependStreamFlag = "source"
	defaultNodeAddr   = "localhost:15030"

	// flag defaults
	defaultListLimit   = 100
	defaultConnTimeout = 60 * time.Second
)

var (
	rootCmd = &cobra.Command{
		Use:           "sonmcli",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	version string

	// flags var
	nodeAddressFlag string
	// deprecated: should be replaced with outputModeJSON
	outputModeFlag string
	outputModeJSON bool
	timeoutFlag    time.Duration
	insecureFlag   bool
	keystoreFlag   string
	configFlag     string
	listLimitFlag  uint64

	// logging flag vars
	logType       string
	since         string
	addTimestamps bool
	follow        bool
	tail          string
	details       bool
	prependStream bool

	// session-related vars
	cfg      = &config.Config{}
	creds    credentials.TransportCredentials
	keystore *accounts.MultiKeystore
)

func getDefaultKey() (*ecdsa.PrivateKey, error) {
	defaultAddr, err := keystore.GetDefaultAddress()
	if err != nil {
		return nil, fmt.Errorf("cannot read default address from keystore: %v", err)
	}

	var pass = cfg.Eth.Passphrase
	if pass == "" {
		pass, err = accounts.NewInteractivePassPhraser().GetPassPhrase()
		if err != nil {
			return nil, fmt.Errorf("cannot read pass phrase: %v", err)
		}
	}

	key, err := keystore.GetKeyWithPass(defaultAddr, pass)
	if err != nil {
		return nil, fmt.Errorf("cannot read default key from keystore: %v", err)
	}

	return key, nil
}

func init() {
	cobra.OnInitialize(func() {
		var err error
		cfg, err = config.NewConfig(configFlag)
		if err != nil {
			fmt.Printf("Cannot load config: %s\r\n", err)
			os.Exit(1)
		}

		if outputModeJSON {
			outputModeFlag = config.OutputModeJSON
		}
	})

	rootCmd.PersistentFlags().StringVar(&nodeAddressFlag, "node", "", "node endpoint")
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", defaultConnTimeout, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&outputModeFlag, "out", "", "Output mode: simple or json (DEPRECATED)")
	rootCmd.PersistentFlags().BoolVar(&outputModeJSON, "json", false, "Show command output in JSON format")
	rootCmd.PersistentFlags().BoolVar(&insecureFlag, "insecure", false, "Disable TLS for connection")
	rootCmd.PersistentFlags().StringVar(&keystoreFlag, "keystore", "", "Keystore dir")
	rootCmd.PersistentFlags().StringVar(&configFlag, "config", "", "Configuration file")

	rootCmd.AddCommand(workerMgmtCmd, orderRootCmd, dealRootCmd, taskRootCmd, blacklistRootCmd)
	rootCmd.AddCommand(loginCmd, tokenRootCmd, versionCmd, autoCompleteCmd, masterRootCmd, profileRootCmd)
}

// Root configure and return root command
func Root(appVersion string) *cobra.Command {
	version = appVersion

	rootCmd.SetOutput(os.Stdout)
	return rootCmd
}

// commandError allow to present any internal error as JSON
type commandError struct {
	rawErr  error
	Error   string `json:"error"`
	Message string `json:"message"`
}

func (ce *commandError) ToJSONString() string {
	if ce.rawErr != nil {
		ce.Error = ce.rawErr.Error()
	}

	j, _ := json.Marshal(ce)
	return string(j)
}

func newCommandError(message string, err error) *commandError {
	return &commandError{rawErr: err, Message: message}
}

// ShowError prints message and chained error in requested format
func ShowError(printer Printer, message string, err error) {
	if isSimpleFormat() {
		if err != nil {
			printer.Printf("[ERR] %s: %s\r\n", message, err.Error())
		} else {
			printer.Printf("[ERR] %s\r\n", message)
		}
	} else {
		printer.Println(newCommandError(message, err).ToJSONString())
	}
}

func showOk(cmd *cobra.Command) {
	if isSimpleFormat() {
		cmd.Println("OK")
	} else {
		showJSON(cmd, map[string]string{"status": "OK"})
	}
}

func isSimpleFormat() bool {
	if outputModeFlag == "" && cfg.OutputFormat() == "" {
		return true
	}

	if outputModeFlag == config.OutputModeJSON || cfg.OutputFormat() == config.OutputModeJSON {
		return false
	}

	return true
}

func nodeAddress() string {
	p := rootCmd.Flag("node").Value.String()
	// flag overrides config
	if p == "" {
		p = cfg.NodeAddr
		// config overrides defaults
		if p == "" {
			p = defaultNodeAddr
		}
	}

	return p
}

func keystorePath() (string, error) {
	var err error
	p := rootCmd.Flag("keystore").Value.String()
	// flag overrides config
	if p == "" {
		p = cfg.Eth.Keystore
		// config overrides defaults ~/.sonm/
		if p == "" {
			p, err = util.GetDefaultKeyStoreDir()
			if err != nil {
				return "", fmt.Errorf("cannot obtain default keystore dir: %v", err)
			}
		}
	}

	expanded, err := homedir.Expand(p)
	if err != nil {
		return "", err
	}

	return expanded, nil
}

func initKeystore(reader accounts.PassPhraser) (*accounts.MultiKeystore, error) {
	keyDir, err := keystorePath()
	if err != nil {
		return nil, err
	}

	return accounts.NewMultiKeystore(&accounts.KeystoreConfig{
		KeyDir:      keyDir,
		PassPhrases: make(map[string]string),
	}, reader)
}

// loadKeyStoreWrapper is matching cobra.Command.PreRunE signature.
// It loads default key from the given keystore and keeps the keystore instance
// in the global variable `keystore` that available for any CLI's sub-commands.
// Note that the keystore must be loaded before any command's logic it started to execute.
func loadKeyStoreWrapper(_ *cobra.Command, _ []string) error {
	var err error
	keystore, err = initKeystore(accounts.NewInteractivePassPhraser())
	if err != nil {
		return fmt.Errorf("cannot init keystore: %v", err)
	}

	// If an insecure flag is set - we do not require TLS auth.
	// But we still need to load keys from a store, somewhere keys are used
	// to sign blockchain transactions, or something like that.
	if !insecureFlag {
		sessionKey, err := getDefaultKey()
		if err != nil {
			return err
		}

		_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), sessionKey)
		if err != nil {
			return err
		}

		creds = auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), crypto.PubkeyToAddress(sessionKey.PublicKey))
	}

	return nil
}

func showJSON(cmd Printer, s interface{}) {
	b, _ := json.Marshal(s)
	cmd.Printf("%s\r\n", b)
}

func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeoutFlag)
}

func argsToBigInts(args []string) ([]*big.Int, error) {
	ints := make([]*big.Int, 0, len(args))
	for _, idStr := range args {
		id, err := util.ParseBigInt(idStr)
		if err != nil {
			return nil, err
		}
		ints = append(ints, id)
	}
	return ints, nil
}
