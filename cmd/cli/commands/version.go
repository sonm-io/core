package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		if isSimpleFormat() {
			fmt.Printf("Version: %s\r\n", version)
		} else {
			v := map[string]string{
				"version": version,
			}
			b, _ := json.Marshal(v)
			fmt.Printf(string(b))
		}
	},
}
