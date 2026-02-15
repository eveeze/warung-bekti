# 🚑 Perbaikan Redis "Unhealthy"

Jika Redis gagal start atau "Unhealthy", biasanya karena permission volume yang rusak (efek disk error kemarin) atau masalah kernel STB.

## Langkah 1: Cek Error Asli (Foreground)

Karena `logs` kosong, kita harus jalankan manual di depan layar untuk melihat errornya.

```bash
# Matikan dulu
docker compose down

# Jalankan Redis saja (tanpa -d)
docker compose up redis
```

**Perhatikan Outputnya:**

- Jika muncul: `Fatal error... Permission denied` -> Lanjut **Langkah 2**.
- Jika muncul: `Ready to accept connections` (diam tidak error), tapi tetap unhealthy nanti -> Lanjut **Langkah 3**.
- Jika error aneh lain, kirim fotonya.

Tekan `Ctrl+C` untuk matikan.

## Langkah 2: Fix Permission Volume

Kita reset permission folder datanya paksa menggunakan container lain.

```bash
# Jalankan perintah ajaib ini
docker run --rm -v warung-bekti_redis_data:/data alpine chown -R 999:999 /data
```

_(User ID 999 adalah user default Redis)._

Setelah itu, coba `docker compose up -d` lagi.

## Langkah 3: Disable Healthcheck (Jurus Terakhir)

Jika Redis sebenarnya jalan (bisa connect) tapi Docker menganggapnya sakit (healthcheck error), kita matikan saja healthcheck-nya.

Edit `docker-compose.yml`:

```bash
nano docker-compose.yml
```

Cari bagian `redis:` dan **HAPUS** (atau beri `#`) pada bagian `healthcheck:` sampai baris `retries`.

Dan **HAPUS** bagian `depends_on` di service `api`:

```yaml
api:
  # ...
  depends_on:
    postgres:
      condition: service_healthy
    # redis:          <-- HAPUS/KOMEN INI
    #   condition: ... <-- HAPUS/KOMEN INI
```

## Langkah 4: Fix Error Code 139 (Segmentation Fault)

Error `exited with code 139` adalah **Segmentation Fault**. Ini SANGAT UMUM terjadi di STB HG680P/B860H karena kernel Linux-nya versi lama dan tidak cocok dengan security profile Docker terbaru.

**Solusinya: Matikan Security Profile (Seccomp) untuk Redis.**

Edit `docker-compose.yml`:

```bash
nano docker-compose.yml
```

Tambahkan `security_opt` di bawah service `redis`:

```yaml
redis:
  image: redis:7-alpine
  container_name: warung-redis
  # ... settingan lain biarkan ...
  security_opt:
    - seccomp:unconfined
```

**Posisi yang benar (Perhatikan Spasi!):**

```yaml
redis:
  image: redis:7-alpine
  # ...
  volumes:
    - redis_data:/data
  security_opt: # <--- Tambah baris ini (sejajar dengan volumes)
    - seccomp:unconfined # <--- Tambah baris ini
  ports:
    - '6379:6379'
```

Setelah itu, **matikan total** dan **hapus container lama** agar config baru terbaca:

```bash
docker compose down
docker compose up -d
```

Dijamin Redis langsung sembuh! 💊
