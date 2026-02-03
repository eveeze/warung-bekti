# Optimistic UI Implementation Guide

Backend WarungOS telah dilengkapi dengan berbagai fitur untuk mendukung penerapan **Optimistic UI** di frontend. Dokumen ini menjelaskan mekanisme dan strategi yang dapat digunakan oleh tim Frontend (Mobile/Web) untuk menciptakan pengalaman pengguna yang instan dan responsif.

## 1. Overview Dukungan Backend

| Fitur                    | Deskripsi                                            | Kegunaan untuk Optimistic UI                                  |
| :----------------------- | :--------------------------------------------------- | :------------------------------------------------------------ |
| **Standard Response**    | Format JSON konsisten (`success`, `data`, `error`)   | Memudahkan _rollback_ state jika operasi gagal.               |
| **ETag Support**         | Header `ETag` pada semua response (GET/POST/PUT/dll) | Sinkronisasi cache lokal dengan server tanpa _over-fetching_. |
| **Server-Sent Events**   | Real-time stream via `/api/v1/events`                | _Background revalidation_ (single source of truth).           |
| **Precondition Helpers** | Error `412 Precondition Failed`                      | (Advanced) Mencegah konflik edit data (`If-Match`).           |

---

## 2. Strategi Implementasi

### A. Optimistic Updates (Mutations)

Saat user melakukan aksi (misal: "Edit Stok"), frontend tidak perlu menunggu loading.

1.  **Update UI Langsung**: Ubah state lokal seketika.
2.  **Kirim Request Background**: Kirim `POST`/`PUT` ke backend.
3.  **Tangkap Respon**:
    - **Sukses (200 OK)**: Backend akan mengembalikan data terbaru beserta header `ETag` baru. Update cache lokal dengan data & ETag tersebut.
    - **Gagal (4xx/5xx)**: Lakukan **Rollback** ke state sebelumnya dan tampilkan pesan error (Toast/Snackbar).

#### Contoh Flow (Update Product)

```javascript
/* Frontend Code Concept */

async function updateProduct(id, newData) {
  // 1. Snapshot previous state
  const previousData = queryClient.getQueryData(['product', id]);

  // 2. Optimistic Update
  queryClient.setQueryData(['product', id], { ...previousData, ...newData });

  try {
    // 3. Send Request
    const response = await api.put(`/products/${id}`, newData);

    // 4. Update with Server Data (Confirmation)
    // Backend returns the definitive state & new ETag
    const serverData = response.data.data;
    const newEtag = response.headers['etag'];

    // Update cache with confirmed data & ETag
    queryClient.setQueryData(['product', id], serverData, { etag: newEtag });
  } catch (error) {
    // 5. Rollback on Error
    queryClient.setQueryData(['product', id], previousData);
    showToast('Update gagal, data dikembalikan.');
  }
}
```

### B. Smart Revalidation (GET with ETag)

Setelah melakukan update optimis, atau saat user kembali ke layar tertentu, frontend harus memastikan data valid.

1.  Simpan `ETag` yang didapat dari request sebelumnya (baik dari GET atau POST/PUT).
2.  Saat melakukan request `GET` ulang, sertakan header `If-None-Match: <saved_etag>`.
3.  **Backend Behavior**:
    - Jika data **belum berubah**: Mengembalikan **`304 Not Modified`** (Body kosong). Frontend tetap gunakan data di cache. Hemat bandwidth & CPU.
    - Jika data **sudah berubah**: Mengembalikan **`200 OK`** dengan Body baru + `ETag` baru.

### C. Background Synchronization (SSE)

Untuk memastikan data "eventually consistent" tanpa user harus refresh manual.

1.  Connect ke endpoint SSE: `GET /api/v1/events`.
2.  Listen event spesifik (contoh: `stock_update`, `order_created`).
3.  Saat event diterima, lakukan "Soft Revalidation":
    - Tandai data di cache sebagai "stale" (kadaluarsa).
    - Trigger refetch background (menggunakan ETag).

```javascript
const evtSource = new EventSource('/api/v1/events');

evtSource.addEventListener('stock_update', function (event) {
  const data = JSON.parse(event.data);
  // Invalidasi cache produk yang berubah
  queryClient.invalidateQueries(['product', data.product_id]);
});
```

---

## 3. Struktur Error Standard

Jika update optimis gagal, backend mengirim error format standar yang dapat di-parse untuk menampilkan pesan spesifik ke user.

```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Stok tidak boleh negatif",
    "details": {
      "stock": "Minimal 0"
    }
  }
}
```

Gunakan `error.message` untuk toast, dan `error.details` untuk highlight field form yang salah.
