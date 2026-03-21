# VoiceBlog

VoiceBlog は、音声ベースのブログプラットフォームを目指すプロジェクトです。
バックエンドは Go (Gin + PostgreSQL)、モバイルアプリは iOS (SwiftUI) で構成されています。

## プロジェクト構成

```
voiceblog/
├── backend/                  # Go API サーバー
│   ├── cmd/api/              # エントリポイント
│   ├── internal/
│   │   ├── config/           # 環境変数・設定
│   │   ├── db/               # DB接続
│   │   ├── di/               # DIコンテナ
│   │   ├── feature/          # 機能単位の実装
│   │   │   ├── auth/         #   認証 feature
│   │   │   │   ├── domain/
│   │   │   │   ├── infra/googleauth/
│   │   │   │   └── usecase/
│   │   │   ├── health/       #   ヘルスチェック feature
│   │   │   └── user/         #   ユーザー feature
│   │   │       ├── domain/
│   │   │       ├── handler/
│   │   │       ├── infra/postgres/
│   │   │       └── usecase/
│   │   ├── handler/          # HTTPハンドラ・ルーティング
│   │   │   └── middleware/   #   認証ミドルウェア
│   │   ├── logger/           # ロギング
│   │   ├── server/           # HTTPサーバー起動
│   │   └── testutil/         # テストヘルパー（testcontainers）
│   └── openapi/              # OpenAPI仕様書
├── ios/                      # iOS アプリ (SwiftUI)
│   ├── Makefile              # build / lint / clean
│   └── VoiceBlog/VoiceBlog/
│       ├── Auth/             # AuthManager, KeychainHelper
│       ├── Models/           # User
│       ├── Views/            # LoginView, HomeView
│       ├── Network/          # APIClient
│       ├── ContentView.swift # ルーティング
│       └── VoiceBlogApp.swift# エントリポイント
└── migrations/               # Atlas マイグレーション
    ├── schema.hcl            # スキーマ定義
    └── migrations/           # SQL ファイル
```

## 技術スタック

| レイヤー | 技術 |
|---|---|
| バックエンド | Go 1.25 / Gin / PostgreSQL |
| ORM / クエリビルダ | bob |
| 認証 | OIDC（Google JWT検証） |
| iOS | SwiftUI / @Observable (iOS 17+) |
| iOS 認証 | Google Sign-In SDK |
| DB マイグレーション | Atlas |
| テスト | testcontainers-go（インテグレーション） |
| コンテナ | Docker Compose / colima |

## セットアップ

### バックエンド (Go)

1. `backend/.env.sample` を `.env` にコピーして環境変数を設定。
2. Docker Desktop または Colima を起動する。
3. Docker で DB を起動し、マイグレーションを適用してサーバーを起動。

```bash
cd backend
cp .env.sample .env
make db-up
make migrate-apply
make dev
```

開発用 PostgreSQL はホスト側 `5433` 番ポートを使います。ローカルの別 PostgreSQL が `5432` を使っていても衝突しません。

#### 主な Make ターゲット

| コマンド | 説明 |
|---|---|
| `make dev` | APIサーバー起動 |
| `make test` | テスト実行 |
| `make lint` | golangci-lint |
| `make fmt` | コードフォーマット |
| `make gen-api` | OpenAPI バンドル |
| `make gen-db` | bobgen でDB型生成 |
| `make db-up` / `db-down` | PostgreSQL起動/停止 |
| `make migrate-apply` | マイグレーション適用 |
| `make migrate-status` | マイグレーション状態確認 |
| `make redoc` | APIドキュメント表示 (localhost:3001) |

### モバイルアプリ (iOS)

Xcode 15 以上が必要。Google Sign-In SDK は SPM で管理。

```bash
cd ios
make build    # コマンドラインビルド（macOS）
make lint     # SwiftLint
make clean    # クリーンビルド
```

または `ios/VoiceBlog/VoiceBlog.xcodeproj` を Xcode で開いて実行。

### データベースマイグレーション (Atlas)

```bash
cd backend
make migrate-status   # 状態確認
make migrate-apply    # 適用
make migrate-diff     # スキーマ差分生成
```

## アーキテクチャ

### バックエンド（Feature Architecture）

```
handler/router → feature/<name>/{handler,usecase,domain,infra}
                         ↑
                    middleware は共通レイヤーに維持
```

- **feature 単位で完結**: `auth`、`user`、`health` ごとに `domain/usecase/infra/handler` を持ち、並行実装時の衝突を減らす
- **router は薄く維持**: ルーティングと middleware の接続だけを担い、feature の組み立ては `di` と各 `feature` package に寄せる
- **middleware は共通のまま**: 認証 middleware は `internal/handler/middleware` に残し、feature 側の usecase / domain にだけ依存させる

### iOS（MVVM）

```
View → ViewModel (AuthManager) → Model / Service (APIClient)
```

- **Models**: データ構造（`Codable`）
- **Views**: SwiftUI の宣言的 UI
- **Auth**: 認証状態管理（ViewModel 相当）
- **Network**: API 通信（`actor` でスレッドセーフ）

## API ドキュメント

OpenAPI 仕様書は `backend/openapi/` に定義。

```bash
cd backend
make redoc
```

http://localhost:3001 でドキュメントを閲覧できます。
