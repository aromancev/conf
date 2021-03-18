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

---- create above / drop below ----

DROP TABLE confas;
DROP TABLE talks;
