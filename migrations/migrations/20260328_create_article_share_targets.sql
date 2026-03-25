CREATE TABLE article_share_targets (
    id           BIGSERIAL PRIMARY KEY,
    article_id   BIGINT NOT NULL REFERENCES articles(id),
    provider     VARCHAR(32) NOT NULL,
    external_id  VARCHAR(255) NOT NULL DEFAULT '',
    external_url TEXT NOT NULL DEFAULT '',
    published_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT article_share_targets_article_id_provider_key UNIQUE (article_id, provider)
);

CREATE INDEX article_share_targets_article_id_idx ON article_share_targets (article_id);
