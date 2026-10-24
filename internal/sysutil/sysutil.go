package sysutil

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// IsRoot returns true if the current process is running with root/sudo privileges.
func IsRoot() bool {
	return os.Geteuid() == 0
}

func init() {
	if runtime.GOOS == "darwin" {
		curPath := os.Getenv("PATH")
		extraPaths := "/opt/homebrew/bin:/opt/homebrew/sbin:/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin"
		if !strings.Contains(curPath, "/opt/homebrew/bin") {
			os.Setenv("PATH", extraPaths+":"+curPath)
		}
	}
}

// EnsureRoot checks if the program is run as root. If not, returns a descriptive error.
func EnsureRoot() error {
	if !IsRoot() {
		return fmt.Errorf("this operation requires root/administrator privileges. Please run with 'sudo'")
	}
	return nil
}

// GetUserConfigDir returns the default user config directory for ezwg.
func GetUserConfigDir() (string, error) {
	if IsRoot() {
		// When running as root on a server, default to standard WireGuard directory
		return "/etc/wireguard", nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		u, errUser := user.Current()
		if errUser != nil {
			return "", fmt.Errorf("unable to determine user home directory: %w", err)
		}
		homeDir = u.HomeDir
	}

	cfgDir := filepath.Join(homeDir, ".config", "ezwg")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create config directory %s: %w", cfgDir, err)
	}
	return cfgDir, nil
}

// GetPublicIP attempts to determine the external IPv4 address by querying trusted endpoints.
func GetPublicIP() (string, error) {
	endpoints := []string{
		"https://api.ipify.org",
		"https://icanhazip.com",
		"https://ifconfig.me/ip",
		"https://checkip.amazonaws.com",
	}

	client := &http.Client{
		Timeout: 4 * time.Second,
	}

	for _, url := range endpoints {
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			cancel()
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			cancel()
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()

		if err == nil {
			ipStr := strings.TrimSpace(string(body))
			ip := net.ParseIP(ipStr)
			if ip != nil && ip.To4() != nil {
				return ipStr, nil
			}
		}
	}

	return "", fmt.Errorf("unable to determine public IP address automatically")
}

// GetDefaultInterface returns the network interface name used for outbound default routing (e.g. eth0, ens3, en0).
func GetDefaultInterface() (string, error) {
	// Dial a remote public IP to discover the local network route without sending packets
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", fmt.Errorf("failed to discover outbound interface: %w", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", fmt.Errorf("failed to list network interfaces: %w", err)
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && ip.Equal(localAddr.IP) {
				return iface.Name, nil
			}
		}
	}

	// Fallback detection
	if runtime.GOOS == "darwin" {
		return "en0", nil
	}
	return "eth0", nil
}

// EnableIPForwarding enables IPv4 forwarding in kernel sysctl.
func EnableIPForwarding() error {
	if runtime.GOOS == "linux" {
		// Set live sysctl
		cmd := exec.Command("sysctl", "-w", "net.ipv4.ip_forward=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set net.ipv4.ip_forward: %s (%w)", strings.TrimSpace(string(out)), err)
		}

		// Persist across reboot in /etc/sysctl.d/99-wireguard-forwarding.conf
		sysctlFile := "/etc/sysctl.d/99-wireguard-forwarding.conf"
		content := "net.ipv4.ip_forward = 1\nnet.ipv6.conf.all.forwarding = 1\n"
		if err := os.WriteFile(sysctlFile, []byte(content), 0644); err != nil {
			// Non-fatal if writing file fails, but warn
			return nil
		}
		return nil
	} else if runtime.GOOS == "darwin" {
		cmd := exec.Command("sysctl", "-w", "net.inet.ip.forwarding=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("failed to set net.inet.ip.forwarding: %s (%w)", strings.TrimSpace(string(out)), err)
		}
		return nil
	}
	return nil
}

// DetectPackageManager returns the primary package manager available on this system.
func DetectPackageManager() string {
	managers := []string{"brew", "apt-get", "dnf", "yum", "pacman", "zypper", "apk"}
	for _, m := range managers {
		if _, err := exec.LookPath(m); err == nil {
			return m
		}
	}
	return ""
}
