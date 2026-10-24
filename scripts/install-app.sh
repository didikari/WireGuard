#!/usr/bin/env bash
#
# Easy WireGuard - Dedicated GUI Client App Installer (macOS & Linux)
#
set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

APP_NAME="Easy WireGuard"
REPO_OWNER=${REPO_OWNER:-"didikari"}
REPO_NAME=${REPO_NAME:-"WireGuard"}
VERSION=${VERSION:-"1.0.0"}

echo -e "${CYAN}${BOLD}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║          Easy WireGuard - Client App Installer             ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_TYPE="$(uname -m)"

case "${OS_TYPE}" in
    darwin*)  OS="darwin" ;;
    linux*)   OS="linux" ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: ${OS_TYPE}${NC}"
        exit 1
        ;;
esac

case "${ARCH_TYPE}" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)
        echo -e "${RED}❌ Unsupported CPU architecture: ${ARCH_TYPE}${NC}"
        exit 1
        ;;
esac

echo -e "Platform: ${GREEN}${OS}/${ARCH}${NC}"

# Check sudo helper
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
    fi
fi

# ==============================================================================
# 1. macOS Installation (Easy WireGuard.app -> /Applications)
# ==============================================================================
if [ "${OS}" = "darwin" ]; then
    echo -e "\n${CYAN}🔍 Checking WireGuard dependencies...${NC}"
    if ! command -v wg >/dev/null 2>&1 || ! command -v wg-quick >/dev/null 2>&1; then
        if command -v brew >/dev/null 2>&1; then
            echo -e "${YELLOW}Installing wireguard-tools via Homebrew...${NC}"
            brew install wireguard-tools wireguard-go
        else
            echo -e "${YELLOW}⚠️ Homebrew not detected. Please make sure wireguard-tools is installed.${NC}"
        fi
    else
        echo -e "${GREEN}✔ WireGuard backend tools ready${NC}"
    fi

    echo -e "\n${CYAN}📦 Installing ${APP_NAME}.app into /Applications...${NC}"
    TARGET_APP="/Applications/${APP_NAME}.app"
    INSTALLED=false

    # Option A: Local dist available
    if [ -d "./dist/${APP_NAME}.app" ]; then
        echo "Installing from local ./dist/${APP_NAME}.app..."
        $SUDO rm -rf "${TARGET_APP}"
        $SUDO cp -r "./dist/${APP_NAME}.app" /Applications/
        INSTALLED=true
    # Option B: Local source available with Go
    elif [ -f "./cmd/ezwg-gui/main.go" ] && command -v go >/dev/null 2>&1; then
        echo "Building local source via Go..."
        ./scripts/package-mac.sh
        $SUDO rm -rf "${TARGET_APP}"
        $SUDO cp -r "./dist/${APP_NAME}.app" /Applications/
        INSTALLED=true
    # Option C: Download from GitHub Releases
    else
        TMP_DIR="$(mktemp -d)"
        ZIP_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/Easy-WireGuard-macOS.zip"
        FALLBACK_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${VERSION}/Easy-WireGuard-macOS.zip"

        echo "Downloading ${ZIP_URL}..."
        if curl -fsSL "${ZIP_URL}" -o "${TMP_DIR}/app.zip" 2>/dev/null || curl -fsSL "${FALLBACK_URL}" -o "${TMP_DIR}/app.zip" 2>/dev/null; then
            unzip -q "${TMP_DIR}/app.zip" -d "${TMP_DIR}"
            if [ -d "${TMP_DIR}/${APP_NAME}.app" ]; then
                $SUDO rm -rf "${TARGET_APP}"
                $SUDO mv "${TMP_DIR}/${APP_NAME}.app" /Applications/
                INSTALLED=true
            fi
        fi
        rm -rf "${TMP_DIR}"
    fi

    if [ "${INSTALLED}" = "true" ]; then
        echo -e "\n${GREEN}${BOLD}🎉 Easy WireGuard.app successfully installed!${NC}"
        echo "================================================================"
        echo -e "You can now open ${BOLD}Easy WireGuard${NC} from:"
        echo "  • Spotlight (Cmd + Space -> 'Easy WireGuard')"
        echo "  • Launchpad / Finder Applications"
        echo "================================================================"
        exit 0
    else
        echo -e "${RED}❌ Failed to download or install ${APP_NAME}.app.${NC}"
        echo -e "You can also download the DMG file manually from:"
        echo "  https://github.com/${REPO_OWNER}/${REPO_NAME}/releases"
        exit 1
    fi
fi

# ==============================================================================
# 2. Linux Desktop Installation (ezwg-gui + desktop launcher)
# ==============================================================================
if [ "${OS}" = "linux" ]; then
    echo -e "\n${CYAN}🔍 Checking Linux dependencies...${NC}"
    if ! command -v wg-quick >/dev/null 2>&1; then
        if command -v apt-get >/dev/null 2>&1; then
            $SUDO apt-get update -qq && $SUDO apt-get install -y -qq wireguard-tools
        elif command -v dnf >/dev/null 2>&1; then
            $SUDO dnf install -y wireguard-tools
        elif command -v pacman >/dev/null 2>&1; then
            $SUDO pacman -Sy --noconfirm wireguard-tools openresolv
        fi
    fi

    echo -e "\n${CYAN}📦 Installing ${APP_NAME} GUI Client...${NC}"
    INSTALL_DIR="/usr/local/bin"
    $SUDO mkdir -p "${INSTALL_DIR}"
    INSTALLED=false

    # Local bin
    if [ -f "./bin/ezwg-gui" ]; then
        $SUDO cp ./bin/ezwg-gui "${INSTALL_DIR}/ezwg-gui"
        INSTALLED=true
    # Compile from source
    elif [ -f "./cmd/ezwg-gui/main.go" ] && command -v go >/dev/null 2>&1; then
        go build -ldflags="-s -w" -o "/tmp/ezwg-gui" ./cmd/ezwg-gui
        $SUDO mv "/tmp/ezwg-gui" "${INSTALL_DIR}/ezwg-gui"
        INSTALLED=true
    # Download from GitHub release
    else
        TAR_NAME="ezwg-gui_${VERSION}_linux_${ARCH}.tar.gz"
        URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${TAR_NAME}"
        TMP_DIR="$(mktemp -d)"
        if curl -fsSL "${URL}" -o "${TMP_DIR}/${TAR_NAME}" 2>/dev/null; then
            tar -xzf "${TMP_DIR}/${TAR_NAME}" -C "${TMP_DIR}"
            $SUDO mv "${TMP_DIR}/ezwg-gui" "${INSTALL_DIR}/ezwg-gui"
            INSTALLED=true
        fi
        rm -rf "${TMP_DIR}"
    fi

    if [ "${INSTALLED}" = "true" ]; then
        $SUDO chmod 755 "${INSTALL_DIR}/ezwg-gui"

        # Install Desktop Icon & Entry
        if [ -d "/usr/share/pixmaps" ] && [ -f "./assets/icon.png" ]; then
            $SUDO cp ./assets/icon.png /usr/share/pixmaps/ezwg.png
        fi

        if [ -d "/usr/share/applications" ]; then
            $SUDO tee /usr/share/applications/ezwg.desktop > /dev/null << 'EOF'
[Desktop Entry]
Name=Easy WireGuard
Comment=WireGuard VPN Client GUI
Exec=/usr/local/bin/ezwg-gui
Icon=ezwg
Terminal=false
Type=Application
Categories=Network;Security;VPN;
EOF
        fi

        echo -e "\n${GREEN}${BOLD}🎉 Easy WireGuard GUI successfully installed!${NC}"
        echo "================================================================"
        echo "You can now launch 'Easy WireGuard' from your Application menu"
        echo "or run 'ezwg-gui' in terminal."
        echo "================================================================"
        exit 0
    else
        echo -e "${RED}❌ Failed to install ezwg-gui.${NC}"
        exit 1
    fi
fi
