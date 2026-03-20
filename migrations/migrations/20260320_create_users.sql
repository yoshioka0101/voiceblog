CREATE TABLE users (
    id            BIGSERIAL    PRIMARY KEY,
    auth_provider VARCHAR(32)  NOT NULL,
    auth_subject  VARCHAR(255) NOT NULL,
    email         VARCHAR(255),
    name          VARCHAR(255),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (auth_provider, auth_subject)
);
