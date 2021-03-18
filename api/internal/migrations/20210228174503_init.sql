-- +goose Up
-- +goose StatementBegin
CREATE TYPE platform AS ENUM ('email');
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);
CREATE TABLE idents (
    id UUID PRIMARY KEY,
    owner UUID REFERENCES users(id) NOT NULL,
    platform platform NOT NULL,
    value VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (platform, value)
);

CREATE TABLE confas (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    handle VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE talks (
    id UUID PRIMARY KEY,
    confa UUID NOT NULL REFERENCES confas (id) ON DELETE CASCADE,
    handle VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (confa, handle)
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE users;
DROP TABLE idents;
DROP TABLE confas;
DROP TABLE talks;
-- +goose StatementEnd
