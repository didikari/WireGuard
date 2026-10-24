package wireguard

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PeerStatus contains runtime stats for a connected peer.
type PeerStatus struct {
	PublicKey           string    `json:"public_key"`
	PresharedKey        string    `json:"preshared_key"`
	Endpoint            string    `json:"endpoint"`
	AllowedIPs          []string  `json:"allowed_ips"`
	LatestHandshake     time.Time `json:"latest_handshake"`
	TransferRx          uint64    `json:"transfer_rx"`
	TransferTx          uint64    `json:"transfer_tx"`
	PersistentKeepalive int       `json:"persistent_keepalive"`
}

// InterfaceStatus contains runtime status for a wireguard interface.
type InterfaceStatus struct {
	Name        string                `json:"name"`
	PublicKey   string                `json:"public_key"`
	ListenPort  int                   `json:"listen_port"`
	FirewallMark string               `json:"fwmark,omitempty"`
	Peers       map[string]PeerStatus `json:"peers"` // keyed by public key
}

// GetInterfaceStatus retrieves the parsed runtime stats from `wg show <iface> dump`.
func GetInterfaceStatus(iface string) (*InterfaceStatus, error) {
	cmd := exec.Command("wg", "show", iface, "dump")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run wg show %s dump: %w", iface, err)
	}

	return ParseWgDump(iface, string(output))
}

// GetAllInterfaces lists all active WireGuard interface names.
func GetAllInterfaces() ([]string, error) {
	if runtime.GOOS == "darwin" {
		var active []string
		stateDirs := []string{
			"/var/run/wireguard",
			"/opt/homebrew/var/run/wireguard",
			"/usr/local/var/run/wireguard",
		}
		for _, dir := range stateDirs {
			entries, err := os.ReadDir(dir)
			if err == nil {
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".name") {
						name := strings.TrimSuffix(entry.Name(), ".name")
						active = append(active, name)
					}
				}
			}
		}
		if len(active) > 0 {
			return active, nil
		}
	}

	cmd := exec.Command("wg", "show", "interfaces")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query wg interfaces: %w", err)
	}
	fields := strings.Fields(string(out))
	return fields, nil
}

// ParseWgDump parses the output of `wg show <iface> dump`.
func ParseWgDump(iface, dumpOutput string) (*InterfaceStatus, error) {
	status := &InterfaceStatus{
		Name:  iface,
		Peers: make(map[string]PeerStatus),
	}

	scanner := bufio.NewScanner(strings.NewReader(dumpOutput))
	isFirstLine := true

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if isFirstLine {
			// Interface line: <private-key>\t<public-key>\t<listen-port>\t<fwmark>
			if len(parts) >= 3 {
				status.PublicKey = parts[1]
				port, _ := strconv.Atoi(parts[2])
				status.ListenPort = port
			}
			if len(parts) >= 4 {
				status.FirewallMark = parts[3]
			}
			isFirstLine = false
			continue
		}

		// Peer line:
		// 0: public-key
		// 1: preshared-key
		// 2: endpoint
		// 3: allowed-ips
		// 4: latest-handshake (epoch)
		// 5: rx-bytes
		// 6: tx-bytes
		// 7: persistent-keepalive
		if len(parts) >= 7 {
			pubKey := parts[0]
			psk := parts[1]
			endpoint := parts[2]
			allowedIPs := splitAndTrim(parts[3], ",")

			var handshakeTime time.Time
			if epoch, err := strconv.ParseInt(parts[4], 10, 64); err == nil && epoch > 0 {
				handshakeTime = time.Unix(epoch, 0)
			}

			rx, _ := strconv.ParseUint(parts[5], 10, 64)
			tx, _ := strconv.ParseUint(parts[6], 10, 64)

			var keepalive int
			if len(parts) >= 8 {
				keepalive, _ = strconv.Atoi(parts[7])
			}

			status.Peers[pubKey] = PeerStatus{
				PublicKey:           pubKey,
				PresharedKey:        psk,
				Endpoint:            endpoint,
				AllowedIPs:          allowedIPs,
				LatestHandshake:     handshakeTime,
				TransferRx:          rx,
				TransferTx:          tx,
				PersistentKeepalive: keepalive,
			}
		}
	}

	return status, nil
}

// FormatBytes formats byte counts into human-readable strings (e.g. 1.25 MiB).
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
