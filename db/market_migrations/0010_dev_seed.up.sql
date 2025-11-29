-- Dev seed: create a seller for user_id=2 and a product for that seller (idempotent without unique constraints)

-- Create seller if not exists
INSERT INTO sellers (user_id, shop_name, is_active, created_at, updated_at)
SELECT 2, 'DevShop (dev seed)', true, NOW(), NOW()
WHERE NOT EXISTS (
	SELECT 1 FROM sellers s WHERE s.user_id = 2
);

-- Create product for that seller if not exists
INSERT INTO products (seller_id, title, price, stock, status, created_at, updated_at)
SELECT s.id, 'Dev Product (dev seed)', 1.00, 100, 'active', NOW(), NOW()
FROM sellers s
WHERE s.user_id = 2
AND NOT EXISTS (
	SELECT 1
	FROM products p
	WHERE p.seller_id = s.id AND p.title = 'Dev Product (dev seed)'
);
