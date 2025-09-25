# 開発コマンド一覧

## mise タスク（推奨）
プロジェクトでは `mise` を使用してタスクを管理しています：

```bash
# 開発サーバーの起動（Genkit UIを含む）
mise run dev

# 開発サーバーの直接起動
mise run start

# 研究関連のポートをクリーンアップ（4033, 4000, 3400）
mise run kill-research-ports

# MCPサーバーの登録
mise run register-mcp

# Goコードのフォーマット
mise run fmt
```

## 直接実行コマンド
```bash
# アプリケーションの直接実行
go run .

# Goコードのフォーマット
go fmt ./...

# ビルド
go build ./...
```

## 開発環境の設定
1. `.env.local.example`を`.env.local`にコピー
2. `GEMINI_API_KEY`を設定
3. Slack統合を使用する場合は`SLACK_OAUTH_TOKEN`と`SLACK_CHANNEL`を設定

## アクセスURL
- HTTP API: http://localhost:3400
- Genkit Developer UI: 自動で開かれる（genkit startコマンド実行時）

## Git フック
- pre-commit: `go fmt ./...` を自動実行
- pre-push: `go build ./...` を自動実行

## システムコマンド（Darwin）
- `lsof -i :ポート番号 -nP`: ポート使用状況の確認
- `kill -9 プロセスID`: プロセス強制終了
- `ls`, `cd`, `grep`, `find`: 基本的なファイル操作