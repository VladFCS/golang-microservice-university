-- name: GetStockByProductID :one
SELECT product_id, available, reserved
FROM inventory_stocks
WHERE product_id = $1
LIMIT 1;

-- name: UpsertStock :one
INSERT INTO inventory_stocks (
    product_id,
    available,
    reserved
) VALUES (
    $1,
    $2,
    $3
)
ON CONFLICT (product_id) DO UPDATE
SET
    available = EXCLUDED.available,
    reserved = EXCLUDED.reserved
RETURNING product_id, available, reserved;

-- name: ReserveStock :one
UPDATE inventory_stocks
SET
    available = available - $2,
    reserved = reserved + $2
WHERE product_id = $1
  AND available >= $2
RETURNING product_id, available, reserved;
