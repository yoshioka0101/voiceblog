-- prompt_run_jobs は非同期ジョブとして設計されたが、ワーカー不在のまま
-- 同期版の /articles/generate に置き換えられたため機能ごと削除する。
ALTER TABLE articles DROP CONSTRAINT IF EXISTS articles_prompt_run_job_id_fkey;
ALTER TABLE articles DROP COLUMN IF EXISTS prompt_run_job_id;
DROP TABLE IF EXISTS prompt_run_jobs;
