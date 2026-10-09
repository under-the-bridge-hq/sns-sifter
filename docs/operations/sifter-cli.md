# sifter CLI リファレンス

sifter は X API v2 からフォロー一覧といいね投稿を取得し、SQLite に保存する Go CLI ツール。ローカルの一覧表示・差分比較・分類・レビュー操作は API を呼ばない。

## ビルド

```bash
make build
```

ビルド結果は `./cmd/sifter/sifter`。

## OAuth 1.0a 認証

X Developer Console でアプリの Consumer Key/Secret を取得し、コールバック URL `http://127.0.0.1:8976/oauth/callback` を登録する。初回のみ次を実行する。

```bash
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
./cmd/sifter/sifter auth login
./cmd/sifter/sifter auth status
```

Consumer Key/Secret は `auth login --consumer-key <key> --consumer-secret <secret>` でも指定できる。ログインは認可 URL を標準出力に表示し、ブラウザは自動で開かない。認可後、Consumer Key/Secret・Access Token/Secret・user_id・screen_name が `~/.sifter/credentials.json` に権限 0600 で保存される。保存先は `SIFTER_CREDENTIALS` で変更できる。`auth status` は保存先・screen_name・user_id を表示する。

コールバックの設定は `SIFTER_OAUTH_CALLBACK_HOST`（既定 `127.0.0.1`）と `SIFTER_OAUTH_CALLBACK_PORT`（既定 `8976`）。パスは `/oauth/callback`、待機時間は 300 秒。

ai-commander 上で Mac のブラウザから認可する場合は、Mac のターミナルで `ssh -L 8976:127.0.0.1:8976 ai-commander` を開始し、その SSH 接続上で `auth login` を実行する。表示された URL を Mac のブラウザで開く。コールバックは SSH トンネル経由で ai-commander に届く。

OAuth 1.0a Access Token は期限切れしないが、ユーザーが失効させることはできる（[X 公式ガイド](https://docs.x.com/fundamentals/authentication/oauth-1-0a/obtaining-user-access-tokens)）。

保存済み資格情報があれば全 API 呼び出しに OAuth 1.0a を使う。ない場合は `SIFTER_BEARER_TOKEN`（または `--token`）を使う。`sync likes` には保存済み資格情報が必要。

## データベース

既定は `~/.sifter/data.db`。`SIFTER_DB` または `--db <path>` で変更できる。

## コマンド

| コマンド | 内容 |
|---|---|
| `auth login` | OAuth 1.0a の認可と資格情報の保存 |
| `auth status` | 保存先・screen_name・user_id の表示 |
| `sync following <username> [--force]` | フォロー一覧の取得とキャッシュ。24 時間以内に同期済みならスキップ |
| `sync likes <username> [--full] [--max N] [--max-pages N]` | いいね投稿の同期。既定は incremental、取得ページごとに保存 |
| `following list <username> [--filter <keyword>]` | キャッシュ済みフォロー一覧の表示 |
| `following diff <username>` | 前回と最新のフォロー一覧の差分 |
| `following compare <base> <ref>` | ref のフォローのうち base にないアカウントを抽出 |
| `likes list <username>` | キャッシュ済み投稿の表示。カテゴリ・未レビュー・キーワードで絞り込み可能 |
| `likes show <username> <tweet_id>` | 投稿の詳細 |
| `likes categorize <username>` | 未分類投稿のキーワード分類 |
| `likes review <username> <tweet_id>...` | レビュー済みとして記録 |
| `likes stats <username>` | 件数・カテゴリ別・未レビュー件数の表示 |
| `history <username>` | 同期履歴の表示 |

例:

```bash
./cmd/sifter/sifter sync following kaz_utb
./cmd/sifter/sifter following list kaz_utb --filter "AI"
./cmd/sifter/sifter following diff kaz_utb
./cmd/sifter/sifter sync likes kaz_utb --max 5 --max-pages 20
./cmd/sifter/sifter likes categorize kaz_utb
./cmd/sifter/sifter likes list kaz_utb --category ai --unreviewed
./cmd/sifter/sifter likes review kaz_utb <tweet_id>
```

## レート制限と料金

HTTP 429 では `x-rate-limit-reset` に従い最大 3 回再試行する。1 回の待機上限は既定で 15 分。`SIFTER_RATE_LIMIT_MAX_WAIT` で変更できる。

[X 公式料金表](https://docs.x.com/x-api/getting-started/pricing) では Post Read は $0.005/件、User Read と Following/Followers Read は各 $0.010/件。認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリの所有者なら、自分の `liked_tweets` や `following` は Owned Reads の $0.001/件。5 件のいいね投稿が返る場合、通常の Post Read は $0.025、Owned Reads は $0.005。1000 件のフォローが返る場合、通常の Following Read は $10.00、Owned Reads は $1.00。

キャッシュ済みの一覧・差分・分類は API コールを行わない。料金は変更されるため、実行時の単価と課金実績は Developer Console で確認する。

## SQLite の主なテーブル

| テーブル | 内容 |
|---|---|
| `users` | ユーザー ID、名前、プロフィール、公開メトリクス |
| `following` | フォロー関係。sync_id ごとに履歴を保持 |
| `liked_posts` | いいね投稿、カテゴリ、レビュー日時 |
| `sync_history` | following / likes の同期日時、件数、API 呼び出し数、状態 |
