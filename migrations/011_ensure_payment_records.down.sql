-- Down migration (Drop if desired, or do nothing if we want to keep)
DROP TABLE IF EXISTS payment_records CASCADE;
