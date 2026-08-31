-- Seed data for the ishopster catalog (dev only).
--
-- Run:
--   docker compose exec -T db psql -U dbuser -d dbname < sql/seed.sql
--
-- Safe to re-run: TRUNCATE ... RESTART IDENTITY wipes both tables and
-- resets the id sequences, so every run reproduces the same ids.
-- CASCADE is needed because products references categories.

BEGIN;

TRUNCATE categories, products RESTART IDENTITY CASCADE;

INSERT INTO categories (name, slug, description) VALUES
  ('Laptops',     'laptops',     'Portable computers for work and study'),
  ('Smartphones', 'smartphones', 'Mobile phones and flagships'),
  ('Accessories', 'accessories', 'Chargers, mice, keyboards and cases'),
  ('Monitors',    'monitors',    'External displays');

-- category_id is resolved from the category slug at insert time, so this
-- does not depend on the generated category ids.
INSERT INTO products (name, slug, category_id, price, description) VALUES
  ('ThinkPad X1 Carbon Gen 12', 'thinkpad-x1-carbon-gen-12',
   (SELECT id FROM categories WHERE slug = 'laptops'), 129990.00, '14" business ultrabook'),
  ('MacBook Air 13 M3', 'macbook-air-13-m3',
   (SELECT id FROM categories WHERE slug = 'laptops'), 119990.00, 'Apple silicon, fanless'),
  ('Acer Aspire 5', 'acer-aspire-5',
   (SELECT id FROM categories WHERE slug = 'laptops'), 54990.00, 'Budget all-rounder'),
  ('iPhone 15', 'iphone-15',
   (SELECT id FROM categories WHERE slug = 'smartphones'), 79990.00, '6.1" OLED, USB-C'),
  ('Google Pixel 8', 'google-pixel-8',
   (SELECT id FROM categories WHERE slug = 'smartphones'), 62990.00, 'Clean Android, great camera'),
  ('Redmi Note 13', 'redmi-note-13',
   (SELECT id FROM categories WHERE slug = 'smartphones'), 19990.00, 'Affordable mid-range'),
  ('Anker USB-C Charger 65W', 'anker-usb-c-charger-65w',
   (SELECT id FROM categories WHERE slug = 'accessories'), 2490.00, 'GaN, 3 ports'),
  ('Logitech Wireless Mouse', 'logitech-wireless-mouse',
   (SELECT id FROM categories WHERE slug = 'accessories'), 1290.00, 'Compact, 2.4 GHz'),
  ('Laptop Sleeve 14"', 'laptop-sleeve-14',
   (SELECT id FROM categories WHERE slug = 'accessories'), 990.00, 'Felt, water-resistant'),
  ('Keychron K2 Keyboard', 'keychron-k2-keyboard',
   (SELECT id FROM categories WHERE slug = 'accessories'), 4990.00, 'Mechanical, hot-swap');

-- Monitors is left empty on purpose (used by exercise 3).

COMMIT;
