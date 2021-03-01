-- +goose Up
-- +goose StatementBegin
CREATE TABLE confas (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    name VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE confas;
-- +goose StatementEnd
