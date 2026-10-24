package wireguard

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// InterfaceConfig holds the [Interface] section of a WireGuard config.
type InterfaceConfig struct {
	PrivateKey string   `json:"private_key"`
	Address    []string `json:"address"`
	ListenPort int      `json:"listen_port,omitempty"`
	DNS        []string `json:"dns,omitempty"`
	MTU        int      `json:"mtu,omitempty"`
	PostUp     []string `json:"post_up,omitempty"`
	PostDown   []string `json:"post_down,omitempty"`
	SaveConfig bool     `json:"save_config,omitempty"`
}

// PeerConfig holds a [Peer] section of a WireGuard config.
type PeerConfig struct {
	Name                string   `json:"name,omitempty"` // Extracted from preceding # Name comment
	PublicKey           string   `json:"public_key"`
	PresharedKey        string   `json:"preshared_key,omitempty"`
	AllowedIPs          []string `json:"allowed_ips"`
	Endpoint            string   `json:"endpoint,omitempty"`
	PersistentKeepalive int      `json:"persistent_keepalive,omitempty"`
}

// Config represents a complete WireGuard configuration file.
type Config struct {
	Interface InterfaceConfig `json:"interface"`
	Peers     []PeerConfig    `json:"peers"`
}

// Serialize converts the Config struct into standard WireGuard INI format.
func (c *Config) Serialize() string {
	var sb strings.Builder

	// [Interface]
	sb.WriteString("[Interface]\n")
	if c.Interface.PrivateKey != "" {
		sb.WriteString(fmt.Sprintf("PrivateKey = %s\n", c.Interface.PrivateKey))
	}
	if len(c.Interface.Address) > 0 {
		sb.WriteString(fmt.Sprintf("Address = %s\n", strings.Join(c.Interface.Address, ", ")))
	}
	if c.Interface.ListenPort > 0 {
		sb.WriteString(fmt.Sprintf("ListenPort = %d\n", c.Interface.ListenPort))
	}
	if len(c.Interface.DNS) > 0 {
		sb.WriteString(fmt.Sprintf("DNS = %s\n", strings.Join(c.Interface.DNS, ", ")))
	}
	if c.Interface.MTU > 0 {
		sb.WriteString(fmt.Sprintf("MTU = %d\n", c.Interface.MTU))
	}
	for _, pu := range c.Interface.PostUp {
		if strings.TrimSpace(pu) != "" {
			sb.WriteString(fmt.Sprintf("PostUp = %s\n", pu))
		}
	}
	for _, pd := range c.Interface.PostDown {
		if strings.TrimSpace(pd) != "" {
			sb.WriteString(fmt.Sprintf("PostDown = %s\n", pd))
		}
	}
	if c.Interface.SaveConfig {
		sb.WriteString("SaveConfig = true\n")
	}

	// [Peer] sections
	for _, peer := range c.Peers {
		sb.WriteString("\n[Peer]\n")
		if peer.Name != "" {
			sb.WriteString(fmt.Sprintf("# Client: %s\n", peer.Name))
		}
		if peer.PublicKey != "" {
			sb.WriteString(fmt.Sprintf("PublicKey = %s\n", peer.PublicKey))
		}
		if peer.PresharedKey != "" {
			sb.WriteString(fmt.Sprintf("PresharedKey = %s\n", peer.PresharedKey))
		}
		if len(peer.AllowedIPs) > 0 {
			sb.WriteString(fmt.Sprintf("AllowedIPs = %s\n", strings.Join(peer.AllowedIPs, ", ")))
		}
		if peer.Endpoint != "" {
			sb.WriteString(fmt.Sprintf("Endpoint = %s\n", peer.Endpoint))
		}
		if peer.PersistentKeepalive > 0 {
			sb.WriteString(fmt.Sprintf("PersistentKeepalive = %d\n", peer.PersistentKeepalive))
		}
	}

	return sb.String()
}

// WriteToFile saves the serialized config to a file path with 0600 permissions.
func (c *Config) WriteToFile(filePath string) error {
	content := c.Serialize()
	return os.WriteFile(filePath, []byte(content), 0600)
}

// ParseConfig parses a WireGuard configuration from an io.Reader.
func ParseConfig(r io.Reader) (*Config, error) {
	scanner := bufio.NewScanner(r)
	cfg := &Config{
		Peers: make([]PeerConfig, 0),
	}

	currentSection := ""
	var currentPeer *PeerConfig
	pendingComment := ""

	for scanner.Scan() {
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)

		// Empty line
		if line == "" {
			continue
		}

		// Comment line
		if strings.HasPrefix(line, "#") {
			commentBody := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			var name string
			if strings.HasPrefix(commentBody, "Client:") {
				name = strings.TrimSpace(strings.TrimPrefix(commentBody, "Client:"))
			} else if strings.HasPrefix(commentBody, "Name:") {
				name = strings.TrimSpace(strings.TrimPrefix(commentBody, "Name:"))
			} else {
				name = commentBody
			}

			if currentPeer != nil && currentPeer.Name == "" {
				currentPeer.Name = name
			} else if pendingComment == "" {
				pendingComment = name
			}
			continue
		}

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			sectionName := strings.ToLower(strings.Trim(line, "[] \t"))
			if currentPeer != nil {
				cfg.Peers = append(cfg.Peers, *currentPeer)
				currentPeer = nil
			}

			if sectionName == "interface" {
				currentSection = "interface"
			} else if sectionName == "peer" {
				currentSection = "peer"
				currentPeer = &PeerConfig{
					Name: pendingComment,
				}
				pendingComment = ""
			} else {
				currentSection = ""
			}
			continue
		}

		// Key = Value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.TrimSpace(parts[1])

		if currentSection == "interface" {
			switch key {
			case "privatekey":
				cfg.Interface.PrivateKey = val
			case "address":
				cfg.Interface.Address = splitAndTrim(val, ",")
			case "listenport":
				var port int
				fmt.Sscanf(val, "%d", &port)
				cfg.Interface.ListenPort = port
			case "dns":
				cfg.Interface.DNS = splitAndTrim(val, ",")
			case "mtu":
				var mtu int
				fmt.Sscanf(val, "%d", &mtu)
				cfg.Interface.MTU = mtu
			case "postup":
				cfg.Interface.PostUp = append(cfg.Interface.PostUp, val)
			case "postdown":
				cfg.Interface.PostDown = append(cfg.Interface.PostDown, val)
			case "saveconfig":
				cfg.Interface.SaveConfig = strings.ToLower(val) == "true"
			}
		} else if currentSection == "peer" && currentPeer != nil {
			switch key {
			case "publickey":
				currentPeer.PublicKey = val
			case "presharedkey":
				currentPeer.PresharedKey = val
			case "allowedips":
				currentPeer.AllowedIPs = splitAndTrim(val, ",")
			case "endpoint":
				currentPeer.Endpoint = val
			case "persistentkeepalive":
				var keepalive int
				fmt.Sscanf(val, "%d", &keepalive)
				currentPeer.PersistentKeepalive = keepalive
			}
		}
	}

	if currentPeer != nil {
		cfg.Peers = append(cfg.Peers, *currentPeer)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config: %w", err)
	}

	return cfg, nil
}

// ParseConfigFile reads and parses a WireGuard config file.
func ParseConfigFile(filePath string) (*Config, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()
	return ParseConfig(f)
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, part := range strings.Split(s, sep) {
		t := strings.TrimSpace(part)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
