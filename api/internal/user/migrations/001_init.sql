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
CREATE TABLE sessions (
    key VARCHAR(128) PRIMARY KEY,
    owner UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

---- create above / drop below ----

DROP TABLE idents;
DROP TABLE sessions;
DROP TABLE users;
DROP TYPE platform;
