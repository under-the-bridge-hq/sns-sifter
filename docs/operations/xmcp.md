# xmcp — X API MCP サーバー

## 概要

[xdevplatform/xmcp](https://github.com/xdevplatform/xmcp) は X API の OpenAPI 仕様を MCP (Model Context Protocol) サーバーとして公開する Python プロジェクト。Claude Code から X API のツールを直接呼び出せる。

本リポジトリでは Git サブモジュールとして `xmcp/` に配置。

## 仕組み

1. 起動時に `https://api.twitter.com/2/openapi.json` から OpenAPI 仕様を取得
2. FastMCP を使って各エンドポイントを MCP ツールとして動的に登録
3. OAuth1 認証フロー（ブラウザ経由）でアクセストークンを取得
4. `http://127.0.0.1:8000/mcp` で Streamable HTTP MCP サーバーとして待機

## セットアップ

```bash
make setup       # Python venv 作成 + 依存パッケージインストール
make xmcp/.env   # .env テンプレートをコピー
```

### 認証情報の設定

[X Developer Platform](https://developer.x.com/en/portal/dashboard) でアプリを作成し、`xmcp/.env` に以下を記入:

```
X_OAUTH_CONSUMER_KEY=<Consumer Key>
X_OAUTH_CONSUMER_SECRET=<Consumer Secret>
X_BEARER_TOKEN=<Bearer Token>
```

アプリの設定画面でコールバック URL を登録:
```
http://127.0.0.1:8976/oauth/callback
```

### App permissions

- **Read only** — 情報収集のみの場合
- **Read and Write** — 投稿・フォロー等の書き込みも行う場合

権限変更後は Consumer Keys と Bearer Token の再生成が必要。

## 起動

```bash
make server
```

ブラウザが開き OAuth1 認証が行われる。認証後 MCP サーバーが起動。

## Claude Code との連携

`.mcp.json` で接続設定済み:

```json
{
  "mcpServers": {
    "x-api": {
      "type": "http",
      "url": "http://127.0.0.1:8000/mcp"
    }
  }
}
```

**重要:** Claude Code は起動時に `.mcp.json` を読み込むため、MCP サーバーを起動した後に Claude Code を再起動する必要がある。

## ツール数の制限

xmcp はデフォルトで 130+ ツールを登録するが、Claude Code のツール読み込みがタイムアウトする。`X_API_TOOL_ALLOWLIST` で必要なツールのみに絞ること。

詳細は [X API リファレンス](x-api.md) を参照。

## トラブルシューティング

| 症状 | 原因 | 対処 |
|------|------|------|
| Claude Code に MCP ツールが表示されない | `.mcp.json` に `"type": "http"` がない | `"type": "http"` を追加して Claude Code 再起動 |
| ツールが認識されない | ツール数が多すぎる | `X_API_TOOL_ALLOWLIST` で絞る |
| 402 Payment Required | API クレジット不足 | Developer Portal でクレジット追加 |
| OAuth 認証が完了しない | コールバック URL 未登録 | アプリ設定で `http://127.0.0.1:8976/oauth/callback` を登録 |
