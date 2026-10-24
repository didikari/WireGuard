package wireguard

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// TunnelManager manages bringing WireGuard interfaces up and down.
type TunnelManager struct{}

// NewTunnelManager creates a new TunnelManager.
func NewTunnelManager() *TunnelManager {
	return &TunnelManager{}
}

// IsWgToolsInstalled checks if `wg` and `wg-quick` binaries are present in PATH.
func (m *TunnelManager) IsWgToolsInstalled() (bool, bool) {
	_, errWg := exec.LookPath("wg")
	_, errWgQuick := exec.LookPath("wg-quick")
	return errWg == nil, errWgQuick == nil
}

// Up brings up a WireGuard interface using wg-quick.
// `target` can be an interface name (e.g. "wg0") or path to a config file ("/etc/wireguard/wg0.conf").
func (m *TunnelManager) Up(target string) error {
	cmd := exec.Command("wg-quick", "up", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg-quick up failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// Down brings down a WireGuard interface using wg-quick.
func (m *TunnelManager) Down(target string) error {
	cmd := exec.Command("wg-quick", "down", target)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg-quick down failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// SyncConfig dynamically updates the runtime config of an active interface without interrupting other peers.
// Uses `wg syncconf <iface> <(wg-quick strip <configFile>)`.
func (m *TunnelManager) SyncConfig(iface, configFile string) error {
	// First, strip wg-quick specific directives (Address, DNS, MTU, PostUp, etc.) using wg-quick strip
	stripCmd := exec.Command("wg-quick", "strip", configFile)
	strippedConfig, err := stripCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to strip config for sync: %w", err)
	}

	// Now pipe into `wg syncconf <iface> /dev/stdin`
	syncCmd := exec.Command("wg", "syncconf", iface, "/dev/stdin")
	syncCmd.Stdin = strings.NewReader(string(strippedConfig))
	out, err := syncCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("wg syncconf failed: %s (%w)", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// IsInterfaceUp checks if the interface is currently active.
func (m *TunnelManager) IsInterfaceUp(iface string) bool {
	if runtime.GOOS == "darwin" {
		stateDirs := []string{
			"/var/run/wireguard",
			"/opt/homebrew/var/run/wireguard",
			"/usr/local/var/run/wireguard",
		}
		for _, dir := range stateDirs {
			targetFile := filepath.Join(dir, fmt.Sprintf("%s.name", iface))
			nameStat, errStat := os.Stat(targetFile)
			if errStat != nil {
				continue
			}

			// Try reading real interface name
			content, err := os.ReadFile(targetFile)
			if err == nil {
				realIface := strings.TrimSpace(string(content))
				if realIface != "" {
					sockFile := filepath.Join(dir, fmt.Sprintf("%s.sock", realIface))
					if _, statErr := os.Stat(sockFile); statErr == nil {
						return true
					}
				}
			}

			// Fallback: match timestamp with any .sock in directory
			entries, readErr := os.ReadDir(dir)
			if readErr == nil {
				for _, entry := range entries {
					if strings.HasSuffix(entry.Name(), ".sock") {
						if sockInfo, errInfo := entry.Info(); errInfo == nil {
							diff := nameStat.ModTime().Sub(sockInfo.ModTime())
							if diff < 10*time.Second && diff > -10*time.Second {
								return true
							}
						}
					}
				}
			}
		}
		return false
	}

	ifaces, err := GetAllInterfaces()
	if err != nil {
		return false
	}
	for _, item := range ifaces {
		if strings.EqualFold(item, iface) {
			return true
		}
	}
	return false
}

// Platform returns the current OS (e.g. "darwin", "linux").
func (m *TunnelManager) Platform() string {
	return runtime.GOOS
}
