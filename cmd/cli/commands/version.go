package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run:   func(cmd *cobra.Command, args []string) { versionCmdRunner(cmd) },
}

func versionCmdRunner(cmd *cobra.Command) {
	if isSimpleFormat() {
		cmd.Printf("Version: %s\r\n", version)
	} else {
		v := map[string]string{
			"version": version,
		}
		b, _ := json.Marshal(v)
		cmd.Printf(string(b))
	}
}
