# CLAUDE.md

## プロジェクト概要

sns-sifter は X (旧Twitter) 上の情報をふるい分け・分析するツール群。
特定ユーザーのフォロー × キーワードのクロス分析を通じて、有用な情報・ナレッジベースを構築する。

## アーキテクチャ

2つのコンポーネントで構成:

- **xmcp MCP サーバー** (Python) — Claude Code から X API を対話的に利用
- **sifter CLI** (Go + SQLite) — フォロー分析のバッチ処理。ローカルキャッシュで API コスト最適化

## 開発ガイド

### ビルド・起動

```bash
make build          # sifter CLI をビルド
make setup          # xmcp の Python 依存をインストール
make server         # xmcp MCP サーバーを起動
```

### sifter CLI の実行

```bash
export SIFTER_BEARER_TOKEN=<token>    # xmcp/.env の X_BEARER_TOKEN と同じ値
./cmd/sifter/sifter sync following <username>
./cmd/sifter/sifter following list <username> --filter "キーワード"
```

### xmcp MCP サーバーの利用

`make server` 後に Claude Code を再起動すると `.mcp.json` 経由で X API ツールが使える。

## コスト注意事項

X API は PPU（Pay-Per-Use）プラン。読み取り $0.005/件。

- **sifter CLI を優先** — ローカルキャッシュがあるため2回目以降は API コール不要
- **MCP 経由は対話的な操作に限定** — 毎回 API コールが発生する
- **大量フォローユーザーの分析は高額** — 1000人 = $5.00/回

## ディレクトリ構成

```
cmd/sifter/         Go CLI エントリポイント
internal/cli/       CLI コマンド実装
internal/xapi/      X API v2 クライアント（net/http 直接）
internal/store/     SQLite 永続化層
internal/domain/    ビジネスロジック（差分検出等）
xmcp/               X API MCP サーバー（サブモジュール）
docs/               ドキュメント
```

## ドキュメント

- [xmcp MCP サーバー](docs/xmcp.md) — セットアップ、認証、トラブルシューティング
- [X API リファレンス](docs/x-api.md) — エンドポイント、料金、ツール一覧
- [sifter CLI リファレンス](docs/sifter-cli.md) — コマンド、スキーマ、コスト最適化
