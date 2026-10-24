package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/ezwg/ezwg/internal/server"
	"github.com/ezwg/ezwg/internal/wireguard"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	serverConfigPath string
	serverPort       int
	serverSubnet     string
	serverPublicIP   string
	serverDNS        string
	serverOutIface   string
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"srv"},
	Short:   "Manage WireGuard VPN server on Linux/macOS",
	Long:    "Initialize, manage peers, and monitor a WireGuard VPN server.",
}

var serverInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize WireGuard server and configure firewall/routing",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerManager(serverConfigPath)
		if s.IsInitialized() {
			color.Yellow("Server is already initialized at: %s", serverConfigPath)
			color.Yellow("To add clients, run: ezwg server add-peer <name>")
			return
		}

		color.Cyan("🚀 Initializing WireGuard VPN Server...")
		opts := server.InitOptions{
			ConfigPath: serverConfigPath,
			Port:       serverPort,
			ServerIPv4: serverSubnet,
			PublicIP:   serverPublicIP,
			DNS:        serverDNS,
			OutIface:   serverOutIface,
		}

		if err := s.Init(opts); err != nil {
			color.Red("❌ Error initializing server: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Server initialized successfully!")
		color.Green("Config saved to: %s", serverConfigPath)
		fmt.Println()
		color.Cyan("Next step: Create client profiles with:")
		color.White("  ezwg server add-peer my-laptop")
	},
}

var serverAddPeerCmd = &cobra.Command{
	Use:     "add-peer <name>",
	Aliases: []string{"add-client", "new-peer"},
	Short:   "Add a new client peer and generate QR code/config",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientName := args[0]
		s := server.NewServerManager(serverConfigPath)

		color.Cyan("Creating client '%s'...", clientName)
		res, err := s.AddPeer(clientName)
		if err != nil {
			color.Red("❌ Error adding client: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Client '%s' added successfully!", res.Name)
		color.White("Assigned VPN IP: %s", res.ClientIPv4)
		color.White("Saved config to: %s", res.ConfigFile)
		fmt.Println()

		if res.QRCodeAscii != "" {
			color.Cyan("📱 Scan QR Code with WireGuard mobile app (iOS/Android):")
			fmt.Println(res.QRCodeAscii)
		}

		color.Cyan("💻 To connect another Mac or Linux computer, copy '%s' or run:", res.ConfigFile)
		color.White("  ezwg client import %s", res.ConfigFile)
		color.White("  ezwg client connect %s", res.Name)
	},
}

var serverRemovePeerCmd = &cobra.Command{
	Use:     "remove-peer <name|public-key>",
	Aliases: []string{"del-peer", "rm-peer"},
	Short:   "Remove and revoke a client peer",
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := args[0]
		s := server.NewServerManager(serverConfigPath)

		if err := s.RemovePeer(target); err != nil {
			color.Red("❌ Error removing peer: %v", err)
			os.Exit(1)
		}

		color.Green("✅ Peer '%s' successfully revoked and removed.", target)
	},
}

var serverListPeersCmd = &cobra.Command{
	Use:     "list-peers",
	Aliases: []string{"peers", "ls-peers"},
	Short:   "List all registered clients and their live traffic stats",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerManager(serverConfigPath)
		peers, err := s.ListPeers()
		if err != nil {
			color.Red("❌ Error listing peers: %v", err)
			os.Exit(1)
		}

		if len(peers) == 0 {
			color.Yellow("No peers registered yet. Add one with: ezwg server add-peer <name>")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tNAME\tALLOWED IP\tENDPOINT\tLAST HANDSHAKE\tRX / TX")

		for _, p := range peers {
			statusStr := color.GreenString("● Online")
			if !p.IsActive {
				statusStr = color.YellowString("○ Idle")
			}
			transfer := fmt.Sprintf("%s / %s", p.TransferRx, p.TransferTx)
			nameDisplay := p.Name
			if nameDisplay == "" {
				nameDisplay = p.PublicKey[:12] + "..."
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				statusStr, nameDisplay, p.AllowedIPs, p.Endpoint, p.LatestHandshake, transfer)
		}
		w.Flush()
	},
}

var serverShowPeerCmd = &cobra.Command{
	Use:   "show-peer <name>",
	Short: "Display config and QR code for an existing client",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		s := server.NewServerManager(serverConfigPath)
		res, err := s.GetClientConfig(name)
		if err != nil {
			color.Red("❌ Error retrieving client config: %v", err)
			os.Exit(1)
		}

		color.Cyan("Config File: %s", res.ConfigFile)
		if res.QRCodeAscii != "" {
			color.Cyan("QR Code:")
			fmt.Println(res.QRCodeAscii)
		}

		fmt.Println("----------------- CONFIG START -----------------")
		fmt.Print(res.ConfigText)
		fmt.Println("------------------ CONFIG END ------------------")
	},
}

var serverStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Display live server interface status and throughput",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerManager(serverConfigPath)
		tm := wireguard.NewTunnelManager()
		if !tm.IsInterfaceUp(s.Interface) {
			color.Yellow("Server interface '%s' is DOWN.", s.Interface)
			color.White("Start it using: ezwg server start")
			return
		}

		status, err := wireguard.GetInterfaceStatus(s.Interface)
		if err != nil {
			color.Red("❌ Failed to query status: %v", err)
			os.Exit(1)
		}

		color.Green("● WireGuard Server '%s' is UP and RUNNING", s.Interface)
		color.White("Listening Port: %d", status.ListenPort)
		color.White("Server Public Key: %s", status.PublicKey)
		color.White("Connected Peers: %d", len(status.Peers))
		fmt.Println()

		// Run list peers table
		serverListPeersCmd.Run(cmd, args)
	},
}

var serverStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Bring up the WireGuard server interface",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerManager(serverConfigPath)
		tm := wireguard.NewTunnelManager()
		if err := tm.Up(s.ConfigPath); err != nil {
			color.Red("❌ Failed to start server: %v", err)
			os.Exit(1)
		}
		color.Green("✅ WireGuard server started successfully.")
	},
}

var serverStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Bring down the WireGuard server interface",
	Run: func(cmd *cobra.Command, args []string) {
		s := server.NewServerManager(serverConfigPath)
		tm := wireguard.NewTunnelManager()
		if err := tm.Down(s.ConfigPath); err != nil {
			color.Red("❌ Failed to stop server: %v", err)
			os.Exit(1)
		}
		color.Green("✅ WireGuard server stopped.")
	},
}

func init() {
	serverInitCmd.Flags().StringVarP(&serverConfigPath, "config", "c", "/etc/wireguard/wg0.conf", "Path to WireGuard configuration file")
	serverInitCmd.Flags().IntVarP(&serverPort, "port", "p", 51820, "UDP port for WireGuard to listen on")
	serverInitCmd.Flags().StringVar(&serverSubnet, "subnet", "10.8.0.1/24", "Server IPv4 address and subnet")
	serverInitCmd.Flags().StringVar(&serverPublicIP, "public-ip", "", "Public IP or domain name for clients to connect to (auto-detected if omitted)")
	serverInitCmd.Flags().StringVar(&serverDNS, "dns", "1.1.1.1, 8.8.8.8", "DNS servers to push to connecting clients")
	serverInitCmd.Flags().StringVar(&serverOutIface, "out-iface", "", "Outbound network interface for NAT masquerade (auto-detected if omitted)")

	serverCmd.PersistentFlags().StringVarP(&serverConfigPath, "config", "c", "/etc/wireguard/wg0.conf", "Path to WireGuard server configuration file")

	serverCmd.AddCommand(serverInitCmd)
	serverCmd.AddCommand(serverAddPeerCmd)
	serverCmd.AddCommand(serverRemovePeerCmd)
	serverCmd.AddCommand(serverListPeersCmd)
	serverCmd.AddCommand(serverShowPeerCmd)
	serverCmd.AddCommand(serverStatusCmd)
	serverCmd.AddCommand(serverStartCmd)
	serverCmd.AddCommand(serverStopCmd)
}
