# 🧪 Panduan Verifikasi Deployment (STB)

Gunakan perintah-perintah ini di **Terminal STB** untuk memastikan API berjalan 100% normal.

## 1. Cek Health (Wajib OK)

```bash
curl -s https://api.warungmanto.store/health | grep "status"
```

Hardapan output: `{"status":"ok", ...}` atau `{"status":"degraded"}` (jika R2/Redis ada isu kecil).

## 2. Test Login (Dapatkan Token)

Ganti `admin` dan `password` dengan user yang ada di database migrasi Anda.

```bash
# Ganti email & password sesuai data asli di laptop Anda
curl -X POST https://api.warungmanto.store/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@warung.com", "password":"password"}' > token.json

# Lihat apakah dapat token (cari field "access_token")
cat token.json
```

Jika berhasil, outputnya JSON berisi `"access_token": "eyJhbGciOi..."`.

## 3. Simpan Token ke Variabel

Supaya tidak copas manual, jalankan ini (pastikan langkah 2 sukses dulu):

```bash
export TOKEN=$(cat token.json | grep -o '"access_token":"[^"]*' | cut -d'"' -f4)
echo "Token tersimpan: $TOKEN"
```

## 4. Test List Products (Data Migrasi)

Cek apakah 50.000 produk tadi muncul.

```bash
# List 5 Produk Aktif
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/products?page=1&per_page=5"

# List 5 Produk NON-AKTIF (Inactive)
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/products?is_active=false&page=1&per_page=5"
```

Harusnya muncul data produk yang `is_active: false`.

## 5. Test Search Produk (Cek Index DB)

Coba cari produk spesifik.

```bash
# Ganti 'kopi' dengan nama barang yang ada
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/products?search=kopi"
```

## 6. Test R2 (Gambar)

Jika produk punya gambar, cek URL-nya.

```bash
# Ambil satu URL gambar dari produk
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/products?per_page=1"
```

Coba buka URL gambar tersebut di browser laptop/HP Anda. Jika muncul, berarti R2 connect!

## 7. Test Categories (Kategori)

Cek endpoint kategori (Admin Only).

```bash
# List Kategori
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/categories"

# Buat Kategori Baru (Optional)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Minuman Segar"}' \
  "https://api.warungmanto.store/api/v1/categories"
```

## 8. Test Customers (Pelanggan)

Cek endpoint pelanggan (Cashier Access).

```bash
# List Pelanggan
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://api.warungmanto.store/api/v1/customers?page=1&per_page=5"

# Buat Pelanggan Baru (Optional)
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Budi Santoso", "phone":"081234567890"}' \
  "https://api.warungmanto.store/api/v1/customers"
```

## 9. Test Transactions (Transaksi)

Coba buat transaksi dummy (Pastikan ID produk valid dari langkah 4).

```bash
# Ganti "PRODUCT_ID_DARI_STEP_4" dengan UUID produk nyata
curl -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": null,
    "items": [
      {"product_id": "PRODUCT_ID_DARI_STEP_4", "quantity": 1}
    ],
    "payment_method": "cash",
    "paid_amount": 50000
  }' \
  "https://api.warungmanto.store/api/v1/transactions"
```

---

## ✅ Checklist Sukses

- [ ] Health Check OK
- [ ] Login Dapat Token (`admin@warung.com` / `password`)
- [ ] List Produk Muncul Data
- [ ] Search Berfungsi
- [ ] Kategori & Pelanggan Bisa Diakses
