env "local" {
  url = env("DB_DSN")
  migration {
    dir = "file://migrations/migrations"
  }
  schema {
    src = "file://migrations/schema.hcl"
  }
}
