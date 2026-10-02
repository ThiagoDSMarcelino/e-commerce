-- name: ListOrders :many
SELECT * FROM orders
LIMIT $2
OFFSET $1;

-- name: Count :one
SELECT count(*) FROM orders;
