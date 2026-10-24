# ezwg (Easy WireGuard) 🚀

> **Aplikasi WireGuard Client GUI & Server Interaktif untuk Linux dan macOS — tinggal install langsung jalan!**
> 
> *The turnkey WireGuard Client Desktop GUI & Interactive Server Manager for Linux and macOS.*

---

## 🌟 Fitur Unggulan

- **🖥️ Client Desktop GUI (Tinggal Klik)**:
  - Tersedia aplikasi visual desktop (**Easy WireGuard.app** untuk macOS dan Desktop Launcher untuk Linux).
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

## 📥 Cara Install (Linux & macOS)

### 1. Menggunakan One-Line Installer (Rekomendasi)

Jalankan perintah ini di terminal **macOS** atau **Linux**:

```bash
curl -fsSL https://raw.githubusercontent.com/ezwg/ezwg/main/scripts/install.sh | bash
```

**Hasil instalasi otomatis:**
* **Di macOS**: Aplikasi **`Easy WireGuard.app`** otomatis terpasang di folder `/Applications` (bisa dibuka langsung dari Launchpad / Spotlight) serta CLI `ezwg` di terminal.
* **Di Linux**: Launcher desktop otomatis terpasang di menu aplikasi (Ubuntu Dash / GNOME / KDE) serta CLI `ezwg` di `/usr/local/bin`.

---

### 2. Install dari Source (Pengembang / Developer)

```bash
git clone https://github.com/ezwg/ezwg.git
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
