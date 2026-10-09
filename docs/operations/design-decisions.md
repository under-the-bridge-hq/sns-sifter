# 設計方針・経緯

## プロジェクトの目的

X (旧Twitter) 上の情報をふるい分け、有用なナレッジベースを構築する。
具体的には「特定ユーザーのフォロー × キーワード」のクロス分析で、AI/Agentic Engineering 領域の有用な情報源を発見・追跡する。

## コンポーネントの棲み分け

### sifter CLI（Go + SQLite）— 情報収集・分析の主軸

- OAuth 1.0a の保存済み資格情報があれば全 API 呼び出しに使用。なければ Bearer Token を使用
- ローカル SQLite にキャッシュし、API コストを最小化
- バッチ処理向き。自動化・定期実行に適する
- 読み取り専用操作のみ

### xmcp 経由から Go 直接呼び出しへの移行

2026 年 4 月の [PR #4](https://github.com/under-the-bridge-hq/sns-sifter/pull/4) で、いいね投稿の収集に xmcp の HTTP MCP クライアントが追加された。その構成では、いいね投稿の同期に xmcp の起動、OAuth 認可、常駐が必要だった。2026 年 10 月、sifter が OAuth 1.0a で X API を直接呼び出す構成に変更し、xmcp を削除することが決定された。日次同期に別プロセスは不要になる。

`sifter auth login` で Access Token/Secret を一度取得して `~/.sifter/credentials.json` に権限 0600 で保存する。いいね投稿の同期には保存済み資格情報が必要。[X 公式認証マッピング](https://docs.x.com/fundamentals/authentication/guides/v2-authentication-mapping) は `GET /2/users/:id/liked_tweets` の OAuth 1.0a User Context 対応を示している。

## 技術選定の理由

### X API クライアント: net/http 直接（SDK 不使用）

- 公式 Go SDK が存在しない
- 使用エンドポイントは 2-3 個と少ない
- サードパーティライブラリの依存を避け、メンテナンスリスクを低減

### SQLite ドライバ: modernc.org/sqlite

- CGo 不要（Pure Go）。クロスコンパイル可能
- `database/sql` 互換。将来のドライバ差し替えが容易

### CLI フレームワーク: 自前（flag + os.Args）

- サブコマンド 4-5 個の規模に cobra 等のフレームワークは不要
- 外部依存を 1 パッケージ（sqlite ドライバ）に抑える方針

### .env 読み込み: 自前パーサー

- `godotenv` 等の依存を避ける
- `KEY=VALUE` のシンプルなパースのみ。既存環境変数は上書きしない

## API コスト最適化

X API は従量課金。[現行の公式料金表](https://docs.x.com/x-api/getting-started/pricing) では Post Read は $0.005/件、User Read と Following/Followers Read は各 $0.010/件。認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリ所有者であれば、自分の `liked_tweets` や `following` などは Owned Reads の $0.001/件が適用される。

### コスト実績

| 操作 | 件数 | コスト |
|------|------|--------|
| kaz_utb フォロー一覧取得 | 57件 | $0.29 |
| zigorou フォロー一覧取得 | 1000件 | $5.00 |

この表は 2026 年 4 月時点の実績。当時は $0.005/件で課金された。現行単価を当てはめた再計算値ではない。

### 現行単価による計算例

| 操作 | 件数 | 通常の読み取り | Owned Reads 適用時 |
|---|---:|---:|---:|
| フォロー一覧取得 | 57件 | $0.57 | $0.057 |
| フォロー一覧取得 | 1000件 | $10.00 | $1.00 |
| いいね投稿取得 | 5件 | $0.025 | $0.005 |

Owned Reads 列は認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリ所有者である場合の計算例。

### 最適化戦略

1. **ページサイズ最大化**: `max_results=1000`
2. **クールダウン**: 24h 以内の再同期はスキップ（`--force` で上書き）
3. **username→ID キャッシュ**: 2回目以降は DB から解決（API コール 0）
4. **差分検出は DB 内で完結**: SQL クエリのみ（API コール 0）
5. **フィルタはローカル**: `following list --filter` は DB 検索のみ

いいね投稿は取得ページごとに SQLite に保存する。

### レート制限の注意

- フォロー API（POST）: 15分あたり約 30 リクエストでレート制限
- フォロー一覧取得（GET）: 比較的緩い
- レート制限到達時は `x-rate-limit-reset` ヘッダに従い待機。最大 3 回再試行し、1 回の待機上限は既定 15 分（`SIFTER_RATE_LIMIT_MAX_WAIT` で変更）

## 今後の拡張予定

### 次のステップ

1. **`sifter search` コマンド** — フォロー中ユーザーの投稿をキーワード検索してキャッシュ（ナレッジ化のコア機能）
2. **`sifter following suggest` コマンド** — 他ユーザーのフォロー × キーワードでフォロー候補を検出
3. **レート制限のペーシング** — バッチ処理時の待機方法の改善

### 将来的な構想

- 定期実行（cron / scheduled agent）でフォロー変動を自動追跡
- 投稿のキャッシュ → キーワード抽出 → ナレッジベース構築
- 複数ユーザーのフォローをクロス分析して情報源のスコアリング
