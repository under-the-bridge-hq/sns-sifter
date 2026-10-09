# 設計方針・経緯

## プロジェクトの目的

X (旧Twitter) 上の情報をふるい分け、有用なナレッジベースを構築する。特定ユーザーのフォローとキーワードを組み合わせて情報源を分析し、いいね投稿をローカルに蓄積する。

## sifter CLI の構成

sifter CLI（Go + SQLite）が X API を直接呼び出す。公開データの取得、フォロー一覧といいね投稿の同期、ローカル検索・分類・レビュー状態の管理を担う。

- `sifter auth login` で OAuth 1.0a の Access Token/Secret を一度取得し、`~/.sifter/credentials.json` に権限 0600 で保存する。
- 保存済み資格情報があれば全 API 呼び出しに OAuth 1.0a を使い、なければ `SIFTER_BEARER_TOKEN` を使う。いいね投稿の同期には保存済み資格情報が必要。
- SQLite キャッシュで重複取得を抑え、フォロー差分とフィルタリングは DB 内で処理する。

### xmcp 経由から Go 直接呼び出しへの移行

2026 年 4 月の [PR #4](https://github.com/under-the-bridge-hq/sns-sifter/pull/4) で、いいね投稿の収集に xmcp の HTTP MCP クライアントが追加された。その構成では、いいね投稿の同期に xmcp の起動、OAuth 認可、常駐が必要だった。2026 年 10 月、sifter が OAuth 1.0a で X API を直接呼び出す構成に変更し、xmcp を削除することが決定された。日次同期に別プロセスは不要になる。

## 技術選定

### X API クライアント

Go の `net/http` から直接呼び出し、OAuth 1.0a User Context と Bearer Token に対応する。OAuth 1.0a での `GET /2/users/:id/liked_tweets` 対応は [X 公式認証マッピング](https://docs.x.com/fundamentals/authentication/guides/v2-authentication-mapping) を参照。

### SQLite ドライバ

`modernc.org/sqlite` を使用する。CGo は不要で、`database/sql` から扱える。

## API コストと同期

[X 公式料金表](https://docs.x.com/x-api/getting-started/pricing) では Post Read は $0.005/件、User Read と Following/Followers Read は各 $0.010/件。認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリ所有者なら、自分の `liked_tweets` と `following` 等は Owned Reads の $0.001/件となる。料金は取得リソースと適用区分によって異なる。

| 計算例 | 件数 | 通常の読み取り | Owned Reads 適用時 |
|---|---:|---:|---:|
| フォロー一覧 | 57件 | $0.57 | $0.057 |
| フォロー一覧 | 1000件 | $10.00 | $1.00 |
| いいね投稿 | 5件 | $0.025 | $0.005 |

表は現在の単価を使った計算例であり、過去の請求実績ではない。

- フォロー一覧は 24 時間以内に同期済みならスキップし、`--force` で再取得できる。
- ユーザー名と ID の対応を SQLite にキャッシュする。
- いいね投稿は既定で incremental sync し、取得ページごとに保存する。
- HTTP 429 は最大 3 回再試行する。1 回の待機上限は既定 15 分で、`SIFTER_RATE_LIMIT_MAX_WAIT` で変更できる。

## 今後の拡張予定

- フォロー中ユーザーの投稿検索とキャッシュ
- 他ユーザーのフォローとキーワードを使った候補の抽出
- 定期実行によるフォロー変動の追跡
