# 🛠️ Perbaikan Error Docker Network (Veth/Bridge)

Error `operation not supported` saat membuat network bridge/veth menandakan kernel STB Anda mungkin belum memuat modul network yang dibutuhkan Docker.

## Langkah 1: Reboot (Wajib)

Karena Anda baru saja menginstall Docker dan mengubah repository, kernel seringkali butuh restart untuk memuat driver baru.

```bash
sudo reboot
```

Tunggu 1-2 menit, lalu SSH lagi. Coba jalankan build lagi. Jika masih gagal, lanjut ke Langkah 2.

## Langkah 2: Cek & Load Kernel Modules

Cek apakah modul `veth` (Virtual Ethernet) ada:

```bash
# Load modul manual
sudo modprobe veth
sudo modprobe overlay
sudo modprobe br_netfilter
sudo modprobe bridge

# Cek apakah berhasil di-load
lsmod | grep veth
```

- Jika tidak ada error (silent), berarti sukses. Coba build lagi.
- Jika error `modprobe: FATAL: Module veth not found`, berarti **Kernel STB Anda tidak Support Docker Network Bridge**.

## Langkah 3: Solusi Alternatif (Jika Langkah 2 Gagal)

Jika kernel tidak support veth, kita tidak bisa build image di dalam STB. Solusinya:

### Opsi A: Gunakan "Host Network" (Bypass Bridge)

Kita ubah `docker-compose.yml` agar menggunakan network host (nebeng network STB langsung), tapi ini hanya untuk _Runtime_. Untuk _Build_, `apk add` tetap butuh network.

Solusi paling ampuh untuk kasus ini adalah **Build di Laptop, Transfer Image ke STB**.

1.  **Di Laptop**, build image untuk ARM64:

    ```bash
    # Di folder project laptop
    docker buildx build --platform linux/arm64 -t warung-api:latest --load -f docker/Dockerfile .
    ```

2.  **Save image ke file:**

    ```bash
    docker save warung-api:latest | gzip > warung-api.tar.gz
    ```

3.  **Kirim ke STB:**

    ```bash
    scp warung-api.tar.gz root@192.168.1.xxx:/mnt/data-warung/app/
    ```

4.  **Di STB**, Load image:

    ```bash
    docker load < warung-api.tar.gz
    ```

5.  **Update `docker-compose.yml` di STB:**
    Ubah bagian `api` agar tidak `build: .` lagi, tapi pakai `image`.

    ```yaml
    # Edit docker-compose.yml
    services:
      api:
        image: warung-api:latest # <-- Pakai image yang sudah diload
        # build: ... (HAPUS BAGIAN BUILD)
        network_mode: 'host' # <-- Gunakan host network agar tidak perlu bridge
    ```

## Langkah 4: Troubleshooting Internet/DNS (Jika Client.Timeout)

Jika muncul error: `Client.Timeout exceeded while awaiting headers` saat pull image, artinya STB (atau Docker) tidak connect internet.

**Diagnosa (Jalankan di STB):**

1.  **Cek Internet Sistem:**

    ```bash
    ping -c 3 google.com
    ```

    - Jika **Reply**: Internet STB aman. Masalah ada di DNS Docker.
    - Jika **Unknown host**: Masalah DNS Sistem (`/etc/resolv.conf`).
    - Jika **Timeout/Unreachable**: Masalah kabel/koneksi.

2.  **Perbaikan DNS (Paling Ampuh):**
    Kita paksa STB menggunakan Google DNS.

    ```bash
    # Edit resolv.conf
    sudo nano /etc/resolv.conf
    ```

    Hapus isinya, ganti dengan:

    ```
    nameserver 8.8.8.8
    nameserver 1.1.1.1
    ```

    Simpan & Exit. Lalu restart docker:

    ```bash
    sudo systemctl restart docker
    ```

3.  **Coba Pull Lagi:**
    ```bash
    docker run --rm hello-world
    ```
