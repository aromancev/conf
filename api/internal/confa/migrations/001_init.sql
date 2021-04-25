CREATE TABLE confas (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    handle VARCHAR(64) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE talks (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    speaker UUID NOT NULL,
    confa UUID NOT NULL REFERENCES confas (id) ON DELETE CASCADE,
    handle VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE (confa, handle)
);

CREATE TABLE claps (
    id UUID PRIMARY KEY,
    owner UUID NOT NULL,
    speaker UUID NOT NULL,
    confa UUID NOT NULL REFERENCES confas (id) ON DELETE CASCADE,
    talk UUID NOT NULL REFERENCES talks (id) ON DELETE CASCADE,
    claps INT8 NOT NULL,
    CONSTRAINT unique_owner_talk UNIQUE (owner, talk)
);

---- create above / drop below ----

<<<<<<< HEAD
DROP TABLE IF EXISTS talks;
DROP TABLE IF EXISTS confas;
=======
DROP TABLE claps;
DROP TABLE talks;
DROP TABLE confas;
>>>>>>> 59617cdbd600be21597da200f9572f62deb2b7d7
