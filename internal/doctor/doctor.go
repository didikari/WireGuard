package doctor

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/ezwg/ezwg/internal/sysutil"
)

// CheckStatus indicates the health of a checked component.
type CheckStatus string

const (
	StatusOK   CheckStatus = "OK"
	StatusWarn CheckStatus = "WARN"
	StatusFail CheckStatus = "FAIL"
)

// CheckItem represents an individual diagnostic check.
type CheckItem struct {
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Details     string      `json:"details"`
	FixCommand  string      `json:"fix_command,omitempty"`
	CanAutoFix  bool        `json:"can_auto_fix"`
}

// Doctor diagnoses system prerequisites for WireGuard on Linux and macOS.
type Doctor struct{}

// NewDoctor creates a new Doctor instance.
func NewDoctor() *Doctor {
	return &Doctor{}
}

// RunDiagnostics executes all system checks and returns the diagnostic report.
func (d *Doctor) RunDiagnostics() []CheckItem {
	var items []CheckItem

	// 1. OS & Architecture Check
	items = append(items, CheckItem{
		Name:    "Operating System",
		Status:  StatusOK,
		Details: fmt.Sprintf("%s (%s)", runtime.GOOS, runtime.GOARCH),
	})

	// 2. wg CLI tool check
	if _, err := exec.LookPath("wg"); err == nil {
		out, _ := exec.Command("wg", "--version").Output()
		items = append(items, CheckItem{
			Name:    "WireGuard CLI (wg)",
			Status:  StatusOK,
			Details: strings.TrimSpace(string(out)),
		})
	} else {
		fixCmd := d.suggestInstallCommand("wireguard-tools")
		items = append(items, CheckItem{
			Name:       "WireGuard CLI (wg)",
			Status:     StatusFail,
			Details:    "Missing 'wg' binary",
			FixCommand: fixCmd,
			CanAutoFix: fixCmd != "",
		})
	}

	// 3. wg-quick tool check
	if _, err := exec.LookPath("wg-quick"); err == nil {
		items = append(items, CheckItem{
			Name:    "WireGuard Quick (wg-quick)",
			Status:  StatusOK,
			Details: "Installed and ready",
		})
	} else {
		fixCmd := d.suggestInstallCommand("wireguard-tools")
		items = append(items, CheckItem{
			Name:       "WireGuard Quick (wg-quick)",
			Status:     StatusFail,
			Details:    "Missing 'wg-quick' script (needed to bring tunnels up/down)",
			FixCommand: fixCmd,
			CanAutoFix: fixCmd != "",
		})
	}

	// 4. Platform-specific checks
	if runtime.GOOS == "linux" {
		d.checkLinux(&items)
	} else if runtime.GOOS == "darwin" {
		d.checkDarwin(&items)
	}

	return items
}

func (d *Doctor) checkLinux(items *[]CheckItem) {
	// IP forwarding check
	forwardData, err := os.ReadFile("/proc/sys/net/ipv4/ip_forward")
	if err == nil && strings.TrimSpace(string(forwardData)) == "1" {
		*items = append(*items, CheckItem{
			Name:    "IPv4 Forwarding",
			Status:  StatusOK,
			Details: "Enabled (net.ipv4.ip_forward = 1)",
		})
	} else {
		*items = append(*items, CheckItem{
			Name:       "IPv4 Forwarding",
			Status:     StatusWarn,
			Details:    "Disabled (required if running ezwg as server)",
			FixCommand: "sysctl -w net.ipv4.ip_forward=1",
			CanAutoFix: true,
		})
	}

	// Firewall / iptables check
	if _, err := exec.LookPath("iptables"); err == nil {
		*items = append(*items, CheckItem{
			Name:    "Firewall (iptables)",
			Status:  StatusOK,
			Details: "Installed",
		})
	} else if _, err := exec.LookPath("nft"); err == nil {
		*items = append(*items, CheckItem{
			Name:    "Firewall (nftables)",
			Status:  StatusOK,
			Details: "Installed",
		})
	} else {
		*items = append(*items, CheckItem{
			Name:       "Firewall (iptables)",
			Status:     StatusWarn,
			Details:    "Neither iptables nor nftables found (needed for server NAT)",
			FixCommand: d.suggestInstallCommand("iptables"),
			CanAutoFix: true,
		})
	}

	// DNS Resolver check (resolvconf or openresolv)
	hasResolv := false
	for _, bin := range []string{"resolvconf", "systemd-resolve", "resolvectl"} {
		if _, err := exec.LookPath(bin); err == nil {
			hasResolv = true
			break
		}
	}
	if hasResolv {
		*items = append(*items, CheckItem{
			Name:    "DNS Resolver Integration",
			Status:  StatusOK,
			Details: "Supported resolver found",
		})
	} else {
		*items = append(*items, CheckItem{
			Name:       "DNS Resolver Integration",
			Status:     StatusWarn,
			Details:    "resolvconf or systemd-resolved not found. Setting custom DNS in client configs might require resolvconf.",
			FixCommand: d.suggestInstallCommand("resolvconf"),
			CanAutoFix: true,
		})
	}
}

func (d *Doctor) checkDarwin(items *[]CheckItem) {
	// Check Homebrew
	if _, err := exec.LookPath("brew"); err == nil {
		*items = append(*items, CheckItem{
			Name:    "Package Manager (Homebrew)",
			Status:  StatusOK,
			Details: "Installed",
		})
	} else {
		*items = append(*items, CheckItem{
			Name:       "Package Manager (Homebrew)",
			Status:     StatusWarn,
			Details:    "Homebrew not detected. Recommended for managing WireGuard tools on macOS.",
			FixCommand: `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`,
			CanAutoFix: false,
		})
	}

	// Check wireguard-go or kernel driver on Mac
	if _, err := exec.LookPath("wireguard-go"); err == nil {
		*items = append(*items, CheckItem{
			Name:    "Userspace Backend (wireguard-go)",
			Status:  StatusOK,
			Details: "Installed",
		})
	} else {
		*items = append(*items, CheckItem{
			Name:       "Userspace Backend (wireguard-go)",
			Status:     StatusWarn,
			Details:    "wireguard-go is recommended on macOS for seamless userspace utun tunnels",
			FixCommand: "brew install wireguard-go",
			CanAutoFix: true,
		})
	}
}

func (d *Doctor) suggestInstallCommand(pkg string) string {
	pm := sysutil.DetectPackageManager()
	switch pm {
	case "brew":
		return fmt.Sprintf("brew install %s", pkg)
	case "apt-get":
		return fmt.Sprintf("apt-get update && apt-get install -y %s", pkg)
	case "dnf":
		return fmt.Sprintf("dnf install -y %s", pkg)
	case "yum":
		return fmt.Sprintf("yum install -y %s", pkg)
	case "pacman":
		return fmt.Sprintf("pacman -Sy --noconfirm %s", pkg)
	case "zypper":
		return fmt.Sprintf("zypper install -y %s", pkg)
	case "apk":
		return fmt.Sprintf("apk add %s", pkg)
	default:
		if runtime.GOOS == "darwin" {
			return fmt.Sprintf("brew install %s", pkg)
		}
		return fmt.Sprintf("sudo apt-get install -y %s", pkg)
	}
}

// AutoFix attempts to execute fix commands for all fixable items that have failed or warned.
func (d *Doctor) AutoFix(items []CheckItem) error {
	for _, item := range items {
		if (item.Status == StatusFail || item.Status == StatusWarn) && item.CanAutoFix && item.FixCommand != "" {
			fmt.Printf("🔧 Auto-fixing: %s...\n", item.Name)
			fmt.Printf("   Running: %s\n", item.FixCommand)

			var cmd *exec.Cmd
			if runtime.GOOS == "darwin" {
				// Homebrew shouldn't run as root
				cmd = exec.Command("sh", "-c", item.FixCommand)
			} else {
				if !sysutil.IsRoot() {
					cmd = exec.Command("sudo", "sh", "-c", item.FixCommand)
				} else {
					cmd = exec.Command("sh", "-c", item.FixCommand)
				}
			}

			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("   ❌ Failed to fix %s: %v\n", item.Name, err)
			} else {
				fmt.Printf("   ✅ Successfully fixed %s\n", item.Name)
			}
		}
	}
	return nil
}
