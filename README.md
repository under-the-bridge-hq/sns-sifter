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

- [sifter CLI リファレンス](docs/sifter-cli.md) — コマンド、SQLite スキーマ、コスト最適化
- [xmcp MCP サーバー](docs/xmcp.md) — セットアップ、Claude Code 連携、トラブルシューティング
- [X API リファレンス](docs/x-api.md) — エンドポイント、料金、ツール一覧
