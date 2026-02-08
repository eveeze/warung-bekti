# 📦 Step-by-Step Implementation Guide: Deploy Backend POS "Warung Bekti" ke STB HG680P

> **Hardware**: STB HG680P (RAM 2GB, Internal 8GB) + MicroSD Class 10 (Min. 32GB, High Endurance recommended)
> **OS**: Armbian Server (Debian/Ubuntu based)
> **Stack**: Golang + PostgreSQL + Redis + Asynq + OneSignal

> **Note**: Penggunaan SD Card dipilih karena port USB STB sering tidak kuat mengangkat SSD eksternal (error -110/-62). Gunakan SD Card berkualitas tinggi (High Endurance/Industrial) agar awet.

---

## Daftar Isi

1. [Bab 1: Persiapan Hardware & Storage (SSD)](#bab-1-persiapan-hardware--storage-ssd)
2. [Bab 2: Environment & Docker Setup](#bab-2-environment--docker-setup)
3. [Bab 3: Networking & Security (Zero Cost)](#bab-3-networking--security-zero-cost)
4. [Bab 4: Implementasi Background Job & Notifikasi](#bab-4-implementasi-background-job--notifikasi)
5. [Bab 5: Deployment Workflow](#bab-5-deployment-workflow)

---

## Bab 1: Persiapan Storage (MicroSD)

### 1.1 Identifikasi SD Card

Masukkan SD Card, lalu cek device:

```bash
lsblk
# Output contoh:
# mmcblk1     179:0    0  29G  0 disk
# └─mmcblk1p1 179:1    0  29G  0 part
```

Usually SD Card terdeteksi sebagai `mmcblk1`. Pastikan bukan `mmcblk0` (ini biasanya internal eMMC).

### 1.2 Format SD Card ke ext4

```bash
# Unmount jika terpasang otomatis
sudo umount /dev/mmcblk1p1

# Format (HATI-HATI: data hilang)
# -L warung-data: Label partisi
sudo mkfs.ext4 -L warung-data /dev/mmcblk1p1
```

### 1.3 Buat Mount Point & Mount Permanen

```bash
# Buat direktori
sudo mkdir -p /mnt/data-warung

# Cari UUID
sudo blkid /dev/mmcblk1p1
# Output: /dev/mmcblk1p1: UUID="aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" TYPE="ext4"

# Edit fstab
sudo nano /etc/fstab
```

**Tambahkan baris berikut di `/etc/fstab`:**

```
UUID=aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee /mnt/data-warung ext4 defaults,noatime,nodiratime,commit=600 0 2
```

> **Optimasi SD Card**:
>
> - `noatime,nodiratime`: Jangan catat waktu akses file (mengurangi write).
> - `commit=600`: Flush data ke disk tiap 10 menit (mengurangi write frequency, tapi risiko data loss 10 menit terakhir jika mati lampu).

**Mount dan verifikasi:**

```bash
sudo mount -a
df -h /mnt/data-warung
```

### 1.4 Pindahkan Docker Data Root ke SSD

Ini krusial agar image & container tidak memenuhi eMMC 8GB.

```bash
# Stop Docker
sudo systemctl stop docker

# Buat direktori Docker di SSD
sudo mkdir -p /mnt/data-warung/docker

# Pindahkan data existing (jika ada)
sudo rsync -aP /var/lib/docker/ /mnt/data-warung/docker/

# Konfigurasi Docker daemon
sudo nano /etc/docker/daemon.json
```

**Isi `/etc/docker/daemon.json`:**

```json
{
  "data-root": "/mnt/data-warung/docker",
  "storage-driver": "overlay2",
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
```

> **log-opts**: Mencegah log membengkak di storage terbatas.

**Restart Docker:**

```bash
sudo systemctl start docker

# Verifikasi
docker info | grep "Docker Root Dir"
# Output: Docker Root Dir: /mnt/data-warung/docker
```

---

## Bab 2: Environment & Docker Setup

### 2.1 Instalasi Docker & Docker Compose di Armbian

```bash
# Update sistem
sudo apt update && sudo apt upgrade -y

# Install dependencies
sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release

# Tambah Docker GPG key
curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Tambah Docker repository (Armbian berbasis Debian)
echo "deb [arch=arm64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Tambahkan user ke group docker (opsional, agar tidak perlu sudo)
sudo usermod -aG docker $USER
```

**Verifikasi:**

```bash
docker --version
docker compose version
```

### 2.2 Struktur Direktori Deployment

```bash
# Buat struktur folder di SSD
sudo mkdir -p /mnt/data-warung/{postgres,redis,app}
sudo chown -R $USER:$USER /mnt/data-warung
```

### 2.3 Docker Compose untuk Production (STB Optimized)

Buat file `/mnt/data-warung/docker-compose.yml`:

```yaml
services:
  # PostgreSQL Database - Volume di SD Card (Rekomendasi: High Endurance)
  postgres:
    image: postgres:16-alpine
    container_name: warung-postgres
    environment:
      POSTGRES_USER: warung
      POSTGRES_PASSWORD: ${DB_PASSWORD:-warung_secret_prod}
      POSTGRES_DB: warung_db
    volumes:
      - /mnt/data-warung/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ['CMD-SHELL', 'pg_isready -U warung -d warung_db']
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    # Tidak expose port ke host, hanya internal network
    networks:
      - warung-internal

  # Redis - Optimasi untuk RAM 2GB
  redis:
    image: redis:7-alpine
    container_name: warung-redis
    command: >
      redis-server
      --appendonly yes
      --maxmemory 256mb
      --maxmemory-policy allkeys-lru
      --save 900 1
      --save 300 10
    volumes:
      - /mnt/data-warung/redis:/data
    healthcheck:
      test: ['CMD', 'redis-cli', 'ping']
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - warung-internal

  # Warung Backend API
  api:
    image: warung-api:latest
    container_name: warung-api
    env_file:
      - .env.production
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    ports:
      - '127.0.0.1:8080:8080' # Hanya localhost, akses via Cloudflare Tunnel
    restart: unless-stopped
    networks:
      - warung-internal
    deploy:
      resources:
        limits:
          memory: 512M # Limit memory untuk stabilitas
        reservations:
          memory: 256M

networks:
  warung-internal:
    driver: bridge
```

### 2.4 File Environment Production

Buat file `/mnt/data-warung/.env.production`:

```bash
# Application
APP_ENV=production
LOG_LEVEL=info

# Server
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# PostgreSQL (internal network)
DB_HOST=postgres
DB_PORT=5432
DB_USER=warung
DB_PASSWORD=GANTI_PASSWORD_YANG_KUAT
DB_NAME=warung_db
DB_SSL_MODE=disable
DB_MAX_OPEN_CONNS=10
DB_MAX_IDLE_CONNS=5

# Redis (internal network)
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# R2/Storage (sesuaikan dengan config Anda)
R2_ACCOUNT_ID=your_account_id
R2_ACCESS_KEY_ID=your_access_key
R2_ACCESS_KEY_SECRET=your_secret_key
R2_BUCKET_NAME=warung-assets
R2_PUBLIC_DOMAIN=https://assets.warungmanto.store

# JWT
JWT_SECRET=GANTI_DENGAN_SECRET_32_KARAKTER_RANDOM
JWT_EXPIRATION_HOURS=24

# OneSignal
ONESIGNAL_APP_ID=your_onesignal_app_id
ONESIGNAL_API_KEY=your_onesignal_rest_api_key
```

### 2.5 Optimasi Redis untuk RAM 2GB

Penjelasan konfigurasi Redis:

| Parameter          | Nilai                                    | Alasan                                         |
| ------------------ | ---------------------------------------- | ---------------------------------------------- |
| `maxmemory`        | 256mb                                    | Menyisakan RAM untuk PostgreSQL & Go API       |
| `maxmemory-policy` | allkeys-lru                              | Hapus key paling jarang diakses jika penuh     |
| `appendonly`       | yes                                      | Persistence, mencegah data hilang saat restart |
| `save 900 1`       | Snapshot tiap 15 menit jika ada 1 change | RDB backup periodik                            |

---

## Bab 3: Networking & Security (Zero Cost)

### 3.1 Instalasi Cloudflare Tunnel (cloudflared)

Cloudflare Tunnel memungkinkan expose API ke internet **tanpa IP Publik** dan **gratis**.

```bash
# Download cloudflared untuk ARM64
curl -L https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-arm64 -o cloudflared
chmod +x cloudflared
sudo mv cloudflared /usr/local/bin/

# Verifikasi
cloudflared --version
```

### 3.2 Login & Buat Tunnel

```bash
# Login ke Cloudflare (buka URL di browser)
cloudflared tunnel login

# Buat tunnel baru
cloudflared tunnel create warung-api

# Output: Tunnel credentials written to /home/user/.cloudflared/xxxxxxxx.json
# Catat TUNNEL_ID dari output
```

### 3.3 Konfigurasi Tunnel

```bash
# Buat config
mkdir -p ~/.cloudflared
nano ~/.cloudflared/config.yml
```

**Isi `/home/user/.cloudflared/config.yml`:**

```yaml
tunnel: <TUNNEL_ID>
credentials-file: /home/<user>/.cloudflared/<TUNNEL_ID>.json

ingress:
  - hostname: api.warungmanto.store
    service: http://localhost:8080
  - service: http_status:404
```

### 3.4 Setup DNS di Cloudflare Dashboard

```bash
# Route DNS ke tunnel
cloudflared tunnel route dns warung-api api.warungmanto.store
```

### 3.5 Jalankan Tunnel sebagai Service

```bash
# Install sebagai systemd service
sudo cloudflared service install

# Enable dan start
sudo systemctl enable cloudflared
sudo systemctl start cloudflared

# Cek status
sudo systemctl status cloudflared
```

> **Benefit**: API sekarang bisa diakses via `https://api.warungmanto.store` dengan SSL otomatis dari Cloudflare.

### 3.6 Setup Tailscale untuk Remote Access

Tailscale digunakan untuk akses SSH dan remote database secara aman (private network).

```bash
# Install Tailscale
curl -fsSL https://tailscale.com/install.sh | sh

# Authenticate
sudo tailscale up

# Output: URL untuk login, buka di browser
```

**Setelah connect, STB akan dapat IP Tailscale (misal: `100.64.x.x`)**

### 3.7 Enable SSH via Tailscale Only (Opsional tapi Recommended)

Edit SSH config untuk hanya menerima koneksi dari Tailscale:

```bash
sudo nano /etc/ssh/sshd_config
```

Tambahkan:

```
ListenAddress 100.64.0.0/10
```

```bash
sudo systemctl restart sshd
```

> **Security**: SSH sekarang hanya bisa diakses dari device yang sudah join Tailscale network Anda.

### 3.8 Remote Database via Tailscale

Dari laptop/HP yang sudah join Tailscale:

```bash
# Connect ke PostgreSQL di STB
psql -h 100.64.x.x -U warung -d warung_db

# Atau gunakan GUI seperti DBeaver dengan host: 100.64.x.x
```

---

## Bab 4: Implementasi Background Job & Notifikasi

### 4.1 Arsitektur Asynq

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   API       │───▶│   Redis     │◀───│   Worker    │
│ (Enqueue)   │    │   (Queue)   │    │ (Process)   │
└─────────────┘    └─────────────┘    └─────────────┘
                                             │
                                             ▼
                                      ┌─────────────┐
                                      │  OneSignal  │
                                      │  Push API   │
                                      └─────────────┘
```

### 4.2 Queue Client (Enqueue Tasks)

File: `internal/platform/queue/client.go`

```go
package queue

import (
    "encoding/json"
    "github.com/hibiken/asynq"
)

// Task Types
const (
    TypeLowStockAlert    = "notification:low_stock"
    TypeNewTransaction   = "notification:new_transaction"
    TypeNotificationSend = "notification:send"
)

// Payloads
type PayloadLowStock struct {
    ProductID   string `json:"product_id"`
    ProductName string `json:"product_name"`
    Stock       int    `json:"stock"`
    UserID      string `json:"user_id"`
}

type PayloadNewTransaction struct {
    TransactionID string  `json:"transaction_id"`
    Total         float64 `json:"total"`
    UserID        string  `json:"user_id"`
}

type Client struct {
    client *asynq.Client
}

func NewClient(redisAddr string, redisPassword string) *Client {
    return &Client{
        client: asynq.NewClient(asynq.RedisClientOpt{
            Addr:     redisAddr,
            Password: redisPassword,
        }),
    }
}

func (c *Client) EnqueueLowStockAlert(payload PayloadLowStock) error {
    data, _ := json.Marshal(payload)
    task := asynq.NewTask(TypeLowStockAlert, data)
    _, err := c.client.Enqueue(task, asynq.ProcessIn(0))
    return err
}
```

### 4.3 Queue Server (Worker)

File: `internal/platform/queue/server.go`

```go
package queue

import (
    "context"
    "log"
    "github.com/hibiken/asynq"
)

type Server struct {
    server *asynq.Server
    mux    *asynq.ServeMux
}

func NewServer(redisAddr, redisPassword string, concurrency int) *Server {
    srv := asynq.NewServer(
        asynq.RedisClientOpt{Addr: redisAddr, Password: redisPassword},
        asynq.Config{
            Concurrency: concurrency,  // Untuk STB: 2-4 workers
            Queues: map[string]int{
                "critical": 6,
                "default":  3,
                "low":      1,
            },
            ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
                log.Printf("ERROR: Task %s failed: %v", task.Type(), err)
            }),
        },
    )
    return &Server{server: srv, mux: asynq.NewServeMux()}
}

func (s *Server) Handle(pattern string, handler func(context.Context, *asynq.Task) error) {
    s.mux.HandleFunc(pattern, handler)
}

func (s *Server) Run() error {
    return s.server.Run(s.mux)
}
```

### 4.4 Struktur Tabel Notifications di PostgreSQL

Migration file: `migrations/020_create_notifications_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,  -- Nullable untuk system/broadcast notifications
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,  -- 'low_stock', 'new_transaction', 'system'
    data JSONB DEFAULT '{}',    -- Untuk deep-link data
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE  -- Soft delete
);

-- Index untuk query performa
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);
CREATE INDEX idx_notifications_type ON notifications(type);
```

### 4.5 OneSignal Client untuk Push Notification

File: `internal/integration/onesignal/client.go`

```go
package onesignal

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type Client struct {
    AppID      string
    APIKey     string
    httpClient *http.Client
}

func NewClient(appID, apiKey string) *Client {
    return &Client{
        AppID:      appID,
        APIKey:     apiKey,
        httpClient: &http.Client{Timeout: 10 * time.Second},
    }
}

type NotificationRequest struct {
    AppID                string                 `json:"app_id"`
    IncludeExternalIDs   []string               `json:"include_external_user_ids,omitempty"`
    IncludedSegments     []string               `json:"included_segments,omitempty"`
    Headings             map[string]string      `json:"headings"`
    Contents             map[string]string      `json:"contents"`
    Data                 map[string]interface{} `json:"data,omitempty"`
}

// SendNotification - Kirim ke specific user atau broadcast
// externalUserIDs adalah UUID user dari database kita
func (c *Client) SendNotification(title, message string, externalUserIDs []string, data map[string]interface{}) error {
    reqBody := NotificationRequest{
        AppID:    c.AppID,
        Headings: map[string]string{"en": title, "id": title},
        Contents: map[string]string{"en": message, "id": message},
        Data:     data,
    }

    if len(externalUserIDs) > 0 {
        reqBody.IncludeExternalIDs = externalUserIDs
    } else {
        reqBody.IncludedSegments = []string{"All"}
    }

    jsonBody, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", "https://onesignal.com/api/v1/notifications", bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Basic "+c.APIKey)

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode >= 400 {
        return fmt.Errorf("onesignal api error: status %d", resp.StatusCode)
    }
    return nil
}
```

### 4.6 Contoh: Handler Low Stock dengan Deep-Link Data

```go
func (s *NotificationService) HandleLowStockTask(ctx context.Context, t *asynq.Task) error {
    var payload queue.PayloadLowStock
    if err := json.Unmarshal(t.Payload(), &payload); err != nil {
        return fmt.Errorf("unmarshal failed: %w", asynq.SkipRetry)
    }

    title := "⚠️ Stok Rendah"
    message := fmt.Sprintf("Produk %s tinggal %d unit. Segera restock!",
        payload.ProductName, payload.Stock)

    // Data untuk deep-link di React Native
    data := map[string]interface{}{
        "type":       "low_stock",
        "product_id": payload.ProductID,
        "action":     "view_product",  // Frontend akan navigate ke product detail
    }

    // Simpan ke database
    notification := &domain.Notification{
        UserID:  &payload.UserID,
        Title:   title,
        Message: message,
        Type:    "low_stock",
        Data:    data,
    }
    s.repo.Create(ctx, notification)

    // Kirim push notification ke OneSignal
    if s.oneSignal != nil {
        s.oneSignal.SendNotification(title, message, []string{payload.UserID}, data)
    }

    return nil
}
```

---

## Bab 5: Deployment Workflow

### 5.1 Cross-Compile dari Laptop ke ARM64

#### Dari Windows (PowerShell/CMD)

```powershell
# Set environment variables
$env:GOOS="linux"
$env:GOARCH="arm64"
$env:CGO_ENABLED="0"

# Build binary
go build -ldflags="-w -s" -o warung-api-arm64 ./cmd/api

# Reset environment (untuk development lokal)
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

#### Dari Linux/Arch Linux

```bash
# One-liner cross compile
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o warung-api-arm64 ./cmd/api

# Atau tambahkan ke Makefile:
build-arm64:
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o bin/warung-api-arm64 ./cmd/api
```

### 5.2 Transfer Binary ke STB

```bash
# Via Tailscale (aman, private network)
scp warung-api-arm64 user@100.64.x.x:/mnt/data-warung/app/

# Atau via SCP biasa jika SSH masih terbuka
scp warung-api-arm64 user@<stb-ip>:/mnt/data-warung/app/
```

### 5.3 Build Docker Image di STB

Jika ingin menggunakan Docker (recommended):

```bash
# Di STB, buat Dockerfile khusus ARM64
cat > /mnt/data-warung/Dockerfile.arm64 << 'EOF'
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY warung-api-arm64 /app/api

RUN addgroup -g 1000 -S appgroup && \
    adduser -u 1000 -S appuser -G appgroup

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/api"]
EOF

# Build image
cd /mnt/data-warung/app
docker build -f ../Dockerfile.arm64 -t warung-api:latest .
```

### 5.4 Startup Script

Buat `/mnt/data-warung/start.sh`:

```bash
#!/bin/bash
set -e

echo "🚀 Starting Warung Backend..."

cd /mnt/data-warung

# Start all services
docker compose up -d

# Wait for health checks
echo "⏳ Waiting for services to be healthy..."
sleep 10

# Verify
docker compose ps

echo "✅ All services started!"
echo ""
echo "📊 Health Check URLs:"
echo "   - API: http://localhost:8080/health"
echo "   - Via Tunnel: https://api.warungmanto.store/health"
```

```bash
chmod +x /mnt/data-warung/start.sh
```

### 5.5 Systemd Service untuk Auto-Start on Boot

```bash
sudo nano /etc/systemd/system/warung-backend.service
```

```ini
[Unit]
Description=Warung Backend Services
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/mnt/data-warung
ExecStart=/usr/bin/docker compose up -d
ExecStop=/usr/bin/docker compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable warung-backend
sudo systemctl start warung-backend
```

### 5.6 Health Check & Verifikasi

```bash
# Cek status container
docker compose ps

# Cek logs
docker compose logs -f api

# Test API health
curl http://localhost:8080/health

# Test via Cloudflare Tunnel
curl https://api.warungmanto.store/health

# Expected response:
# {"status":"ok","timestamp":"2026-02-08T14:00:00Z"}
```

### 5.7 Monitoring Commands

```bash
# Resource usage
docker stats

# Check Asynq queue
docker compose exec redis redis-cli KEYS "asynq:*"

# PostgreSQL connection test
docker compose exec postgres psql -U warung -d warung_db -c "SELECT 1;"

# View recent notifications
docker compose exec postgres psql -U warung -d warung_db -c "SELECT id, title, type, created_at FROM notifications ORDER BY created_at DESC LIMIT 5;"
```

---

## 📋 Quick Reference

### File Locations

| Path                                  | Deskripsi                |
| ------------------------------------- | ------------------------ |
| `/mnt/data-warung/`                   | Root directory di SSD    |
| `/mnt/data-warung/postgres/`          | PostgreSQL data          |
| `/mnt/data-warung/redis/`             | Redis data               |
| `/mnt/data-warung/docker-compose.yml` | Docker Compose config    |
| `/mnt/data-warung/.env.production`    | Environment variables    |
| `~/.cloudflared/config.yml`           | Cloudflare Tunnel config |

### Useful Commands

```bash
# Start everything
cd /mnt/data-warung && docker compose up -d

# Stop everything
docker compose down

# View logs
docker compose logs -f

# Restart API only
docker compose restart api

# Run database migration
docker compose exec api /app/api migrate up

# Backup database
docker compose exec postgres pg_dump -U warung warung_db > backup_$(date +%Y%m%d).sql
```

### Troubleshooting

| Issue                       | Solution                                          |
| --------------------------- | ------------------------------------------------- |
| Container restart loop      | Check `docker compose logs api` untuk error       |
| Database connection refused | Pastikan PostgreSQL healthy: `docker compose ps`  |
| Tunnel tidak connect        | Check `sudo systemctl status cloudflared`         |
| Out of memory               | Kurangi `maxmemory` Redis atau resource limit API |

---

> **📚 Dokumentasi Terkait:**
>
> - [Notifications API](./notifications/README.md)
> - [OneSignal Setup Guide](./notifications/ONESIGNAL_SETUP_GUIDE.md)
