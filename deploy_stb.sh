#!/bin/bash

# ==========================================
# Script Deploy API directly to STB (Fast & Safe)
# ==========================================

STB_USER="root"
STB_IP="100.115.147.87"
BINARY_NAME="warung-api"

echo "========================================="
echo "🚀 BUILD & DEPLOY KE STB DIMULAI..."
echo "Metode: Lokal Kompilasi + SCP Binary"
echo "========================================="

# 1. Kompilasi GO murni (linux/arm64)
echo ""
echo "[1/4] 🏗️ Build Binary Go untuk linux/arm64..."
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o $BINARY_NAME ./cmd/api
if [ $? -ne 0 ]; then
    echo "❌ Build Gagal!"
    exit 1
fi
echo "✅ Build Sukses: File $BINARY_NAME siap dikirim."

# 2. Transfer via SCP
echo ""
echo "[2/4] 🚚 Mengirim file Binary ke STB ($STB_IP)..."
echo "Silakan masukkan password STB jika diminta."
scp $BINARY_NAME $STB_USER@$STB_IP:~/warung-bekti/$BINARY_NAME
if [ $? -ne 0 ]; then
    echo "❌ Transfer Gagal!"
    exit 1
fi

# 3. Build Docker Image LOKAL di STB (Hanya menyalin binary, nggak makan RAM)
echo ""
echo "[3/4] 🔄 Mem-build ulang image ringan dan menjalankan container di STB..."
echo "Silakan masukkan password STB sekali lagi."
ssh $STB_USER@$STB_IP << 'ENDSSH'
    cd ~/warung-bekti || exit
    
    echo "-> Mengambil update terbaru dari repositori..."
    git pull origin main
    
    echo "-> Menyesuaikan permission execution..."
    chmod +x warung-api
    
    echo "-> Mem-build Docker image secara kilat menggunakan Dockerfile.release..."
    # Build image warung-api dari binary yang dikirim
    docker build -t warung-api:latest -f docker/Dockerfile.release .
    
    echo "-> Menghentikan container lama..."
    docker compose down
    
    echo "-> Menjalankan service dengan container baru..."
    docker compose up -d
    
    echo "-> Membersihkan image tak terpakai & binary instalasi..."
    docker image prune -f
    rm warung-api
    
    echo "✅ DEPLOYMENT SELESAI!"
ENDSSH

echo ""
echo "========================================="
echo "🎉 PROSES SELESAI!"
echo "Cek logs di STB dengan:"
echo "ssh root@100.115.147.87 'cd ~/warung-bekti && docker compose logs -f api'"
echo "========================================="
