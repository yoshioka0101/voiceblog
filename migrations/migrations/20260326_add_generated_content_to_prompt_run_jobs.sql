ALTER TABLE prompt_run_jobs
    ADD COLUMN generated_title VARCHAR(255),
    ADD COLUMN generated_content TEXT;
