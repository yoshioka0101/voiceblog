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
