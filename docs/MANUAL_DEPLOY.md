# 🚀 Deployment Manual (Jalur Cepat)

Karena disk STB terindikasi korup/unstable saat melakukan operasi berat (seperti `go mod download`), solusi terbaik adalah **Build di Laptop, Kirim Binary ke STB**.

Ini jauh lebih cepat dan hemat stress.

## 1. Di Laptop: Build Binary

Buka terminal laptop (di folder project), jalankan:

```bash
# 1. Pastikan sudah di folder project warung-backend
# 2. Build untuk arsitektur STB (ARM64)
make build-arm64
```

_Output: akan membuat file `bin/warung-api-arm64`._

## 2. Di Laptop: Kirim Binary

Kirim file tersebut ke STB menggunakan `scp`.
(Ganti `192.168.1.49` dengan IP STB Anda).

```bash
scp bin/warung-api-arm64 root@192.168.1.49:/mnt/data-warung/app/warung-bekti/warung-api
```

_(Masukkan password STB saat diminta)_.

## 3. Di STB: Update Dockerfile

Sekarang kita ubah Dockerfile agar **tidak perlu download/compile lagi**, cukup pakai file yang kita kirim tadi.

Masuk SSH STB:

```bash
ssh root@192.168.1.49
cd /mnt/data-warung/app/warung-bekti
```

Edit Dockerfile:

```bash
nano docker/Dockerfile
```

**Hapus SEMUA isinya**, ganti dengan yang simple ini:

```dockerfile
# Start from API image
FROM alpine:3.19

WORKDIR /app

# Install dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary yang dikirim dari laptop
COPY warung-api /app/api

# Copy migrations dan .env
COPY migrations /app/migrations
# COPY .env.production /app/.env (Opsional, biasanya dimount via docker-compose)

# User setup
RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup
USER appuser

EXPOSE 8080

CMD ["/app/api"]
```

_(Simpan: Ctrl+O, Enter, Ctrl+X)_.

## 4. Di STB: Deploy!

Sekarang jalankan lagi. Karena tidak ada download/compile, prosesnya akan sangat cepat (detik).

```bash
docker compose up -d --build
```

**Cek Log:**

```bash
docker compose logs -f api
```

Jika muncul `Starting server on 0.0.0.0:8080...` berarti **SUKSES!** 🎉
