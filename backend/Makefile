APP_NAME=voiceblog-api
.DEFAULT_GOAL := help

.PHONY: help dev server lint test fmt gen-api redoc

help:
	@printf "使い方: make <ターゲット>\n\n"
	@printf "ターゲット一覧:\n"
	@printf "  dev       APIサーバー起動（serverのエイリアス）\n"
	@printf "  server    APIサーバー起動\n"
	@printf "  lint      golangci-lint 実行\n"
	@printf "  test      go test 実行\n"
	@printf "  fmt       Goファイルを gofmt\n"
	@printf "  gen-api   OpenAPI の bundle を生成\n"
	@printf "  redoc     OpenAPI ドキュメントを localhost:9999 で配信\n"

dev:
	go run ./cmd/api

server:
	go run ./cmd/api

lint:
	golangci-lint run ./...

test:
	go test ./...

fmt:
	gofmt -w $(shell rg --files -g '*.go')

gen-api:
	npx @redocly/cli@latest bundle openapi/openapi.yaml -o openapi/gen-openapi.yaml

redoc:
	npx @redocly/cli@1.34.2 preview-docs -p 3001 openapi/openapi.yaml

