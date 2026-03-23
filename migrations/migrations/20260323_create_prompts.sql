CREATE TABLE prompts (
    id         BIGSERIAL    PRIMARY KEY,
    user_id    BIGINT       REFERENCES users(id),
    name       VARCHAR(255) NOT NULL,
    body       TEXT         NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    is_default BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX prompts_user_id_is_active_created_at_desc_idx
    ON prompts (user_id, is_active, created_at DESC);

INSERT INTO prompts (user_id, name, body, is_active, is_default)
VALUES (
    NULL,
    'ブログ記事作成（標準）',
    '以下の文字起こしをもとに、読みやすい日本語のブログ記事を作成してください。タイトルと導入文も含めてください。',
    TRUE,
    TRUE
);
