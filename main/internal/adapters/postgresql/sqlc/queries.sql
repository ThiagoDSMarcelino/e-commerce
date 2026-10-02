-- name: ListOrders :many
SELECT * FROM orders
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: ListOrderItemsByOrderIDs :many
SELECT * FROM order_items
WHERE order_id = ANY(sqlc.arg('order_ids')::uuid[]);

-- name: Count :one
SELECT count(*) FROM orders;

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
