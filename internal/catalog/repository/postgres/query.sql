-- name: GetProductByID :one
SELECT id, name, description, price_cents, currency
FROM products
WHERE id = $1
LIMIT 1;

-- name: UpsertProduct :one
INSERT INTO products (
    id,
    name,
    description,
    price_cents,
    currency
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
ON CONFLICT (id) DO UPDATE
SET
    name = EXCLUDED.name,
    description = EXCLUDED.description,
    price_cents = EXCLUDED.price_cents,
    currency = EXCLUDED.currency
RETURNING id, name, description, price_cents, currency;
