# WarungOS Backend

Backend untuk aplikasi kasir warung kelontong, dibangun dengan Golang vanilla (tanpa framework), PostgreSQL raw SQL, Redis, dan Cloudflare R2.

## Features

- **Product Management** dengan variable pricing (harga bertingkat: eceran, grosir, promo)
- **Point of Sale (POS)** dengan kalkulasi harga otomatis
- **Kasbon (Debt) System** untuk pelanggan dengan credit limit
- **Hybrid Stock Management** (tracked/untracked items)
- **Inventory Management** dengan stock movement audit trail
- **Reports & Analytics** (omzet harian, profit, top products)

## Tech Stack

- **Language**: Go 1.24+
- **Database**: PostgreSQL 16 (raw SQL, no ORM)
- **Cache**: Redis 7
- **Object Storage**: Cloudflare R2 (opsional untuk development lokal)
- **Containerization**: Docker & Docker Compose

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make (optional)

### 1. Setup development lokal

```bash
make setup-local
```

Perintah ini membuat `.env` jika belum ada, menyalakan PostgreSQL dan Redis lokal,
menunggu keduanya siap, menjalankan migrasi, membuat akun development, dan memasang
Air versi tetap ke `bin/air`. `.env` yang sudah ada tidak ditimpa. Kredensial database
di Compose mengikuti `.env`; mengubahnya setelah volume dibuat memerlukan perubahan
user/database di PostgreSQL juga.

### 2. Jalankan API dengan hot reload

```bash
make dev
```

Alamat default:

| Service | Alamat |
| --- | --- |
| API | `http://127.0.0.1:8080` |
| PostgreSQL | `127.0.0.1:5435` |
| Redis | `127.0.0.1:6381` |

Akun development: `admin@warung.com`, `cashier@warung.com`, dan
`inventory@warung.com`, semuanya dengan password `password`. Akun ini dibuat untuk
database development lokal. Data STB/Aiven tidak otomatis disalin.

### 3. Verifikasi dan kontrol service

```bash
curl http://127.0.0.1:8080/health
make migrate-status
make dev-down       # Hentikan infrastruktur; volume data tetap tersimpan
make dev-up         # Nyalakan lagi
make run            # Alternatif menjalankan API tanpa hot reload
make test-auto      # Test dengan database/Redis terpisah di port 5436/6382
```

Upload gambar membutuhkan kredensial `R2_*` di `.env`; jika kosong, storage tidak
diaktifkan dan fitur API lainnya bisa dikembangkan. Pembayaran Midtrans dan push
notification OneSignal membutuhkan kredensial sandbox/akun development masing-masing.
Untuk akses dari perangkat lain di jaringan lokal, ubah `SERVER_HOST=0.0.0.0` di `.env`
dan gunakan IP LAN komputer sebagai base URL.

### 4. Deployment saat ini (STB)

```bash
# Compose ini memakai .env.production dan image warung-api:latest
make docker-up

# View logs
make docker-logs
```

`docker-compose.yml` masih merupakan konfigurasi deployment STB dengan PostgreSQL
eksternal dan R2. Development lokal menggunakan `docker-compose.dev.yml`.

## API Endpoints

### Health Check
- `GET /health` - Health status of all services
- `GET /ready` - Readiness check
- `GET /live` - Liveness check

### Products
- `GET /api/v1/products` - List products
- `POST /api/v1/products` - Create product
- `GET /api/v1/products/{id}` - Get product by ID
- `PUT /api/v1/products/{id}` - Update product
- `DELETE /api/v1/products/{id}` - Delete product
- `GET /api/v1/products/search?barcode=xxx` - Search by barcode
- `GET /api/v1/products/low-stock` - Low stock products
- `POST /api/v1/products/{id}/pricing-tiers` - Add pricing tier

### Customers
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `GET /api/v1/customers/{id}` - Get customer
- `PUT /api/v1/customers/{id}` - Update customer
- `GET /api/v1/customers/{id}/kasbon` - Kasbon history
- `POST /api/v1/customers/{id}/kasbon/pay` - Record payment

### Transactions (POS)
- `POST /api/v1/transactions` - Create transaction (checkout)
- `POST /api/v1/transactions/calculate` - Calculate cart (preview)
- `GET /api/v1/transactions` - List transactions
- `GET /api/v1/transactions/{id}` - Get transaction
- `POST /api/v1/transactions/{id}/cancel` - Cancel transaction

### Inventory
- `POST /api/v1/inventory/restock` - Restock product
- `POST /api/v1/inventory/adjust` - Manual stock adjustment
- `GET /api/v1/inventory/{productId}/movements` - Stock movement history
- `GET /api/v1/inventory/report` - Inventory report

### Reports
- `GET /api/v1/reports/daily` - Daily sales report
- `GET /api/v1/reports/kasbon` - Outstanding debts report
- `GET /api/v1/reports/inventory` - Stock report
- `GET /api/v1/reports/dashboard` - Dashboard summary

## Variable Pricing Example

Produk dapat memiliki multiple pricing tiers:

```json
{
  "name": "Indomie Goreng",
  "base_price": 3500,
  "pricing_tiers": [
    {"name": "Eceran", "min_quantity": 1, "max_quantity": 2, "price": 3500},
    {"name": "Promo 3+", "min_quantity": 3, "max_quantity": 11, "price": 3000},
    {"name": "Grosir", "min_quantity": 12, "price": 2800}
  ]
}
```

Saat checkout 5 pcs → Tier "Promo 3+" aktif → 5 × Rp 3.000 = Rp 15.000

## Project Structure

```
warung-backend/
├── cmd/api/               # Application entry point
├── internal/
│   ├── config/            # Configuration management
│   ├── database/          # PostgreSQL, Redis, migrations
│   ├── domain/            # Domain models & entities
│   ├── handler/           # HTTP handlers
│   ├── middleware/        # HTTP middleware
│   ├── pkg/               # Shared utilities
│   ├── repository/        # Data access layer
│   ├── router/            # HTTP routing
│   ├── service/           # Business logic
│   └── storage/           # Minio client
├── migrations/            # SQL migration files
├── docker/                # Docker configurations
├── docker-compose.yml
├── Makefile
└── README.md
```

## Development

```bash
# Run tests with isolated local infrastructure
make test-auto

# Run with coverage
make test-coverage

# Lint code
make lint

# Build binary
make build
```

## License

MIT
