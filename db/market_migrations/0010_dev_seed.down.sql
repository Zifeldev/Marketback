-- Rollback dev seed: remove dev product and seller
DELETE FROM products WHERE title = 'Dev Product (dev seed)';
DELETE FROM sellers WHERE user_id = 2 AND shop_name = 'DevShop (dev seed)';
