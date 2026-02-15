# 🛠️ Perbaikan Error Build Docker (Corrupt & Arch)

Error yang Anda alami (`expected 'package', found trucode`) menandakan 2 hal:

1.  **Cache Korup**: File library Go yang didownload rusak (efek disk error tadi).
2.  **Salah Arsitektur**: Dockerfile diset untuk `amd64` (Laptop/Server Biasa), padahal STB Anda `arm64`.

Silakan ikuti 3 langkah ini di terminal SSH:

## Langkah 1: Fix Arsitektur di Dockerfile

Kita harus ubah target build jadi `arm64`.

```bash
nano docker/Dockerfile
```

Cari baris ini (sekitar baris 17):

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
```

Ubah **`amd64`** menjadi **`arm64`**:

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build \
```

_(Simpan: Ctrl+O, Enter, Ctrl+X)_

## Langkah 2: Bersihkan Cache Rusak (PENTING)

Karena file download sebelumnya error/korup, kita harus hapus cache build docker.

```bash
docker builder prune -a -f
```

_(Tunggu sampai selesai)_.

## Langkah 3: Build Ulang (Fresh)

Sekarang build lagi dengan memaksa download ulang.

```bash
docker compose build --no-cache api
```

Jika berhasil (muncul `Writing image ... done`), baru jalankan semuanya:

```bash
docker compose up -d
```
