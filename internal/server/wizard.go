package server

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/ezwg/ezwg/internal/sysutil"
	"github.com/fatih/color"
)

// RunInteractiveInit runs an interactive setup wizard asking questions with auto-detected defaults.
func (s *ServerManager) RunInteractiveInit() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println()
	color.Cyan("╔════════════════════════════════════════════════════════════╗")
	color.Cyan("║          Easy WireGuard (ezwg) Server Setup Wizard         ║")
	color.Cyan("╚════════════════════════════════════════════════════════════╝")
	color.White("This wizard will configure WireGuard VPN server, routing,")
	color.White("and firewall NAT rules automatically on your machine.\n")

	// 1. Detect and prompt for Public IP
	color.Yellow("Detecting public IP address...")
	detectedIP, err := sysutil.GetPublicIP()
	if err != nil {
		detectedIP = ""
	}
	publicIPPrompt := "Enter server Public IP or Domain"
	if detectedIP != "" {
		publicIPPrompt = fmt.Sprintf("Enter server Public IP or Domain [%s]", color.GreenString(detectedIP))
	}
	fmt.Printf("%s: ", publicIPPrompt)
	inputIP, _ := reader.ReadString('\n')
	inputIP = strings.TrimSpace(inputIP)
	if inputIP == "" {
		inputIP = detectedIP
	}
	if inputIP == "" {
		return fmt.Errorf("public IP or domain is required for clients to connect")
	}

	// 2. Detect and prompt for Network Interface
	detectedIf, _ := sysutil.GetDefaultInterface()
	if detectedIf == "" {
		detectedIf = "eth0"
	}
	fmt.Printf("Outbound Network Interface for NAT [%s]: ", color.GreenString(detectedIf))
	inputIf, _ := reader.ReadString('\n')
	inputIf = strings.TrimSpace(inputIf)
	if inputIf == "" {
		inputIf = detectedIf
	}

	// 3. Port
	defaultPort := 51820
	fmt.Printf("WireGuard UDP Listen Port [%s]: ", color.GreenString("%d", defaultPort))
	inputPortStr, _ := reader.ReadString('\n')
	inputPortStr = strings.TrimSpace(inputPortStr)
	port := defaultPort
	if inputPortStr != "" {
		if p, err := strconv.Atoi(inputPortStr); err == nil && p > 0 && p < 65536 {
			port = p
		}
	}

	// 4. Subnet
	defaultSubnet := "10.8.0.1/24"
	fmt.Printf("VPN IPv4 Subnet [%s]: ", color.GreenString(defaultSubnet))
	inputSubnet, _ := reader.ReadString('\n')
	inputSubnet = strings.TrimSpace(inputSubnet)
	if inputSubnet == "" {
		inputSubnet = defaultSubnet
	}

	// 5. DNS
	defaultDNS := "1.1.1.1, 8.8.8.8"
	fmt.Printf("DNS Resolvers for Clients [%s]: ", color.GreenString(defaultDNS))
	inputDNS, _ := reader.ReadString('\n')
	inputDNS = strings.TrimSpace(inputDNS)
	if inputDNS == "" {
		inputDNS = defaultDNS
	}

	fmt.Println()
	color.Cyan("Applying configuration...")

	opts := InitOptions{
		ConfigPath: s.ConfigPath,
		Interface:  s.Interface,
		Port:       port,
		ServerIPv4: inputSubnet,
		PublicIP:   inputIP,
		DNS:        inputDNS,
		OutIface:   inputIf,
	}

	if err := s.Init(opts); err != nil {
		return err
	}

	color.Green("🎉 WireGuard Server successfully initialized and running!")
	color.White("Server Endpoint: %s:%d", inputIP, port)
	color.White("Subnet:          %s", inputSubnet)
	color.White("Config:          %s", s.ConfigPath)
	fmt.Println()
	return nil
}

// RunInteractiveMenu presents an interactive management console for the WireGuard server.
func (s *ServerManager) RunInteractiveMenu() error {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println()
		color.Cyan("════════════════════════════════════════════════════════════")
		color.Cyan("                 ezwg Server Control Panel                  ")
		color.Cyan("════════════════════════════════════════════════════════════")

		peers, _ := s.ListPeers()
		statusText := color.GreenString("● Running")
		if !s.tm.IsInterfaceUp(s.Interface) {
			statusText = color.RedString("○ Stopped")
		}

		fmt.Printf("Interface: %s (%s)  |  Registered Clients: %d\n\n",
			color.WhiteString(s.Interface), statusText, len(peers))

		fmt.Println("  1) ➕ Add a New Client")
		fmt.Println("  2) 📋 List Clients & Realtime Traffic")
		fmt.Println("  3) 📱 Show Client QR Code & Configuration")
		fmt.Println("  4) ❌ Revoke / Delete a Client")
		fmt.Println("  5) 🔄 Restart WireGuard Server")
		fmt.Println("  6) 🛑 Stop WireGuard Server")
		fmt.Println("  7) 🚪 Exit")
		fmt.Println("────────────────────────────────────────────────────────────")
		fmt.Print("Select an option [1-7]: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("\nEnter client name (e.g. laptop-andi): ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)
			if name == "" {
				color.Yellow("Cancelled: Client name cannot be empty.")
				continue
			}

			color.Cyan("Generating keys and client configuration...")
			res, err := s.AddPeer(name)
			if err != nil {
				color.Red("❌ Error: %v", err)
				continue
			}

			color.Green("\n✅ Client '%s' successfully created!", res.Name)
			color.White("Assigned VPN IP: %s", res.ClientIPv4)
			color.White("Configuration:   %s", res.ConfigFile)
			fmt.Println()

			if res.QRCodeAscii != "" {
				color.Cyan("Scan with mobile WireGuard app:")
				fmt.Println(res.QRCodeAscii)
			}

		case "2":
			fmt.Println()
			peers, err := s.ListPeers()
			if err != nil {
				color.Red("❌ Error: %v", err)
				continue
			}
			if len(peers) == 0 {
				color.Yellow("No registered clients yet.")
				continue
			}

			color.White("STATUS     NAME             ALLOWED IP    LAST HANDSHAKE   TRANSFER (RX / TX)")
			color.White("-----------------------------------------------------------------------------")
			for _, p := range peers {
				st := color.GreenString("● Online")
				if !p.IsActive {
					st = color.YellowString("○ Idle  ")
				}
				name := p.Name
				if len(name) > 16 {
					name = name[:16]
				}
				fmt.Printf("%s   %-16s %-13s %-16s %s / %s\n",
					st, name, p.AllowedIPs, p.LatestHandshake, p.TransferRx, p.TransferTx)
			}

		case "3":
			peers, _ := s.ListPeers()
			if len(peers) == 0 {
				color.Yellow("No clients available to display.")
				continue
			}
			fmt.Println("\nAvailable Clients:")
			for i, p := range peers {
				fmt.Printf("  %d) %s (%s)\n", i+1, p.Name, p.AllowedIPs)
			}
			fmt.Print("Select client number or type name: ")
			sel, _ := reader.ReadString('\n')
			sel = strings.TrimSpace(sel)

			targetName := sel
			if idx, err := strconv.Atoi(sel); err == nil && idx >= 1 && idx <= len(peers) {
				targetName = peers[idx-1].Name
			}

			res, err := s.GetClientConfig(targetName)
			if err != nil {
				color.Red("❌ Error: %v", err)
				continue
			}
			fmt.Println()
			if res.QRCodeAscii != "" {
				fmt.Println(res.QRCodeAscii)
			}
			color.Cyan("Config File: %s\n", res.ConfigFile)
			fmt.Println(res.ConfigText)

		case "4":
			peers, _ := s.ListPeers()
			if len(peers) == 0 {
				color.Yellow("No clients to remove.")
				continue
			}
			fmt.Println("\nSelect client to revoke:")
			for i, p := range peers {
				fmt.Printf("  %d) %s (%s)\n", i+1, p.Name, p.AllowedIPs)
			}
			fmt.Print("Select client number or name: ")
			sel, _ := reader.ReadString('\n')
			sel = strings.TrimSpace(sel)

			target := sel
			if idx, err := strconv.Atoi(sel); err == nil && idx >= 1 && idx <= len(peers) {
				target = peers[idx-1].Name
			}

			fmt.Printf("Are you sure you want to revoke '%s'? [y/N]: ", target)
			confirm, _ := reader.ReadString('\n')
			if strings.ToLower(strings.TrimSpace(confirm)) == "y" {
				if err := s.RemovePeer(target); err != nil {
					color.Red("❌ Error removing peer: %v", err)
				} else {
					color.Green("✅ Client '%s' successfully revoked.", target)
				}
			}

		case "5":
			color.Cyan("Restarting WireGuard server...")
			_ = s.tm.Down(s.ConfigPath)
			if err := s.tm.Up(s.ConfigPath); err != nil {
				color.Red("❌ Failed to start server: %v", err)
			} else {
				color.Green("✅ Server restarted.")
			}

		case "6":
			color.Cyan("Stopping WireGuard server...")
			if err := s.tm.Down(s.ConfigPath); err != nil {
				color.Red("❌ Error stopping server: %v", err)
			} else {
				color.Yellow("Server stopped.")
			}

		case "7", "exit", "q":
			color.White("Goodbye!")
			return nil

		default:
			color.Yellow("Invalid option. Please enter 1-7.")
		}
	}
}
