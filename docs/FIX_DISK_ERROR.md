# 🚑 Perbaikan Disk Error (EXT4 Corruption)

Error `EXT4-fs error ... inode check` menandakan ada kerusakan struktur file (file system corruption) pada SSD Anda (`/dev/sda1`). Ini biasanya terjadi karena mati listrik mendadak, kabel kendor, atau reboot paksa saat disk sibuk.

## Solusi: Jalankan fsck (File System Check)

Anda perlu menjalankan "dokter disk" linux (`fsck`) untuk memperbaikinya.

### 1. Unmount Disk (Wajib)

Disk tidak boleh sedang dipakai saat diperbaiki.

```bash
# Pastikan tidak ada service yang pakai disk
sudo systemctl stop docker
sudo systemctl stop docker.socket

# Lepas mount
sudo umount /mnt/data-warung
# Jika error "target is busy", paksa dengan:
sudo umount -l /mnt/data-warung
```

### 2. Jalankan Perbaikan

Gunakan `fsck` dengan opsi `-y` (Auto Yes) agar memperbaiki semua error secara otomatis.

```bash
sudo fsck.ext4 -fv -y /dev/sda1
```

- `-f`: Force check (cek paksa meski terlihat bersih)
- `-v`: Verbose (tampilkan detail)
- `-y`: Yes to all repairs

Tunggu sampai proses selesai. Anda akan melihat banyak laporan "Fixing...", "Freeing...", atau "Recovering". Itu normal.

### 3. Mount Kembali & Cek

Setelah selesai dan bersih:

```bash
# Mount ulang
sudo mount -a

# Cek apakah bisa diakses
ls -l /mnt/data-warung/

# Nyalakan docker lagi
sudo systemctl start docker
```

---

## ⚠️ Jika STB Gagal Booting (Stuck di Layar Hitam)

Jika karena error ini STB Anda tidak mau masuk ke menu utama (stuck saat booting karena gagal mount `/etc/fstab`), Anda perlu masuk ke **Emergency Mode** atau **Recovery**:

1.  Jika punya akses keyboard di TV, masukkan password root jika diminta.
2.  Edit fstab untuk mematikan mount otomatis sementara:
    ```bash
    nano /etc/fstab
    ```
    Beri tanda `#` di depan baris UUID sda1:
    ```
    # UUID=xxxx-xxxx... /mnt/data-warung ...
    ```
3.  Reboot (`reboot`).
4.  Setelah masuk sistem normal, lakukan langkah perbaikan `fsck` di atas.
5.  Setelah sembuh, hilangkan tanda `#` di fstab.
