CREATE TABLE
  orders (
    order_id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL,
    buyer_id TEXT NOT NULL,
    quantity INTEGER NOT NULL,
    total_price NUMERIC(10, 2) NOT NULL,
    timestamp TIMESTAMP NOT NULL
  );