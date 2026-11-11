#!/usr/bin/env python3
"""
High-Performance Batch Product Pricing & Stock Enricher for WarungOS Live Neon DB.
Fills realistic market prices, wholesale costs, and initial stock for all 50,596 items.
"""
import sys
import time
import psycopg2
from psycopg2.extras import execute_batch
from test_pricing_engine import estimate_product_price

DEFAULT_DSN = "postgresql://neondb_owner:npg_gyj8tN1IkAwn@ep-bitter-mouse-a8r2sapk-pooler.eastus2.azure.neon.tech/neondb?sslmode=require"

def main():
    dsn = DEFAULT_DSN
    print("Connecting to Neon PostgreSQL...")
    conn = psycopg2.connect(dsn)
    cur = conn.cursor()

    print("Fetching remaining products with price or stock <= 0...")
    t0 = time.time()
    cur.execute("""
        SELECT p.id, p.name, p.unit, COALESCE(c.name, ''), p.base_price, p.cost_price, p.current_stock
        FROM products p
        LEFT JOIN categories c ON p.category_id = c.id
        WHERE p.base_price <= 0 OR p.current_stock <= 0
    """)
    products = cur.fetchall()
    print(f"Fetched {len(products)} products in {time.time() - t0:.2f}s")

    updates = []
    skipped_existing_prices = 0

    for pid, name, unit, cat_name, cur_base_price, cur_cost_price, cur_stock in products:
        # If product already has positive price, preserve it
        if cur_base_price and cur_base_price > 0:
            bp = cur_base_price
            cp = cur_cost_price if (cur_cost_price and cur_cost_price > 0) else int(round((bp * 0.82) / 100.0) * 100)
            skipped_existing_prices += 1
        else:
            bp, cp = estimate_product_price(name, unit or "pcs", cat_name or "")

        # Ensure stock is initialized so POS checkout doesn't fail
        stock = cur_stock if (cur_stock and cur_stock > 0) else 25
        min_stock = 5

        updates.append((bp, cp, stock, min_stock, pid))

    print(f"Prepared {len(updates)} updates ({skipped_existing_prices} preserved existing prices).")
    print("Executing batch update...")

    t1 = time.time()
    update_sql = """
        UPDATE products AS p
        SET base_price = u.base_price,
            cost_price = u.cost_price,
            current_stock = u.current_stock,
            min_stock_alert = u.min_stock_alert,
            updated_at = NOW()
        FROM (VALUES %s) AS u(base_price, cost_price, current_stock, min_stock_alert, id)
        WHERE p.id = u.id::uuid;
    """

    # Process in chunks of 5000 using execute_values
    from psycopg2.extras import execute_values
    chunk_size = 5000
    for i in range(0, len(updates), chunk_size):
        chunk = updates[i:i + chunk_size]
        execute_values(cur, update_sql, chunk)
        conn.commit()
        print(f"Updated {min(i + chunk_size, len(updates))}/{len(updates)} products...")

    print(f"All updates completed in {time.time() - t1:.2f}s")

    # Verification query
    cur.execute("""
        SELECT 
            count(*) as total,
            count(CASE WHEN base_price <= 0 THEN 1 END) as zero_price,
            count(CASE WHEN cost_price <= 0 THEN 1 END) as zero_cost,
            count(CASE WHEN current_stock <= 0 THEN 1 END) as zero_stock,
            min(base_price) as min_price,
            avg(base_price)::int as avg_price,
            max(base_price) as max_price
        FROM products
    """)
    row = cur.fetchone()
    print("\n=== POST-ENRICHMENT VERIFICATION ===")
    print(f"Total Products: {row[0]}")
    print(f"Products with Price <= 0: {row[1]}")
    print(f"Products with Cost <= 0:  {row[2]}")
    print(f"Products with Stock <= 0: {row[3]}")
    print(f"Price Range: Rp {row[4]:,d} - Rp {row[6]:,d} (Avg: Rp {row[5]:,d})")

    # Sample check
    cur.execute("""
        SELECT name, base_price, cost_price, current_stock
        FROM products
        ORDER BY random()
        LIMIT 10
    """)
    print("\n=== RANDOM 10 PRODUCTS SAMPLE ===")
    for p in cur.fetchall():
        print(f"{p[0]:<40} | Jual: Rp {p[1]:>7,d} | Pokok: Rp {p[2]:>7,d} | Stok: {p[3]}")

    cur.close()
    conn.close()

if __name__ == "__main__":
    main()
