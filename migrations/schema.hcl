schema "voiceblog" {
}

table "users" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "auth_provider" {
    type = varchar(32)
  }
  column "auth_subject" {
    type = varchar(255)
  }
  column "email" {
    type    = varchar(255)
    null    = true
  }
  column "name" {
    type    = varchar(255)
    null    = true
  }
  column "created_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  column "updated_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  primary_key {
    columns = [column.id]
  }
  unique "users_auth_provider_auth_subject_key" {
    columns = [column.auth_provider, column.auth_subject]
  }
}

table "transcriptions" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "user_id" {
    type = bigint
  }
  column "full_text" {
    type = text
  }
  column "segments_json" {
    type = jsonb
  }
  column "created_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  column "updated_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "transcriptions_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "transcriptions_user_id_created_at_desc_idx" {
    columns = [column.user_id, column.created_at]
    desc    = [false, true]
  }
}

table "prompts" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "user_id" {
    type = bigint
    null = true
  }
  column "name" {
    type = varchar(255)
  }
  column "body" {
    type = text
  }
  column "is_active" {
    type    = boolean
    default = true
  }
  column "is_default" {
    type    = boolean
    default = false
  }
  column "created_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  column "updated_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "prompts_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  index "prompts_user_id_is_active_created_at_desc_idx" {
    columns = [column.user_id, column.is_active, column.created_at]
    desc    = [false, false, true]
  }
}

table "prompt_run_jobs" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "transcription_id" {
    type = bigint
  }
  column "prompt_id" {
    type = bigint
  }
  column "status" {
    type = varchar(32)
  }
  column "attempt_count" {
    type    = integer
    default = 0
  }
  column "next_run_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  column "error_message" {
    type = text
    null = true
  }
  column "generated_title" {
    type = varchar(255)
    null = true
  }
  column "generated_content" {
    type = text
    null = true
  }
  column "created_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "prompt_run_jobs_transcription_id_fkey" {
    columns     = [column.transcription_id]
    ref_columns = [table.transcriptions.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "prompt_run_jobs_prompt_id_fkey" {
    columns     = [column.prompt_id]
    ref_columns = [table.prompts.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  check "prompt_run_jobs_status_check" {
    expr = "((status)::text = ANY ((ARRAY['pending'::character varying, 'running'::character varying, 'completed'::character varying, 'failed'::character varying])::text[]))"
  }
  index "prompt_run_jobs_status_next_run_at_created_at_idx" {
    columns = [column.status, column.next_run_at, column.created_at]
  }
}

table "articles" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "user_id" {
    type = bigint
  }
  column "prompt_run_job_id" {
    type = bigint
    null = true
  }
  column "title" {
    type = varchar(255)
  }
  column "content" {
    type = text
  }
  column "deleted_at" {
    type = timestamptz
    null = true
  }
  column "created_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  column "updated_at" {
    type    = timestamptz
    default = sql("NOW()")
  }
  primary_key {
    columns = [column.id]
  }
  foreign_key "articles_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  foreign_key "articles_prompt_run_job_id_fkey" {
    columns     = [column.prompt_run_job_id]
    ref_columns = [table.prompt_run_jobs.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  unique "articles_prompt_run_job_id_key" {
    columns = [column.prompt_run_job_id]
  }
  index "articles_user_id_deleted_at_updated_at_desc_idx" {
    columns = [column.user_id, column.deleted_at, column.updated_at]
    desc    = [false, false, true]
  }
}
