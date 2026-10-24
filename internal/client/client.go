package client

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ezwg/ezwg/internal/sysutil"
	"github.com/ezwg/ezwg/internal/wireguard"
)

// ProfileInfo holds basic metadata for a client profile.
type ProfileInfo struct {
	Name       string
	Path       string
	IsActive   bool
	ServerIP   string
	ClientIP   string
	Endpoint   string
	ModifiedAt time.Time
}

// ClientManager handles client-side profiles and tunnels.
type ClientManager struct {
	ProfilesDir string
	tm          *wireguard.TunnelManager
}

// NewClientManager creates a ClientManager with profile directory setup.
func NewClientManager() (*ClientManager, error) {
	baseDir, err := sysutil.GetUserConfigDir()
	if err != nil {
		return nil, err
	}

	profilesDir := filepath.Join(baseDir, "profiles")
	if err := os.MkdirAll(profilesDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create profiles directory: %w", err)
	}

	return &ClientManager{
		ProfilesDir: profilesDir,
		tm:          wireguard.NewTunnelManager(),
	}, nil
}

// ImportProfile imports a WireGuard config from a file path or URL.
func (c *ClientManager) ImportProfile(name, source string) (string, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		return "", fmt.Errorf("profile name cannot be empty")
	}

	var content []byte
	var err error

	if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
		// Download from URL
		resp, errGet := http.Get(source)
		if errGet != nil {
			return "", fmt.Errorf("failed to download profile from URL: %w", errGet)
		}
		defer resp.Body.Close()
		content, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}
	} else if strings.Contains(source, "[Interface]") {
		// Raw config content passed directly
		content = []byte(source)
	} else {
		// Local file path
		content, err = os.ReadFile(source)
		if err != nil {
			return "", fmt.Errorf("failed to read profile file '%s': %w", source, err)
		}
	}

	// Validate config syntax
	cfg, err := wireguard.ParseConfig(strings.NewReader(string(content)))
	if err != nil {
		return "", fmt.Errorf("invalid wireguard configuration: %w", err)
	}

	if cfg.Interface.PrivateKey == "" {
		return "", fmt.Errorf("invalid configuration: missing Interface.PrivateKey")
	}

	destPath := filepath.Join(c.ProfilesDir, fmt.Sprintf("%s.conf", name))
	if err := os.WriteFile(destPath, content, 0600); err != nil {
		return "", fmt.Errorf("failed to save profile to %s: %w", destPath, err)
	}

	return destPath, nil
}

// Connect activates a profile using wg-quick.
func (c *ClientManager) Connect(name string) error {
	profilePath, resolvedName, err := c.resolveProfile(name)
	if err != nil {
		return err
	}

	// Check if already active
	if c.tm.IsInterfaceUp(resolvedName) {
		return fmt.Errorf("profile '%s' is already connected and active", resolvedName)
	}

	// Bring up the tunnel
	if sysutil.IsRoot() {
		return c.tm.Up(profilePath)
	}

	// If not root, execute with sudo
	cmd := exec.Command("sudo", "wg-quick", "up", profilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to connect profile '%s' via sudo wg-quick: %w", resolvedName, err)
	}

	return nil
}

// Disconnect deactivates a profile using wg-quick.
func (c *ClientManager) Disconnect(name string) error {
	profilePath, resolvedName, err := c.resolveProfile(name)
	if err != nil {
		return err
	}

	if !c.tm.IsInterfaceUp(resolvedName) {
		return fmt.Errorf("profile '%s' is not currently active", resolvedName)
	}

	if sysutil.IsRoot() {
		return c.tm.Down(profilePath)
	}

	cmd := exec.Command("sudo", "wg-quick", "down", profilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to disconnect profile '%s': %w", resolvedName, err)
	}

	return nil
}

// ListProfiles lists all imported profiles and flags active ones.
func (c *ClientManager) ListProfiles() ([]ProfileInfo, error) {
	entries, err := os.ReadDir(c.ProfilesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read profiles dir: %w", err)
	}

	activeIfaces, _ := wireguard.GetAllInterfaces()
	activeMap := make(map[string]bool)
	for _, iface := range activeIfaces {
		activeMap[strings.ToLower(iface)] = true
	}

	profiles := make([]ProfileInfo, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".conf")
		fullPath := filepath.Join(c.ProfilesDir, entry.Name())
		info, _ := entry.Info()

		cfg, err := wireguard.ParseConfigFile(fullPath)
		clientIP := "-"
		endpoint := "-"

		if err == nil {
			if len(cfg.Interface.Address) > 0 {
				clientIP = strings.Join(cfg.Interface.Address, ", ")
			}
			if len(cfg.Peers) > 0 {
				endpoint = cfg.Peers[0].Endpoint
			}
		}

		modTime := time.Now()
		if info != nil {
			modTime = info.ModTime()
		}

		profiles = append(profiles, ProfileInfo{
			Name:       name,
			Path:       fullPath,
			IsActive:   activeMap[strings.ToLower(name)],
			ClientIP:   clientIP,
			Endpoint:   endpoint,
			ModifiedAt: modTime,
		})
	}

	return profiles, nil
}

// DeleteProfile removes a saved profile.
func (c *ClientManager) DeleteProfile(name string) error {
	profilePath, resolvedName, err := c.resolveProfile(name)
	if err != nil {
		return err
	}

	if c.tm.IsInterfaceUp(resolvedName) {
		return fmt.Errorf("cannot delete active profile '%s'. Please disconnect first", resolvedName)
	}

	return os.Remove(profilePath)
}

// Status returns runtime stats for a connected profile.
func (c *ClientManager) Status(name string) (*wireguard.InterfaceStatus, error) {
	_, resolvedName, err := c.resolveProfile(name)
	if err != nil {
		return nil, err
	}

	if !c.tm.IsInterfaceUp(resolvedName) {
		return nil, fmt.Errorf("profile '%s' is disconnected", resolvedName)
	}

	return wireguard.GetInterfaceStatus(resolvedName)
}

// GetActiveProfile returns the name of any currently active profile from the profiles dir.
func (c *ClientManager) GetActiveProfile() (string, error) {
	profiles, err := c.ListProfiles()
	if err != nil {
		return "", err
	}
	for _, p := range profiles {
		if p.IsActive {
			return p.Name, nil
		}
	}
	return "", nil
}

func (c *ClientManager) resolveProfile(name string) (string, string, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		// If name is omitted, try finding the currently active profile or sole profile
		profiles, err := c.ListProfiles()
		if err == nil {
			for _, p := range profiles {
				if p.IsActive {
					return p.Path, p.Name, nil
				}
			}
			if len(profiles) == 1 {
				return profiles[0].Path, profiles[0].Name, nil
			}
		}
		return "", "", fmt.Errorf("no profile specified, and multiple or zero profiles exist")
	}

	confPath := filepath.Join(c.ProfilesDir, fmt.Sprintf("%s.conf", name))
	if _, err := os.Stat(confPath); err != nil {
		// Also check if user passed full file path directly
		if _, errDirect := os.Stat(name); errDirect == nil {
			base := filepath.Base(name)
			clean := strings.TrimSuffix(base, filepath.Ext(base))
			return name, clean, nil
		}
		return "", "", fmt.Errorf("profile '%s' not found in %s", name, c.ProfilesDir)
	}

	return confPath, name, nil
}

func sanitizeProfileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.TrimSuffix(name, ".conf")
	// Keep alphanumeric, dash, and underscore
	var clean strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			clean.WriteRune(r)
		}
	}
	return clean.String()
}

// PingServer tests latency to server endpoint or IP.
func PingServer(target string) string {
	// Strip port if present
	host := target
	if strings.Contains(target, ":") {
		h, _, err := strings.Cut(target, ":")
		if err {
			host = h
		}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("ping", "-c", "1", "-W", "2000", host)
	} else {
		cmd = exec.Command("ping", "-c", "1", "-W", "2", host)
	}

	out, err := cmd.Output()
	if err != nil {
		return "unreachable"
	}

	// Parse avg time
	outputStr := string(out)
	if idx := strings.Index(outputStr, "avg"); idx != -1 {
		line := outputStr[idx:]
		parts := strings.Split(line, "/")
		if len(parts) >= 5 {
			return fmt.Sprintf("%s ms", parts[4])
		}
	}
	return "reachable"
}
