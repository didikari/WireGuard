package server

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ezwg/ezwg/internal/sysutil"
	"github.com/ezwg/ezwg/internal/wireguard"
)

// InitOptions contains settings for initializing a WireGuard server.
type InitOptions struct {
	ConfigPath string // e.g. "/etc/wireguard/wg0.conf"
	Interface  string // e.g. "wg0"
	Port       int    // e.g. 51820
	ServerIPv4 string // e.g. "10.8.0.1/24"
	PublicIP   string // e.g. "203.0.113.1" or domain
	DNS        string // e.g. "1.1.1.1, 8.8.8.8"
	OutIface   string // e.g. "eth0"
}

// PeerResult represents a newly created client peer.
type PeerResult struct {
	Name         string
	ClientIPv4   string
	ConfigFile   string
	ConfigText   string
	QRCodeAscii  string
	PublicKey    string
	PresharedKey string
}

// PeerSummary provides a detailed view of a registered peer.
type PeerSummary struct {
	Name            string
	PublicKey       string
	AllowedIPs      string
	Endpoint        string
	LatestHandshake string
	TransferRx      string
	TransferTx      string
	IsActive        bool
}

// ServerManager handles server-side WireGuard operations.
type ServerManager struct {
	ConfigPath string
	Interface  string
	ClientsDir string
	tm         *wireguard.TunnelManager
}

// NewServerManager creates an instance of ServerManager.
func NewServerManager(configPath string) *ServerManager {
	if configPath == "" {
		configPath = "/etc/wireguard/wg0.conf"
	}
	base := filepath.Base(configPath)
	iface := strings.TrimSuffix(base, filepath.Ext(base))
	clientsDir := filepath.Join(filepath.Dir(configPath), "clients")

	return &ServerManager{
		ConfigPath: configPath,
		Interface:  iface,
		ClientsDir: clientsDir,
		tm:         wireguard.NewTunnelManager(),
	}
}

// IsInitialized checks if the server config file already exists.
func (s *ServerManager) IsInitialized() bool {
	info, err := os.Stat(s.ConfigPath)
	return err == nil && !info.IsDir()
}

// Init initializes the WireGuard server configuration, enables routing, and starts the service.
func (s *ServerManager) Init(opts InitOptions) error {
	if s.IsInitialized() {
		return fmt.Errorf("server is already initialized at %s", s.ConfigPath)
	}

	if err := sysutil.EnsureRoot(); err != nil {
		return err
	}

	// Default fallback values
	if opts.Port <= 0 {
		opts.Port = 51820
	}
	if opts.ServerIPv4 == "" {
		opts.ServerIPv4 = "10.8.0.1/24"
	}
	if opts.DNS == "" {
		opts.DNS = "1.1.1.1, 8.8.8.8"
	}
	if opts.PublicIP == "" {
		pubIP, err := sysutil.GetPublicIP()
		if err != nil {
			return fmt.Errorf("could not auto-detect public IP. Please specify --public-ip: %w", err)
		}
		opts.PublicIP = pubIP
	}
	if opts.OutIface == "" {
		outIf, err := sysutil.GetDefaultInterface()
		if err != nil {
			opts.OutIface = "eth0"
		} else {
			opts.OutIface = outIf
		}
	}

	// Generate Server Key Pair
	serverKeys, err := wireguard.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("failed to generate server keys: %w", err)
	}

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(s.ConfigPath), 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(s.ConfigPath), err)
	}
	if err := os.MkdirAll(s.ClientsDir, 0700); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", s.ClientsDir, err)
	}

	// Formulate firewall rules
	postUp := []string{}
	postDown := []string{}

	if runtime.GOOS == "linux" {
		postUp = append(postUp,
			fmt.Sprintf("iptables -A FORWARD -i %s -j ACCEPT", s.Interface),
			fmt.Sprintf("iptables -A FORWARD -o %s -j ACCEPT", s.Interface),
			fmt.Sprintf("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", opts.OutIface),
		)
		postDown = append(postDown,
			fmt.Sprintf("iptables -D FORWARD -i %s -j ACCEPT", s.Interface),
			fmt.Sprintf("iptables -D FORWARD -o %s -j ACCEPT", s.Interface),
			fmt.Sprintf("iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", opts.OutIface),
		)
	}

	// Enable kernel IP forwarding
	if err := sysutil.EnableIPForwarding(); err != nil {
		fmt.Printf("Warning: failed to enable IP forwarding automatically: %v\n", err)
	}

	dnsList := []string{}
	for _, d := range strings.Split(opts.DNS, ",") {
		if t := strings.TrimSpace(d); t != "" {
			dnsList = append(dnsList, t)
		}
	}

	cfg := &wireguard.Config{
		Interface: wireguard.InterfaceConfig{
			PrivateKey: serverKeys.PrivateKey,
			Address:    []string{opts.ServerIPv4},
			ListenPort: opts.Port,
			DNS:        dnsList,
			PostUp:     postUp,
			PostDown:   postDown,
			SaveConfig: false,
		},
		Peers: []wireguard.PeerConfig{},
	}

	if err := cfg.WriteToFile(s.ConfigPath); err != nil {
		return fmt.Errorf("failed to save server config to %s: %w", s.ConfigPath, err)
	}

	// Save server metadata (public key and endpoint)
	metaPath := filepath.Join(filepath.Dir(s.ConfigPath), ".server.env")
	metaContent := fmt.Sprintf("SERVER_PUBKEY=%s\nSERVER_ENDPOINT=%s:%d\nSERVER_DNS=%s\n",
		serverKeys.PublicKey, opts.PublicIP, opts.Port, opts.DNS)
	_ = os.WriteFile(metaPath, []byte(metaContent), 0600)

	// Start server service
	if runtime.GOOS == "linux" {
		// Use systemd if available
		if _, err := exec.LookPath("systemctl"); err == nil {
			_ = exec.Command("systemctl", "enable", fmt.Sprintf("wg-quick@%s", s.Interface)).Run()
			cmd := exec.Command("systemctl", "restart", fmt.Sprintf("wg-quick@%s", s.Interface))
			if out, err := cmd.CombinedOutput(); err != nil {
				return fmt.Errorf("failed to start systemd service wg-quick@%s: %s (%w)", s.Interface, strings.TrimSpace(string(out)), err)
			}
			return nil
		}
	}

	// Fallback to wg-quick up
	if err := s.tm.Up(s.ConfigPath); err != nil {
		return fmt.Errorf("server config written, but wg-quick up failed: %w", err)
	}

	return nil
}

// AddPeer adds a new client peer to the server configuration.
func (s *ServerManager) AddPeer(name string) (*PeerResult, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server is not initialized yet. Run 'ezwg server init' first")
	}

	if err := sysutil.EnsureRoot(); err != nil {
		return nil, err
	}

	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("client name cannot be empty")
	}

	cfg, err := wireguard.ParseConfigFile(s.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server config: %w", err)
	}

	// Check if peer name already exists
	for _, p := range cfg.Peers {
		if strings.EqualFold(p.Name, name) {
			return nil, fmt.Errorf("a client named '%s' already exists", name)
		}
	}

	// Allocate next available IP in server subnet
	clientIP, err := allocateNextIP(cfg)
	if err != nil {
		return nil, fmt.Errorf("ip allocation failed: %w", err)
	}

	// Generate Client Keys and PSK
	clientKeys, err := wireguard.GenerateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate client keypair: %w", err)
	}

	psk, err := wireguard.GeneratePresharedKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate preshared key: %w", err)
	}

	// Derive Server Public Key
	serverPubKey, err := wireguard.DerivePublicKey(cfg.Interface.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive server public key: %w", err)
	}

	// Append peer to server config
	peerCfg := wireguard.PeerConfig{
		Name:         name,
		PublicKey:    clientKeys.PublicKey,
		PresharedKey: psk,
		AllowedIPs:   []string{fmt.Sprintf("%s/32", clientIP.String())},
	}
	cfg.Peers = append(cfg.Peers, peerCfg)

	if err := cfg.WriteToFile(s.ConfigPath); err != nil {
		return nil, fmt.Errorf("failed to write updated server config: %w", err)
	}

	// Hot reload WireGuard server if active
	if s.tm.IsInterfaceUp(s.Interface) {
		_ = s.tm.SyncConfig(s.Interface, s.ConfigPath)
	}

	// Determine server endpoint and DNS
	serverEndpoint := s.getStoredEndpoint(cfg.Interface.ListenPort)
	dnsServers := cfg.Interface.DNS
	if len(dnsServers) == 0 {
		dnsServers = []string{"1.1.1.1", "8.8.8.8"}
	}

	// Create client config
	clientConfig := &wireguard.Config{
		Interface: wireguard.InterfaceConfig{
			PrivateKey: clientKeys.PrivateKey,
			Address:    []string{fmt.Sprintf("%s/32", clientIP.String())},
			DNS:        dnsServers,
		},
		Peers: []wireguard.PeerConfig{
			{
				PublicKey:           serverPubKey,
				PresharedKey:        psk,
				Endpoint:            serverEndpoint,
				AllowedIPs:          []string{"0.0.0.0/0", "::/0"},
				PersistentKeepalive: 25,
			},
		},
	}

	clientConfText := clientConfig.Serialize()
	clientConfPath := filepath.Join(s.ClientsDir, fmt.Sprintf("%s.conf", name))
	_ = os.MkdirAll(s.ClientsDir, 0700)
	if err := clientConfig.WriteToFile(clientConfPath); err != nil {
		return nil, fmt.Errorf("failed to write client config file: %w", err)
	}

	// Generate QR code ASCII
	qrAscii, _ := wireguard.GenerateTerminalQRCode(clientConfText)

	return &PeerResult{
		Name:         name,
		ClientIPv4:   clientIP.String(),
		ConfigFile:   clientConfPath,
		ConfigText:   clientConfText,
		QRCodeAscii:  qrAscii,
		PublicKey:    clientKeys.PublicKey,
		PresharedKey: psk,
	}, nil
}

// RemovePeer removes a client peer by name or public key.
func (s *ServerManager) RemovePeer(nameOrPubKey string) error {
	if !s.IsInitialized() {
		return fmt.Errorf("server is not initialized")
	}
	if err := sysutil.EnsureRoot(); err != nil {
		return err
	}

	cfg, err := wireguard.ParseConfigFile(s.ConfigPath)
	if err != nil {
		return fmt.Errorf("failed to parse server config: %w", err)
	}

	foundIdx := -1
	var targetPubKey string
	var targetName string

	for i, p := range cfg.Peers {
		if strings.EqualFold(p.Name, nameOrPubKey) || p.PublicKey == nameOrPubKey {
			foundIdx = i
			targetPubKey = p.PublicKey
			targetName = p.Name
			break
		}
	}

	if foundIdx == -1 {
		return fmt.Errorf("peer '%s' not found", nameOrPubKey)
	}

	// Remove from slice
	cfg.Peers = append(cfg.Peers[:foundIdx], cfg.Peers[foundIdx+1:]...)

	if err := cfg.WriteToFile(s.ConfigPath); err != nil {
		return fmt.Errorf("failed to save updated config: %w", err)
	}

	// Dynamically remove from live interface
	if s.tm.IsInterfaceUp(s.Interface) {
		cmd := exec.Command("wg", "set", s.Interface, "peer", targetPubKey, "remove")
		_ = cmd.Run()
	}

	// Delete client conf if present
	if targetName != "" {
		confPath := filepath.Join(s.ClientsDir, fmt.Sprintf("%s.conf", targetName))
		_ = os.Remove(confPath)
	}

	return nil
}

// ListPeers lists all registered peers with their runtime stats if the interface is active.
func (s *ServerManager) ListPeers() ([]PeerSummary, error) {
	if !s.IsInitialized() {
		return nil, fmt.Errorf("server is not initialized")
	}

	cfg, err := wireguard.ParseConfigFile(s.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var liveStatus *wireguard.InterfaceStatus
	if s.tm.IsInterfaceUp(s.Interface) {
		liveStatus, _ = wireguard.GetInterfaceStatus(s.Interface)
	}

	summaries := make([]PeerSummary, 0, len(cfg.Peers))

	for _, p := range cfg.Peers {
		summary := PeerSummary{
			Name:            p.Name,
			PublicKey:       p.PublicKey,
			AllowedIPs:      strings.Join(p.AllowedIPs, ", "),
			Endpoint:        "none",
			LatestHandshake: "never",
			TransferRx:      "0 B",
			TransferTx:      "0 B",
			IsActive:        false,
		}

		if liveStatus != nil {
			if stat, ok := liveStatus.Peers[p.PublicKey]; ok {
				if stat.Endpoint != "" {
					summary.Endpoint = stat.Endpoint
				}
				if !stat.LatestHandshake.IsZero() {
					diff := time.Since(stat.LatestHandshake)
					if diff < 3*time.Minute {
						summary.IsActive = true
					}
					summary.LatestHandshake = formatTimeAgo(diff)
				}
				summary.TransferRx = wireguard.FormatBytes(stat.TransferRx)
				summary.TransferTx = wireguard.FormatBytes(stat.TransferTx)
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// GetClientConfig returns the raw config and QR code for an existing client.
func (s *ServerManager) GetClientConfig(name string) (*PeerResult, error) {
	confPath := filepath.Join(s.ClientsDir, fmt.Sprintf("%s.conf", name))
	content, err := os.ReadFile(confPath)
	if err != nil {
		return nil, fmt.Errorf("could not find client file for '%s': %w", name, err)
	}

	text := string(content)
	qr, _ := wireguard.GenerateTerminalQRCode(text)

	return &PeerResult{
		Name:        name,
		ConfigFile:  confPath,
		ConfigText:  text,
		QRCodeAscii: qr,
	}, nil
}

func (s *ServerManager) getStoredEndpoint(port int) string {
	metaPath := filepath.Join(filepath.Dir(s.ConfigPath), ".server.env")
	if data, err := os.ReadFile(metaPath); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "SERVER_ENDPOINT=") {
				return strings.TrimPrefix(line, "SERVER_ENDPOINT=")
			}
		}
	}

	pubIP, err := sysutil.GetPublicIP()
	if err != nil {
		pubIP = "YOUR_SERVER_IP"
	}
	return fmt.Sprintf("%s:%d", pubIP, port)
}

func allocateNextIP(cfg *wireguard.Config) (net.IP, error) {
	if len(cfg.Interface.Address) == 0 {
		return nil, fmt.Errorf("server interface has no address configured")
	}

	serverIP, serverNet, err := net.ParseCIDR(cfg.Interface.Address[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse server address CIDR: %w", err)
	}

	usedIPs := make(map[string]bool)
	usedIPs[serverIP.String()] = true

	for _, p := range cfg.Peers {
		for _, allowed := range p.AllowedIPs {
			ip, _, err := net.ParseCIDR(allowed)
			if err == nil {
				usedIPs[ip.String()] = true
			}
		}
	}

	// Iterate through the subnet to find the lowest available host IP
	curIP := make(net.IP, len(serverIP))
	copy(curIP, serverIP.To4())

	// Start searching from host 2 up to host 254 in /24
	for host := 2; host < 255; host++ {
		candidate := net.IPv4(curIP[0], curIP[1], curIP[2], byte(host))
		if serverNet.Contains(candidate) && !usedIPs[candidate.String()] {
			return candidate, nil
		}
	}

	return nil, fmt.Errorf("no available IP addresses remaining in subnet %s", serverNet.String())
}

func formatTimeAgo(d time.Duration) string {
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds ago", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm ago", mins)
	}
	hrs := mins / 60
	if hrs < 24 {
		return fmt.Sprintf("%dh ago", hrs)
	}
	days := hrs / 24
	return fmt.Sprintf("%dd ago", days)
}
