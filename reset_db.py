import os
import psycopg2
from urllib.parse import urlparse

# Load .env manually or use defaults
DB_HOST = "localhost"
DB_PORT = "5432"
DB_USER = "postgres"
DB_PASS = "postgres"
DB_NAME = "warung"

# Try to read .env
if os.path.exists(".env"):
    with open(".env", "r") as f:
        for line in f:
            if line.strip() and not line.startswith("#"):
                key, val = line.strip().split("=", 1)
                if key == "DB_HOST": DB_HOST = val
                if key == "DB_PORT": DB_PORT = val
                if key == "DB_USER": DB_USER = val
                if key == "DB_PASSWORD": DB_PASS = val
                if key == "DB_NAME": DB_NAME = val

try:
    conn = psycopg2.connect(
        host=DB_HOST,
        port=DB_PORT,
        user=DB_USER,
        password=DB_PASS,
        dbname=DB_NAME
    )
    cur = conn.cursor()
    
    # List of tables to truncate (order matters due to FK, or use CASCADE)
    tables = [
        "schema_migrations", # Don't truncate migrations? But if we re-run verify... keep migrations.
        # But we want to keep User 'admin'?
        # If we truncate 'users', we lose admin.
        # Seeder handles admin creation.
        # So it's safe to truncate all and re-seed.
    ]
    
    # Actually, TRUNCATE users CASCADE will clear everything linked.
    print("Truncating all data...")
    cur.execute("""
        DO $$ DECLARE
            r RECORD;
        BEGIN
            FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename != 'schema_migrations') LOOP
                EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
            END LOOP;
        END $$;
    """)
    
    conn.commit()
    print("Database reset complete.")
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error resetting DB: {e}")
