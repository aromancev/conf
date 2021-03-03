-- +goose Up
-- +goose StatementBegin
CREATE TABLE "user"
(
    id UUID PRIMARY KEY
);

CREATE TABLE confa (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    handle VARCHAR(32) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "user";
DROP TABLE confa;
-- +goose StatementEnd
