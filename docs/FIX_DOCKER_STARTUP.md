# 🚑 Perbaikan Docker Gagal Start

Karena tadi sempat ada error disk dan repo, kita harus pastikan statusnya sekarang.

## Langkah 0: Cek Status (Lakukan via SSH)

Jalankan perintah ini satu per satu untuk melihat kondisi terakhir:

1.  **Cek Mount Point (Wajib ada):**

    ```bash
    df -h /mnt/data-warung
    ```

    - Jika muncul `/dev/sda1`, lanjut ke no 2.
    - Jika error/kosong, **STOP**. Docker tidak akan bisa jalan. Lakukan `sudo mount -a`.

2.  **Cek Service Docker:**
    ```bash
    sudo systemctl status docker
    ```

    - Jika **`Active: active (running)`**: Aman! Langsung tes `docker run hello-world`.
    - Jika **`inactive (dead)`**: Coba start: `sudo systemctl start docker`.
    - Jika **`failed`**: Lanjut ke **Langkah 1** di bawah.

## Langkah 1: Cek Penyebab Error (Jika Failed)

Jalankan perintah ini untuk melihat pesan error sebenarnya:

```bash
sudo journalctl -xu docker.service --no-pager | tail -n 20
```

Cari baris berwarna merah atau pesan seperti provided:

- `unable to configure the Docker daemon with file /etc/docker/daemon.json`: Salah ketik di config.
- `overlay2: ... not supported`: Masalah driver storage (efek disk error tadi).
- `failed to start daemon: ... address already in use`: Ada process docker nyangkut.

## Solusi Umum

### A. Reset Config daemon.json (Paling Sering)

Jika Anda sebelumnya mengedit `/etc/docker/daemon.json`, mungkin ada salah koma atau kurung. Coba reset ke default atau perbaiki.

```bash
# Backup dulu
sudo mv /etc/docker/daemon.json /etc/docker/daemon.json.bak

# Buat config minimal yang aman
sudo bash -c 'cat > /etc/docker/daemon.json <<EOF
{
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF'

# Coba start lagi
sudo systemctl restart docker
```

Note: Config di atas **sementara menghapus** setting `data-root` ke SSD. Jika docker berhasil start dengan ini, berarti masalahnya ada di **path SSD yang tidak mount** atau **salah ketik**.

### B. Reset Docker Data (Jika Disk Korup)

Jika errornya `graphdriver` atau `overlay2`, data docker lama mungkin rusak karena error disk tadi.

```bash
# Stop service
sudo systemctl stop docker
sudo systemctl stop docker.socket

# Rename folder data lama (mulai dari nol)
sudo mv /var/lib/docker /var/lib/docker.bak
# ATAU jika sudah pindah ke SSD:
sudo mv /mnt/data-warung/docker /mnt/data-warung/docker.bak

# Start lagi (akan buat folder baru bersih)
sudo systemctl start docker
```

### C. Cek Mount Point SSD

Pastikan SSD benar-benar termount sebelum Docker jalan (jika Anda pakai data-root di SSD).

```bash
df -h /mnt/data-warung
```

Jika kosong/tidak muncul, Docker akan gagal start karena foldernya hilang. Mount dulu: `sudo mount -a`.
