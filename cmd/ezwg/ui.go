package main

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var guiCmd = &cobra.Command{
	Use:     "gui",
	Aliases: []string{"ui", "app"},
	Short:   "Launch the Easy WireGuard GUI desktop application",
	Run: func(cmd *cobra.Command, args []string) {
		color.Cyan("🚀 Launching Easy WireGuard Desktop GUI...")

		// 1. Check if macOS .app exists
		if runtime.GOOS == "darwin" {
			appPaths := []string{
				"/Applications/Easy WireGuard.app",
				"dist/Easy WireGuard.app",
			}
			for _, appPath := range appPaths {
				if info, err := os.Stat(appPath); err == nil && info.IsDir() {
					execCmd := exec.Command("open", appPath)
					if err := execCmd.Start(); err == nil {
						color.Green("✅ Opened %s", appPath)
						return
					}
				}
			}
		}

		// 2. Check ezwg-gui binary in PATH or local bin
		binCandidates := []string{
			"ezwg-gui",
			"/usr/local/bin/ezwg-gui",
			"bin/ezwg-gui",
			"./dist/Easy WireGuard.app/Contents/MacOS/ezwg-gui",
		}

		for _, bin := range binCandidates {
			if path, err := exec.LookPath(bin); err == nil {
				execCmd := exec.Command(path)
				execCmd.Stdout = os.Stdout
				execCmd.Stderr = os.Stderr
				if err := execCmd.Start(); err == nil {
					color.Green("✅ Running GUI from %s", path)
					return
				}
			}
		}

		color.Yellow("GUI executable 'ezwg-gui' not found.")
		color.White("To build and package the GUI app, run:")
		color.Cyan("  make gui")
	},
}

func init() {
	rootCmd.AddCommand(guiCmd)
}
