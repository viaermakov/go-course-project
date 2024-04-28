-- +goose Up
-- +goose StatementBegin
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
drop table if exists orders_events cascade;
-- +goose StatementEnd
