# sns-sifter
SNS投稿をふるい分け・分析するツール

## セットアップ

### 1. 依存パッケージのインストール

```bash
make setup
```

### 2. X API 認証情報の設定

[X Developer Platform](https://developer.x.com/en/portal/dashboard) でアプリを作成し、`xmcp/.env` に認証情報を設定する。

```bash
make xmcp/.env   # テンプレートからコピー
vi xmcp/.env     # 認証情報を記入
```

必須項目:
- `X_OAUTH_CONSUMER_KEY`
- `X_OAUTH_CONSUMER_SECRET`
- `X_BEARER_TOKEN`

X Developer App の設定画面でコールバックURLを登録:
```
http://127.0.0.1:8976/oauth/callback
```

### 3. MCP サーバーの起動

```bash
make server
```

起動するとブラウザが開き OAuth1 認証が行われる。認証後、`http://127.0.0.1:8000/mcp` でMCPサーバーが利用可能になる。

## 使い方

MCPサーバー起動後、Claude Code から X API のツールが利用可能になる（`.mcp.json` で設定済み）。

### 情報収集に使える主なツール

| ツール | 用途 |
|--------|------|
| `searchPostsRecent` | 最近の投稿を検索 |
| `getPostsById` | 特定の投稿を取得 |
| `getUsersByUsername` | ユーザー情報を取得 |
| `searchUsers` | ユーザーを検索 |
| `getUsersFollowers` / `getUsersFollowing` | フォロワー/フォロー一覧 |
| `getTrendsByWoeid` | トレンドを取得 |
| `getPostsLikingUsers` | いいねしたユーザー一覧 |

### ツールの制限

`xmcp/.env` の `X_API_TOOL_ALLOWLIST` で使用可能なツールを制限できる。
