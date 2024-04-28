-- +goose Up
-- +goose StatementBegin
create table if not exists orders_info
(
    order_id bigint,
    status   int    not null,
    user_id  bigint not null,
    primary key (order_id)
);

create table if not exists orders_items
(
    order_id bigint not null,
    item_id  bigint not null,
    count    int    not null,
    primary key (order_id, item_id)
);

create table if not exists stocks
(
    sku_id    bigint not null,
    available bigint not null,
    reserved  bigint not null,
    primary key (sku_id)
);

insert into stocks (sku_id, available, reserved)
values
    (773297411, 140, 10),
    (1002, 180, 20),
    (1003, 250, 30),
    (1004, 260, 40),
    (1005, 300, 50);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists orders_info cascade;
drop table if exists orders_items cascade;
drop table if exists stocks cascade;
-- +goose StatementEnd
