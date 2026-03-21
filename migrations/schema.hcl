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
