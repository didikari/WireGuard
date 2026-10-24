package main

import (
	"fmt"
	"text/tabwriter"

	"github.com/ezwg/ezwg/internal/doctor"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var doctorAutoFix bool

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health and WireGuard requirements for Linux/macOS",
	Long:  "Inspects kernel modules, tools (wg, wg-quick), firewall, DNS resolver, and permissions.",
	Run: func(cmd *cobra.Command, args []string) {
		doc := doctor.NewDoctor()
		items := doc.RunDiagnostics()

		color.Cyan("🩺 Running ezwg System Diagnostics...")
		fmt.Println()

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "STATUS\tCHECK\tDETAILS")

		hasIssues := false

		for _, item := range items {
			var statusStr string
			switch item.Status {
			case doctor.StatusOK:
				statusStr = color.GreenString("✔ OK")
			case doctor.StatusWarn:
				statusStr = color.YellowString("⚠ WARN")
				hasIssues = true
			case doctor.StatusFail:
				statusStr = color.RedString("✖ FAIL")
				hasIssues = true
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", statusStr, item.Name, item.Details)
		}
		w.Flush()
		fmt.Println()

		if hasIssues {
			color.Yellow("💡 Recommendations / Fixes:")
			for _, item := range items {
				if item.Status != doctor.StatusOK && item.FixCommand != "" {
					color.White("  - %s: %s", item.Name, color.CyanString(item.FixCommand))
				}
			}
			fmt.Println()

			if doctorAutoFix {
				color.Cyan("Attempting auto-remediation (--fix enabled)...")
				_ = doc.AutoFix(items)
			} else {
				color.White("Tip: You can attempt automatic installation of missing tools with:")
				color.Cyan("  ezwg doctor --fix")
			}
		} else {
			color.Green("🎉 All system checks passed! Your machine is ready for WireGuard.")
		}
	},
}

func init() {
	doctorCmd.Flags().BoolVarP(&doctorAutoFix, "fix", "f", false, "Automatically attempt to install missing dependencies")
}
