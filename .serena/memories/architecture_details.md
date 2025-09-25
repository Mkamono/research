# アーキテクチャ詳細

## システム構成

### メイン構成要素
1. **Genkit初期化**: Google AIプラグインとGemini 2.5 Flash Liteモデル
2. **MCPホスト**: 外部ツールとの統合
3. **HTTPサーバー**: ポート3400でAPI提供

### フロー詳細

#### RecipeGeneratorFlow
- **入力**: RecipeInput（材料、食事制限）
- **出力**: Recipe（構造化されたレシピデータ）
- **処理**: プロンプトベースでのレシピ生成

#### SimpleFlow  
- **入力**: SimpleInput（テキスト入力）
- **出力**: プレーンテキスト
- **特徴**: MCPツールアクセス可能

#### DeepResearchFlow
- **入力**: DeepResearchInput（トピック、目的、範囲、言語）
- **出力**: DeepResearchResult（包括的な研究レポート）
- **段階**: 
  1. 計画フェーズ
  2. 計画確認フェーズ  
  3. 研究フェーズ（Webサーチ使用）
  4. 統合フェーズ
  5. レポート提供フェーズ

### MCP統合

#### MCPサーバー設定
- **ask-me**: インタラクティブチャットサーバー
  - `chat`: Slackを通じたユーザーとの質疑応答
  - `get_thread_history`: スレッド履歴の取得

#### データフロー
1. MCPサーバーの接続（local_mcp.goで定義）
2. アクティブツールの取得
3. Genkitアクションとしての登録
4. フロー内でのツール使用

### データ構造

#### 入力/出力スキーマ
- JSON形式での構造化データ
- バリデーション付きスキーマ定義
- エラーハンドリングとレスポンス形式の統一

#### 環境設定
- `.env.local`: 環境変数（API キー等）
- `mise.toml`: 開発ツールとタスク定義
- `lefthook.yml`: Git フック設定

### HTTP エンドポイント構成
```
POST /recipeGeneratorFlow    -> RecipeGeneratorFlow
POST /simpleFlow            -> SimpleFlow  
POST /deepResearchFlow      -> DeepResearchFlow
```

### 依存関係管理
- go.mod/go.sumによる依存関係管理
- Firebase Genkit Go SDK
- Google AI API
- MCP Go library
- Slack API (WebSocket通信)