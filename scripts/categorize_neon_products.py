#!/usr/bin/env python3
"""
Batch Product Categorizer for WarungOS -> Neon PostgreSQL
Categorizes 50,600+ warung products into standardized retail categories in batches.
"""

import os
import sys
import time
import argparse
import psycopg2
from psycopg2.extras import execute_values

DEFAULT_DSN = "postgresql://neondb_owner:npg_gyj8tN1IkAwn@ep-bitter-mouse-a8r2sapk-pooler.eastus2.azure.neon.tech/neondb?sslmode=require"

CATEGORIES = [
    {
        "name": "Makanan & Mie Instan",
        "description": "Mie instan, pasta, bihun, sarden, kornet, sosis, dan makanan cepat saji",
        "keywords": [
            "MIE", "NOODLE", "INDOMIE", "SARIMI", "SUPERMI", "SEDAP", "SEDAAP", "POP MIE",
            "BIHUN", "SOUN", "SOHUN", "MISUA", "PASTA", "SPAGHETTI", "SPAGETI", "MACARONI",
            "MAKARONI", "LAFONTE", "SARDEN", "SARDINE", "CORNED", "KORNET", "PRONAS", "ABON",
            "SOSIS", "SAUSAGE", "NUGGET", "BAKSO", "BASO", "SOMAY", "SIOMAY", "SOTO", "GULAI",
            "KARI", "RENDANG", "BUBUR", "SHIRATAKI", "KONJAC", "NATA DE COCO", "SEAWEED", "NORI",
            "MAMEE", "GAGA", "GEKIKARA"
        ]
    },
    {
        "name": "Minuman",
        "description": "Kopi, teh, susu, air mineral, sirup, jus, dan minuman kemasan",
        "keywords": [
            "KOPI", "COFFEE", "NESCAFE", "LUWAK", "KAPAL API", "TORABIKA", "TOP KOPI", "GOOD DAY",
            "WHITE KOFFIE", "ESPRESSO", "LATTE", "CAPPUCHINO", "TEH", "TEA", "TONG TJI", "SOSRO",
            "PUCUK", "SARIWANGI", "SARI WANGI", "2 TANG", "KEPALA JENGGOT", "SUSU", "MILK",
            "INDOMILK", "FRISIAN", "DANCOW", "ULTRA", "MILO", "BEAR BRAND", "BENDERA", "CREAMER",
            "KRIMER", "AIR", "MINERAL", "AQUA", "LE MINERALE", "CLEO", "VIT", "ADES", "CLUB",
            "PRISTIN", "OASIS", "SIRUP", "SYRUP", "MARJAN", "SQUASH", "FREISS", "JUS", "JUICE",
            "BUAVITA", "FLORIDINA", "PULPY", "ALE-ALE", "ALE ALE", "MARIMAS", "NUTRISARI", "TEAJUS",
            "JASJUS", "FANTA", "COCA COLA", "COKE", "SPRITE", "PEPSI", "POCARI", "HYDRO COCO",
            "YOU C1000", "KRATINDAENG", "EXTRA JOSS", "CINCAU", "LARUTAN", "ADEM SARI", "LASEGAR",
            "CAP KAKI TIGA", "BADAK", "SARSAPARILLA", "TEH KOTAK", "TEH BOTOL", "TEH GELAS",
            "OKKY JELLY", "ICHITAN", "MILKU"
        ]
    },
    {
        "name": "Makanan Ringan & Biskuit",
        "description": "Snack, keripik, biskuit, wafer, permen, dan cokelat",
        "keywords": [
            "CHIKI", "TARO", "CHITATO", "LAYS", "LAY", "CHEETOS", "POTABEE", "QTELA", "KUSUKA",
            "KERIPIK", "KRIPIK", "KERUPUK", "KRUPUK", "EMPING", "PEYEK", "KACANG", "GARUDA",
            "DUA KELINCI", "SUKRO", "ROSTA", "ATOM", "BISKUIT", "BISCUIT", "COOKIES", "COOKIE",
            "WAFER", "ROMA", "OREO", "NABATI", "TANGO", "BENG BENG", "CHOCOLATOS", "REGAL",
            "GANDUM", "GOOD TIME", "KHONG GUAN", "NISSIN", "MALKIST", "SLAI OLAI", "SELAMAT",
            "ASTOR", "TWISTER", "TIM TAM", "GERY", "PERMEN", "CANDY", "YUPI", "MENTOS", "KOPIKO",
            "RELAXA", "SUGUS", "FOX", "BLASTER", "CHUPA", "LOLLIPOP", "COKELAT", "COKLAT",
            "CHOCOLATE", "SILVERQUEEN", "KITKAT", "CADBURY", "DELFI", "JELLY", "AGAR", "PUDDING",
            "PUDING", "NUTRIJELL", "POPCORN", "KUACI", "GUM", "CHEWING GUM", "TIC TAC", "PIATTO",
            "MAXICORN", "SNACK"
        ]
    },
    {
        "name": "Sembako & Bahan Pokok",
        "description": "Beras, minyak goreng, gula, tepung, telur, margarin, dan keju",
        "keywords": [
            "BERAS", "RICE", "KETAN", "MINYAK", "BIMOLI", "FILMA", "SANIA", "FORTUNE", "TROPICAL",
            "KUNCIMAS", "SOVIA", "SUNCO", "ROSE BRAND", "GULA", "SUGAR", "GULAKU", "TEPUNG",
            "FLOUR", "TERIGU", "SEGITIGA BIRU", "CAKRA KEMBAR", "KUNCI BIRU", "TAPIOKA", "MAIZENA",
            "SAGU", "MENTEGA", "BUTTER", "MARGARIN", "MARGARINE", "BLUE BAND", "FORVITA", "SIMAS",
            "GARAM", "SALT", "DOLPHIN", "TELUR", "TELOR", "EGG", "KEJU", "CHEESE", "KRAFT",
            "PROCHIZ", "RAGI", "FERMIPAN", "BAKING POWDER", "SODA KUE", "VANILI", "KENTANG SEGAR"
        ]
    },
    {
        "name": "Bumbu & Penyedap Dapur",
        "description": "Bumbu instan, saus, kecap, santan, dan rempah-rempah",
        "keywords": [
            "BUMBU", "SEASON", "RACIK", "MASAKO", "ROYCO", "SASA", "AJINOMOTO", "MAGGI",
            "PENYEDAP", "KALDU", "MICIN", "VETSIN", "KECAP", "BANGO", "SAUS", "SAOS", "SAUCE",
            "CATSUP", "SAMBAL", "CHILI", "DELMONTE", "BELIBIS", "TERASI", "PETIS", "SANTAN",
            "KARA", "LADA", "PEPPER", "MERICA", "LADAKU", "DESAKU", "KETUMBAR", "KUNYIT", "JAHE",
            "LENGKUAS", "KENCUR", "PALA", "CENGKEH", "KAYU MANIS", "SERAI", "KEMIRI", "CUKA",
            "VINEGAR", "MAYONAIS", "MAYONNAISE", "SAORI", "SAUS TIRAM", "WIJEN", "TAUCO", "PETIS"
        ]
    },
    {
        "name": "Rokok & Tembakau",
        "description": "Rokok, cerutu, tembakau, korek api, dan aksesoris perokok",
        "keywords": [
            "ROKOK", "CIGARETTE", "DJARUM", "SAMPOERNA", "SURYA", "GUDANG GARAM", "MARLBORO",
            "MAGNUM", "MILD", "FILTER", "KRETEK", "LA LIGHTS", "LA BOLD", "ESSE", "CAMEL",
            "DUNHILL", "WINSTON", "DIPLOMAT", "CLAS MILD", "PRO MILD", "DJI SAM SOE", "234",
            "CERUTU", "KOREK", "LIGHTER", "TOKAI", "CRICKET", "TEMBAKAU", "PAPIR", "VIP MILD",
            "SEJATI", "GUDANG BARU"
        ]
    },
    {
        "name": "Perawatan Tubuh & Kebersihan",
        "description": "Sabun mandi, sampo, pasta gigi, deodoran, parfum, dan kosmetik",
        "keywords": [
            "SABUN", "SOAP", "BODY WASH", "SHOWER", "LIFEBUOY", "DETTOL", "LUX", "GIV", "NUVO",
            "BIORE", "SHINZUI", "DOVE", "SAMPO", "SHAMPOO", "SUNSILK", "PANTENE", "CLEAR", "ZINC",
            "REJOICE", "CONDITIONER", "KONDISIONER", "SIKAT GIGI", "TOOTHBRUSH", "ODOL", "PASTA GIGI",
            "TOOTHPASTE", "PEPSODENT", "CIPTADENT", "CLOSE UP", "FORMULA", "SENSODYNE", "MOUTHWASH",
            "LISTERINE", "DEODORAN", "DEODORANT", "REXONA", "NIVEA", "AXE", "GATSBY", "PARFUM",
            "PERFUME", "COLOGNE", "BODY MIST", "MARINA", "VASELINE", "LOTION", "CITRA", "WARDAH",
            "GARNIER", "PONDS", "BEDAK", "POWDER", "CALADINE", "PEMBALUT", "CHARM", "SOFTEX",
            "LAURIER", "KAPAS", "COTTON", "GILLETTE", "CUKUR", "RAZOR", "POMADE", "HAIR GEL",
            "JERGENS", "SATTO", "EMERON", "ELLIPS"
        ]
    },
    {
        "name": "Pembersih & Rumah Tangga",
        "description": "Deterjen, pencuci piring, pembersih lantai, tisu, kantong plastik, lampu, baterai",
        "keywords": [
            "RINSO", "SO KLIN", "SOKLIN", "DAIA", "BOOM", "ATTACK", "DETERJEN", "DETERGENT", "MOLTO",
            "DOWNY", "KISPRAY", "RAPIKA", "BAYCLIN", "VANISH", "PROCLIN", "SUNLIGHT", "MAMA LEMON",
            "MAMA LIME", "CUCI PIRING", "DISHWASH", "WIPOL", "SUPER PEL", "VIXAL", "HARPIC",
            "KARBOL", "KAMPER", "KAPUR BARUS", "STELLA", "GLADE", "BAYFRESH", "BAYGON", "HIT",
            "VAPE", "OBAT NYAMUK", "TISU", "TISSUE", "PASEO", "NICE", "JOLLY", "KRESEK", "PLASTIK",
            "TRASH BAG", "SPONS", "SPON", "SPONGE", "SCOTCH BRITE", "SAPU", "PEL", "KANEBO",
            "SEDOTAN", "WAX PAPER", "REYNOLDS", "LAMPU", "BULB", "LED", "PHILIPS", "BOHLAM",
            "BATERAI", "BATTERY", "ENERGIZER", "ABC BATERAI", "PANASONIC BATERAI", "LILIN",
            "LAKBAN", "ISOLASI", "KARET GELANG", "TUSUK GIGI"
        ]
    },
    {
        "name": "Obat & Kesehatan",
        "description": "Obat bebas, minyak angin, balsem, vitamin, dan plester",
        "keywords": [
            "OBAT", "MEDICINE", "BODREX", "PARAMEX", "PANADOL", "OSKADON", "SARIDON", "INZA",
            "MIXAGRIP", "PROCOLD", "ULTRAFLU", "FLU", "BATUK", "KONIDIN", "VICKS", "KOMIX",
            "WOODS", "OBH", "PROMAG", "ANTASIDA", "MYLANTA", "ENTROSTOP", "DIATABS", "DIAPET",
            "TOLAK ANGIN", "ANTANGIN", "KAYU PUTIH", "CAJUPUT", "MINYAK ANGIN", "FRESHCARE",
            "SAFE CARE", "BALSEM", "GELIGA", "COUNTERPAIN", "SALONPAS", "KOYO", "HANSAPLAST",
            "PLESTER", "BETADINE", "ANTISEPTIK", "ALKOHOL", "PERBAN", "INSTO", "ROHTO", "ENERVO",
            "CDR", "REDOXON", "VITAMIN", "MADU", "MADURASA", "TJ MADU", "CURCUMA", "SANGOBION"
        ]
    },
    {
        "name": "Perlengkapan Bayi & Anak",
        "description": "Popok bayi, minyak telon, makanan bayi, dan perawatan bayi",
        "keywords": [
            "MAMYPOKO", "PAMPERS", "SWEETY", "MERRIES", "BABY HAPPY", "POPOK", "DIAPER", "DIAPERS",
            "MINYAK TELON", "TELON", "ZWITSAL", "CUSSONS", "JOHNSON", "PIGEON", "MY BABY TELON",
            "KONICARE", "BABY OIL", "BABY BATH", "BABY SHAMPOO", "BEDAK BAYI", "TISU BASAH",
            "WET WIPES", "BABY WIPES", "CERELAC", "PROMINA", "SUN BUBUR", "BUBUR BAYI", "SGM",
            "BEBELAC", "MORINAGA", "CHIL KID", "PEDIASURE", "DOT", "BOTOL SUSU", "GERBER", "MILNA"
        ]
    },
    {
        "name": "Alat Tulis & Kantor",
        "description": "Buku tulis, pulpen, pensil, kertas, map, dan perlengkapan sekolah/kantor",
        "keywords": [
            "BUKU", "BOOK", "SIDU", "KIKY", "SINAR DUNIA", "BINDER", "LOOSE LEAF", "PULPEN", "PEN",
            "PENA", "BOLPOIN", "PILOT", "STANDARD", "SNOWMAN", "FASTER", "FABER CASTELL", "JOYKO",
            "KENKO", "PENSIL", "PENCIL", "2B", "MEKANIK", "SPIDOL", "MARKER", "WHITEBOARD",
            "PENGHAPUS", "ERASER", "TIPE-X", "TIP EX", "PENGGARIS", "RULER", "LEM", "GLUE",
            "GLUKOL", "UHU", "STEPLES", "STAPLES", "HEKTER", "AMPLOP", "MAP", "KERTAS", "PAPER",
            "HVS", "FOLIO", "GUNTING", "CUTTER", "RAUTAN", "RIBBON", "E-PRINT", "EPSON RF"
        ]
    },
    {
        "name": "Pakaian & Sandang",
        "description": "Sarung, kaos kaki, sandal, pakaian dalam, dan tekstil",
        "keywords": [
            "SARUNG", "KAOS KAKI", "SANDAL", "ANDO", "SWALLOW", "SEPATU", "CELANA", "BAJU",
            "KAOS", "SINGLET", "HANDUK", "SELIMUT"
        ]
    },
    {
        "name": "Lain-lain / Serba Serbi",
        "description": "Produk umum lainnya",
        "keywords": []
    }
]

def classify_product(name_upper):
    for cat in CATEGORIES[:-1]:  # exclude Lain-lain
        for kw in cat["keywords"]:
            if kw in name_upper:
                return cat["name"]
    return "Lain-lain / Serba Serbi"

def main():
    parser = argparse.ArgumentParser(description="Categorize Products in Neon DB")
    parser.add_argument("--dsn", default=os.getenv("DATABASE_URL", DEFAULT_DSN), help="Neon PostgreSQL DSN")
    parser.add_argument("--batch-size", type=int, default=5000, help="Batch size for processing")
    args = parser.parse_args()

    print("=" * 65)
    print("🏷️  BATCH PRODUCT CATEGORIZER -> NEON POSTGRESQL")
    print("=" * 65)
    print(f"📦 Batch Size : {args.batch_size:,} items")
    print(f"🌐 Target DB  : {args.dsn.split('@')[1] if '@' in args.dsn else 'configured DSN'}")

    start_time = time.time()

    # 1. Connect
    print("\n[1/4] 🔌 Menghubungkan ke Neon PostgreSQL...", flush=True)
    conn = psycopg2.connect(args.dsn)
    cur = conn.cursor()
    print("✓ Berhasil terhubung!", flush=True)

    # 2. Ensure Categories Exist in DB
    print("\n[2/4] 📁 Sinkronisasi 13 Kategori Master di Database...", flush=True)
    category_map = {}  # name -> UUID
    for cat in CATEGORIES:
        cur.execute(
            """
            INSERT INTO categories (name, description, is_active)
            VALUES (%s, %s, true)
            ON CONFLICT DO NOTHING
            RETURNING id;
            """,
            (cat["name"], cat["description"])
        )
        row = cur.fetchone()
        if row:
            category_map[cat["name"]] = row[0]
        else:
            cur.execute("SELECT id FROM categories WHERE name = %s;", (cat["name"],))
            category_map[cat["name"]] = cur.fetchone()[0]

    conn.commit()
    for name, cid in category_map.items():
        print(f"  ✓ [{name}] -> {cid}", flush=True)

    # 3. Read All Products needing categorization
    print("\n[3/4] 🔍 Mengambil produk dari database...", flush=True)
    cur.execute("SELECT id, name FROM products WHERE category_id IS NULL ORDER BY created_at ASC;")
    all_products = cur.fetchall()
    total_products = len(all_products)
    print(f"✓ Ditemukan {total_products:,} produk tanpa kategori untuk diproses.", flush=True)

    if total_products == 0:
        print("Semua produk sudah memiliki kategori. Selesai!")
        cur.close()
        conn.close()
        return

    # 4. Batch Classify and Update
    print(f"\n[4/4] ⚡ Memproses klasifikasi per batch ({args.batch_size:,} item/batch)...", flush=True)
    total_batches = (total_products + args.batch_size - 1) // args.batch_size
    processed = 0
    category_stats = {}

    update_sql = """
        UPDATE products AS p
        SET category_id = c.cat_id::uuid
        FROM (VALUES %s) AS c(prod_id, cat_id)
        WHERE p.id = c.prod_id::uuid;
    """

    for b in range(total_batches):
        b_start = b * args.batch_size
        b_end = min((b + 1) * args.batch_size, total_products)
        chunk = all_products[b_start:b_end]
        b_timer = time.time()

        update_values = []
        for prod_id, prod_name in chunk:
            cat_name = classify_product(prod_name.upper())
            cat_id = category_map[cat_name]
            update_values.append((str(prod_id), str(cat_id)))
            category_stats[cat_name] = category_stats.get(cat_name, 0) + 1

        execute_values(cur, update_sql, update_values)
        conn.commit()
        processed += len(chunk)
        duration_b = time.time() - b_timer
        print(f"  → Batch {b+1:2d}/{total_batches} ({processed:6,d}/{total_products:,d}) selesai ({duration_b:.2f}s)", flush=True)

    # 5. Final Report
    cur.close()
    conn.close()
    total_duration = time.time() - start_time

    print("\n" + "=" * 65)
    print("🎉 SELURUH PRODUK TELAH BERHASIL DIKATEGORISASIKAN!")
    print("=" * 65)
    print(f"⏱️  Total Durasi : {total_duration:.2f} detik")
    print(f"📊 Distribusi Kategori:")
    print("-" * 65)
    for cat_name, count in sorted(category_stats.items(), key=lambda x: x[1], reverse=True):
        pct = (count / total_products) * 100
        print(f"  {cat_name:32s}: {count:6,d} produk ({pct:5.1f}%)")
    print("=" * 65)

if __name__ == "__main__":
    main()
