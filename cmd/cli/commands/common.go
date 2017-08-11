package commands

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/spf13/cobra"
)

const (
	appName        = "sonm"
	hubAddressFlag = "addr"
	hubTimeoutFlag = "timeout"
)

var (
	rootCmd = &cobra.Command{Use: appName}
	gctx    = context.Background()

	version    string
	hubAddress string
	timeout    = 60 * time.Second
	cfg        config.Config

	errHubAddressRequired   = errors.New("--addr flag is required")
	errMinerAddressRequired = errors.New("Miner address is required")
	errTaskIDRequired       = errors.New("Task ID is required")
	errTaskFileRequired     = errors.New("Task definition file is required")
)

func init() {
	rootCmd.PersistentFlags().StringVar(&hubAddress, hubAddressFlag, "", "hub addr")
	rootCmd.PersistentFlags().DurationVar(&timeout, hubTimeoutFlag, 60*time.Second, "Connection timeout")
	rootCmd.AddCommand(hubRootCmd, minerRootCmd, tasksRootCmd, versionCmd)
}

// Root configure and return root command
func Root(c config.Config) *cobra.Command {
	cfg = c
	hubAddress = cfg.HubAddress()
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
	ce.Error = ce.rawErr.Error()
	j, _ := json.Marshal(ce)
	return string(j)
}

func newCommandError(message string, err error) *commandError {
	return &commandError{rawErr: err, Message: message}
}

func showError(cmd *cobra.Command, message string, err error) {
	if cfg.OutputFormat() == config.OutputModeSimple {
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
	if cfg.OutputFormat() == config.OutputModeSimple {
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
	return cfg.OutputFormat() == config.OutputModeSimple
}
