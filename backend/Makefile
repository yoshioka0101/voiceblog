APP_NAME=voiceblog-api
-include .env
export
POSTGRES_PORT ?= 5433
DB_DSN ?= postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@127.0.0.1:$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable
ATLAS_MIGRATION_DIR := file://migrations/migrations
ATLAS_SCHEMA_FILE := file://migrations/schema.hcl
ATLAS_DEV_URL ?= docker://postgres/16/dev?search_path=public
.DEFAULT_GOAL := help

.PHONY: help dev server lint test fmt gen-api gen-db redoc db-up db-down migrate-diff migrate-apply migrate-status seed

help:
	@printf "使い方: make <ターゲット>\n\n"
	@printf "ターゲット一覧:\n"
	@printf "  dev              APIサーバー起動（serverのエイリアス）\n"
	@printf "  server           APIサーバー起動\n"
	@printf "  lint             golangci-lint 実行\n"
	@printf "  test             go test 実行\n"
	@printf "  fmt              Goファイルを gofmt\n"
	@printf "  gen-api          OpenAPI の bundle を生成\n"
	@printf "  gen-db           bobgen-psql で DB モデルを生成\n"
	@printf "  redoc            OpenAPI ドキュメントを localhost:3001 で配信\n"
	@printf "  db-up            PostgreSQL を起動\n"
	@printf "  db-down          PostgreSQL を停止\n"
	@printf "  migrate-diff     Atlas マイグレーション差分を生成\n"
	@printf "  migrate-apply    Atlas マイグレーションを適用\n"
	@printf "  migrate-status   Atlas マイグレーションの状態を確認\n"
	@printf "  seed             初期データを投入\n"

dev:
	go run ./cmd/api

server:
	go run ./cmd/api

lint:
	go tool golangci-lint run ./...

test:
	go test ./...

fmt:
	gofmt -w $(shell rg --files -g '*.go')

gen-api:
	npx @redocly/cli@latest bundle openapi/openapi.yaml -o openapi/gen-openapi.yaml
	go tool oapi-codegen -config openapi/oapi-codegen.yaml openapi/gen-openapi.yaml

gen-db: ## bobgen-psql で DB モデルを生成
	PSQL_DSN="$(DB_DSN)" go tool bobgen-psql

redoc:
	npx @redocly/cli@1.34.2 preview-docs -p 3001 openapi/openapi.yaml

db-up:
	docker compose -f docker-compose.yml up -d

db-down:
	docker compose -f docker-compose.yml down

migrate-diff:
	cd .. && atlas migrate diff $(name) --dir "$(ATLAS_MIGRATION_DIR)" --to "$(ATLAS_SCHEMA_FILE)" --dev-url "$(ATLAS_DEV_URL)"

migrate-apply:
	cd .. && atlas migrate apply --dir "$(ATLAS_MIGRATION_DIR)" --url "$(DB_DSN)"

migrate-status:
	cd .. && atlas migrate status --dir "$(ATLAS_MIGRATION_DIR)" --url "$(DB_DSN)"

seed:
	go run ./cmd/seed
