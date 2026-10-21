# X API v2 リファレンス

## アクセスと認証

sifter CLI は Go の HTTP クライアントから X API v2 を直接呼び出す。保存済み OAuth 1.0a 資格情報があれば全 API 呼び出しに使い、なければ `SIFTER_BEARER_TOKEN` による App-only 認証を使う。`sync likes` は OAuth 1.0a 資格情報を必要とする。

`sifter auth login` は Consumer Key/Secret を使って認可 URL を表示し、OAuth 1.0a の認可後に Access Token/Secret を `~/.sifter/credentials.json`（権限 0600）へ保存する。保存先は `SIFTER_CREDENTIALS` で変更できる。OAuth 1.0a Access Token は期限切れしないが、利用者が失効させることはできる。[X 公式 OAuth 1.0a ガイド](https://docs.x.com/fundamentals/authentication/oauth-1-0a/obtaining-user-access-tokens) を参照。

X の [認証マッピング](https://docs.x.com/fundamentals/authentication/guides/v2-authentication-mapping) は `GET /2/users/:id/liked_tweets` の OAuth 1.0a User Context 対応を示している。

## 料金

[X 公式料金表](https://docs.x.com/x-api/getting-started/pricing) の従量課金区分を参照。料金は変更されるため、実行前に Developer Console の表示を確認する。

| 対象 | 単価 |
|---|---:|
| Post Read | $0.005/件 |
| Owned Reads | $0.001/件 |
| User Read | $0.010/件 |
| Following/Followers Read | $0.010/件 |

Owned Reads は認証ユーザー ID がパスの `{id}` と一致し、そのユーザーが開発アプリの所有者である場合、自分の `liked_tweets` や `following` などに適用される。月間上限は 300 万 Post Read。

| 操作 | 件数 | 通常の読み取り | Owned Reads の条件を満たす場合 |
|---|---:|---:|---:|
| いいね投稿の取得 | 5件 | $0.025 | $0.005 |
| フォロー一覧の取得 | 100件 | $1.00 | $0.10 |
| フォロー一覧の取得 | 1000件 | $10.00 | $1.00 |
| ユーザー情報の取得 | 1件 | $0.010 | 対象外 |

表は単価と返却件数による計算例。X の 24 時間 UTC 単位の重複課金排除など、実際の請求条件は公式料金表を参照する。

## 使用するエンドポイント

| エンドポイント | 用途 | 主なパラメータ |
|---|---|---|
| `GET /2/users/by/username/:username` | ユーザー名から ID を解決 | `user.fields` |
| `GET /2/users/:id/following` | フォロー一覧を取得 | `max_results`, `pagination_token`, `user.fields` |
| `GET /2/users/:id/liked_tweets` | いいね投稿を取得 | `max_results`（5～100）, `pagination_token`, `tweet.fields`, `expansions`, `user.fields` |

ページングはレスポンスの `meta.next_token` を次リクエストの `pagination_token` に渡す。いいね投稿はページごとに SQLite へ保存する。

## レート制限

HTTP 429 が返った場合、sifter は `x-rate-limit-reset` に従って待機し、最大 3 回再試行する。1 回の待機上限は既定で 15 分。環境変数 `SIFTER_RATE_LIMIT_MAX_WAIT` で変更できる。
