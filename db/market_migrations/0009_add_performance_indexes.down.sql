-- Откат добавления индексов

DROP INDEX IF EXISTS idx_products_category_status;
DROP INDEX IF EXISTS idx_products_seller_status;
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_stock;
DROP INDEX IF EXISTS idx_products_created_at;

DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_name;

DROP INDEX IF EXISTS idx_sellers_status;
DROP INDEX IF EXISTS idx_sellers_user_status;
DROP INDEX IF EXISTS idx_sellers_store_name;

DROP INDEX IF EXISTS idx_orders_user_created;
DROP INDEX IF EXISTS idx_orders_seller_status;
DROP INDEX IF EXISTS idx_orders_seller_created;
DROP INDEX IF EXISTS idx_orders_total_amount;

DROP INDEX IF EXISTS idx_order_items_product_id;
DROP INDEX IF EXISTS idx_order_items_order_product;

DROP INDEX IF EXISTS idx_cart_items_cart_product;

DROP INDEX IF EXISTS idx_carts_user_id;
DROP INDEX IF EXISTS idx_carts_session_id;
DROP INDEX IF EXISTS idx_carts_updated_at;

DROP INDEX IF EXISTS idx_product_images_display_order;

DROP INDEX IF EXISTS idx_sellers_active;
DROP INDEX IF EXISTS idx_products_approved;

DROP INDEX IF EXISTS idx_products_sizes_gin;
DROP INDEX IF EXISTS idx_products_title_trgm;
DROP INDEX IF EXISTS idx_products_description_trgm;

DROP EXTENSION IF EXISTS pg_trgm;
