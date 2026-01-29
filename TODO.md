# Warung Backend Development Roadmap (TODO)

This document outlines the steps required to transform the current backend into a fully autonomous Point of Sale (POS) and Stock Management system.

## 1. Payment Integration (Midtrans QRIS) ✅
Integrasi pembayaran menggunakan **Midtrans** untuk QRIS statis maupun dinamis.
- [x] **Setup Midtrans**:
  - [x] Register Account & Get Server Key / Client Key.
  - [x] Configure `config.go` to load Midtrans credentials.
- [x] **Implement Payment Service**:
  - [x] Create `PaymentRepository` to store payment gateway responses (Snap token/redirect URL).
  - [x] Implement `GenerateSnapToken(transactionID, amount)` using Midtrans Go SDK.
  - [x] Create API Endpoint for **Midtrans Notification Webhook** to auto-update transaction status to `PAID`.
- [x] **Manual Verification Fallback**: Feature to manually mark as paid if webhook fails.

## 2. Advanced Stock Management ✅
To manage the store without help, the system must track inventory precisely.
- [x] **Stock Opname (Stock Taking)**:
  - [x] Create features to adjust physical stock vs system stock.
  - [x] Generate "Variance Report" (Lost/Stolen items).
- [x] **Low Stock Alerts**:
  - [x] Daily report of items below `min_stock_alert`.
  - [x] Auto-generate "Shopping List" (Restock Plan).
- [x] **Expiry Date Tracking (Optional)**:
  - [x] *Note: Only for perishable items (e.g., Bread, Milk) to avoid input fatigue.*
  - [x] Add `expiry_date` column to `stock_movements` (incoming) but make it optional.
  - [x] Create "Near Expiry" alerts only if date is set.

## 3. Financial & Cash Management ✅
Detailed tracking of money to ensure no leakage.
- [x] **Cash Flow / Petty Cash**:
  - [x] Track money taken out for operational costs (e.g., pay electricity, trash, buy plastic bags) directly from the drawer.
- [x] **Daily Profit/Loss**:
  - [x] Real-time profit calculation (Revenue - HPP).
  - [x] Sales report by payment method (Cash vs QRIS vs Kasbon).

## 4. Kasbon (Debt) Management System ✅
Critical for Warung Kelontong.
- [x] **Debt Limits**: Enforce `credit_limit` validation automatically at checkout.
- [x] **Billing/Reminder**: Generate a "Tagihan" list to send via WhatsApp to customers.
- [x] **Installment Payments**: Allow paying debt in parts (Cicilan) and track balance history.

## 5. POS Operational Features ✅
Features to speed up the cashier process.
- [x] **Barcode Scanner Support**: Verify API searches by barcode efficiently (Scanning must be instant).
- [x] **Hold/Resume Transaction**: Allow saving a cart temporarily if a customer leaves to pick up another item.
- [x] **Refund/Return**: Handling customers returning defective items.

## 6. System Reliability (No Parents Needed) ✅
- [x] **Database Backup**: Automated daily backup (e.g., to Google Drive or local storage cron job).
- [x] **Transaction Logs**: detailed logs of who modified what (Audit Trail).

## 7. Warung Special Features (Requested) ✅
- [x] **Consignment System (Titip Jual)**:
  - [x] **Supplier Tracking**: Track who owns the product (e.g., "Kue Basah Bu Tejo").
  - [x] **Commission Support**: Auto-calculate profit share (e.g., Warung takes 10%, Supplier 90%).
  - [x] **Settlement (Setoran)**: Generate report of how much to pay the supplier based on sales.
- [x] **Gas & Galon Tracking (Refillables)**:
  - [x] **Container Management**: Track "Tabung Kosong" vs "Isi".
  - [x] **Exchange Logic**: When selling "Gas Isi", system auto-adds 1 "Tabung Kosong" to inventory.
  - [x] **Restock Logic**: When buying "Gas Isi" from agent, system auto-deducts "Tabung Kosong".
- [x] **Wholesale Pricing (Grosir)**:
  - [x] **Tiered Pricing**: Automatic price adjustment based on quantity (e.g., Buy 10 = cheaper).



## 8. Product Image Management (MinIO) ✅
Implement image handling with best practices to save storage and ensure performance.
- [x] **MinIO Integration**:
  - [x] Setup MinIO Client / Service.
  - [x] Implement `UploadImage` with image compression/resizing (WebP format preferred).
  - [x] Implement `DeleteImage` to remove files from MinIO when product is deleted or image updated (Storage Optimization).
- [x] **Product API Update**:
  - [x] Update `Product` struct to include `image_url`.
  - [x] Update `Create/Update` endpoints to handle file uploads (multipart/form-data).


## 9. Performance Optimization (Redis) ✅
Use caching to speed up frequent read operations.
- [x] **Redis Integration**:
  - [x] Setup Redis Client.
  - [x] Implement Cache Service (Set, Get, Delete, Invalidate).
- [x] **Caching Targets**:
  - [x] **Products List**: Cache `GET /products` results (invalidate on Create/Update/StockChange).
  - [ ] **Dashboard Stats**: Cache daily dashboard metrics (invalidate on new transaction).
  - [ ] **Reports**: Cache heavy reports (e.g., Monthly Sales).


## 10. Restock PDF Generator 📄
Generate professional PDF for low-stock products to facilitate quick wholesale purchasing.
- [x] **PDF Generation**:
  - [x] Create `restock.go` with well-formatted table (Product Name, Stock, Min, Deficit, Suggested Order).
  - [x] Include store info, generation date, and summary.
  - [x] Smart order suggestions (rounded to 5/10 with safety margin).
- [x] **API Endpoint**:
  - [x] `GET /api/v1/inventory/restock-list/pdf` - Downloads PDF for low-stock products.
  - [x] Inventory role access control.

