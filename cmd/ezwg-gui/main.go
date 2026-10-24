package main

import (
	_ "embed"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/ezwg/ezwg/internal/client"
	"github.com/ezwg/ezwg/internal/wireguard"
)

//go:embed icon.png
var appIconBytes []byte

func main() {
	myApp := app.NewWithID("com.ezwg.client")
	iconResource := fyne.NewStaticResource("icon.png", appIconBytes)
	myApp.SetIcon(iconResource)

	myWindow := myApp.NewWindow("Easy WireGuard")
	myWindow.SetIcon(iconResource)
	myWindow.Resize(fyne.NewSize(520, 560))
	myWindow.SetFixedSize(false)

	cm, err := client.NewClientManager()
	if err != nil {
		dialog.ShowError(fmt.Errorf("failed to initialize client manager: %w", err), myWindow)
	}

	// Status header widgets
	statusIcon := widget.NewIcon(theme.CancelIcon())
	statusLabel := widget.NewLabelWithStyle("Disconnected", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	statusContainer := container.NewHBox(statusIcon, statusLabel)

	// Profile selector widgets
	var profileSelect *widget.Select
	selectedProfile := ""

	// Detail Labels
	ipLabel := widget.NewLabel("-")
	endpointLabel := widget.NewLabel("-")
	transferLabel := widget.NewLabel("- / -")
	handshakeLabel := widget.NewLabel("-")
	pingLabel := widget.NewLabel("-")

	// Connect/Disconnect Button
	var toggleBtn *widget.Button

	refreshProfiles := func() {
		if cm == nil {
			return
		}
		profiles, err := cm.ListProfiles()
		if err != nil {
			return
		}
		var names []string
		activeProfile := ""
		for _, p := range profiles {
			names = append(names, p.Name)
			if p.IsActive {
				activeProfile = p.Name
			}
		}

		profileSelect.Options = names
		if activeProfile != "" {
			selectedProfile = activeProfile
			profileSelect.SetSelected(activeProfile)
		} else if len(names) > 0 && selectedProfile == "" {
			selectedProfile = names[0]
			profileSelect.SetSelected(names[0])
		} else if len(names) == 0 {
			profileSelect.ClearSelected()
			selectedProfile = ""
		}
		profileSelect.Refresh()
	}

	updateStatusDisplay := func() {
		if cm == nil {
			return
		}
		active, _ := cm.GetActiveProfile()
		if active != "" {
			statusIcon.SetResource(theme.ConfirmIcon())
			statusLabel.SetText(fmt.Sprintf("Connected to: %s", active))
			toggleBtn.SetText("Disconnect")
			toggleBtn.Importance = widget.DangerImportance
			toggleBtn.SetIcon(theme.MediaStopIcon())

			// Fetch live stats
			status, err := cm.Status(active)
			if err == nil && status != nil {
				for _, p := range status.Peers {
					endpointLabel.SetText(p.Endpoint)
					transferLabel.SetText(fmt.Sprintf("%s RX / %s TX",
						wireguard.FormatBytes(p.TransferRx),
						wireguard.FormatBytes(p.TransferTx)))
					if !p.LatestHandshake.IsZero() {
						handshakeLabel.SetText(fmt.Sprintf("%s (%s ago)",
							p.LatestHandshake.Format("15:04:05"),
							strings.TrimSpace(formatSince(p.LatestHandshake))))
					} else {
						handshakeLabel.SetText("Never")
					}
					if p.Endpoint != "" {
						go func(ep string) {
							lat := client.PingServer(ep)
							pingLabel.SetText(lat)
						}(p.Endpoint)
					}
					break
				}
			}
		} else {
			statusIcon.SetResource(theme.CancelIcon())
			statusLabel.SetText("Disconnected")
			toggleBtn.SetText("Connect")
			toggleBtn.Importance = widget.HighImportance
			toggleBtn.SetIcon(theme.MediaPlayIcon())

			endpointLabel.SetText("-")
			transferLabel.SetText("-")
			handshakeLabel.SetText("-")
			pingLabel.SetText("-")
		}

		if selectedProfile != "" {
			profiles, _ := cm.ListProfiles()
			for _, p := range profiles {
				if p.Name == selectedProfile {
					ipLabel.SetText(p.ClientIP)
					if active == "" {
						endpointLabel.SetText(p.Endpoint)
					}
					break
				}
			}
		} else {
			ipLabel.SetText("-")
		}

		toggleBtn.Refresh()
		statusContainer.Refresh()
	}

	profileSelect = widget.NewSelect([]string{}, func(s string) {
		selectedProfile = s
		updateStatusDisplay()
	})

	toggleBtn = widget.NewButtonWithIcon("Connect", theme.MediaPlayIcon(), func() {
		if cm == nil {
			return
		}
		active, _ := cm.GetActiveProfile()
		if active != "" {
			// Disconnect
			toggleBtn.Disable()
			go func() {
				defer toggleBtn.Enable()
				err := cm.Disconnect(active)
				if err != nil {
					dialog.ShowError(fmt.Errorf("disconnect failed: %w", err), myWindow)
				}
				updateStatusDisplay()
			}()
		} else {
			// Connect
			if selectedProfile == "" {
				dialog.ShowInformation("No Profile", "Please import or select a VPN profile first.", myWindow)
				return
			}
			toggleBtn.Disable()
			go func() {
				defer toggleBtn.Enable()
				err := cm.Connect(selectedProfile)
				if err != nil {
					dialog.ShowError(fmt.Errorf("connection failed: %w\n(Ensure WireGuard tools or sudo permissions are available)", err), myWindow)
				}
				updateStatusDisplay()
			}()
		}
	})
	toggleBtn.Importance = widget.HighImportance

	// Import Profile Button
	importBtn := widget.NewButtonWithIcon("Import .conf", theme.FileIcon(), func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil || reader == nil {
				return
			}
			defer reader.Close()

			fileName := reader.URI().Name()
			baseName := strings.TrimSuffix(fileName, filepath.Ext(fileName))

			path := reader.URI().Path()
			_, importErr := cm.ImportProfile(baseName, path)
			if importErr != nil {
				dialog.ShowError(fmt.Errorf("import failed: %w", importErr), myWindow)
				return
			}

			dialog.ShowInformation("Success", fmt.Sprintf("Profile '%s' imported successfully!", baseName), myWindow)
			selectedProfile = baseName
			refreshProfiles()
			updateStatusDisplay()
		}, myWindow)

		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".conf"}))
		fileDialog.Show()
	})

	// Delete Profile Button
	deleteBtn := widget.NewButtonWithIcon("Delete", theme.DeleteIcon(), func() {
		if selectedProfile == "" {
			return
		}
		dialog.ShowConfirm("Delete Profile", fmt.Sprintf("Are you sure you want to delete '%s'?", selectedProfile), func(ok bool) {
			if !ok {
				return
			}
			err := cm.DeleteProfile(selectedProfile)
			if err != nil {
				dialog.ShowError(err, myWindow)
				return
			}
			selectedProfile = ""
			refreshProfiles()
			updateStatusDisplay()
		}, myWindow)
	})

	// Details Card
	detailsForm := widget.NewForm(
		widget.NewFormItem("VPN Assigned IP:", ipLabel),
		widget.NewFormItem("Server Endpoint:", endpointLabel),
		widget.NewFormItem("Transfer (RX/TX):", transferLabel),
		widget.NewFormItem("Latest Handshake:", handshakeLabel),
		widget.NewFormItem("Latency / Ping:", pingLabel),
	)

	// Build user interface layout
	header := container.NewVBox(
		widget.NewLabelWithStyle("WireGuard Connection Manager", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		container.NewHBox(widget.NewLabel("Status:"), statusContainer),
		widget.NewSeparator(),
	)

	profileRow := container.NewBorder(
		nil, nil,
		widget.NewLabel("Profile:"),
		container.NewHBox(importBtn, deleteBtn),
		profileSelect,
	)

	btnContainer := container.NewVBox(
		widget.NewSeparator(),
		toggleBtn,
	)

	content := container.NewVBox(
		header,
		profileRow,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Tunnel Statistics", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		detailsForm,
		btnContainer,
	)

	myWindow.SetContent(container.NewPadded(content))

	// Initial load
	refreshProfiles()
	updateStatusDisplay()

	// Periodic ticker to refresh statistics
	ticker := time.NewTicker(3 * time.Second)
	go func() {
		for range ticker.C {
			updateStatusDisplay()
		}
	}()

	myWindow.ShowAndRun()
}

func formatSince(t time.Time) string {
	d := time.Since(t)
	secs := int(d.Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	mins := secs / 60
	if mins < 60 {
		return fmt.Sprintf("%dm", mins)
	}
	hrs := mins / 60
	return fmt.Sprintf("%dh", hrs)
}
