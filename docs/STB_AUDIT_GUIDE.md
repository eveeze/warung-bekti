# 🔍 Panduan Audit & Pembersihan Sistem STB (Armbian)

Sebelum deploy sistem penting seperti Warung Bekti, sangat disarankan untuk melakukan **audit** terhadap sistem operasi bawaan STB, terutama jika STB dibeli dalam kondisi _pre-installed_. Hal ini untuk memastikan sistem dalam keadaan bersih dari bloatware, malware, atau konfigurasi yang tidak aman.

Jalankan perintah berikut melalui SSH di terminal STB Anda.

## 1. Cek Informasi Sistem

### Versi OS & Kernel

```bash
cat /etc/os-release
uname -a
```

_Pastikan Anda menggunakan versi Armbian/Debian/Ubuntu yang stabil (misal: Bullseye/Jammy)._

### Cek User Aktif

```bash
cat /etc/passwd | grep /home
```

_Hanya harus ada user `root` dan user Anda (misal: `warung` atau `opi`). Jika ada user asing, segera periksa._

## 2. Audit Software Terinstall

### Cek Paket yang Terinstall (APT)

```bash
dpkg --get-selections | grep -v deinstall
```

_Lihat daftar ini sekilas. Jika ada nama paket yang mencurigakan (seperti `miner`, `proxy`, atau tools hacking tak dikenal), catat untuk dihapus._

### Cek Snap Packages (Jika ada)

```bash
snap list
```

_Armbian server biasanya tidak pre-installed snap. Jika ada dan tidak digunakan, pertimbangkan remove untuk hemat RAM._

## 3. Cek Service yang Berjalan (PENTING)

### Service Aktif

```bash
systemctl list-units --type=service --state=running
```

_Perhatikan service yang tidak standar. Contoh service standar: `ssh`, `cron`, `getty`, `networking`, `systemd-journald`._

### Cek Port yang Terbuka (Listening Ports)

```bash
sudo ss -tulpn
# atau
sudo netstat -tulpn
```

_Output ini menunjukkan aplikasi apa saja yang sedang "mendengarkan" koneksi dari luar.
Contoh yang wajar: port `22` (SSH)._
_Jika ada port aneh (misal: `8080`, `3000`, `9000`) padahal Anda belum install apa-apa, waspadai._

## 4. Cek Penggunaan Resource Awal

### RAM & CPU

```bash
htop
```

_(Install dulu jika belum: `sudo apt install htop`)_
_Idealnya penggunaan RAM saat idle < 300MB._

### Disk Space

```bash
df -h
lsblk
```

_Pastikan `/` (rootfs) tidak penuh. Cek juga apakah ada partisi aneh yang ter-mount._

## 5. Pembersihan (Cleanup)

Jika Anda menemukan software yang tidak diinginkan, hapus dengan cara:

### Update Repository Dulu

```bash
sudo apt update
```

### Hapus Paket Tidak Perlu

```bash
sudo apt remove --purge <nama_paket>
sudo apt autoremove
sudo apt clean
```

### Kasus Khusus: Menghapus CasaOS

Jika STB Anda datang dengan **CasaOS**, disarankan untuk **MENGHAPUSNYA** demi menghemat RAM (sangat berharga di STB 2GB) dan menghindari konflik port. Backend kita akan berjalan "headless" (tanpa UI) agar performa maksimal.

**Cara Uninstall CasaOS:**

```bash
casaos-uninstall
# atau
curl -fsSL https://get.casaos.io/uninstall.sh | sudo bash
```

_Pastikan process `casaos` sudah tidak berjalan di `htop` setelah uninstall._

### Cek & Hapus Docker Container/Image Lama (Jika ada docker sebelumnya)

```bash
docker ps -a
docker images
docker system prune -a --volumes
```

## 6. Security Hardening Awal

### Ganti Password Default

Jika password masih bawaan penjual (misal: `1234`), **SEGERA GANTI**.

```bash
passwd
```

### Cek SSH Keys Asing

```bash
ls -la ~/.ssh/authorized_keys
cat ~/.ssh/authorized_keys
```

_Jika ada public key yang bukan milik Anda di file ini, HAPUS. Penjual mungkin meninggalkan key mereka untuk akses remote._

---

Setelah langkah-langkah ini selesai dan sistem dirasa bersih, Anda siap lanjut ke **Bab 1: Persiapan Hardware & Storage**.
