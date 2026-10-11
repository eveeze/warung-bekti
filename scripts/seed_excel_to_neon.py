#!/usr/bin/env python3
"""
High-Performance Batch Seeder for WarungOS -> Neon PostgreSQL
Imports 50,600+ product catalog items from Excel in seconds.
"""

import os
import sys
import time
import argparse
import openpyxl
import psycopg2
from psycopg2.extras import execute_values

DEFAULT_DSN = "postgresql://neondb_owner:npg_gyj8tN1IkAwn@ep-bitter-mouse-a8r2sapk-pooler.eastus2.azure.neon.tech/neondb?sslmode=require"
DEFAULT_FILE = "/home/eveeze/.t3/userdata/attachments/2fa25d1d-2b2a-46fc-926f-f94ddd312570-e971c6fc-6124-404d-b2a1-a044e8610833-xlsx.xlsx"

def clean_str(val):
    if val is None:
        return None
    s = str(val).strip()
    if s == "" or s.upper() in ["NA", "NONE", "-", "NULL"]:
        return None
    return s

def clean_int(val, default=0):
    if val is None:
        return default
    try:
        if isinstance(val, (int, float)):
            return max(0, int(val))
        s = str(val).replace(",", "").replace(".", "").strip()
        if not s:
            return default
        return max(0, int(s))
    except Exception:
        return default

def main():
    parser = argparse.ArgumentParser(description="Seed Excel Data to Neon DB")
    parser.add_argument("--dsn", default=os.getenv("DATABASE_URL", DEFAULT_DSN), help="Postgres connection DSN")
    parser.add_argument("--file", default=DEFAULT_FILE, help="Path to Excel (.xlsx) file")
    parser.add_argument("--batch-size", type=int, default=2000, help="Batch size for bulk insertion")
    args = parser.parse_args()

    print("=" * 60)
    print("🚀 SEEDING 50,600+ WARUNG PRODUCTS TO NEON POSTGRESQL")
    print("=" * 60)
    print(f"📁 Source File : {args.file}")
    print(f"📦 Batch Size  : {args.batch_size}")
    print(f"🌐 Database    : {args.dsn.split('@')[1] if '@' in args.dsn else 'configured DSN'}")

    start_time = time.time()

    # 1. Connect to Database
    print("\n[1/3] 🔌 Connecting to Neon PostgreSQL...")
    conn = psycopg2.connect(args.dsn)
    cur = conn.cursor()
    print("✓ Connected successfully!")

    # 2. Read Excel
    print("\n[2/3] 📖 Reading Excel workbook...")
    read_start = time.time()
    wb = openpyxl.load_workbook(args.file, read_only=True)
    sheet = wb.active

    rows = []
    seen_barcodes = set()
    seen_skus = set()
    total_raw_rows = 0

    for i, row in enumerate(sheet.iter_rows(values_only=True)):
        if i == 0:
            continue
        total_raw_rows += 1

        sku = clean_str(row[0])
        barcode = clean_str(row[1])
        name = clean_str(row[2])
        unit = clean_str(row[5]) or "pcs"
        cost_price = clean_int(row[6])
        base_price = clean_int(row[7])
        current_stock = clean_int(row[8])
        min_stock = clean_int(row[9])

        if not name:
            continue

        # Avoid duplicate barcodes or SKUs in same import batch
        if barcode:
            if barcode in seen_barcodes:
                barcode = None
            else:
                seen_barcodes.add(barcode)

        if sku:
            if sku in seen_skus:
                sku = None
            else:
                seen_skus.add(sku)

        # Tuple matching products table columns:
        # (name, barcode, sku, unit, base_price, cost_price, is_stock_active, current_stock, min_stock_alert, is_active)
        rows.append((
            name,
            barcode,
            sku,
            unit.lower() if len(unit) <= 20 else "pcs",
            base_price,
            cost_price,
            True,           # is_stock_active
            current_stock,
            min_stock,
            True            # is_active
        ))

    wb.close()
    print(f"✓ Read {len(rows)} products from {total_raw_rows} rows in {time.time() - read_start:.2f}s")

    # 3. Batch Insert
    cur.execute("SELECT count(*) FROM products;")
    existing_count = cur.fetchone()[0]
    if existing_count > 0:
        print(f"ℹ️  Found {existing_count} products already in database, continuing from row {existing_count}...", flush=True)
        rows_to_insert = rows[existing_count:]
    else:
        rows_to_insert = rows

    print(f"\n[3/3] ⚡ Batch inserting {len(rows_to_insert)} products into Neon...", flush=True)
    insert_sql = """
        INSERT INTO products (
            name, barcode, sku, unit, base_price, cost_price, is_stock_active, current_stock, min_stock_alert, is_active
        ) VALUES %s
        ON CONFLICT DO NOTHING
    """

    inserted = 0
    total_batches = (len(rows_to_insert) + args.batch_size - 1) // args.batch_size

    for b in range(total_batches):
        b_start = b * args.batch_size
        b_end = min((b + 1) * args.batch_size, len(rows_to_insert))
        chunk = rows_to_insert[b_start:b_end]

        execute_values(cur, insert_sql, chunk)
        conn.commit()
        inserted += len(chunk)
        print(f"  → Batch {b+1:2d}/{total_batches} ({inserted:5d}/{len(rows_to_insert)}) committed", flush=True)

    # 4. Verification
    cur.execute("SELECT count(*) FROM products;")
    db_count = cur.fetchone()[0]

    cur.execute("SELECT count(*) FROM products WHERE barcode IS NOT NULL;")
    barcode_count = cur.fetchone()[0]

    cur.execute("SELECT count(*) FROM products WHERE is_active = true;")
    active_count = cur.fetchone()[0]

    cur.close()
    conn.close()

    total_duration = time.time() - start_time
    print("\n" + "=" * 60)
    print("🎉 SEEDING COMPLETED SUCCESSFULLY!")
    print("=" * 60)
    print(f"⏱️  Total Duration : {total_duration:.2f} seconds")
    print(f"📦 Total in DB    : {db_count} products")
    print(f"🏷️  With Barcodes  : {barcode_count} products")
    print(f"✅ Active Items   : {active_count} products")
    print("=" * 60)

if __name__ == "__main__":
    main()
