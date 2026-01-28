# Warung Backend Development Roadmap (TODO)

This document outlines the steps required to transform the current backend into a fully autonomous Point of Sale (POS) and Stock Management system.

## 1. Payment Integration (Midtrans QRIS)
Integrasi pembayaran menggunakan **Midtrans** untuk QRIS statis maupun dinamis.
- [ ] **Setup Midtrans**:
  - [ ] Register Account & Get Server Key / Client Key.
  - [ ] Configure `config.go` to load Midtrans credentials.
- [ ] **Implement Payment Service**:
  - [ ] Create `PaymentRepository` to store payment gateway responses (Snap token/redirect URL).
  - [ ] Implement `GenerateSnapToken(transactionID, amount)` using Midtrans Go SDK.
  - [ ] Create API Endpoint for **Midtrans Notification Webhook** to auto-update transaction status to `PAID`.
- [ ] **Manual Verification Fallback**: Feature to manually mark as paid if webhook fails.

## 2. Advanced Stock Management
To manage the store without help, the system must track inventory precisely.
- [ ] **Stock Opname (Stock Taking)**:
  - [ ] Create features to adjust physical stock vs system stock.
  - [ ] Generate "Variance Report" (Lost/Stolen items).
- [ ] **Low Stock Alerts**:
  - [ ] Daily report of items below `min_stock_alert`.
  - [ ] Auto-generate "Shopping List" (Restock Plan).
- [ ] **Expiry Date Tracking (Optional)**:
  - [ ] *Note: Only for perishable items (e.g., Bread, Milk) to avoid input fatigue.*
  - [ ] Add `expiry_date` column to `stock_movements` (incoming) but make it optional.
  - [ ] Create "Near Expiry" alerts only if date is set.

## 3. Financial & Cash Management
Detailed tracking of money to ensure no leakage.
- [ ] **Cash Flow / Petty Cash**:
  - [ ] Track money taken out for operational costs (e.g., pay electricity, trash, buy plastic bags) directly from the drawer.
- [ ] **Daily Profit/Loss**:
  - [ ] Real-time profit calculation (Revenue - HPP).
  - [ ] Sales report by payment method (Cash vs QRIS vs Kasbon).

## 4. Kasbon (Debt) Management System
Critical for Warung Kelontong.
- [ ] **Debt Limits**: Enforce `credit_limit` validation automatically at checkout.
- [ ] **Billing/Reminder**: Generate a "Tagihan" list to send via WhatsApp to customers.
- [ ] **Installment Payments**: Allow paying debt in parts (Cicilan) and track balance history.

## 5. POS Operational Features
Features to speed up the cashier process.
- [ ] **Barcode Scanner Support**: Verify API searches by barcode efficiently (Scanning must be instant).
- [ ] **Hold/Resume Transaction**: Allow saving a cart temporarily if a customer leaves to pick up another item.
- [ ] **Refund/Return**: Handling customers returning defective items.

## 6. System Reliability (No Parents Needed)
- [ ] **Database Backup**: Automated daily backup (e.g., to Google Drive or local storage cron job).
- [ ] **Transaction Logs**: detailed logs of who modified what (Audit Trail).
