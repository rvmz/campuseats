-- Session 3: sqlc queries for the order transaction.
-- Session 6: InsertOutbox and the unpublished/mark-published queries.
-- Session 7: CreateOrder uses ON CONFLICT (user_id, idempotency_key) DO NOTHING.
-- Accept and the other staff commands update status only from the legal previous status.

-- name: CreateOrder :one
-- insert into orders and return the new row.

-- name: CreateOrderItem :exec
-- insert one order line. Price comes from menu_items, not from the client.
