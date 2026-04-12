# X API v2 リファレンス

## 概要

X (旧Twitter) API v2 は REST ベースの API。sns-sifter では以下の2つの方法でアクセスする:
- **xmcp MCP サーバー** — Claude Code から対話的に利用
- **sifter CLI** — Go で直接 API を呼び出し、SQLite にキャッシュ

## 認証

### Bearer Token (App-only)

sifter CLI で使用。公開ユーザーのフォロー一覧取得等に十分。

```
Authorization: Bearer <token>
```

### OAuth 1.0a (User Context)

xmcp MCP サーバーで使用。投稿・いいね等のユーザー操作に必要。
起動時にブラウザで認証フローが実行される。

必要な認証情報:
- `X_OAUTH_CONSUMER_KEY`
- `X_OAUTH_CONSUMER_SECRET`
- `X_BEARER_TOKEN`

## 料金（PPU プラン）

2026年2月以降、新規開発者は Pay-Per-Use が唯一の選択肢。

| 操作 | 単価 |
|------|------|
| 投稿読み取り | $0.005/件 |
| 投稿作成 | $0.01/件 |

月間上限: 200万投稿読み取り（超過は Enterprise が必要）

### コスト例

| 操作 | 件数 | コスト |
|------|------|--------|
| フォロー一覧取得（1000人） | 1000 | $5.00 |
| フォロー一覧取得（100人） | 100 | $0.50 |
| ユーザー情報取得（1件） | 1 | $0.005 |

## sifter CLI が使用するエンドポイント

### GET /2/users/by/username/:username

ユーザー名から ID を解決する。

```
GET /2/users/by/username/kaz_utb?user.fields=id,username,name,description,public_metrics
```

### GET /2/users/:id/following

指定ユーザーのフォロー一覧を取得する。

```
GET /2/users/1836555519273910272/following?user.fields=id,username,name,description,public_metrics&max_results=1000
```

- `max_results`: 最大1000（ページサイズ最大化でコスト削減）
- `pagination_token`: 次ページのトークン（レスポンスの `meta.next_token`）

### レート制限

- ヘッダ `x-rate-limit-remaining` が 0 になったら `x-rate-limit-reset` まで待機
- sifter CLI は自動待機を実装済み

## xmcp MCP サーバーが提供するツール

### 読み取り系
| ツール | エンドポイント | 用途 |
|--------|---------------|------|
| `searchPostsRecent` | POST /2/tweets/search/recent | 最近7日間の投稿検索 |
| `getPostsById` | GET /2/tweets/:id | 特定投稿の取得 |
| `getUsersMe` | GET /2/users/me | 認証ユーザー情報 |
| `getUsersById` | GET /2/users/:id | ユーザー情報取得 |
| `getUsersByUsername` | GET /2/users/by/username/:username | ユーザー名で取得 |
| `searchUsers` | GET /2/users/search | ユーザー検索 |
| `getUsersFollowers` | GET /2/users/:id/followers | フォロワー一覧 |
| `getUsersFollowing` | GET /2/users/:id/following | フォロー一覧 |
| `getTrendsByWoeid` | GET /2/trends/by/woeid/:woeid | トレンド取得 |
| `getPostsLikingUsers` | GET /2/tweets/:id/liking_users | いいねユーザー |
| `getPostsQuotedPosts` | GET /2/tweets/:id/quote_tweets | 引用ポスト |
| `getPostsReposts` | GET /2/tweets/:id/retweets | リポスト |

### 書き込み系
| ツール | エンドポイント | 用途 |
|--------|---------------|------|
| `createPosts` | POST /2/tweets | 投稿作成 |
| `followUser` | POST /2/users/:id/following | フォロー |
| `unfollowUser` | DELETE /2/users/:source_id/following/:target_id | フォロー解除 |
| `likePost` | POST /2/users/:id/likes | いいね |
| `repostPost` | POST /2/users/:id/retweets | リポスト |

### ツール制限の設定

`xmcp/.env` の `X_API_TOOL_ALLOWLIST` でカンマ区切りで指定:

```
X_API_TOOL_ALLOWLIST=searchPostsRecent,getPostsById,getUsersById,...
```

全ツール（130+）を有効にすると Claude Code の読み込みがタイムアウトするため、必要なツールのみに絞ること。
