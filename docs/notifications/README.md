# Notifications API

Endpoint untuk mengelola notifikasi push dan riwayat notifikasi pengguna.

## Base URL

```
/api/v1/notifications
```

## Authentication

Semua endpoint memerlukan **Bearer Token** di header:

```
Authorization: Bearer <access_token>
```

---

## Endpoints

### 1. Get Notifications

Mengambil riwayat notifikasi untuk user yang sedang login.

**Endpoint:** `GET /api/v1/notifications`

**Query Parameters:**

- `limit` (optional, default: 20, max: 100): Jumlah notifikasi per halaman
- `offset` (optional, default: 0): Offset untuk pagination

**Response Success (200):**

```json
{
  "success": true,
  "message": "Notifications retrieved",
  "data": {
    "notifications": [
      {
        "id": "uuid",
        "user_id": "uuid",
        "title": "Low Stock Alert",
        "message": "Product Indomie is running low (4 left). Reorder soon!",
        "type": "low_stock",
        "data": {
          "product_id": "uuid",
          "current_stock": 4
        },
        "is_read": false,
        "created_at": "2026-02-04T16:20:24Z"
      }
    ],
    "limit": 20,
    "offset": 0
  }
}
```

**Notification Types:**

- `low_stock`: Notifikasi stok rendah
- `new_transaction`: Notifikasi transaksi baru
- `system`: Notifikasi sistem lainnya

---

### 2. Mark Notification as Read

Menandai satu notifikasi sebagai sudah dibaca.

**Endpoint:** `PATCH /api/v1/notifications/:id/read`

**Path Parameters:**

- `id` (required): UUID notifikasi

**Response Success (200):**

```json
{
  "success": true,
  "message": "Notification marked as read",
  "data": null
}
```

**Response Error (400):**

```json
{
  "success": false,
  "message": "Invalid notification ID"
}
```

---

### 3. Mark All as Read

Menandai semua notifikasi user sebagai sudah dibaca.

**Endpoint:** `PATCH /api/v1/notifications/read-all`

**Response Success (200):**

```json
{
  "success": true,
  "message": "All notifications marked as read",
  "data": null
}
```

---

## Frontend Integration Example

### Fetch Notifications

```typescript
const fetchNotifications = async (limit = 20, offset = 0) => {
  const response = await fetch(
    `/api/v1/notifications?limit=${limit}&offset=${offset}`,
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  );
  return response.json();
};
```

### Mark as Read

```typescript
const markAsRead = async (notificationId: string) => {
  await fetch(`/api/v1/notifications/${notificationId}/read`, {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
};
```

### Mark All as Read

```typescript
const markAllAsRead = async () => {
  await fetch('/api/v1/notifications/read-all', {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
};
```

---

## React Native Integration

### Using with OneSignal

Kombinasikan dengan OneSignal untuk real-time push notification:

1. **Receive Push** → OneSignal mengirim ke device
2. **User Click** → App membuka halaman detail
3. **Fetch History** → Call `GET /notifications` untuk history
4. **Mark as Read** → Call `PATCH /notifications/:id/read`

### Example Component

```tsx
import { useEffect, useState } from 'react';
import { OneSignal } from 'react-native-onesignal';

export function NotificationsScreen() {
  const [notifications, setNotifications] = useState([]);

  useEffect(() => {
    // Fetch notification history
    fetchNotifications().then(setNotifications);

    // Listen for new push notifications
    const listener = OneSignal.Notifications.addEventListener('click', (event) => {
      // Refresh notification list when user clicks push
      fetchNotifications().then(setNotifications);
    });

    return () => {
      OneSignal.Notifications.removeEventListener('click', listener);
    };
  }, []);

  const handleMarkAsRead = async (id: string) => {
    await markAsRead(id);
    // Refresh list
    fetchNotifications().then(setNotifications);
  };

  return (
    // Your UI here
  );
}
```
