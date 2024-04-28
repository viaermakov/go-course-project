-- name: AddEvent :exec
insert into orders_events (order_id, order_status, created_at, send_status, send_at)
values ($1, $2, $3, $4, $5);

-- name: UpdateEvent :exec
update orders_events
set send_status=$1, send_at=$2
where order_id = $3;

-- name: GetScheduledEvents :many
select order_id, order_status, created_at, send_status, send_at
from orders_events
where send_status = $1;