# CLAUDE.md

## プロジェクト概要

sns-sifter は X (旧Twitter) 上の情報をふるい分け・分析するツール群。
特定ユーザーのフォロー × キーワードのクロス分析を通じて、有用な情報・ナレッジベースを構築する。

## アーキテクチャ

2つのコンポーネントで構成。役割を明確に分離する:

- **sifter CLI** (Go + SQLite) — 収集・蓄積・CLI レイヤー。バッチ収集、SQLite 永続化、フィルタ・分類、CLI を提供
- **xmcp MCP サーバー** (Python) — X API ゲートウェイ。OAuth 認証 (1.0a User Context など) を集約。Claude Code から MCP として、sifter CLI から HTTP (JSON-RPC) として利用される

### 設計方針

- **xmcp は X API ゲートウェイ** — OAuth 認証を集約。Claude Code から MCP として、sifter から HTTP として利用される
- **sifter は収集・蓄積・CLI レイヤー** — SQLite 永続化、バッチ収集、フィルタ・分類、CLI を提供
- **認証の使い分け**:
  - Bearer Token (App-Only): public timeline 等の読み取り。`sifter sync following` で直接使用
  - OAuth 1.0a User Context: liked_tweets 等 user context 必須の API。**xmcp 経由**で叩く
  - OAuth 2.0 PKCE: bookmarks 等。将来 xmcp に追加
- sifter は user context が必要な API を呼ぶ場合、`http://127.0.0.1:8000/mcp` (環境変数 `XMCP_URL`) の xmcp に JSON-RPC でアクセスする
- cron 運用 (現在は手動 sync 中。cron は様子見): ai-commander で xmcp を zellij 常駐 + cron で sifter 実行する想定 ([docs/operations/cron-setup.md](docs/operations/cron-setup.md))

### ナレッジベース構築フロー (1on1 review ワークフロー)

sns-sifter の主用途は、kaz_utb の X いいね投稿を一次情報として AI/組織論のナレッジを蓄積すること。以下の二層構造で運用する:

```
[X likes] ──→ sifter sync likes ──→ [liked_posts (SQLite)]
                                        ↓ likes categorize (キーワード分類)
                                        ↓ ai / work / other
                                        ↓
                                  [1on1 review セッション]
                                        ↓ ↘
                       即蒸留 → notes.md       蓄積待ち → そのまま残す (reviewed しない)
                       (陳腐化早いもの)             ↓ サンプル N 件溜まったら
                                              テーマ別 1on1 で深掘り
                                                    ↓
                                            独立ドキュメント
                                            (claude-code-features.md, managed-agents.md, etc.)
```

**振り分けの原則**:
- **即蒸留** (notes.md に追記 + reviewed マーク): Claude Code 新機能、Managed Agents、AI 駆動開発 Tips、周辺ツール — 単発で意味が完結し、深掘り議論を必要としないもの
- **継続蓄積記事** (`docs/<topic>.md` に追記 + reviewed マーク): 同じテーマで情報流入が続くもの (Claude Code 新機能、Managed Agents)
- **蓄積待ち** (reviewed しない、次回再登場): 類似サンプルを増やしてから整理したい深掘りテーマ
  - **Agent Harness 論** (現在 5 件) — 共通概念 "harness" / Evals / prompting as skill
  - **組織への AI 導入** (現在 11 件) — DeNA / Timee / LayerX 対談 / kaz_utb 自身の登壇

**ドキュメント役割の使い分け**:
- `notes.md` — 流入の窓口 + 陳腐化早いものの一過性メモ + 蓄積待ちトラッキング
- `docs/<topic>.md` — 継続蓄積される独立記事 (例: claude-code-features, managed-agents)
- `docs/<topic>-evaluation.md` — 特定の問いへの静的な調査回答 (例: managed-agents-evaluation)

**1on1 review コマンド**:
```bash
./cmd/sifter/sifter likes list kaz_utb --category ai --unreviewed
./cmd/sifter/sifter likes show kaz_utb <tweet_id>
./cmd/sifter/sifter likes review kaz_utb <tweet_id>...   # レビュー後マーク
```

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
export XMCP_URL=http://127.0.0.1:8000/mcp  # likes 系で使用 (デフォルト値)

# フォロー一覧 (Bearer Token で直接 X API)
./cmd/sifter/sifter sync following <username>
./cmd/sifter/sifter following list <username> --filter "キーワード"

# いいね投稿 (xmcp 経由 / OAuth 1.0a)
./cmd/sifter/sifter sync likes <username>            # incremental
./cmd/sifter/sifter sync likes <username> --full     # full re-sync
./cmd/sifter/sifter likes categorize <username>      # キーワード分類
./cmd/sifter/sifter likes list <username> --category ai
./cmd/sifter/sifter likes review <username> <tweet_id>...  # 1on1 後マーク
./cmd/sifter/sifter likes stats <username>
```

### xmcp MCP サーバーの利用

`make server` 後に Claude Code を再起動すると `.mcp.json` 経由で X API ツールが使える。

#### リモート（ai-commander）で起動する場合

OAuth1 認証のリダイレクトを Mac のブラウザで完結させるため、SSH ローカルポートフォワードを使う。X11 forwarding は不要。

```bash
# Mac のターミナルから: コールバックポート(8976)をフォワード
ssh -L 8976:127.0.0.1:8976 ai-commander

# ai-commander 内: zellij/tmux で分離してから起動（SSH 切断対策）
zellij attach xmcp || zellij -s xmcp
cd ~/git/github.com/under-the-bridge-hq/sns-sifter
make server
```

`make server` の stdout に OAuth 認可 URL が表示されるので、Mac のブラウザにコピペして開く。X 側で許可するとリダイレクト先 `http://127.0.0.1:8976/oauth/callback` が SSH トンネル経由で ai-commander の xmcp に戻り、OAuth が完走する。

完走後、`Ctrl+p d` で zellij をデタッチすれば SSH を切ってもサーバは生存する。Claude Code を再起動すると `.mcp.json` 経由で X API ツールが利用可能になる。

`XMCP_AUTO_OPEN_BROWSER=1` を設定するとローカルブラウザの自動起動を試みる（GUI 環境向け）。

## コスト注意事項

X API は PPU（Pay-Per-Use）プラン。読み取り $0.005/件 (= レスポンスに含まれた tweet 数で課金)。

- **sifter CLI のキャッシュを優先** — フォロー一覧等は2回目以降 DB から
- **likes sync は incremental + max=5 推奨** — `--max 5` で 1 sync 最低 $0.025、新規 N 件で `(N+5) × $0.005`
- **大量フォローユーザーの分析は高額** — 1000人 = $5.00/回
- **試算例 (1日1回 sync, 新規10件/日想定)**: 月 $2.25, 年 $27 — [docs/operations/cron-setup.md](docs/operations/cron-setup.md) 参照

## ディレクトリ構成

```
cmd/sifter/         Go CLI エントリポイント
internal/cli/       CLI コマンド実装 (sync, following, likes, history)
internal/xapi/      X API v2 クライアント（net/http 直接, Bearer Token 用）
internal/mcpclient/ xmcp MCP HTTP クライアント (JSON-RPC + SSE)
internal/store/     SQLite 永続化層 (users, following, liked_posts, knowledge_articles)
internal/domain/    ビジネスロジック (diff, classifier)
xmcp/               X API MCP サーバー（サブモジュール）
docs/operations/    運用・設計ドキュメント
docs/knowledge/     1on1 review で育てる知識ベース
```

## ドキュメント

運用・設計 (`docs/operations/`):
- [設計方針・経緯](docs/operations/design-decisions.md) — アーキテクチャ判断、コスト最適化、今後の拡張予定
- [sifter CLI リファレンス](docs/operations/sifter-cli.md) — コマンド、スキーマ、コスト最適化
- [X API リファレンス](docs/operations/x-api.md) — エンドポイント、料金、ツール一覧
- [xmcp MCP サーバー](docs/operations/xmcp.md) — セットアップ、認証、トラブルシューティング
- [cron セットアップ](docs/operations/cron-setup.md) — ai-commander での定期 sync 手順 (Ansible 対応)

知識ベース (`docs/knowledge/` — 1on1 review で育てる):
- [メモ・気付き](docs/knowledge/notes.md) — 流入の窓口、一過性メモ、蓄積待ちトラッキング
- [Claude Code 新機能ノート](docs/knowledge/claude-code-features.md) — 新機能の継続蓄積記事
- [Managed Agents](docs/knowledge/managed-agents.md) — Managed Agents 入門/実践の継続記事
- [Managed Agents 評価](docs/knowledge/managed-agents-evaluation.md) — 非エンジニア向け業務自動化の適合性調査
