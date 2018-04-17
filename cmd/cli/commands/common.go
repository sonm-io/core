package commands

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"os"
	"time"

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
)

var (
	rootCmd = &cobra.Command{Use: "sonmcli"}
	version string

	// flags var
	nodeAddressFlag string
	outputModeFlag  string
	timeoutFlag     = 60 * time.Second
	insecureFlag    bool

	// logging flag vars
	logType       string
	since         string
	addTimestamps bool
	follow        bool
	tail          string
	details       bool

	// session-related vars
	cfg        config.Config
	sessionKey *ecdsa.PrivateKey = nil
	creds      credentials.TransportCredentials
)

func init() {
	rootCmd.PersistentFlags().StringVar(&nodeAddressFlag, "node", "localhost:15030", "node endpoint")
	rootCmd.PersistentFlags().DurationVar(&timeoutFlag, "timeout", 60*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&outputModeFlag, "out", "", "Output mode: simple or json")
	rootCmd.PersistentFlags().BoolVar(&insecureFlag, "insecure", false, "Disable TLS for connection")

	rootCmd.AddCommand(workerMgmtCmd, marketRootCmd, nodeDealsRootCmd, taskRootCmd)
	rootCmd.AddCommand(loginCmd, approveTokenCmd, getTokenCmd, getBalanceCmd, versionCmd, autoCompleteCmd)
}

// Root configure and return root command
func Root(appVersion string, c config.Config) *cobra.Command {
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

func showError(cmd *cobra.Command, message string, err error) {
	if isSimpleFormat() {
		showErrorInSimple(cmd, message, err)
	} else {
		showErrorInJSON(cmd, message, err)
	}
}

func showErrorInSimple(cmd *cobra.Command, message string, err error) {
	if err != nil {
		cmd.Printf("[ERR] %s: %s\r\n", message, err.Error())
	} else {
		cmd.Printf("[ERR] %s\r\n", message)
	}
}

func showErrorInJSON(cmd *cobra.Command, message string, err error) {
	jerr := newCommandError(message, err)
	cmd.Println(jerr.ToJSONString())
}

func showOk(cmd *cobra.Command) {
	if isSimpleFormat() {
		showOkSimple(cmd)
	} else {
		showOkJson(cmd)
	}
}

func showOkSimple(cmd *cobra.Command) {
	cmd.Println("OK")
}

func showOkJson(cmd *cobra.Command) {
	r := map[string]string{"status": "OK"}
	j, _ := json.Marshal(r)
	cmd.Println(string(j))
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

// loadKeyStoreWrapper implemented to match cobra.Command.PreRun signature.
//
// Function loads and opens keystore. Also, storing opened key in "sessionKey" var
// to be able to reuse it into cli during one session.
func loadKeyStoreWrapper(cmd *cobra.Command, _ []string) {
	ko, err := accounts.DefaultKeyOpener(accounts.NewSilentPrinter(), cfg.KeyStore(), cfg.PassPhrase())
	if err != nil {
		showError(cmd, err.Error(), nil)
		os.Exit(1)
	}

	_, err = ko.OpenKeystore()
	if err != nil {
		showError(cmd, err.Error(), nil)
		os.Exit(1)
	}

	key, err := ko.GetKey()
	if err != nil {
		showError(cmd, err.Error(), nil)
		os.Exit(1)
	}

	sessionKey = key

	// If an insecure flag is set - we do not require TLS auth.
	// But we still need to load keys from a store, somewhere keys are used
	// to sign blockchain transactions, or something like that.
	if !insecureFlag {
		_, TLSConfig, err := util.NewHitlessCertRotator(context.Background(), sessionKey)
		if err != nil {
			showError(cmd, err.Error(), nil)
			os.Exit(1)
		}

		creds = auth.NewWalletAuthenticator(util.NewTLS(TLSConfig), util.PubKeyToAddr(sessionKey.PublicKey))
	}
}

// loadKeyStoreIfRequired loads eth keystore if `insecure` flag is not set.
// this wrapper is required for any command that not require eth keys implicitly
// but may use TLS to connect to the Node.
func loadKeyStoreIfRequired(cmd *cobra.Command, _ []string) {
	if !insecureFlag {
		loadKeyStoreWrapper(cmd, nil)
	}
}

func showJSON(cmd *cobra.Command, s interface{}) {
	b, _ := json.Marshal(s)
	cmd.Printf("%s\r\n", b)
}

func newTimeoutContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeoutFlag)
}
