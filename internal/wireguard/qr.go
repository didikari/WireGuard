package wireguard

import (
	"fmt"

	"github.com/skip2/go-qrcode"
)

// GenerateTerminalQRCode generates an ASCII string of the QR code suitable for printing to the terminal.
func GenerateTerminalQRCode(content string) (string, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %w", err)
	}
	// ToSmallString returns an inverted or regular half-block string that prints cleanly on dark and light terminals
	return qr.ToSmallString(false), nil
}

// GenerateQRCodePNG generates a PNG byte slice for the given config content.
func GenerateQRCodePNG(content string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return nil, fmt.Errorf("failed to encode QR code to PNG: %w", err)
	}
	return png, nil
}
