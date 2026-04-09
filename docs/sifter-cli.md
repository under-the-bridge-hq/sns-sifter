# sifter CLI リファレンス

## 概要

sifter は X API v2 を直接呼び出し、取得データを SQLite にキャッシュする Go CLI ツール。MCP 経由と異なり、ローカルキャッシュにより API コストを最小化できる。

## ビルド

```bash
make build
# または
go build -o cmd/sifter/sifter ./cmd/sifter/
```

## 設定

### Bearer Token

以下のいずれかで設定:

```bash
# 環境変数
export SIFTER_BEARER_TOKEN=<token>

# コマンドラインフラグ
sifter --token <token> ...
```

### データベース

デフォルト: `~/.sifter/data.db`

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

### sifter history \<username\>

同期履歴を表示する。各同期の件数、APIコール数、推定コストを確認できる。

```bash
sifter history kaz_utb
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
| sync_type | TEXT | 'following' |
| started_at | TEXT | 開始日時 |
| completed_at | TEXT | 完了日時 |
| total_fetched | INTEGER | 取得件数 |
| api_calls | INTEGER | APIコール数 |
| status | TEXT | running/completed/failed |

## コスト最適化戦略

1. **ページサイズ最大化** — `max_results=1000` で1回の API コールで最大件数を取得
2. **クールダウン** — 24時間以内の再同期をスキップ
3. **username→ID キャッシュ** — 2回目以降は DB から解決（API コール0）
4. **差分は DB 内で完結** — SQL クエリで新規/解除を検出（API コール0）
5. **フィルタはローカル** — `following list --filter` は DB 検索のみ
