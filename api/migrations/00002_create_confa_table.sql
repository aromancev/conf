-- +goose Up
-- +goose StatementBegin
create table confa
(
    tag        varchar(50),
    created_at date default now(),
    id         uuid not null
        constraint id
            primary key,
    owner      uuid
);

alter table confa
    owner to confa;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table confa
-- +goose StatementEnd
