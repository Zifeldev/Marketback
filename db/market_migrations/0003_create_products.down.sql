-- Drop products table
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_seller_id;
DROP TABLE IF EXISTS products;
