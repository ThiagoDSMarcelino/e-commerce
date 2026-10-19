-- name: ListOrders :many
SELECT * FROM orders
WHERE client_id = sqlc.arg('client_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[]);

-- name: CountOrders :one
SELECT count(*) FROM orders
WHERE client_id = sqlc.arg('client_id');

-- name: CreateOrder :one
INSERT INTO orders (
    id,
    client_id,
    status
) VALUES ($1, $2, $3) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (
    order_id,
    product_id,
    quantity
) VALUES ($1, $2, $3) RETURNING *;


-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = sqlc.arg('status')
WHERE id = sqlc.arg('order_id')::uuid;

-- name: ListRegistredEmail :many
SELECT * FROM promotion_registrations
WHERE client_id = sqlc.arg('client_id')
ORDER BY created_at DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountRegistrations :one
SELECT count(*) FROM promotion_registrations
WHERE client_id = sqlc.arg('client_id');
