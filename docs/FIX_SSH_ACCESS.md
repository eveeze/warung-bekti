# 🚑 Perbaikan Akses SSH (Reset ke Default)

Jika Anda terkunci dari SSH (Connection Refused / Timeout), ikuti langkah ini langsung di **Terminal STB (Layar TV)** untuk mereset konfigurasi ke awal yang aman.

## Langkah 1: Reset Config SSH

Kita akan mengizinkan SSH dari mana saja (semua IP) untuk sementara, agar Anda bisa masuk dulu.

```bash
# 1. Edit file config
nano /etc/ssh/sshd_config
```

**Cari baris `ListenAddress` dan HAPUS atau beri tanda `#` di depannya.**

Contoh SALAH (Penyebab Error):

```
ListenAddress 100.115.147.87/10  <-- INI SALAH (Ada /10)
ListenAddress 100.115.147.87     <-- Ini benar tapi membatasi
```

**Ubah Menjadi (Default):**

```
#ListenAddress 0.0.0.0
#ListenAddress ::
```

_(Beri tanda `#` di depan semua baris ListenAddress, atau hapus barisnya. Ini akan membuat SSH listen di semua IP)._

Simpan: `Ctrl+O` -> Enter -> `Ctrl+X`.

## Langkah 2: Restart Service SSH (Penting!)

Di Ubuntu/Armbian, nama servicenya **`ssh`**, bukan `sshd`.

```bash
# Restart service
service ssh restart
# ATAU
systemctl restart ssh

# Cek status (harus "active (running)" warna hijau)
service ssh status
```

## Langkah 3: Cek IP Address

Pastikan Anda tahu IP STB yang benar.

```bash
ip addr | grep "inet 192"
```

Catat IP-nya (misal: `192.168.1.15`).

## Langkah 4: Login dari Laptop

Sekarang coba SSH dari laptop ke IP tersebut (bukan IP Tailscale dulu, tapi IP Lokal WiFi/LAN).

## Langkah 5: Restore Config (Jika File Kosong/Rusak)

Jika saat dibuka file `sshd_config` kosong melompong, kemungkinan file-nya korup. Tulis ulang 3 baris penting ini saja:

```bash
# 1. Hapus file lama
rm /etc/ssh/sshd_config

# 2. Buat baru (ketik manual pelan-pelan)
nano /etc/ssh/sshd_config
```

**Isi dengan 3 baris ini (WAJIB):**

```
Port 22
PermitRootLogin yes
PasswordAuthentication yes
```

_(Jangan tambah aneh-aneh dulu)_.

Simpan (`Ctrl+O` -> Enter -> `Ctrl+X`), lalu restart:

```bash
service ssh restart
```
123123123
## Langkah 6: Troubleshooting Koneksi (Jika Masih Gagal)

Jika Anda mengetik command SSH tapi error, cek pesan errornya:

### A. "Connection Timed Out" (Paling Sering)

Artinya laptop tidak bisa menghubungi STB.

1.  **Cek IP Real**: Jangan pakai IP contoh `192.168.1.15`! Cek IP asli di layar TV:
    ```bash
    ip addr | grep "inet 192"
    ```
    _(Angka di belakang `inet` adalah IP Anda, misal `192.168.1.8`)_
2.  **Satu Jaringan**: Pastikan Laptop dan STB connect ke WiFi/Router yang **SAMA**.
3.  **Ping**: Dari terminal laptop, coba `ping <IP_STB>`. Jika tidak reply, berarti beda jaringan.

### B. "Connection Refused"

Artinya STB menolak, biasanya service SSH belum jalan.

1.  Cek status di TV: `service ssh status`.
2.  Jika error config, ulangi **Langkah 5 (Restore Config)** di atas.

### C. "WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!"

Artinya laptop bingung karena "sidik jari" STB berubah (karena install ulang/reset).
**Solusi di Laptop:**

```bash
ssh-keygen -R <IP_STB>
# Contoh: ssh-keygen -R 192.168.1.8
```

Lalu coba SSH lagi.

---

**Command SSH yang Benar:**

```bash
ssh root@<IP_ASLI_DARI_TV>
```

Contoh: `ssh root@192.168.100.20`
_(Masukkan password root STB saat diminta)_
