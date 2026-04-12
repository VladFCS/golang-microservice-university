CREATE TABLE IF NOT EXISTS products (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL,
    price_cents BIGINT NOT NULL,
    currency TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS inventory_stocks (
    product_id TEXT PRIMARY KEY,
    available BIGINT NOT NULL,
    reserved BIGINT NOT NULL
);
