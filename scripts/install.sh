#!/usr/bin/env bash
#
# ezwg - Universal 1-Click Installer for Linux & macOS
#
set -euo pipefail

CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m' # No Color

APP_NAME="ezwg"
INSTALL_DIR="/usr/local/bin"
REPO_OWNER=${REPO_OWNER:-"didikariyanto"}
REPO_NAME=${REPO_NAME:-"WireGuard"}
VERSION=${VERSION:-"1.0.0"}

echo -e "${CYAN}${BOLD}"
echo "  _____ ______          _______ "
echo " / ____|___  /         / ____(_)"
echo "| |  __   / / |  _ _ _| |  __ _ "
echo "| | |_ | / /| | | | '_| | |_ | |"
echo "| |__| |/ /_| |_| | | | |__| | |"
echo " \_____/_____|__,_|_|  \_____|_|"
echo -e "${NC}"
echo -e "${BOLD}Easy WireGuard (ezwg) - Universal Installer for Linux & macOS${NC}"
echo "================================================================"

# 1. Detect OS
OS_TYPE="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH_TYPE="$(uname -m)"

case "${OS_TYPE}" in
    darwin*)  OS="darwin" ;;
    linux*)   OS="linux" ;;
    *)
        echo -e "${RED}❌ Unsupported operating system: ${OS_TYPE}${NC}"
        echo "ezwg currently supports macOS (Darwin) and Linux."
        exit 1
        ;;
esac

# 2. Detect Architecture
case "${ARCH_TYPE}" in
    x86_64|amd64)   ARCH="amd64" ;;
    arm64|aarch64)  ARCH="arm64" ;;
    *)
        echo -e "${RED}❌ Unsupported CPU architecture: ${ARCH_TYPE}${NC}"
        exit 1
        ;;
esac

echo -e "Detected System: ${GREEN}${OS}/${ARCH}${NC}"

# Check for root/sudo helper
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    if command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
    else
        echo -e "${RED}❌ Root privileges or 'sudo' is required to install to ${INSTALL_DIR}${NC}"
        exit 1
    fi
fi

# 3. Check & Install WireGuard Prerequisites
echo -e "\n${CYAN}🔍 Checking WireGuard system dependencies...${NC}"

if [ "${OS}" = "darwin" ]; then
    if ! command -v wg >/dev/null 2>&1 || ! command -v wg-quick >/dev/null 2>&1; then
        echo -e "${YELLOW}wireguard-tools not found.${NC}"
        if command -v brew >/dev/null 2>&1; then
            echo -e "Installing WireGuard tools via Homebrew..."
            brew install wireguard-tools wireguard-go
        else
            echo -e "${YELLOW}⚠️ Homebrew is not installed. Please install Homebrew or install wireguard-tools manually.${NC}"
        fi
    else
        echo -e "${GREEN}✔ WireGuard tools found (wg, wg-quick)${NC}"
    fi
elif [ "${OS}" = "linux" ]; then
    MISSING_PKGS=()
    if ! command -v wg >/dev/null 2>&1; then MISSING_PKGS+=("wireguard-tools"); fi
    if ! command -v iptables >/dev/null 2>&1 && ! command -v nft >/dev/null 2>&1; then MISSING_PKGS+=("iptables"); fi

    if [ ${#MISSING_PKGS[@]} -gt 0 ]; then
        echo -e "${YELLOW}Installing missing dependencies: ${MISSING_PKGS[*]}...${NC}"
        if command -v apt-get >/dev/null 2>&1; then
            $SUDO apt-get update -qq && $SUDO apt-get install -y -qq wireguard wireguard-tools resolvconf iptables
        elif command -v dnf >/dev/null 2>&1; then
            $SUDO dnf install -y wireguard-tools iptables
        elif command -v yum >/dev/null 2>&1; then
            $SUDO yum install -y epel-release && $SUDO yum install -y wireguard-tools iptables
        elif command -v pacman >/dev/null 2>&1; then
            $SUDO pacman -Sy --noconfirm wireguard-tools iptables openresolv
        fi
    else
        echo -e "${GREEN}✔ Linux WireGuard tools found${NC}"
    fi
fi

# 4. Install ezwg CLI & GUI binary
echo -e "\n${CYAN}📦 Installing ${APP_NAME} to ${INSTALL_DIR}...${NC}"
$SUDO mkdir -p "${INSTALL_DIR}"

INSTALLED=false

# Option A: Local bin already compiled
if [ -f "./bin/ezwg" ]; then
    echo "Installing from local ./bin/ezwg..."
    $SUDO cp ./bin/ezwg "${INSTALL_DIR}/${APP_NAME}"
    if [ -f "./bin/ezwg-gui" ]; then
        $SUDO cp ./bin/ezwg-gui "${INSTALL_DIR}/ezwg-gui"
    fi
    INSTALLED=true
# Option B: Go is available in the current project directory
elif [ -f "./cmd/ezwg/root.go" ] && command -v go >/dev/null 2>&1; then
    echo "Compiling from local source using Go..."
    go build -ldflags="-s -w" -o "/tmp/${APP_NAME}" ./cmd/ezwg
    $SUDO mv "/tmp/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
    if [ -d "./cmd/ezwg-gui" ]; then
        echo "Compiling ezwg-gui client..."
        go build -ldflags="-s -w" -o "/tmp/ezwg-gui" ./cmd/ezwg-gui || true
        if [ -f "/tmp/ezwg-gui" ]; then
            $SUDO mv "/tmp/ezwg-gui" "${INSTALL_DIR}/ezwg-gui"
        fi
    fi
    INSTALLED=true
# Option C: Download prebuilt release from GitHub
else
    TAR_NAME="${APP_NAME}_${VERSION}_${OS}_${ARCH}.tar.gz"
    RELEASE_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${TAR_NAME}"
    FALLBACK_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/v${VERSION}/${TAR_NAME}"
    TMP_DIR="$(mktemp -d)"

    echo "Downloading ${TAR_NAME} from GitHub Releases..."
    DOWNLOAD_SUCCESS=false

    if command -v curl >/dev/null 2>&1; then
        if curl -fsSL "${RELEASE_URL}" -o "${TMP_DIR}/${TAR_NAME}" 2>/dev/null; then
            DOWNLOAD_SUCCESS=true
        elif curl -fsSL "${FALLBACK_URL}" -o "${TMP_DIR}/${TAR_NAME}" 2>/dev/null; then
            DOWNLOAD_SUCCESS=true
        fi
    elif command -v wget >/dev/null 2>&1; then
        if wget -q "${RELEASE_URL}" -O "${TMP_DIR}/${TAR_NAME}" 2>/dev/null; then
            DOWNLOAD_SUCCESS=true
        elif wget -q "${FALLBACK_URL}" -O "${TMP_DIR}/${TAR_NAME}" 2>/dev/null; then
            DOWNLOAD_SUCCESS=true
        fi
    fi

    if [ "${DOWNLOAD_SUCCESS}" = "true" ] && [ -s "${TMP_DIR}/${TAR_NAME}" ]; then
        tar -xzf "${TMP_DIR}/${TAR_NAME}" -C "${TMP_DIR}"
        if [ -f "${TMP_DIR}/${APP_NAME}_${VERSION}_${OS}_${ARCH}/${APP_NAME}" ]; then
            $SUDO mv "${TMP_DIR}/${APP_NAME}_${VERSION}_${OS}_${ARCH}/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
            rm -rf "${TMP_DIR}"
            INSTALLED=true
        fi
    fi

    # Fallback to Go install if release binary download was unavailable
    if [ "${INSTALLED}" = "false" ] && command -v go >/dev/null 2>&1; then
        echo "Release binary not yet published on GitHub. Compiling directly via Go..."
        GOBIN="/tmp" go install "github.com/${REPO_OWNER}/${REPO_NAME}/cmd/ezwg@latest" || true
        if [ -f "/tmp/${APP_NAME}" ]; then
            $SUDO mv "/tmp/${APP_NAME}" "${INSTALL_DIR}/${APP_NAME}"
            INSTALLED=true
        fi
    fi
fi

if [ "${INSTALLED}" = "false" ]; then
    echo -e "${RED}❌ Could not install binary. Please make sure Go is installed or download prebuilt binaries from releases.${NC}"
    exit 1
fi

$SUDO chmod 755 "${INSTALL_DIR}/${APP_NAME}"
if [ -f "${INSTALL_DIR}/ezwg-gui" ]; then
    $SUDO chmod 755 "${INSTALL_DIR}/ezwg-gui"
fi

# Create alias symlink ezwireguard -> ezwg
$SUDO ln -sf "${INSTALL_DIR}/${APP_NAME}" "${INSTALL_DIR}/ezwireguard"

# Install macOS .app bundle to /Applications if available
if [ "${OS}" = "darwin" ]; then
    if [ -d "./dist/Easy WireGuard.app" ]; then
        echo "Installing Easy WireGuard.app to /Applications..."
        $SUDO rm -rf "/Applications/Easy WireGuard.app"
        $SUDO cp -r "./dist/Easy WireGuard.app" /Applications/
    fi
elif [ "${OS}" = "linux" ]; then
    if [ -f "./assets/ezwg.desktop" ] && [ -d "/usr/share/applications" ]; then
        echo "Installing Linux desktop launcher..."
        $SUDO cp ./assets/ezwg.desktop /usr/share/applications/
    fi
fi

echo -e "${GREEN}✔ ${APP_NAME} successfully installed in ${INSTALL_DIR}/${APP_NAME}!${NC}"

# 5. Run Doctor Check
echo -e "\n${CYAN}🩺 Running system diagnostic check...${NC}"
"${INSTALL_DIR}/${APP_NAME}" doctor || true

echo -e "\n${GREEN}${BOLD}🎉 Installation Complete!${NC}"
echo "================================================================"
echo -e "You can now use ${BOLD}ezwg${NC} anywhere in your terminal or launch the GUI app."
echo ""
echo -e "${BOLD}🖥️ Client Desktop GUI:${NC}"
if [ "${OS}" = "darwin" ]; then
    echo "  Open 'Easy WireGuard' from your Applications / Launchpad"
    echo "  Or in terminal: ezwg gui"
else
    echo "  Open 'Easy WireGuard' from your Application menu"
    echo "  Or in terminal: ezwg gui"
fi
echo ""
echo -e "${BOLD}💻 Client CLI (Quick Commands):${NC}"
echo "  ezwg client import /path/to/profile.conf"
echo "  ezwg connect"
echo "  ezwg status"
echo "  ezwg disconnect"
echo ""
echo -e "${BOLD}⚙️ Server Interactive Setup (on VPS):${NC}"
echo "  sudo ezwg server         # Interactive setup wizard & menu"
echo "================================================================"
