# プロジェクト概要

## プロジェクトの目的
このプロジェクトは、Go言語とFirebase Genkitを使用したAI駆動のワークフローアプリケーションです。レシピ生成、シンプルなテキスト処理、詳細な研究の3つの主要な機能を提供します。MCP（Model Context Protocol）統合により外部ツールへのアクセスも可能です。

## 主な機能
1. **RecipeGeneratorFlow**: 材料と食事制限を入力として、構造化されたレシピデータを生成
2. **SimpleFlow**: MCPツールアクセス付きの基本的なAI対話
3. **DeepResearchFlow**: ユーザーとの対話とWebサーチを使用した包括的な研究レポート作成

## アーキテクチャ
- **main.go**: Genkitの初期化、MCPホストの設定、HTTPサーバー（ポート3400）の起動
- **flow/**: 各種フローの実装
  - recipe.go: レシピ生成フロー
  - simple.go: シンプルテキスト生成フロー
  - deepresearch.go: 多段階研究フロー
- **mcp/**: MCP（Model Context Protocol）関連のコード
  - ask-me/: Slack統合を含むMCPサーバー
  - local_mcp.go: MCPサーバー設定

## 技術スタック
- Go 1.24.5
- Firebase Genkit Go SDK
- Google AI (Gemini 2.5 Flash Lite) - デフォルトモデル
- MCP Go library
- Slack API (MCPツール用)

## 依存関係
- GEMINI_API_KEY環境変数（.env.localに設定）
- SLACK_OAUTH_TOKEN、SLACK_CHANNEL（MCP機能用）
- npm:genkit-cli（mise経由でインストール）

## API エンドポイント
- POST /recipeGeneratorFlow: レシピ生成
- POST /simpleFlow: MCPツールサポート付き基本テキスト生成
- POST /deepResearchFlow: ユーザー対話とWebサーチを使用した包括的な研究