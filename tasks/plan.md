# Implementation Plan: ezwg (Easy WireGuard Suite for Linux & macOS)

## Goal
Build a complete, user-friendly, cross-platform WireGuard CLI suite in Go called `ezwg` (Easy WireGuard), featuring:
1. One-click install script (`curl -fsSL ... | bash`) for macOS (Intel & Apple Silicon) and Linux (x86_64 & ARM64).
2. Server management: initialize server on VPS, configure firewall/NAT, add/remove peers, generate client `.conf` & QR codes.
3. Client management: import profiles, connect/disconnect tunnels with one command, inspect status & transfer stats.
4. System doctor / diagnostic: auto-detect missing dependencies (wg-quick, resolvconf, kernel modules) and install or guide resolution.

## Architecture & Components
- **`cmd/ezwg`**: Main entrypoint with Cobra CLI framework (`server`, `client`, `doctor`, `version`).
- **`internal/wireguard`**:
  - Pure Go key generation (Curve25519/X25519 clamp, Base64, PresharedKey).
  - Config parser/serializer for `.conf` files.
  - Terminal QR code rendering (`skip2/go-qrcode`).
  - WireGuard process & interface controller (`wg`, `wg-quick`, `systemctl`).
- **`internal/server`**:
  - VPS network interface & public IP discovery.
  - Server initialization (subnet CIDR, listen port, DNS, iptables/nftables NAT rules).
  - Peer IP allocator & lifecycle (add, remove, list, export).
- **`internal/client`**:
  - Profile storage (`~/.config/ezwg/` or `/etc/wireguard/`).
  - Connect, disconnect, list profiles, real-time status.
- **`internal/sysutil`**:
  - Cross-platform OS detection (Darwin vs Linux), distro package manager detection (brew, apt, dnf, pacman).
  - IP forwarding sysctl manipulation.
  - Elevation / sudo check.
- **`scripts/install.sh`**:
  - Universal one-liner installer for macOS and Linux.
  - Detects OS & Arch, installs dependencies if requested, installs binary to `/usr/local/bin/ezwg`.
- **`scripts/build.sh` & `Makefile`**:
  - Cross-compilation for `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`.

## Step-by-Step Milestones
1. Initialize Go module & dependencies (`github.com/spf13/cobra`, `golang.org/x/crypto`, `skip2/go-qrcode`, etc.).
2. Core Cryptography & Config Parsing (`internal/wireguard`).
3. System & Network Utilities (`internal/sysutil`).
4. Server Suite (`internal/server`).
5. Client Suite (`internal/client`).
6. Diagnostics (`internal/doctor`).
7. Cobra CLI integration (`cmd/ezwg`).
8. Unit & Integration tests for config generation, key derivation, and CLI commands.
9. Universal installer (`scripts/install.sh`) and build scripts.
10. Documentation, verification, and end-to-end testing.
