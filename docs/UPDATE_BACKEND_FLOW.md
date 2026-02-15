# 🔄 Panduan Update Backend (Laptop -> STB)

Ikuti flow ini jika Anda mengubah kode Go di laptop dan ingin update ke STB.

Metode ini menggunakan **Cross-Compilation** di laptop (cepat) lalu kirim binary jadi ke STB (ringan). STB tidak perlu compile ulang kode.

## Persiapan Awal (Sekali Saja)

Pastikan file `docker/Dockerfile.release` sudah ada di laptop Anda. Jika belum, buat isinya seperti ini:

_(File ini sudah otomatis dibuatkan oleh AI di folder `docker/Dockerfile.release`)_

---

## Langkah Rutin (Setiap Update)

### 1. Build Binary di Laptop

Jalankan perintah ini di terminal laptop (folder `warung-backend`):

```bash
# Compile kode Go menjadi binary Linux ARM64
make build-arm64
```

_(Hasilnya adalah file `bin/warung-api-arm64`)_

### 2. Kirim Binary & Dockerfile ke STB

Upload binary baru dan Dockerfile release ke STB.

```bash
# Kirim Binary (Rename jadi 'warung-api' saat sampai di sana)
scp bin/warung-api-arm64 root@192.168.1.49:/mnt/data-warung/app/warung-bekti/warung-api

# Kirim Dockerfile Release (Sekali saja cukup, kecuali ada update Dockerfile)
scp docker/Dockerfile.release root@192.168.1.49:/mnt/data-warung/app/warung-bekti/
```

### 3. Restart Service di STB

Login ke STB (SSH) dan jalankan:

```bash
ssh root@192.168.1.49

# Masuk folder app
cd /mnt/data-warung/app/warung-bekti

# 1. Build Image Baru (Cepat, cuma copy file)
docker build -f Dockerfile.release -t warung-api:latest .

# 2. Restart Container API (Otomatis pakai image baru)
docker compose up -d --force-recreate --no-deps api
```

### 4. Verifikasi

Cek logs untuk memastikan aplikasi jalan normal:

```bash
docker logs -f warung-api --tail 50
```

---

## ✅ Ringkasan Command (Cheat Sheet)

Kalau sudah biasa, cukup copy-paste 3 baris ini di terminal laptop:

```bash
make build-arm64
scp bin/warung-api-arm64 root@192.168.1.49:/mnt/data-warung/app/warung-bekti/warung-api
ssh root@192.168.1.49 "cd /mnt/data-warung/app/warung-bekti && docker build -f Dockerfile.release -t warung-api:latest . && docker compose up -d --force-recreate --no-deps api"
```

_(Refresh sambil ngopi ☕, update selesai dalam < 1 menit)_
