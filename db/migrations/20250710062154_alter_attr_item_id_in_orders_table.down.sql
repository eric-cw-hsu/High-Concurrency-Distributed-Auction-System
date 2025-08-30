ALTER TABLE orders DROP CONSTRAINT fk_orders_stock_id;

ALTER TABLE orders ADD COLUMN item_id TEXT;

UPDATE orders SET item_id = stock_id::text;

ALTER TABLE orders DROP COLUMN stock_id;