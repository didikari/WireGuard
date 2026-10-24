package main

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version = "1.0.0"
	commit  = "dev"
	date    = "2026-10-24"
)

var rootCmd = &cobra.Command{
	Use:   "ezwg",
	Short: "ezwg - The hassle-free WireGuard Server & Client Manager for Linux & macOS",
	Long: color.CyanString(`
  _____ ______          _______ 
 / ____|___  /         / ____(_)
| |  __   / / |  _ _ _| |  __ _ 
| | |_ | / /| | | | '_| | |_ | |
| |__| |/ /_| |_| | | | |__| | |
 \_____/_____|__,_|_|  \_____|_|
`) + color.WhiteString(`
Easy WireGuard (ezwg) is an all-in-one CLI designed to make installing,
configuring, and using WireGuard on Linux and macOS completely turnkey.
`),
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(doctorCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(clientCmd)

	// Quick shortcuts at root level
	rootCmd.AddCommand(quickConnectCmd)
	rootCmd.AddCommand(quickDisconnectCmd)
	rootCmd.AddCommand(quickStatusCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print ezwg version and build info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ezwg version %s (%s) built at %s\n", version, commit, date)
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
