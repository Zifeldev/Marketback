-- Добавление дополнительных индексов для оптимизации запросов

-- Включаем расширение для полнотекстового поиска (должно быть в начале)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Products: составные индексы для частых запросов (избегаем дублирования с миграцией 0003)
CREATE INDEX IF NOT EXISTS idx_products_category_status ON products(category_id, status);
CREATE INDEX IF NOT EXISTS idx_products_seller_status ON products(seller_id, status);
CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock) WHERE stock > 0;
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at DESC);

-- Categories: индекс для иерархии (дополнительный к существующему)
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- Sellers: составной индекс для поиска
CREATE INDEX IF NOT EXISTS idx_sellers_user_active ON sellers(user_id, is_active);

-- Orders: составные индексы для частых запросов
CREATE INDEX IF NOT EXISTS idx_orders_user_created ON orders(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_payment ON orders(status, payment_status);
CREATE INDEX IF NOT EXISTS idx_orders_total_amount ON orders(total_amount);

-- Order Items: индекс для агрегаций (дополнительно к существующим)
CREATE INDEX IF NOT EXISTS idx_order_items_order_product ON order_items(order_id, product_id);

-- Cart Items: составной индекс для быстрого поиска
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_product ON cart_items(cart_id, product_id);

-- Carts: дополнительный индекс для обновлений
CREATE INDEX IF NOT EXISTS idx_carts_updated_at ON carts(updated_at DESC);

-- Partial index для активных продавцов
CREATE INDEX IF NOT EXISTS idx_sellers_active ON sellers(id) WHERE is_active = true;

-- Partial index для одобренных товаров
CREATE INDEX IF NOT EXISTS idx_products_approved ON products(id, category_id, seller_id) WHERE status = 'approved';

-- Индекс для поиска по товарам с размерами (JSONB)
CREATE INDEX IF NOT EXISTS idx_products_sizes_gin ON products USING gin(sizes);

-- Полнотекстовый поиск по описанию товаров (теперь pg_trgm уже создан)
CREATE INDEX IF NOT EXISTS idx_products_description_trgm ON products USING gin(description gin_trgm_ops);
