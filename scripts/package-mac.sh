#!/usr/bin/env bash
set -euo pipefail

APP_NAME="Easy WireGuard"
DIST_DIR="dist"
BUNDLE_DIR="${DIST_DIR}/${APP_NAME}.app"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"

echo "📦 Packaging macOS Desktop App: ${APP_NAME}.app..."

mkdir -p "${MACOS_DIR}" "${RESOURCES_DIR}"

# Build the GUI binary
echo "  -> Compiling ezwg-gui binary for macOS..."
go build -ldflags="-s -w" -o "${MACOS_DIR}/ezwg-gui" ./cmd/ezwg-gui

# Copy AppIcon.icns if available
if [ -f "assets/AppIcon.icns" ]; then
    cp assets/AppIcon.icns "${RESOURCES_DIR}/AppIcon.icns"
fi

# Create Info.plist
cat << EOF > "${CONTENTS_DIR}/Info.plist"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>ezwg-gui</string>
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
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

# Create DMG Installer
echo "  -> Generating macOS DMG Installer (Easy-WireGuard.dmg)..."
hdiutil create -volname "${APP_NAME}" -srcfolder "${BUNDLE_DIR}" -ov -format UDZO "${DIST_DIR}/Easy-WireGuard.dmg"

# Create ZIP archive for automated installer
echo "  -> Generating macOS ZIP Archive (Easy-WireGuard-macOS.zip)..."
(cd "${DIST_DIR}" && zip -rq "Easy-WireGuard-macOS.zip" "${APP_NAME}.app")

echo "✅ macOS App packages created in ${DIST_DIR}/:"
ls -lh "${DIST_DIR}"/Easy-WireGuard*
