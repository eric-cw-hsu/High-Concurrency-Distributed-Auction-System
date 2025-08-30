CREATE TABLE
  products (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP
    WITH
      TIME ZONE NOT NULL DEFAULT NOW (),
      updated_at TIMESTAMP
    WITH
      TIME ZONE NOT NULL DEFAULT NOW ()
  );