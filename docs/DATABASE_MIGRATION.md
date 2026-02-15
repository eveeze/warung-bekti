# 📦 Panduan Migrasi Database (Laptop -> STB)

Panduan ini untuk memindahkan **SEMUA DATA** (Produk, Transaksi, User) dari database di Laptop Anda ke Database Production di STB.

## Syarat

1.  Laptop dan STB terhubung jaringan (WiFi sama atau Tailscale).
2.  Docker di STB sudah jalan (`warung-postgres` Up).

## Metode 1: Dump & Restore (Paling Aman)

Kita akan backup dulu database di laptop jadi file, kirim ke STB, lalu restore.

### Langkah 1: Backup dari Laptop

Buka terminal **di Laptop**, jalankan:

```bash
# Sesuaikan 'warung_db' dengan nama database di laptop Anda
pg_dump -h localhost -U postgres warung_db > backup_warung.sql
```

_(Masukkan password DB laptop jika diminta)._

### Langkah 2: Kirim ke STB

Gunakan `scp` untuk mengirim file `backup_warung.sql` ke STB.

```bash
# Ganti 192.168.1.49 dengan IP STB Anda
scp backup_warung.sql root@192.168.1.49:/mnt/data-warung/
```

### Langkah 3: Restore di STB

Masuk SSH ke STB, lalu jalankan perintah restore ini:

```bash
ssh root@192.168.1.49

# Masuk folder data
cd /mnt/data-warung

# RESTORE (Ini akan menimpa data di STB)
# 'warung-postgres' adalah nama container
# '-U warung' adalah user DB di container
# '-d warung_db' adalah nama database target

cat backup_warung.sql | docker exec -i warung-postgres psql -U warung -d warung_db
```

_(Tidak perlu password jika user 'warung' adalah owner DB)._

---

## Metode Manual (Paling Stabil ⭐️)

Jika cara "Direct Pipe" di atas gagal atau password terasa "stuck" (padahal itu cuma silent input), gunakan cara file ini.

### 1. Dump Data di Laptop

Jalankan di terminal laptop project Anda:

```bash
docker exec -i warung-postgres pg_dump -U warung --no-owner --no-acl warung_db > backup_warung.sql
```

_(File `backup_warung.sql` akan muncul di folder project)._

### 2. Upload ke STB

```bash
scp backup_warung.sql root@192.168.1.49:/mnt/data-warung/
```

### 3. Restore di STB

Masuk SSH STB, lalu jalankan:

```bash
# Hapus DB lama (biar bersih)
docker exec -i warung-postgres psql -U warung -d template1 -c "DROP DATABASE warung_db;"
docker exec -i warung-postgres psql -U warung -d template1 -c "CREATE DATABASE warung_db;"

# Restore
cat /mnt/data-warung/backup_warung.sql | docker exec -i warung-postgres psql -U warung -d warung_db
```

---

## Troubleshooting Error

### 1. "Role/User does not exist"

Jika `pg_dump` dari laptop membawa nama user (misal `postgres`) yang tidak ada di STB (user STB adalah `warung`), tambahkan opsi `--no-owner --no-acl` saat dump.

**Revisi Langkah 1:**

```bash
pg_dump -h localhost -U postgres --no-owner --no-acl warung_db > backup_warung.sql
```

### 2. "Database does not exist"

Pastikan container postgres di STB sudah jalan dan database `warung_db` sudah dibuat (biasanya otomatis oleh docker-compose).

### 3. Reset Total (Hapus Data Lama STB)

Jika restore error karena tabrakan data ("Duplicate Key"), hapus dulu data di STB:

```bash
# Di SSH STB:
docker exec -it warung-postgres psql -U warung -d template1 -c "DROP DATABASE warung_db;"
docker exec -it warung-postgres psql -U warung -d template1 -c "CREATE DATABASE warung_db;"
```

Lalu restore lagi langkah 3.
