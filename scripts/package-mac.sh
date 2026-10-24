#!/usr/bin/env bash
set -euo pipefail

APP_NAME="Easy WireGuard"
BUNDLE_DIR="dist/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "📦 Packaging macOS Desktop App: ${APP_NAME}.app..."

mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

# Build the GUI binary
echo "  -> Compiling ezwg-gui binary..."
go build -ldflags="-s -w" -o "${MACOS_DIR}/ezwg-gui" ./cmd/ezwg-gui

# Create Info.plist
cat << EOF > "${CONTENTS_DIR}/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>ezwg-gui</string>
    <key>CFBundleIdentifier</key>
    <string>com.ezwg.client</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0.0</string>
    <key>CFBundleVersion</key>
    <string>1</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.15</string>
    <key>NSHighResolutionCapable</key>
    <true/>
</dict>
</plist>
EOF

chmod +x "${MACOS_DIR}/ezwg-gui"

echo "✅ App bundle created at: ${BUNDLE_DIR}"
echo "You can move it to /Applications with: cp -r '${BUNDLE_DIR}' /Applications/"
