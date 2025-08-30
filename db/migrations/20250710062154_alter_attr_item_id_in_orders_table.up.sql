ALTER TABLE orders ADD COLUMN stock_id UUID;

UPDATE orders SET stock_id = item_id::uuid;

ALTER TABLE orders DROP COLUMN item_id;

ALTER TABLE orders
ADD CONSTRAINT fk_orders_stock_id
FOREIGN KEY (stock_id)
REFERENCES stocks(id)
ON DELETE RESTRICT;