CREATE TABLE
  stocks (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    price NUMERIC(10, 2) NOT NULL,
    quantity INT NOT NULL CHECK (quantity >= 0),
    created_at TIMESTAMP
    WITH
      TIME ZONE NOT NULL DEFAULT NOW (),
      updated_at TIMESTAMP
    WITH
      TIME ZONE NOT NULL DEFAULT NOW ()
  );

CREATE INDEX idx_stocks_product_id ON stocks (product_id);

CREATE INDEX idx_stocks_seller_id ON stocks (seller_id);

CREATE INDEX idx_stocks_product_seller ON stocks (product_id, seller_id);