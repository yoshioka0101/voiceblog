CREATE TABLE transcriptions (
    id            BIGSERIAL   PRIMARY KEY,
    user_id       BIGINT      NOT NULL REFERENCES users(id),
    full_text     TEXT        NOT NULL,
    segments_json JSONB       NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX transcriptions_user_id_created_at_desc_idx
    ON transcriptions (user_id, created_at DESC);
