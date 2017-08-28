package commands

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/spf13/cobra"
)

const (
	appName        = "sonm"
	hubAddressFlag = "addr"
	hubTimeoutFlag = "timeout"
	outputModeFlag = "out"

	// log flag names
	logTypeFlag       = "type"
	sinceFlag         = "since"
	addTimestampsFlag = "ts"
	followFlag        = "follow"
	tailFlag          = "tail"
	detailsFlag       = "detailed"
)

var (
	rootCmd    = &cobra.Command{Use: appName}
	version    string
	hubAddress string
	outputMode string
	timeout    = 60 * time.Second
	cfg        config.Config

	// logging flag vars
	logType       string
	since         string
	addTimestamps bool
	follow        bool
	tail          string
	details       bool

	// errors
	errHubAddressRequired   = errors.New("--addr flag is required")
	errMinerAddressRequired = errors.New("Miner address is required")
	errTaskIDRequired       = errors.New("Task ID is required")
	errTaskFileRequired     = errors.New("Task definition file is required")
)

func init() {
	rootCmd.PersistentFlags().StringVar(&hubAddress, hubAddressFlag, "", "hub addr")
	rootCmd.PersistentFlags().DurationVar(&timeout, hubTimeoutFlag, 60*time.Second, "Connection timeout")
	rootCmd.PersistentFlags().StringVar(&outputMode, outputModeFlag, "", "Output mode: simple or json")
	rootCmd.AddCommand(hubRootCmd, minerRootCmd, tasksRootCmd, versionCmd)
}

// Root configure and return root command
func Root(c config.Config) *cobra.Command {
	cfg = c
	hubAddress = cfg.HubAddress()
	rootCmd.SetOutput(os.Stdout)
	return rootCmd
}

func checkHubAddressIsSet(cmd *cobra.Command, _ []string) error {
	if cmd.Flag(hubAddressFlag).Value.String() == "" {
		return errHubAddressRequired
	}
	return nil
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
	if outputMode == "" && cfg.OutputFormat() == "" {
		return true
	}

	if outputMode == config.OutputModeJSON || cfg.OutputFormat() == config.OutputModeJSON {
		return false
	}

	return true
}
