# sns-sifter

X (旧Twitter) 上の情報をふるい分け・分析するツール群。
特定ユーザーのフォロー × キーワードのクロス分析で、有用な情報源を発見・追跡する。

## コンポーネント

| コンポーネント | 言語 | 用途 |
|---------------|------|------|
| **sifter CLI** | Go + SQLite | フォロー分析のバッチ処理。ローカルキャッシュで API コスト最適化 |
| **xmcp MCP サーバー** | Python | Claude Code から X API を対話的に利用 |

## クイックスタート

### sifter CLI

```bash
make build
export SIFTER_BEARER_TOKEN=<Bearer Token>

# フォロー一覧を取得・キャッシュ
./cmd/sifter/sifter sync following kaz_utb

# キーワードでフィルタ（APIコール不要）
./cmd/sifter/sifter following list kaz_utb --filter "AI"

# 前回との差分を表示
./cmd/sifter/sifter following diff kaz_utb

# 同期履歴・コスト確認
./cmd/sifter/sifter history kaz_utb

# いいね投稿の収集 (xmcp 経由 / OAuth 1.0a 必須)
./cmd/sifter/sifter sync likes kaz_utb
./cmd/sifter/sifter likes categorize kaz_utb
./cmd/sifter/sifter likes list kaz_utb --category ai --unreviewed
./cmd/sifter/sifter likes stats kaz_utb
```

### xmcp MCP サーバー

```bash
make setup                # Python 依存インストール
make xmcp/.env            # 認証情報テンプレートをコピー
vi xmcp/.env              # 認証情報を記入
make server               # MCP サーバー起動（ブラウザで OAuth 認証）
```

起動後、Claude Code を再起動すると X API ツールが利用可能になる。

## セットアップ

### X API 認証情報

[X Developer Platform](https://developer.x.com/en/portal/dashboard) でアプリを作成し、以下を取得:

- `X_OAUTH_CONSUMER_KEY` / `X_OAUTH_CONSUMER_SECRET` — OAuth1 認証用（xmcp）
- `X_BEARER_TOKEN` — Bearer Token 認証用（sifter CLI / xmcp 共通）

アプリ設定でコールバック URL を登録:
```
http://127.0.0.1:8976/oauth/callback
```

App permissions を **Read and Write** に設定（投稿・フォロー操作を行う場合）。

### コスト

X API は PPU（Pay-Per-Use）プラン。読み取り $0.005/件、書き込み $0.01/件。
sifter CLI のローカルキャッシュにより、2回目以降のフィルタ・差分検出は API コール不要。

## ドキュメント

運用・設計 (`docs/operations/`):
- [sifter CLI リファレンス](docs/operations/sifter-cli.md) — コマンド、SQLite スキーマ、コスト最適化
- [xmcp MCP サーバー](docs/operations/xmcp.md) — セットアップ、Claude Code 連携、トラブルシューティング
- [X API リファレンス](docs/operations/x-api.md) — エンドポイント、料金、ツール一覧
- [cron セットアップ](docs/operations/cron-setup.md) — ai-commander での定期 sync 手順

知識ベース (`docs/knowledge/`):
- [メモ・気付き](docs/knowledge/notes.md) — 開発中の気付き、試したいアイデア
- [Claude Code 新機能ノート](docs/knowledge/claude-code-features.md) — 新機能の継続蓄積
- [Managed Agents](docs/knowledge/managed-agents.md) — Managed Agents 入門/実践の継続記事
- [Managed Agents 評価](docs/knowledge/managed-agents-evaluation.md) — 非エンジニア向け業務自動化の適合性調査
