# ezwg (Easy WireGuard) 🚀

> **Aplikasi WireGuard Server & Client serba praktis untuk Linux dan macOS — tinggal install langsung jalan!**
> 
> *The turnkey, cross-platform WireGuard CLI manager for Linux and macOS.*

---

## 🌟 Fitur Utama (Key Features)

- **Instalasi 1-Klik**: Skrip instalasi otomatis untuk Linux (Ubuntu/Debian, CentOS/RHEL, Fedora, Arch) dan macOS (Intel & Apple Silicon).
- **Client VPN Mudah**:
  - `ezwg client import <profile.conf>`: Import profil WireGuard secara lokal atau langsung dari URL.
  - `ezwg connect`: Langsung konek ke VPN.
  - `ezwg status`: Cek status koneksi, IP tunnel, bytes transferred (RX/TX), handshake time, dan ping latency.
  - `ezwg disconnect`: Putuskan koneksi VPN.
- **Server VPN Lengkap**:
  - `ezwg server init`: Setup server WireGuard di VPS dengan otomatisasi IP forwarding, iptables/nftables firewall, NAT masquerading, dan pembuatan kunci.
  - `ezwg server add-peer <nama>`: Tambah client baru dengan alokasi IP otomatis, langsung tampilkan **QR Code** untuk HP dan file `.conf` untuk Mac/Linux.
  - `ezwg server list-peers`: Monitor daftar pengguna aktif beserta statistik pemakaian kuota/bandwidth secara realtime.
  - `ezwg server remove-peer <nama>`: Hapus hak akses client secara instan (hot-reload tanpa restart server).
- **ezwg Doctor (`ezwg doctor`)**:
  - Mendiagnosa kesiapan sistem: kernel module WireGuard, `wg-quick`, `resolvconf`, iptables, sysctl IP forward.
  - Opsi auto-fix (`ezwg doctor --fix`) untuk memasang dependensi yang kurang secara otomatis.

---

## 📥 Cara Install (Installation)

### 1. Menggunakan One-Line Installer (Rekomendasi)

Jalankan perintah ini di terminal **macOS** atau **Linux**:

```bash
curl -fsSL https://raw.githubusercontent.com/ezwg/ezwg/main/scripts/install.sh | bash
```

Skrip ini akan otomatis:
1. Mendeteksi sistem operasi (macOS / Linux) dan arsitektur CPU (`amd64` / `arm64`).
2. Memasang dependensi WireGuard (`wireguard-tools`, `wireguard-go`) jika belum ada.
3. Memasang binary `ezwg` ke `/usr/local/bin/ezwg`.
4. Menjalankan diagnosa awal `ezwg doctor`.

---

### 2. Install dari Source (Jika sudah ada Go)

```bash
git clone https://github.com/ezwg/ezwg.git
cd ezwg
make build
sudo make install
```

---

## 📖 Panduan Penggunaan (Usage Guide)

### A. Sebagai Client VPN (Koneksi ke VPN)

Jika Anda memiliki file konfigurasi `.conf` dari kantor, VPS pribadi, atau penyedia VPN:

#### 1. Import Profil
```bash
# Dari file lokal
ezwg client import ~/Downloads/kantor.conf --name kantor

# Atau dari URL
ezwg client import https://my-server.com/client.conf --name myvpn
```

#### 2. Konek ke VPN
```bash
# Menggunakan shortcut
ezwg connect kantor

# Atau jika hanya ada 1 profil, cukup:
ezwg connect
```

#### 3. Cek Status & Kecepatan
```bash
ezwg status
```
*Output menampilkan IP tunnel, public key, server endpoint, total data RX/TX, dan latency ping.*

#### 4. Putuskan Koneksi
```bash
ezwg disconnect
```

#### 5. Lihat Semua Profil Tersimpan
```bash
ezwg client list
```

---

### B. Sebagai Server VPN (Setup di Linux VPS)

Jika Anda ingin menjadikan VPS (Ubuntu, Debian, dsb.) sebagai server VPN pribadi:

#### 1. Inisialisasi Server
Jalankan perintah berikut di VPS Anda:
```bash
sudo ezwg server init
```
*`ezwg` akan otomatis mendeteksi IP publik server, network interface (misal: `eth0`), mengaktifkan `net.ipv4.ip_forward=1`, mengatur NAT iptables, dan menjalankan service WireGuard.*

Opsi kustomisasi (opsional):
```bash
sudo ezwg server init --port 51820 --subnet 10.8.0.1/24 --dns "1.1.1.1, 8.8.8.8"
```

#### 2. Tambah Pengguna / Client Baru
```bash
sudo ezwg server add-peer macbook-didik
```
**Output:**
- Menghasilkan file konfigurasi di `/etc/wireguard/clients/macbook-didik.conf`.
- Menampilkan gambar **QR Code** di terminal (tinggal scan pakai kamera aplikasi WireGuard di iPhone/Android).
- Menghasilkan konfigurasi siap pakai untuk Mac/Linux.

#### 3. Monitor Pengguna yang Sedang Online
```bash
sudo ezwg server list-peers
```
Contoh tampilan:
```text
STATUS     NAME             ALLOWED IP    ENDPOINT               LAST HANDSHAKE   RX / TX
● Online   macbook-didik    10.8.0.2/32   180.252.10.15:54321    15s ago          2.34 MiB / 14.81 MiB
○ Idle     hp-android       10.8.0.3/32   none                   never            0 B / 0 B
```

#### 4. Cabut / Hapus Pengguna
```bash
sudo ezwg server remove-peer macbook-didik
```

---

### C. System Doctor & Perbaikan Otomatis

Untuk memastikan perangkat Anda memiliki semua dependensi yang dibutuhkan:
```bash
ezwg doctor
```

Jika ada dependensi yang kurang, jalankan dengan bendera `--fix`:
```bash
ezwg doctor --fix
```

---

## 🛠️ Cross-Compilation & Build Release

Untuk menghasilkan paket rilis mandiri bagi semua platform (macOS x86_64, macOS Apple Silicon M1/M2/M3/M4, Linux x86_64, Linux ARM64):

```bash
make dist
```

File output akan berada di folder `dist/`:
- `ezwg_1.0.0_darwin_amd64.tar.gz` (Mac Intel)
- `ezwg_1.0.0_darwin_arm64.tar.gz` (Mac Apple Silicon)
- `ezwg_1.0.0_linux_amd64.tar.gz` (Linux x86_64 / PC / Server)
- `ezwg_1.0.0_linux_arm64.tar.gz` (Linux ARM64 / Raspberry Pi / AWS Graviton)
- `checksums.txt` (SHA256 checksums)

---

## 📄 Lisensi

MIT License © 2026 ezwg contributors.
