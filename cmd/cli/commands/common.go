package commands

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
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
	outputModeFlag  string
	timeoutFlag     = 60 * time.Second
	insecureFlag    bool
	keystoreFlag    string

	// logging flag vars
	logType       string
	since         string
	addTimestamps bool
	follow        bool
	tail          string
	details       bool
	prependStream bool

	// session-related vars
	cfg      *config.Config
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
	rootCmd.PersistentFlags().StringVar(&nodeAddressFlag, "node", "localhost:15030", "node endpoint")
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", 60*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&outputModeFlag, "out", "", "Output mode: simple or json")
	rootCmd.PersistentFlags().BoolVar(&insecureFlag, "insecure", false, "Disable TLS for connection")
	rootCmd.PersistentFlags().StringVar(&keystoreFlag, "keystore", "", "Keystore dir")

	rootCmd.AddCommand(workerMgmtCmd, orderRootCmd, dealRootCmd, taskRootCmd, blacklistRootCmd)
	rootCmd.AddCommand(loginCmd, tokenRootCmd, versionCmd, autoCompleteCmd, masterRootCmd, profileRootCmd)
}

// Root configure and return root command
func Root(appVersion string, c *config.Config) *cobra.Command {
	version = appVersion
	cfg = c

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
func ShowError(cmd *cobra.Command, message string, err error) {
	if isSimpleFormat() {
		if err != nil {
			cmd.Printf("[ERR] %s: %s\r\n", message, err.Error())
		} else {
			cmd.Printf("[ERR] %s\r\n", message)
		}
	} else {
		cmd.Println(newCommandError(message, err).ToJSONString())
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

func initKeystore() (*accounts.MultiKeystore, error) {
	keyDir, err := keystorePath()
	if err != nil {
		return nil, err
	}

	return accounts.NewMultiKeystore(&accounts.KeystoreConfig{
		KeyDir:      keyDir,
		PassPhrases: make(map[string]string),
	}, accounts.NewInteractivePassPhraser())
}

// loadKeyStoreWrapper implemented to match cobra.Command.PreRun signature.
//
// Function loads and opens keystore. Also, storing opened key in "sessionKey" var
// to be able to reuse it into cli during one session.
func loadKeyStoreWrapper(_ *cobra.Command, _ []string) error {
	var err error
	keystore, err = initKeystore()
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

// loadKeyStoreIfRequired loads eth keystore if `insecure` flag is not set.
// this wrapper is required for any command that not require eth keys implicitly
// but may use TLS to connect to the Node.
func loadKeyStoreIfRequired(cmd *cobra.Command, _ []string) error {
	if !insecureFlag {
		return loadKeyStoreWrapper(cmd, nil)
	}

	return nil
}

func showJSON(cmd *cobra.Command, s interface{}) {
	b, _ := json.Marshal(s)
	cmd.Printf("%s\r\n", b)
}

func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeoutFlag)
}
