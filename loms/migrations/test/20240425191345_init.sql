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

create table if not exists orders_events
(
    order_id     bigint    not null,
    order_status int       not null,
    created_at   timestamp not null,
    send_status  int       not null,
    send_at      timestamp,
    primary key (order_id, order_status)
);

create index ids_created_at on orders_events (created_at desc);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists orders_info cascade;
drop table if exists orders_items cascade;
drop table if exists stocks cascade;
drop table if exists orders_events cascade;
-- +goose StatementEnd
