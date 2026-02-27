# Landing Page Backend API Guide

Panduan lengkap untuk menggunakan public API endpoints backend Warung Manto pada landing page website. Endpoint ini **tidak memerlukan autentikasi** (no JWT token) dan hanya menampilkan data publik (tanpa informasi sensitif seperti `cost_price`, `current_stock`, dll).

---

## Base URL

```
Production: https://api.warungmanto.store
Local:      http://localhost:8080
```

---

## CORS

Backend sudah mengizinkan `Access-Control-Allow-Origin: *`, sehingga bisa dipanggil dari domain manapun.

---

## Available Endpoints

### 1. `GET /public/products` — List Produk Aktif

Mengembalikan produk-produk aktif dengan pagination, search, dan filter kategori.

#### Query Parameters

| Parameter     | Type   | Default | Deskripsi                                       |
| ------------- | ------ | ------- | ----------------------------------------------- |
| `search`      | string | —       | Cari berdasarkan nama produk (case-insensitive) |
| `category_id` | UUID   | —       | Filter berdasarkan ID kategori                  |
| `page`        | int    | 1       | Halaman (1-indexed)                             |
| `per_page`    | int    | 20      | Jumlah produk per halaman (max 100)             |

#### Contoh Request

```bash
# Semua produk (halaman pertama)
GET /public/products

# Search "indomie"
GET /public/products?search=indomie

# Filter kategori + pagination
GET /public/products?category_id=550e8400-e29b-41d4-a716-446655440000&page=2&per_page=10
```

#### Response Format

```json
{
  "status": "success",
  "message": "Products retrieved",
  "data": {
    "products": [
      {
        "id": "uuid-produk",
        "name": "Indomie Goreng",
        "description": "Mie instan goreng original",
        "unit": "pcs",
        "base_price": 3500,
        "image_url": "https://assets.warungmanto.store/products/indomie.jpg",
        "category": {
          "id": "uuid-kategori",
          "name": "Makanan"
        },
        "pricing_tiers": [
          {
            "name": "Grosir",
            "min_quantity": 10,
            "max_quantity": 49,
            "price": 3200
          },
          {
            "name": "Grosir Besar",
            "min_quantity": 50,
            "max_quantity": null,
            "price": 3000
          }
        ]
      }
    ],
    "total": 150,
    "page": 1,
    "per_page": 20
  }
}
```

#### Field Descriptions

| Field                          | Type    | Deskripsi                                              |
| ------------------------------ | ------- | ------------------------------------------------------ |
| `id`                           | UUID    | ID unik produk                                         |
| `name`                         | string  | Nama produk                                            |
| `description`                  | string? | Deskripsi produk (nullable)                            |
| `unit`                         | string  | Satuan produk: `pcs`, `kg`, `liter`, `pack`, dll       |
| `base_price`                   | int64   | Harga dasar dalam Rupiah (eceran)                      |
| `image_url`                    | string? | URL gambar produk (nullable, hosted di Cloudflare R2)  |
| `category`                     | object? | Kategori produk (nullable jika tanpa kategori)         |
| `category.id`                  | UUID    | ID kategori                                            |
| `category.name`                | string  | Nama kategori                                          |
| `pricing_tiers`                | array   | Daftar tier harga (bisa kosong jika hanya harga dasar) |
| `pricing_tiers[].name`         | string? | Nama tier: "Grosir", "Promo 3+", dll                   |
| `pricing_tiers[].min_quantity` | int     | Minimum jumlah untuk mendapatkan harga tier ini        |
| `pricing_tiers[].max_quantity` | int?    | Maksimum jumlah (null = unlimited)                     |
| `pricing_tiers[].price`        | int64   | Harga per unit pada tier ini (Rupiah)                  |

---

### 2. `GET /public/categories` — List Kategori Aktif

Mengembalikan semua kategori aktif yang tersedia.

#### Contoh Request

```bash
GET /public/categories
```

#### Response Format

```json
{
  "status": "success",
  "message": "Categories retrieved",
  "data": [
    {
      "id": "uuid-kategori-1",
      "name": "Makanan",
      "description": "Berbagai produk makanan"
    },
    {
      "id": "uuid-kategori-2",
      "name": "Minuman",
      "description": "Berbagai produk minuman"
    },
    {
      "id": "uuid-kategori-3",
      "name": "Kebutuhan Rumah Tangga",
      "description": null
    }
  ]
}
```

---

## Pricing Tiers Logic (Simulasi Keranjang)

Logika untuk menghitung harga berdasarkan jumlah pesanan. Ini HARUS diimplementasikan di frontend:

```javascript
/**
 * Menghitung harga per unit berdasarkan quantity dan pricing tiers.
 * Tier dipilih berdasarkan tier dengan min_quantity terbesar yang masih <= quantity.
 *
 * @param {Object} product - Produk dari API
 * @param {number} quantity - Jumlah yang dipesan
 * @returns {{ pricePerUnit: number, tierName: string }}
 */
function calculatePrice(product, quantity) {
  if (!product.pricing_tiers || product.pricing_tiers.length === 0) {
    return { pricePerUnit: product.base_price, tierName: 'Harga Dasar' };
  }

  // Sort descending by min_quantity to find the best (highest) matching tier
  const sortedTiers = [...product.pricing_tiers].sort(
    (a, b) => b.min_quantity - a.min_quantity,
  );

  for (const tier of sortedTiers) {
    if (quantity >= tier.min_quantity) {
      // Check upper bound if max_quantity exists
      if (tier.max_quantity !== null && quantity > tier.max_quantity) {
        continue;
      }
      return {
        pricePerUnit: tier.price,
        tierName: tier.name || 'Tier',
      };
    }
  }

  // No tier matches → use base price
  return { pricePerUnit: product.base_price, tierName: 'Harga Dasar' };
}

// Contoh penggunaan:
// const product = { base_price: 3500, pricing_tiers: [...] };
// const { pricePerUnit, tierName } = calculatePrice(product, 15);
// const subtotal = pricePerUnit * 15;
```

---

## Format Harga (Rupiah)

Semua harga dari API dalam satuan **Rupiah (integer, tanpa desimal)**. Format di frontend:

```javascript
function formatRupiah(amount) {
  return 'Rp ' + amount.toLocaleString('id-ID');
}
// formatRupiah(3500)  → "Rp 3.500"
// formatRupiah(50000) → "Rp 50.000"
```

---

## WhatsApp Checkout (Simulasi Belanja)

Setelah user mengisi keranjang, generate pesan WhatsApp:

```javascript
/**
 * Generate WhatsApp checkout URL dari cart items.
 *
 * @param {Array} cartItems - Array of { product, quantity, pricePerUnit, tierName }
 * @param {Object} customerInfo - { name: string, address: string } (optional)
 * @param {string} waNumber - Nomor WA toko (format: 6281234567890)
 * @returns {string} WhatsApp URL
 */
function generateWhatsAppURL(cartItems, customerInfo, waNumber) {
  let message = 'Halo Warung Manto! 👋\nSaya ingin memesan:\n\n';

  let grandTotal = 0;
  cartItems.forEach((item, index) => {
    const subtotal = item.pricePerUnit * item.quantity;
    grandTotal += subtotal;

    const tierLabel =
      item.tierName !== 'Harga Dasar' ? ` (${item.tierName})` : '';
    message += `${index + 1}. ${item.product.name} x${item.quantity}${tierLabel}`;
    message += ` — Rp ${subtotal.toLocaleString('id-ID')}\n`;
  });

  message += `\n*Total: Rp ${grandTotal.toLocaleString('id-ID')}*\n`;

  if (customerInfo?.name) {
    message += `\nNama: ${customerInfo.name}`;
  }
  if (customerInfo?.address) {
    message += `\nAlamat: ${customerInfo.address}`;
  }

  const encoded = encodeURIComponent(message);
  return `https://api.whatsapp.com/send?phone=${waNumber}&text=${encoded}`;
}

// Buka di tab baru:
// window.open(generateWhatsAppURL(cart, info, '6281234567890'), '_blank');
```

---

## Implementasi Fetch di Frontend

### Fetch Products

```javascript
const API_BASE = 'https://api.warungmanto.store'; // atau http://localhost:8080

async function fetchProducts({
  search = '',
  categoryId = '',
  page = 1,
  perPage = 20,
} = {}) {
  const params = new URLSearchParams();
  if (search) params.set('search', search);
  if (categoryId) params.set('category_id', categoryId);
  params.set('page', page);
  params.set('per_page', perPage);

  const res = await fetch(`${API_BASE}/public/products?${params}`);
  const json = await res.json();
  return json.data; // { products: [...], total, page, per_page }
}
```

### Fetch Categories

```javascript
async function fetchCategories() {
  const res = await fetch(`${API_BASE}/public/categories`);
  const json = await res.json();
  return json.data; // [ { id, name, description }, ... ]
}
```

---

## Cart State Management (localStorage)

Contoh implementasi cart yang persist di browser:

```javascript
const CART_KEY = 'warung_cart';

function getCart() {
  try {
    return JSON.parse(localStorage.getItem(CART_KEY)) || [];
  } catch {
    return [];
  }
}

function saveCart(cart) {
  localStorage.setItem(CART_KEY, JSON.stringify(cart));
}

function addToCart(product, quantity = 1) {
  const cart = getCart();
  const existing = cart.find((item) => item.productId === product.id);

  if (existing) {
    existing.quantity += quantity;
  } else {
    cart.push({
      productId: product.id,
      name: product.name,
      unit: product.unit,
      basePrice: product.base_price,
      imageUrl: product.image_url,
      pricingTiers: product.pricing_tiers || [],
      quantity: quantity,
    });
  }

  saveCart(cart);
}

function removeFromCart(productId) {
  const cart = getCart().filter((item) => item.productId !== productId);
  saveCart(cart);
}

function updateQuantity(productId, quantity) {
  const cart = getCart();
  const item = cart.find((i) => i.productId === productId);
  if (item) {
    item.quantity = Math.max(1, quantity);
  }
  saveCart(cart);
}

function clearCart() {
  localStorage.removeItem(CART_KEY);
}

function getCartTotal(cart) {
  return cart.reduce((total, item) => {
    const { pricePerUnit } = calculatePrice(
      { base_price: item.basePrice, pricing_tiers: item.pricingTiers },
      item.quantity,
    );
    return total + pricePerUnit * item.quantity;
  }, 0);
}
```

---

## Halaman & Sections yang Diperlukan

Landing page ini adalah **single-page website** dengan navigasi scroll-to-section. Berikut adalah struktur lengkap halaman dan apa yang harus ada di setiap section:

### Section 1: Navigation Bar (Fixed)

- Logo toko "Warung Manto" di kiri (teks saja, tidak perlu gambar logo)
- Menu navigasi di kanan: Beranda, Tentang, Belanja, Kontak
- Tombol CTA kecil: ikon keranjang dengan badge counter jumlah item
- Navbar **transparan** di hero, lalu berubah jadi **solid dengan blur/glass** saat scroll (sticky)
- Pada mobile: hamburger menu

### Section 2: Hero

- **Full-viewport height** (100vh)
- Background: gradient gelap (deep green ke hitam) atau bisa pakai gambar warung/toko dengan overlay gelap
- Typography hero yang **sangat besar** (display size, 5rem+):
  - Headline: "Warung Manto" atau "Belanja Mudah, Harga Bersahabat"
  - Sub-headline kecil di bawahnya yang menjelaskan lokasi/tagline toko
- Tombol CTA utama: "Mulai Belanja →" (scroll ke section belanja)
- Subtle scroll indicator (animated chevron di bawah)

### Section 3: Tentang Kami / About

- Layout: **split 2 kolom** (teks di kiri, visual/stats di kanan)
- Konten kiri:
  - Judul besar: "Tentang Kami"
  - Paragraf singkat tentang warung (2-3 kalimat): toko kelontong, melayani eceran dan grosir, buka setiap hari
- Konten kanan — stats/keunggulan dalam card/pill format:
  - "Harga Grosir Tersedia"
  - "Buka Setiap Hari"
  - "Pesan via WhatsApp"
- Section ini harus punya **padding besar** dan terasa lapang (whitespace)

### Section 4: Simulasi Belanja (Section Utama)

Ini section terpenting. Berisi:

**4a. Header Section**

- Judul besar: "Belanja" atau "Produk Kami"
- Search bar besar di tengah (full-width, prominent)
- Filter kategori: baris horizontal pill/chip buttons yang bisa diklik (data dari `GET /public/categories`). Ada chip "Semua" yang aktif by default.

**4b. Product Grid**

- Grid responsive: 4 kolom desktop, 3 tablet, 2 mobile
- Setiap product card berisi:
  - Gambar produk (dari `image_url`, pakai placeholder warna jika null)
  - Nama produk
  - Harga: tampilkan `base_price` diformat Rupiah
  - Jika ada `pricing_tiers`: tampilkan badge kecil "Harga Grosir" di pojok card
  - Unit: "/ pcs", "/ kg", dll
  - Tombol "+" untuk tambah ke keranjang
- **Infinite scroll** atau tombol "Load More" (pagination dari API)
- **Skeleton loading** saat fetch data
- Jika produk tidak ada: tampilkan empty state "Produk tidak ditemukan"

**4c. Cart Sidebar / Drawer**

- Muncul dari kanan saat tombol keranjang diklik
- Overlay gelap di belakang (backdrop)
- Isi sidebar:
  - Header: "Keranjang Belanja" + tombol close (×)
  - List item di keranjang:
    - Nama produk
    - Quantity stepper (- / jumlah / +)
    - Harga per unit (otomatis berubah berdasarkan tier quantity)
    - Jika harga berubah karena tier: tampilkan label tier (contoh: "Harga Grosir")
    - Subtotal per item
    - Tombol hapus
  - Divider
  - Grand total
  - Form sederhana: Nama (required), Alamat (optional)
  - Tombol besar: "Pesan via WhatsApp" (warna hijau WA)
  - Klik tombol → buka link `api.whatsapp.com` dengan pesan terformat

### Section 5: Kontak / Footer

- Background gelap (deep green/hitam)
- Layout 2-3 kolom:
  - Kolom 1: "Warung Manto" + alamat toko + jam operasional
  - Kolom 2: Quick links (Beranda, Belanja, Kontak)
  - Kolom 3: Kontak (WhatsApp link, telepon/HP)
- Copyright di bagian paling bawah

---

## Design Direction (SANGAT PENTING)

**JANGAN** buat desain yang terlihat seperti hasil generasi AI biasa (generic, Bootstrap-like, banyak warna norak, terlalu ramai). Desain harus terlihat seperti **website profesional dari Framer/Awwwards**.

### Visual Identity

| Aspek           | Spesifikasi                                                  |
| --------------- | ------------------------------------------------------------ |
| **Mood**        | Premium, clean, modern tapi warm (bukan corporate-cold)      |
| **Inspiration** | Framer templates, Awwwards nominees, editorial landing pages |
| **Approach**    | Minimalis tapi tidak kosong — setiap elemen punya purpose    |

### Color Palette

```
Primary:      #1B4332 (deep forest green — untuk heading, navbar, footer bg)
Secondary:    #2D6A4F (medium green — untuk hover states, accents)
Background:   #FAFAF7 (warm off-white — page background, bukan pure white)
Surface:      #FFFFFF (white — untuk cards)
Text Primary: #1A1A1A (near-black — heading text)
Text Body:    #4A4A4A (dark gray — body text)
Accent:       #D4A843 (warm gold — CTA buttons, badges, highlights)
Accent Hover: #C4922E (darker gold — hover state)
WhatsApp:     #25D366 (hijau WA — untuk tombol WhatsApp checkout)
Border:       #E8E8E4 (subtle warm gray — card borders, dividers)
```

### Typography

- **Heading font**: `Outfit` dari Google Fonts — weight 600-800
- **Body font**: `Inter` dari Google Fonts — weight 400-500
- **Scaling**: Gunakan fluid typography atau clamp()
  - Hero headline: `clamp(3rem, 8vw, 6rem)` — harus SANGAT besar
  - Section heading: `clamp(2rem, 4vw, 3.5rem)`
  - Card title: `1rem - 1.125rem`
  - Body/price: `0.875rem - 1rem`
- **Letter spacing**: Heading sedikit tight (-0.02em), body normal
- **Line height**: Heading 1.1, body 1.6

### Layout & Spacing

- **Max content width**: 1200px, centered
- **Section padding**: Minimal 80px-120px vertical
- **Generous whitespace** — biarkan elemen "bernapas"
- **Grid gap**: 20px-24px antar product cards
- **Border radius**: 12px-16px pada cards, 8px pada buttons, 100px pada pills/chips

### Micro-Interactions & Animations

- **Page load**: Fade-in + subtle slide-up pada setiap section saat scroll masuk viewport (Intersection Observer)
- **Navbar**: Transisi smooth dari transparan ke glass-blur saat scroll
- **Product cards**: `transform: scale(1.02)` dan shadow lebih dalam saat hover, transisi 0.3s ease
- **Cart sidebar**: Slide-in dari kanan dengan ease-out, plus backdrop fade
- **Buttons**: Subtle lift effect (translateY(-1px) + shadow pada hover)
- **Category pills**: Background-color transition smooth saat active/hover
- **Skeleton loading**: Shimmer/pulse animation saat loading produk
- **Quantity stepper**: Angka berubah dengan micro-fade
- **Scroll**: Smooth scroll (`scroll-behavior: smooth`) saat navigasi

### JANGAN Lakukan

- ❌ Jangan pakai warna-warna norak atau neon
- ❌ Jangan pakai gradient pelangi atau terlalu banyak warna
- ❌ Jangan pakai font default browser (Times New Roman, Arial)
- ❌ Jangan buat layout yang terlalu padat/crowded
- ❌ Jangan pakai border-radius terlalu besar (30px+) pada cards
- ❌ Jangan pakai stock photo generic — lebih baik pakai warna solid/gradient sebagai placeholder jika tidak ada gambar
- ❌ Jangan pakai ikon yang terlalu banyak warna — gunakan single-color icons (Lucide/Phosphor icons)
- ❌ Jangan pakai shadow yang terlalu tebal/gelap — subtle shadows saja

### HARUS Dilakukan

- ✅ Big typography yang bold pada hero dan section headings
- ✅ Banyak whitespace — ratio content vs space harus seimbang
- ✅ Konsisten: semua cards ukuran dan style sama
- ✅ Dark footer yang kontras dengan body page
- ✅ Smooth scroll dan transisi pada semua interaksi
- ✅ Loading state (skeleton) saat fetch data dari API
- ✅ Responsive sempurna: test di 320px (mobile kecil), 768px (tablet), 1024px+ (desktop)
- ✅ Accessible: warna kontras cukup, tombol cukup besar untuk touch, focus states pada semua interactive elements

---

## Tech Stack Rekomendasi untuk Frontend

Gunakan salah satu:

- **Vite + Vanilla JS + CSS** — paling simple, paling cepat
- **Next.js** — jika butuh SSR/SEO lebih bagus
- **Astro** — ideal untuk landing page statis dengan sedikit interaksi

Tidak perlu pakai UI library berat (Material UI, Ant Design, dll). Custom CSS/Tailwind saja sudah cukup.

---

## Ringkasan Perubahan Backend

File-file yang ditambahkan/dimodifikasi di backend untuk mendukung landing page:

| File                                   | Perubahan                                                                                                     |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `internal/handler/public_handler.go`   | **[NEW]** Handler untuk endpoint publik, menampilkan data produk/kategori tanpa info sensitif                 |
| `internal/repository/product_repo.go`  | **[MODIFIED]** Ditambahkan `PublicProductFilter` struct dan method `ListPublic()` yang join dengan categories |
| `internal/repository/category_repo.go` | **[MODIFIED]** Ditambahkan method `ListActive()` untuk mengambil semua kategori aktif                         |
| `internal/router/router.go`            | **[MODIFIED]** Ditambahkan 2 public routes: `GET /public/products` dan `GET /public/categories`               |
