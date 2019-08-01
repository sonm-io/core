package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sonm-io/core/insonmnia/auth"
	"github.com/sonm-io/core/secsh/secshc"
	"github.com/spf13/cobra"
)

var (
	configPath string
)

func init() {
	rootCmd.Flags().StringVar(&configPath, "config", "etc/secterm.yaml", "Path to the configuration file")
}

func runSecTerm(v string) error {
	addr, err := auth.ParseAddr(v)
	if err != nil {
		return fmt.Errorf("failed to parse target address: %v", err)
	}

	cfg, err := secshc.NewRPTYConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	tty, err := secshc.NewRemotePTY(cfg)
	if err != nil {
		return fmt.Errorf("failed to construct remote PTY: %v", err)
	}

	if err := tty.Run(context.Background(), *addr); err != nil {
		return fmt.Errorf("failed to run remote PTY: %v", err)
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "secterm ADDR",
	Short: "Secure PTY",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSecTerm(args[0]); err != nil {
			fmt.Printf("ERROR: %v\n\r", err)
			os.Exit(1)
		}
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("ERROR: %v\n\r", err)
		os.Exit(1)
	}
}
