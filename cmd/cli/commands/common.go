package commands

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"fmt"

	"github.com/sonm-io/core/cmd/cli/config"
	"github.com/spf13/cobra"

	pb "github.com/sonm-io/core/proto"
)

const (
	appName        = "sonm"
	hubAddressFlag = "addr"
	hubTimeoutFlag = "timeout"

	registryNameFlag     = "registry"
	registryUserFlag     = "user"
	registryPasswordFlag = "password"
)

var (
	rootCmd = &cobra.Command{Use: appName}
	gctx    = context.Background()

	version          string
	hubAddress       string
	timeout          = 60 * time.Second
	registryName     string
	registryUser     string
	registryPassword string
	cfg              config.Config

	errMinerAddressRequired = errors.New("Miner address is required")
	errTaskIDRequired       = errors.New("Task ID is required")
	errImageNameRequired    = errors.New("Image name is required")
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

func encodeRegistryAuth(login, password string) string {
	data := fmt.Sprintf("%s:%s", login, password)
	return b64.StdEncoding.EncodeToString([]byte(data))
}

func getMinerStatusByID(status *pb.TaskStatusReply) string {
	statusName, ok := pb.TaskStatusReply_Status_name[int32(status.Status)]
	if !ok {
		statusName = "UNKNOWN"
	}

	return statusName
}

func checkHubAddressIsSet(cmd *cobra.Command, args []string) error {
	if cmd.Flag(hubAddressFlag).Value.String() == "" {
		return fmt.Errorf("--%s flag is required", hubAddressFlag)
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

func showError(message string, err error) {
	if cfg.OutputFormat() == config.OutputModeSimple {
		showErrorInSimple(message, err)
	} else {
		showErrorInJSON(message, err)

	}
}

func showErrorInSimple(message string, err error) {
	if err != nil {
		fmt.Printf("[ERR] %s: %s\r\n", message, err.Error())
	} else {
		fmt.Printf("[ERR] %s\r\n", message)
	}
}

func showErrorInJSON(message string, err error) {
	jerr := newCommandError(message, err)
	fmt.Println(jerr.ToJSONString())
}

func showOk() {
	if cfg.OutputFormat() == config.OutputModeSimple {
		showOkSimple()
	} else {
		showOkJson()
	}
}

func showOkSimple() {
	fmt.Println("OK")
}

func showOkJson() {
	r := map[string]string{"status": "OK"}
	j, _ := json.Marshal(r)
	fmt.Println(string(j))
}

func isSimpleFormat() bool {
	return cfg.OutputFormat() == config.OutputModeSimple
}
