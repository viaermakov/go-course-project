-- name: AddStock :exec
insert into stocks (sku_id, available, reserved)
values ($1, $2, $3);

-- name: RemoveStock :exec
update stocks
set reserved=$1
where sku_id = $2;

-- name: ReserveStock :exec
update stocks
set reserved=$1, available=$2
where sku_id = $3;

-- name: GetStock :one
select sku_id, available, reserved
from stocks
where sku_id = $1;