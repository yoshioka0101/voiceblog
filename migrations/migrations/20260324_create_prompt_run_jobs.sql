CREATE TABLE prompt_run_jobs (
    id               BIGSERIAL   PRIMARY KEY,
    transcription_id BIGINT      NOT NULL REFERENCES transcriptions(id),
    prompt_id        BIGINT      NOT NULL REFERENCES prompts(id),
    status           VARCHAR(32) NOT NULL,
    attempt_count    INTEGER     NOT NULL DEFAULT 0,
    next_run_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    error_message    TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT prompt_run_jobs_status_check
        CHECK (status IN ('pending', 'running', 'completed', 'failed'))
);

CREATE INDEX prompt_run_jobs_status_next_run_at_created_at_idx
    ON prompt_run_jobs (status, next_run_at, created_at);
