# ezwg (Easy WireGuard) 🚀

> **Aplikasi WireGuard Client GUI & Server Interaktif untuk Linux dan macOS — tinggal install langsung jalan!**
> 
> *The turnkey WireGuard Client Desktop GUI & Interactive Server Manager for Linux and macOS.*

---

## 🌟 Fitur Unggulan

- **🖥️ Client Desktop GUI (Tinggal Klik)**:
  - Tersedia aplikasi visual desktop (**Easy WireGuard.app** untuk macOS dan Desktop Launcher untuk Linux).
  - **Menu Bar / System Tray Support (macOS)**: Berjalan di status bar atas dengan indikator titik hijau aktif, 1-click Connect/Disconnect langsung dari menu bar, dan fitur *close-to-tray* (koneksi tetap aktif di background saat jendela ditutup).
  - Otorisasi aman terintegrasi dengan Touch ID / Administrator Dialog macOS.
  - Import file `.conf` hanya dengan klik tombol file picker.
  - Tombol **Connect / Disconnect** besar dengan indikator status visual (Hijau / Merah).
  - Pemantauan metrik realtime: IP tunnel, transfer data RX/TX, handshake terakhir, dan ping latency.
- **⚙️ Server Setup Interaktif (Anti Ribet)**:
  - Cukup jalankan `sudo ezwg server` di VPS Linux Anda.
  - **Setup Wizard Otomatis**: Mendeteksi IP publik server, network interface (eth0/ens3), mengatur firewall NAT iptables, dan mengaktifkan IP forwarding secara otomatis.
  - **Interactive Management Menu**: Menambah client baru, menampilkan QR Code untuk HP, melihat daftar pengguna yang sedang online, atau mencabut akses user hanya dengan memilih menu angka (1-7).
- **📥 One-Click Installer**:
  - Skrip universal yang otomatis mendeteksi arsitektur (Apple Silicon M1/M2/M3/M4, Intel x86_64, Linux AMD64 & ARM64) dan menyiapkan seluruh dependensi WireGuard secara otomatis.
- **🩺 ezwg Doctor (`ezwg doctor`)**:
  - Diagnostik kesehatan sistem dan modul WireGuard dengan opsi perbaikan otomatis (`ezwg doctor --fix`).

---

## 📥 Cara Install Terpisah (Client vs Server)

### A. Install Client Desktop GUI (Untuk Laptop Mac / Linux Pengguna)

#### Opsi 1: Download Installer DMG (macOS - Paling Mudah)
Pengguna Mac cukup mengunduh file **`Easy-WireGuard.dmg`** dari [GitHub Releases](https://github.com/didikari/WireGuard/releases), buka filenya, lalu seret ke folder **Applications**.

#### Opsi 2: Menggunakan One-Liner Script (macOS & Linux)
Jalankan perintah ini di terminal laptop Anda:
```bash
curl -fsSL https://raw.githubusercontent.com/didikari/WireGuard/main/scripts/install-app.sh | bash
```
* **Di Mac**: Otomatis memasang **Easy WireGuard.app** ke folder `/Applications` dan menyiapkan dependensi WireGuard via Homebrew jika diperlukan.
* **Di Linux**: Otomatis memasang launcher **Easy WireGuard** ke Menu Aplikasi desktop.

---

### B. Install Server & CLI (Untuk Server VPS Linux)

Khusus di server VPS Linux (headless / tanpa tampilan grafis), pasang CLI & Server Suite ringan dengan satu baris perintah:

```bash
curl -fsSL https://raw.githubusercontent.com/didikari/WireGuard/main/scripts/install.sh | bash
```

Setelah selesai, langsung buka wizard kontrol server dengan:
```bash
sudo ezwg server
```

---

### 2. Install dari Source (Pengembang / Developer)

```bash
git clone https://github.com/didikari/WireGuard.git
cd ezwg
make build
make gui
sudo make install
```

---

## 💻 Panduan Penggunaan Client

### Opsi A: Menggunakan Tampilan GUI (Desktop)
1. Buka aplikasi **Easy WireGuard** dari Launchpad / Applications di Mac, atau Menu Aplikasi di Linux.
   *(Atau dari terminal cukup ketik: `ezwg gui`)*.
2. Klik tombol **`Import .conf`** dan pilih file konfigurasi WireGuard Anda.
3. Klik tombol **`Connect`**!
4. Status akan langsung berubah menjadi hijau, dan metrik kecepatan serta transfer data akan tampil secara realtime.

### Opsi B: Menggunakan Terminal (CLI)
```bash
# Import profil konfigurasi
ezwg client import ~/Downloads/kantor.conf --name kantor

# Konek ke VPN
ezwg connect kantor

# Cek status & latency
ezwg status

# Putuskan koneksi
ezwg disconnect
```

---

## 🖥️ Panduan Penggunaan Server (di VPS Linux)

Kini Anda tidak perlu lagi menghafal perintah atau konfigurasi manual di VPS.

### 1. Jalankan Server Wizard & Control Panel
Di server VPS Anda, cukup ketik:
```bash
sudo ezwg server
```

#### Jika Server Baru Pertama Kali Dijalankan:
Aplikasi akan membuka **Interactive Setup Wizard**:
```text
╔════════════════════════════════════════════════════════════╗
║          Easy WireGuard (ezwg) Server Setup Wizard         ║
╚════════════════════════════════════════════════════════════╝
Enter server Public IP or Domain [103.125.10.25]: <Tekan Enter>
Outbound Network Interface for NAT [eth0]: <Tekan Enter>
WireGuard UDP Listen Port [51820]: <Tekan Enter>
VPN IPv4 Subnet [10.8.0.1/24]: <Tekan Enter>
DNS Resolvers for Clients [1.1.1.1, 8.8.8.8]: <Tekan Enter>

Applying configuration...
🎉 WireGuard Server successfully initialized and running!
```

#### Setelah Server Aktif:
Menjalankan `sudo ezwg server` akan langsung membuka **Interactive Menu**:
```text
════════════════════════════════════════════════════════════
                 ezwg Server Control Panel                  
════════════════════════════════════════════════════════════
Interface: wg0 (● Running)  |  Registered Clients: 2

  1) ➕ Add a New Client
  2) 📋 List Clients & Realtime Traffic
  3) 📱 Show Client QR Code & Configuration
  4) ❌ Revoke / Delete a Client
  5) 🔄 Restart WireGuard Server
  6) 🛑 Stop WireGuard Server
  7) 🚪 Exit
────────────────────────────────────────────────────────────
Select an option [1-7]: 
```

* **Pilih 1**: Masukkan nama user (misal: `macbook-didik`). Sistem akan otomatis mengalokasikan IP VPN, membuat kunci enkripsi, dan langsung mencetak **QR Code** di terminal (bisa langsung di-scan pakai HP) serta menyimpan file `.conf`.
* **Pilih 2**: Menampilkan status online/idle, waktu koneksi terakhir, dan kuota upload/download masing-masing client.
* **Pilih 3**: Menampilkan kembali QR code dan file `.conf` client tertentu.
* **Pilih 4**: Mencabut akses VPN client secara instan tanpa perlu restart server.

---

## 🌐 Integrasi MikroTik (Remote MikroTik Tanpa IP Publik)

Anda dapat menghubungkan router **MikroTik (RouterOS v7)** ke server `ezwg` agar router cabang, kantor, atau rumah dapat di-remote via **WinBox**, **WebFig**, maupun **SSH** dari mana saja — **meskipun MikroTik berada di belakang NAT / ISP tanpa IP Publik (CGNAT seperti IndiHome, GSM 4G/5G, dsb.)**.

### Topologi Hub-and-Spoke
```text
[ Laptop / WinBox Anda ] (10.8.0.2)
          │
          ▼ (Easy WireGuard VPN)
[ Server ezwg Pusat ] (10.8.0.1) ◄─── IP Publik Server VPS
          ▲
          │ (WireGuard Tunnel via NAT)
[ MikroTik Remote / Cabang ] (10.8.0.5) ─── Tanpa butuh IP Publik!
```

### Langkah Setup MikroTik (RouterOS v7):

#### 1. Buat Peer untuk MikroTik di Server ezwg
Di server VPS Anda, jalankan:
```bash
sudo ezwg server
# Pilih menu 1 (Add a New Client) -> masukkan nama: mikrotik-cabang
```
Catat informasi yang dihasilkan:
- **Client IP**: misalnya `10.8.0.5/32`
- **PrivateKey**: kunci privat client
- **Server PublicKey**: kunci publik server
- **Server Endpoint**: `IP_SERVER:51820`

#### 2. Konfigurasi di Terminal MikroTik
Buka New Terminal di WinBox atau SSH MikroTik remote, lalu jalankan:

```routeros
# 1. Buat Interface WireGuard
/interface wireguard add name=wg-ezwg private-key="<PRIVATE_KEY_CLIENT>" listen-port=13231 mtu=1420

# 2. Pasang IP Address VPN pada interface
/ip address add address=10.8.0.5/24 interface=wg-ezwg

# 3. Tambahkan Peer ke Server Pusat
/interface wireguard peers add interface=wg-ezwg \
    public-key="<PUBLIC_KEY_SERVER>" \
    endpoint-address=<IP_PUBLIK_SERVER> \
    endpoint-port=51820 \
    allowed-address=10.8.0.0/24 \
    persistent-keepalive=25s
```

> [!NOTE]
> Parameter `persistent-keepalive=25s` sangat penting agar router MikroTik menjaga koneksi tunnel UDP tetap terbuka menembus firewall provider secara terus-menerus.

#### 3. Remote MikroTik dari Laptop Anda
1. Hubungkan VPN pada laptop menggunakan aplikasi **Easy WireGuard**.
2. Buka **WinBox**, lalu pada kolom *Connect To*, masukkan IP VPN MikroTik:
   ```text
   10.8.0.5
   ```
3. Masukkan user & password router MikroTik -> **Connect**. Router langsung terbuka secara instan dan aman!

---

## 🛠️ Perintah Diagnostik (`ezwg doctor`)

Untuk memeriksa kesiapan perangkat Anda (kernel module, DNS resolver, firewall iptables):
```bash
ezwg doctor

# Memperbaiki dependensi yang kurang secara otomatis:
ezwg doctor --fix
```

---

## 📄 Lisensi

MIT License © 2026 ezwg contributors.
