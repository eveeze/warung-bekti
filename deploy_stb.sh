#!/bin/bash

# ==========================================
# Script Deploy API directly to STB
# ==========================================

STB_USER="root"
STB_IP="100.115.147.87"
IMAGE_NAME="warung-api:latest"
TAR_FILE="warung-api-stb.tar"

echo "========================================="
echo "🚀 BUILD & DEPLOY KE STB DIMULAI..."
echo "========================================="

# 1. Pastikan build platform adalah linux/arm64
echo ""
echo "[1/4] 🏗️ Membangun Docker Image (linux/arm64)..."
docker buildx build --platform linux/arm64 -t $IMAGE_NAME -f docker/Dockerfile .
if [ $? -ne 0 ]; then
    echo "❌ Build Gagal!"
    exit 1
fi

# 2. Save Image ke file tar
echo ""
echo "[2/4] 📦 Menyimpan Image ke file archive ($TAR_FILE)..."
docker save -o $TAR_FILE $IMAGE_NAME
if [ $? -ne 0 ]; then
    echo "❌ Gagal menyimpan image!"
    exit 1
fi

# 3. Transfer via SCP
echo ""
echo "[3/4] 🚚 Mentransfer file sebesar ~30MB ke STB ($STB_IP)..."
echo "Silakan masukkan password STB jika diminta."
scp $TAR_FILE $STB_USER@$STB_IP:~/$TAR_FILE
if [ $? -ne 0 ]; then
    echo "❌ Transfer Gagal!"
    exit 1
fi

# 4. Load & Run di STB
echo ""
echo "[4/4] 🔄 Memuat dan menjalankan container di STB..."
echo "Silakan masukkan password STB sekali lagi."
ssh $STB_USER@$STB_IP << 'ENDSSH'
    echo "-> Memuat image di STB..."
    docker load -i ~/warung-api-stb.tar
    
    echo "-> Menghapus file archive..."
    rm ~/warung-api-stb.tar
    
    echo "-> Pindah ke direktori project..."
    cd ~/warung-bekti || exit
    
    echo "-> Menghentikan container lama..."
    docker compose down
    
    echo "-> Menjalankan container baru (tanpa build ulang)..."
    # Re-create up without building since we loaded the image
    docker compose up -d
    
    echo "-> Membersihkan sisa docker lama..."
    docker system prune -f
    
    echo "✅ DEPLOYMENT SELESAI!"
ENDSSH

echo ""
echo "========================================="
echo "🎉 PROSES SELESAI!"
echo "Cek logs di STB dengan:"
echo "ssh root@100.115.147.87 'cd ~/warung-bekti && docker compose logs -f api'"
echo "========================================="
