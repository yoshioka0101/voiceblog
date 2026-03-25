CREATE TABLE external_tokens (
    id              BIGSERIAL PRIMARY KEY,
    user_id         BIGINT NOT NULL REFERENCES users(id),
    provider        VARCHAR(32) NOT NULL,
    encrypted_token BYTEA NOT NULL,
    nonce           BYTEA NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT external_tokens_user_id_provider_key UNIQUE (user_id, provider)
);

CREATE INDEX external_tokens_user_id_idx ON external_tokens (user_id);
