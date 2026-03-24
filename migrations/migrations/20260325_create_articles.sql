CREATE TABLE articles (
    id                BIGSERIAL    PRIMARY KEY,
    user_id           BIGINT       NOT NULL REFERENCES users(id),
    prompt_run_job_id BIGINT       REFERENCES prompt_run_jobs(id) UNIQUE,
    title             VARCHAR(255) NOT NULL,
    content           TEXT         NOT NULL,
    deleted_at        TIMESTAMPTZ,
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX articles_user_id_deleted_at_updated_at_desc_idx
    ON articles (user_id, deleted_at, updated_at DESC);
