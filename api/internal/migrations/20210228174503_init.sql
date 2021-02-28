-- +goose Up
-- +goose StatementBegin
CREATE TABLE confas (
    id UUID PRIMARY KEY
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE confas;
-- +goose StatementEnd
