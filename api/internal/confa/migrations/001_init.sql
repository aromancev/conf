CREATE TABLE confas (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    handle VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

---- create above / drop below ----

DROP TABLE confas;
