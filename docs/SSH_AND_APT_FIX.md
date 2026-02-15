# 🔧 Panduan SSH & Perbaikan APT Error

## Bagian 1: Mengaktifkan & Mengakses SSH

Agar lebih nyaman (bisa copy-paste command), sebaiknya akses STB dari laptop via SSH.

### 1. Di Terminal STB (Layar TV):

Cek IP Address STB Anda:

```bash
ip addr show end0
# atau
ip addr show eth0
```

_Cari bagian `inet 192.168.1.xxx`. Catat IP-nya._

Misal IP-nya: **192.168.1.15**
192.168.1.49
### 2. Di Terminal Laptop (Windows/Mac/Linux):

Buka PowerShell / Terminal, lalu ketik:

```bash
ssh root@192.168.1.15
# atau jika user Anda bukan root (misal warung/opi)
ssh warung@192.168.1.15
```

Jika diminta password, masukkan password login STB Anda.

---

## Bagian 2: Memperbaiki Error APT (Docker Source List)l

Error `Malform entry 2 in list file /etc/apt/sources.list.d/docker.list` berarti ada kesalahan penulisan di file tersebut (mungkin karena copy-paste manual sebelumnya).

### Solusi Perbaikan:

Koneksi SSH dulu ke STB, lalu jalankan perintah ini untuk menghapus file yang rusak dan membuatnya ulang dengan benar.

**1. Hapus file yang rusak:**

```bash
sudo rm /etc/apt/sources.list.d/docker.list
```

**2. Tambahkan ulang repository Docker (Cara Paling Aman):**
Jangan edit manual. Gunakan perintah ini (copy-paste semua blok):

```bash
# Setup Docker's GPG key
sudo apt-get update
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
sudo chmod a+r /etc/apt/keyrings/docker.gpg

# Setup the repository
echo \
  "deb [arch="$(dpkg --print-architecture)" signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian \
  "$(. /etc/os-release && echo "$VERSION_CODENAME")" stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
```

**3. Update APT:**

```bash
sudo apt-get update
```

Jika tidak ada error lagi, silakan lanjutkan install docker:

```bash
sudo apt-get install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
```
