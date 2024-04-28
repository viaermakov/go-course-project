-- name: InsertOrderInfo :exec
insert into orders_info (order_id, status, user_id)
values ($1, $2, $3);

-- name: InsertOrderItem :exec
insert into orders_items (order_id, item_id, count)
values ($1, $2, $3);

-- name: UpdateOrderStatus :exec
update orders_info
set status=$1
where order_id = $2;

-- name: GetOrderInfo :one
select order_id, status, user_id from orders_info
where order_id=$1;

-- name: GetOrderItem :many
select order_id, item_id, count from orders_items
where order_id=$1;

-- name: GetLastOrderItem :one
select order_id, item_id, count from orders_items
order by order_id DESC
limit 1;
