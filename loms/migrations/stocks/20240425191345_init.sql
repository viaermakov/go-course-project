-- +goose Up
-- +goose StatementBegin
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
drop table if exists stocks cascade;
-- +goose StatementEnd
