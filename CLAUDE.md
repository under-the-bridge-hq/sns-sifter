# CLAUDE.md

## プロジェクト概要

sns-sifter は X (旧Twitter) 上の情報をふるい分け・分析する CLI ツール。
特定ユーザーのフォロー × キーワードのクロス分析を通じて、有用な情報・ナレッジベースを構築する。

## アーキテクチャ

sifter CLI (Go + SQLite) が X API を直接呼び出し、収集・蓄積・フィルタ・分類を行う。

### 設計方針

- X API 呼び出しは sifter の Go クライアントに集約する。
- `sifter auth login` で OAuth 1.0a の認可を一度行い、資格情報を `~/.sifter/credentials.json` に保存する（権限 0600）。
- 資格情報がある場合は全 API 呼び出しに OAuth 1.0a User Context を使う。ない場合は `SIFTER_BEARER_TOKEN` を使う。`sync likes` には保存済み資格情報が必要。
- cron の設定例は [docs/operations/cron-setup.md](docs/operations/cron-setup.md) を参照。

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
```

### sifter CLI の実行

```bash
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
./cmd/sifter/sifter auth login
./cmd/sifter/sifter auth status

# 保存済み OAuth 資格情報がない場合の公開データ取得用
# export SIFTER_BEARER_TOKEN=<bearer-token>

# フォロー一覧
./cmd/sifter/sifter sync following <username>
./cmd/sifter/sifter following list <username> --filter "キーワード"

# いいね投稿 (保存済み OAuth 1.0a 資格情報が必要)
./cmd/sifter/sifter sync likes <username>            # incremental
./cmd/sifter/sifter sync likes <username> --full     # full re-sync
./cmd/sifter/sifter likes categorize <username>      # キーワード分類
./cmd/sifter/sifter likes list <username> --category ai
./cmd/sifter/sifter likes review <username> <tweet_id>...  # 1on1 後マーク
./cmd/sifter/sifter likes stats <username>
```

### ai-commander での OAuth 1.0a ログイン

X Developer Console に `http://127.0.0.1:8976/oauth/callback` をコールバック URL として登録する。ai-commander 上でログインする場合は、Mac のブラウザからコールバックへ到達できるように SSH ポートフォワードを使う。

```bash
# Mac のターミナルから: コールバックポート(8976)をフォワード
ssh -L 8976:127.0.0.1:8976 ai-commander

# ai-commander 内: 認証情報を設定してログイン
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
./cmd/sifter/sifter auth login
```

stdout に表示された認可 URL を Mac のブラウザで開く。認可後のコールバックは SSH トンネル経由で ai-commander の sifter に届く。保存内容は `sifter auth status` で確認する。

## コスト注意事項

X API は従量課金。Post Read は $0.005/件、User Read と Following Read は各 $0.010/件。認証ユーザーと対象 ID が一致し、そのユーザーが開発アプリの所有者なら、自分の `liked_tweets` は Owned Reads の $0.001/件が適用される（[公式料金表](https://docs.x.com/x-api/getting-started/pricing)）。

- **sifter CLI のキャッシュを優先** — フォロー一覧等は2回目以降 DB から
- **likes sync は incremental + max=5** — 5件返れば Owned Reads は $0.005、通常の Post Read は $0.025。実際の課金は返却件数と適用区分による
- **大量フォローユーザーの分析** — 1000人の通常の Following Read は $10.00。Owned Reads の条件を満たす自分の following なら $1.00
- **日次 sync の試算** — [docs/operations/cron-setup.md](docs/operations/cron-setup.md) を参照

## ディレクトリ構成

```
cmd/sifter/         Go CLI エントリポイント
internal/cli/       CLI コマンド実装 (sync, following, likes, history)
internal/xapi/      X API v2 クライアント（OAuth 1.0a / Bearer Token）
internal/store/     SQLite 永続化層 (users, following, liked_posts, knowledge_articles)
internal/domain/    ビジネスロジック (diff, classifier)
docs/operations/    運用・設計ドキュメント
docs/knowledge/     1on1 review で育てる知識ベース
```

## ドキュメント

運用・設計 (`docs/operations/`):
- [設計方針・経緯](docs/operations/design-decisions.md) — アーキテクチャ判断、コスト最適化、今後の拡張予定
- [sifter CLI リファレンス](docs/operations/sifter-cli.md) — コマンド、スキーマ、コスト最適化
- [X API リファレンス](docs/operations/x-api.md) — エンドポイント、認証、料金
- [cron セットアップ](docs/operations/cron-setup.md) — ai-commander での定期 sync 手順 (Ansible 対応)

知識ベース (`docs/knowledge/` — 1on1 review で育てる):
- [メモ・気付き](docs/knowledge/notes.md) — 流入の窓口、一過性メモ、蓄積待ちトラッキング
- [Claude Code 新機能ノート](docs/knowledge/claude-code-features.md) — 新機能の継続蓄積記事
- [Managed Agents](docs/knowledge/managed-agents.md) — Managed Agents 入門/実践の継続記事
- [Managed Agents 評価](docs/knowledge/managed-agents-evaluation.md) — 非エンジニア向け業務自動化の適合性調査
