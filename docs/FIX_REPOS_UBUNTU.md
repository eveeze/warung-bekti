# 🛠️ Perbaikan Repository (Ubuntu Noble / Armbian)

Sistem Anda terdeteksi berbasis **Ubuntu 24.04 (Noble)**, bukan Debian. Karena itu perintah sebelumnya gagal (karena mencari repo Debian).

Silakan jalankan perintah perbaikan "Sapu Jagat" ini untuk memperbaiki repo Docker & Tailscale sekaligus:

```bash
# 1. Hapus config yang salah (bersih-bersih)
sudo rm /etc/apt/sources.list.d/docker.list
sudo rm /etc/apt/sources.list.d/tailscale.list

# --- PERBAIKAN DOCKER (Versi Ubuntu) ---

# Setup GPG Key Docker
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc

# Add Repo Docker (Paksa ke Ubuntu Noble)
echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu \
  noble stable" | \
  sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# --- PERBAIKAN TAILSCALE (Versi Ubuntu Noble) ---

# Add GPG Key Tailscale
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/noble.noarmor.gpg | sudo tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null

# Add Repo Tailscale
curl -fsSL https://pkgs.tailscale.com/stable/ubuntu/noble.tailscale-keyring.list | sudo tee /etc/apt/sources.list.d/tailscale.list

# 3. Update & Install Ulang
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin tailscale -y
```

Setelah ini sukses, baru jalankan:

```bash
sudo tailscale up
```
