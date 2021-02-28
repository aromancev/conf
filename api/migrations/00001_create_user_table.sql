-- +goose Up
-- +goose StatementBegin
create table "user"
(
    id uuid not null
        constraint user_pk
            primary key
);

alter table "user"
    owner to confa;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table "user"
-- +goose StatementEnd
