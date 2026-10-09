# sifter CLI リファレンス

## 概要

sifter は X API v2 を直接呼び出し、取得データを SQLite にキャッシュする Go CLI ツール。ローカルキャッシュにより再取得を抑える。

## ビルド

```bash
make build
# または
go build -o cmd/sifter/sifter ./cmd/sifter/
```

## 設定

### OAuth 1.0a User Context

`sync likes` の初回実行前に、X Developer Console でアプリの Consumer Key/Secret を取得し、コールバック URL `http://127.0.0.1:8976/oauth/callback` を登録する。次のログインは一度だけ実行する。

```bash
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
sifter auth login
sifter auth status
```

Consumer Key/Secret は `sifter auth login --consumer-key <key> --consumer-secret <secret>` でも指定できる。認可 URL が標準出力に表示されるのでブラウザで開く。ブラウザの自動起動は行わない。認可後、Consumer Key/Secret・Access Token/Secret・user_id・screen_name が `~/.sifter/credentials.json` に権限 0600 で保存される。保存先は `SIFTER_CREDENTIALS` で変更できる。`auth status` は保存先・screen_name・user_id を表示する。

コールバックは `SIFTER_OAUTH_CALLBACK_HOST`（既定 `127.0.0.1`）と `SIFTER_OAUTH_CALLBACK_PORT`（既定 `8976`）で設定できる。パスは `/oauth/callback`、待機時間は 300 秒。

ai-commander で Mac のブラウザを使う場合は、Mac のターミナルから `ssh -L 8976:127.0.0.1:8976 ai-commander` で接続し、その接続上で `sifter auth login` を実行する。ブラウザからのコールバックは SSH トンネル経由で ai-commander に届く。

保存済み OAuth 1.0a 資格情報があれば全 API 呼び出しに使用し、なければ `SIFTER_BEARER_TOKEN` による App-only 認証を使う。`sync likes` には保存済み資格情報が必要。OAuth 1.0a Access Token は期限切れしないが、利用者が失効させることはできる（[X 公式ガイド](https://docs.x.com/fundamentals/authentication/oauth-1-0a/obtaining-user-access-tokens)）。

### Bearer Token

保存済み OAuth 1.0a 資格情報がない場合、公開データの取得に使用する。以下のいずれかで設定:

```bash
# 環境変数
export SIFTER_BEARER_TOKEN=<token>

# コマンドラインフラグ
sifter --token <token> ...
```

### データベース

デフォルト: `~/.sifter/data.db`
`SIFTER_DB` または次のフラグで変更できる。

```bash
sifter --db /path/to/data.db ...
```

## コマンド

### sifter sync following \<username\>

指定ユーザーのフォロー一覧を X API から取得し、SQLite にキャッシュする。

```bash
sifter sync following kaz_utb
```

- 24時間以内に同期済みの場合はスキップ（`--force` で強制再取得）
- ユーザー名 → ID の解決結果もキャッシュ（2回目以降 API コール不要）
- フォロー数やAPIコール数、推定コストを表示

### sifter sync likes \<username\>

指定ユーザーのいいね投稿を X API から取得し、SQLite にページごとに保存する。保存済み OAuth 1.0a 資格情報が必要。

```bash
sifter sync likes kaz_utb
sifter sync likes kaz_utb --max 5 --max-pages 20
sifter sync likes kaz_utb --full
```

- incremental が既定。取得済みの投稿を見つけると終了
- `--full` は全件取得、`--max` はページあたり 5〜100 件、`--max-pages` はページ数の上限
- 取得したページは都度保存

### sifter following list \<username\>

キャッシュ済みのフォロー一覧を表示する（API コール不要）。

```bash
sifter following list kaz_utb                    # 全件表示
sifter following list kaz_utb --filter "AI"      # description/name/usernameでフィルタ
sifter following list kaz_utb --filter "agent"   # Agentic系に絞り込み
```

### sifter following diff \<username\>

前回と最新の同期結果を比較し、フォロー追加/解除を表示する。

```bash
sifter following diff kaz_utb
```

2回以上 `sync` を実行した後に利用可能。

### sifter following compare \<base\> \<ref\>

`ref` がフォローし、`base` がフォローしていないユーザーをローカル DB から抽出する。

```bash
sifter following compare kaz_utb zigorou
```

### sifter likes list / show / categorize / review / stats

いいね投稿の表示・分類・レビューはローカル DB で完結する。

```bash
sifter likes list kaz_utb --category ai --unreviewed
sifter likes list kaz_utb --keyword "agent" --limit 20
sifter likes show kaz_utb <tweet_id>
sifter likes categorize kaz_utb
sifter likes review kaz_utb <tweet_id>
sifter likes stats kaz_utb
```

### sifter history \<username\>

同期履歴を表示する。各同期の件数、APIコール数、推定コストを確認できる。

```bash
sifter history kaz_utb
```

### sifter debug following-page \<username\>

キャッシュ済みユーザー ID を使ってフォロー一覧の 1 ページを取得し、レスポンスを表示する。

```bash
sifter debug following-page kaz_utb --max-results 5
```

## SQLite スキーマ

### users テーブル

ユーザー情報のキャッシュ。UPSERT で常に最新に更新。

| カラム | 型 | 説明 |
|--------|-----|------|
| id | TEXT PK | X API ユーザーID |
| username | TEXT | @ハンドル |
| name | TEXT | 表示名 |
| description | TEXT | プロフィール文 |
| public_metrics_json | TEXT | フォロワー数等（JSON） |
| fetched_at | TEXT | 取得日時（RFC3339） |

### following テーブル

フォロー関係。sync_id と紐づけて世代管理。

| カラム | 型 | 説明 |
|--------|-----|------|
| source_user_id | TEXT | フォローしている側 |
| target_user_id | TEXT | フォローされている側 |
| sync_id | INTEGER | 同期ID |

### sync_history テーブル

同期の実行履歴。

| カラム | 型 | 説明 |
|--------|-----|------|
| id | INTEGER PK | 自動採番 |
| source_user_id | TEXT | 対象ユーザーID |
| sync_type | TEXT | 'following' / 'likes' |
| started_at | TEXT | 開始日時 |
| completed_at | TEXT | 完了日時 |
| total_fetched | INTEGER | 取得件数 |
| api_calls | INTEGER | APIコール数 |
| status | TEXT | running/completed/failed |

### liked_posts テーブル

取得したいいね投稿とローカルの分類・レビュー状態を保持する。

| カラム | 型 | 説明 |
|---|---|---|
| user_id | TEXT | いいねしたユーザー ID |
| tweet_id | TEXT | 投稿 ID |
| author_id | TEXT | 投稿者 ID |
| text | TEXT | 投稿本文 |
| created_at | TEXT | 投稿日時 |
| fetched_at | TEXT | 取得日時 |
| category | TEXT | 分類 |
| reviewed_at | TEXT | レビュー日時 |
| raw_json | TEXT | 元データ |

### レート制限

HTTP 429 では `x-rate-limit-reset` に従って最大 3 回再試行する。1 回の待機上限は既定 15 分で、`SIFTER_RATE_LIMIT_MAX_WAIT` により変更できる。

## コスト最適化戦略

1. **ページサイズ最大化** — `max_results=1000` で1回の API コールで最大件数を取得
2. **クールダウン** — 24時間以内の再同期をスキップ
3. **username→ID キャッシュ** — 2回目以降は DB から解決（API コール0）
4. **差分は DB 内で完結** — SQL クエリで新規/解除を検出（API コール0）
5. **フィルタはローカル** — `following list --filter` は DB 検索のみ
6. **いいね投稿は incremental** — 取得済みの投稿で停止し、ページごとに保存

[X 公式料金表](https://docs.x.com/x-api/getting-started/pricing) では Post Read は $0.005/件、User Read と Following/Followers Read は各 $0.010/件。認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリの所有者なら、自分の `liked_tweets` や `following` は Owned Reads の $0.001/件。5 件のいいね投稿が返れば、通常の Post Read は $0.025、Owned Reads は $0.005。1000 件のフォローが返れば、通常の Following Read は $10.00、Owned Reads は $1.00。実際の課金は返却件数と適用区分で決まる。
