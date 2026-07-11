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
    on {
      column = column.user_id
    }
    on {
      column = column.created_at
      desc   = true
    }
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
    on {
      column = column.user_id
    }
    on {
      column = column.is_active
    }
    on {
      column = column.created_at
      desc   = true
    }
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
  index "articles_user_id_deleted_at_updated_at_desc_idx" {
    on {
      column = column.user_id
    }
    on {
      column = column.deleted_at
    }
    on {
      column = column.updated_at
      desc   = true
    }
  }
}

table "external_tokens" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "user_id" {
    type = bigint
  }
  column "provider" {
    type = varchar(32)
  }
  column "encrypted_token" {
    type = bytea
  }
  column "nonce" {
    type = bytea
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
  foreign_key "external_tokens_user_id_fkey" {
    columns     = [column.user_id]
    ref_columns = [table.users.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  unique "external_tokens_user_id_provider_key" {
    columns = [column.user_id, column.provider]
  }
  index "external_tokens_user_id_idx" {
    columns = [column.user_id]
  }
}

table "article_share_targets" {
  schema = schema.voiceblog
  column "id" {
    type = bigserial
  }
  column "article_id" {
    type = bigint
  }
  column "provider" {
    type = varchar(32)
  }
  column "external_id" {
    type    = varchar(255)
    default = ""
  }
  column "external_url" {
    type    = text
    default = ""
  }
  column "published_at" {
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
  foreign_key "article_share_targets_article_id_fkey" {
    columns     = [column.article_id]
    ref_columns = [table.articles.column.id]
    on_update   = NO_ACTION
    on_delete   = NO_ACTION
  }
  unique "article_share_targets_article_id_provider_key" {
    columns = [column.article_id, column.provider]
  }
  index "article_share_targets_article_id_idx" {
    columns = [column.article_id]
  }
}
