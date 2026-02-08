# OneSignal Setup & React Native Implementation Guide (Android)

Panduan langkah demi langkah untuk setup akun OneSignal, mendapatkan API Keys, dan implementasi Push Notification di aplikasi React Native untuk **Android**.

> **📱 Catatan iOS**: Push notification untuk iOS memerlukan Apple Developer Account ($99/tahun). Untuk development awal, fokus ke Android dulu yang 100% gratis. Kode yang kita buat sudah support iOS, tinggal tambahkan APNs certificate nanti saat sudah siap.

## Bagian 1: Setup Dashboard OneSignal (Android)

Lakukan ini untuk mendapatkan `App ID` dan `Rest API Key`.

1.  **Daftar / Login**:
    - Buka [onesignal.com](https://onesignal.com).
    - Sign Up (Gratis) atau Login.

2.  **Buat Aplikasi Baru**:
    - Klik tombol **"New App/Website"**.
    - **Name of App**: Isi nama aplikasi (misal: `WarungOS Mobile`).
    - **Select Organization**: Pilih organisasi default Anda.
    - Klik **Next**.

3.  **Pilih Platform**:
    - Pilih **"Google Android (FCM)"**.
    - Klik **Next**.

4.  **Konfigurasi Firebase (FCM)**:
    - OneSignal membutuhkan akses ke Firebase Cloud Messaging.
    - Buka [Firebase Console](https://console.firebase.google.com/).
    - Buat Project baru atau pilih yang sudah ada.
    - Masuk ke **Project Settings** -> **Service Accounts**.
    - Klik **"Generate New Private Key"** -> Ini akan mendownload file JSON.
    - Kembali ke OneSignal setup:
      - Upload file JSON tersebut di kolom **"Service Account JSON"**.
    - Klik **Save & Continue**.

5.  **Pilih Target SDK**:
    - Pilih **"React Native / Expo"**.
    - Klik **Save & Continue**.

6.  **Dapatkan App ID**:
    - Setelah setup FCM, Anda akan melihat **OneSignal App ID** di dashboard.
    - Copy dan simpan (Anda sudah punya: ``).

7.  **Dapatkan REST API Key** (PENTING untuk Backend):
    - Di dashboard OneSignal, klik menu **Settings** (navigasi atas).
    - Pilih **Keys & IDs**.
    - Cari bagian **REST API Key**.
    - Klik icon "eye" atau "copy" untuk melihat/copy key tersebut.
    - **JANGAN SHARE KEY INI KE PUBLIK!**

8.  **Update Backend**:
    - Buka file `.env` di backend Go Anda.
    - Isi kedua nilai:
      ```env
      ONESIGNAL_APP_ID=
      ONESIGNAL_API_KEY=PASTE_REST_API_KEY_DISINI
      ```

---

## Bagian 2: Implementasi Frontend (React Native)

Berikut adalah _best practices_ untuk implementasi di React Native menggunakan `react-native-onesignal`.

### 1. Instalasi Library

**Untuk Expo** (Managed Workflow):

```bash
npx expo install react-native-onesignal
```

**Untuk React Native CLI**:

```bash
npm install react-native-onesignal
# Android auto-link, tidak perlu langkah tambahan
```

> **Catatan**: Jika nanti mau support iOS, baru jalankan `cd ios && pod install`.

### 2. Inisialisasi (App.tsx)

Lakukan inisialisasi di root component (`App.tsx` atau `_layout.tsx`).

```tsx
import { LogLevel, OneSignal } from 'react-native-onesignal';
import { Constants } from 'expo-constants'; // Atau config environment Anda

// Inisialisasi di luar komponen
OneSignal.Debug.setLogLevel(LogLevel.Verbose); // Hapus saat production
OneSignal.initialize('MASUKKAN_ONESIGNAL_APP_ID_DISINI');

// Request Permission (Wajib untuk Android 13+)
OneSignal.Notifications.requestPermission(true);

export default function App() {
  // ...
}
```

### 3. Best Practice: Login User (Linking User)

Agar backend bisa mengirim notifikasi ke user tertentu (bukan broadcast ke semua), kita harus menghubungkan `UserID` dari database kita ke OneSignal.

Gunakan `OneSignal.login(externalId)` saat user berhasil login di aplikasi.

**File: `src/auth/AuthContext.tsx` atau saat Login Success:**

```tsx
const handleLoginSuccess = async (userData) => {
  // userData.id adalah UUID dari database PostgreSQL Anda
  OneSignal.login(userData.id);

  // Opsional: Kirim email/phone jika ingin notifikasi multi-channel
  // OneSignal.User.addEmail(userData.email);
};
```

**Saat Logout:**

```tsx
const handleLogout = () => {
  OneSignal.logout();
};
```

### 4. Handling Deep Links (Klik Notifikasi)

Ketika user mengklik notifikasi (misal: "Stok Rendah"), aplikasi harus membuka halaman yang relevan.

```tsx
import { useEffect } from 'react';
import { useRouter } from 'expo-router'; // Contoh pakai Expo Router

export default function RootLayout() {
  const router = useRouter();

  useEffect(() => {
    // Listener untuk klik notifikasi
    const clickListener = (event) => {
      const data = event.notification.additionalData;

      console.log('Notification Clicked:', data);

      // Data 'type' dikirim dari backend Go (notification_svc.go)
      if (data && data.type === 'low_stock') {
        // Navigasi ke detail produk
        router.push(`/inventory/products/${data.product_id}`);
      }

      if (data && data.type === 'new_transaction') {
        router.push(`/transactions/${data.transaction_id}`);
      }
    };

    OneSignal.Notifications.addEventListener('click', clickListener);

    return () => {
      // Cleanup listener
      OneSignal.Notifications.removeEventListener('click', clickListener);
    };
  }, []);

  return <Slot />;
}
```

---

## Ringkasan Alur Data

1.  **Trigger**: Backend mendeteksi stok habis.
2.  **Job Queue**: Backend memasukkan job ke Redis.
3.  **Worker**: Memproses job -> Mengambil UserID (Owner) -> Memanggil OneSignal API.
    - Backend mengirim payload: `include_external_user_ids: ["UUID-USER-OWNER"]`.
4.  **OneSignal**: Mencari device HP yang login dengan UUID tersebut.
5.  **Device**: Menerima notifikasi.
6.  **User Klik**: Aplikasi React Native menangkap event -> Membuka halaman Detail Produk.
