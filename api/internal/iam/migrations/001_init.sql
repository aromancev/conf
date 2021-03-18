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

---- create above / drop below ----

DROP TABLE idents;
DROP TABLE users;
DROP TYPE platform;
