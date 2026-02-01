
import os
import psycopg2

try:
    conn = psycopg2.connect(
        host="localhost",
        database="warung_db",
        user="warung",
        password="warung_secret"
    )
    cur = conn.cursor()
    cur.execute("SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'users';")
    rows = cur.fetchall()
    print("Columns in users table:")
    for row in rows:
        print(f" - {row[0]} ({row[1]})")
    
    cur.close()
    conn.close()
except Exception as e:
    print(f"Error: {e}")
