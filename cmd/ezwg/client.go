package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/ezwg/ezwg/internal/client"
	"github.com/ezwg/ezwg/internal/wireguard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var clientProfileName string

var clientCmd = &cobra.Command{
	Use:     "client",
	Aliases: []string{"cli"},
	Short:   "Manage and connect to WireGuard VPN tunnels on Mac & Linux",
	Long:    "Import configuration profiles, connect/disconnect tunnels, and view real-time metrics.",
}

var clientImportCmd = &cobra.Command{
	Use:   "import <file-path-or-url>",
	Short: "Import a WireGuard .conf file or download from URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		source := args[0]
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		name := clientProfileName
		if name == "" {
			base := filepath.Base(source)
			name = strings.TrimSuffix(base, filepath.Ext(base))
		}

		savedPath, err := cm.ImportProfile(name, source)
		if err != nil {
			color.Red("❌ Failed to import profile: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Profile '%s' successfully imported!", name)
		color.White("Location: %s", savedPath)
		fmt.Println()
		color.Cyan("To connect now, simply run:")
		color.White("  ezwg connect %s", name)
	},
}

var clientConnectCmd = &cobra.Command{
	Use:     "connect [profile-name]",
	Aliases: []string{"up", "on"},
	Short:   "Connect to a WireGuard VPN tunnel",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		color.Cyan("Connecting WireGuard tunnel...")
		if err := cm.Connect(target); err != nil {
			color.Red("❌ Connection failed: %v", err)
			os.Exit(1)
		}

		color.Green("🔒 Connected successfully!")
	},
}

var clientDisconnectCmd = &cobra.Command{
	Use:     "disconnect [profile-name]",
	Aliases: []string{"down", "off"},
	Short:   "Disconnect an active WireGuard VPN tunnel",
	Args:    cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		target := ""
		if len(args) > 0 {
			target = args[0]
		}

		if err := cm.Disconnect(target); err != nil {
			color.Red("❌ Disconnect failed: %v", err)
			os.Exit(1)
		}

		color.Green("🔓 Disconnected from WireGuard VPN.")
	},
}

var clientListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all saved client VPN profiles",
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		profiles, err := cm.ListProfiles()
		if err != nil {
			color.Red("❌ Failed to list profiles: %v", err)
			os.Exit(1)
		}

		if len(profiles) == 0 {
			color.Yellow("No profiles found. Import one using:")
			color.White("  ezwg client import /path/to/wg0.conf")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tPROFILE\tCLIENT IP\tSERVER ENDPOINT")

		for _, p := range profiles {
			status := color.YellowString("○ Disconnected")
			if p.IsActive {
				status = color.GreenString("● Connected")
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", status, p.Name, p.ClientIP, p.Endpoint)
		}
		w.Flush()
	},
}

var clientStatusCmd = &cobra.Command{
	Use:   "status [profile-name]",
	Short: "Show connection status, latency, and transfer stats",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		target := ""
		if len(args) > 0 {
			target = args[0]
		} else {
			active, _ := cm.GetActiveProfile()
			if active != "" {
				target = active
			}
		}

		if target == "" {
			color.Yellow("No active WireGuard connection found.")
			color.White("View saved profiles: ezwg client list")
			return
		}

		status, err := cm.Status(target)
		if err != nil {
			color.Yellow("Tunnel '%s' is not connected: %v", target, err)
			return
		}

		color.Green("🔒 Tunnel: %s (CONNECTED)", target)
		color.White("Public Key: %s", status.PublicKey)

		for _, p := range status.Peers {
			fmt.Println()
			color.Cyan("Peer Gateway:")
			color.White("  Endpoint:        %s", p.Endpoint)
			color.White("  Allowed IPs:     %s", strings.Join(p.AllowedIPs, ", "))
			color.White("  Data Received:   %s", wireguard.FormatBytes(p.TransferRx))
			color.White("  Data Sent:       %s", wireguard.FormatBytes(p.TransferTx))
			if !p.LatestHandshake.IsZero() {
				color.White("  Latest Handshake:%s (%s ago)", p.LatestHandshake.Format("15:04:05"), strings.TrimSpace(formatDuration(p.LatestHandshake)))
			}
			if p.Endpoint != "" {
				latency := client.PingServer(p.Endpoint)
				color.White("  Endpoint Ping:   %s", latency)
			}
		}
	},
}

var clientDeleteCmd = &cobra.Command{
	Use:     "delete <profile-name>",
	Aliases: []string{"rm"},
	Short:   "Delete a saved client profile",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cm, err := client.NewClientManager()
		if err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}

		if err := cm.DeleteProfile(args[0]); err != nil {
			color.Red("❌ Error: %v", err)
			os.Exit(1)
		}
		color.Green("✅ Profile '%s' removed.", args[0])
	},
}

// Root-level shortcuts
var quickConnectCmd = &cobra.Command{
	Use:   "connect [profile-name]",
	Short: "Shortcut to connect to a WireGuard profile",
	Args:  cobra.MaximumNArgs(1),
	Run:   clientConnectCmd.Run,
}

var quickDisconnectCmd = &cobra.Command{
	Use:   "disconnect [profile-name]",
	Short: "Shortcut to disconnect active WireGuard profile",
	Args:  cobra.MaximumNArgs(1),
	Run:   clientDisconnectCmd.Run,
}

var quickStatusCmd = &cobra.Command{
	Use:   "status [profile-name]",
	Short: "Shortcut to view active VPN status",
	Args:  cobra.MaximumNArgs(1),
	Run:   clientStatusCmd.Run,
}

func init() {
	clientImportCmd.Flags().StringVarP(&clientProfileName, "name", "n", "", "Custom name for the profile")

	clientCmd.AddCommand(clientImportCmd)
	clientCmd.AddCommand(clientConnectCmd)
	clientCmd.AddCommand(clientDisconnectCmd)
	clientCmd.AddCommand(clientListCmd)
	clientCmd.AddCommand(clientStatusCmd)
	clientCmd.AddCommand(clientDeleteCmd)
}

func formatDuration(t time.Time) string {
	d := time.Since(t)
	if d.Seconds() < 60 {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	}
	if d.Minutes() < 60 {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	}
	return fmt.Sprintf("%d hours", int(d.Hours()))
}
