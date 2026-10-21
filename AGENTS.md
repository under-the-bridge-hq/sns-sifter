# AGENTS.md

このファイルは sns-sifter を扱うすべてのエージェント（Claude Code、Codex、Grok Bot、Cursor Cloud Agent 等）向けの共通指示。ツール固有の設定は各ツールのディレクトリ（`.claude/`、`.codex/`）に置き、ここにはツールに依存しない内容だけを書く。

## プロジェクト概要

sns-sifter は X 上の情報をふるい分け、AI・Coding Agent・組織での AI 利用に関するナレッジベース（`docs/knowledge/`）を育てるリポジトリ。

2026 年 10 月から、情報収集と一次整理をクラウドエージェント（Grok Bot + Cursor Cloud Agent）に任せる構成へ移行中。X API を直接呼ぶ sifter CLI は補助ルートとして残す。設計は [docs/design/cloud-agent-workflow.md](docs/design/cloud-agent-workflow.md) を参照。

## 2 つの収集ルート

| ルート | 役割 | 実行場所 | 状態 |
|---|---|---|---|
| クラウドエージェント | Grok Bot が X からテーマ別にキュレートし、Cursor Cloud Agent が repo へ PR を作る。蒸留・整理も Cloud Agent の定期実行で行う | xAI / Cursor のクラウド | PoC 準備中 |
| sifter CLI | X API（OAuth 1.0a）でいいね投稿・フォロー一覧を SQLite に同期し、1on1 review で人手で蒸留する | kaz の Mac（M5Max） | 稼働可能。補助ルート |

sifter CLI は kaz の Mac で実行する。資格情報と DB は Mac 側に置く。手順は [docs/sifter-cli/README.md](docs/sifter-cli/README.md)。

## キュレーションのテーマ

- AI 全般
- Coding Agent（Claude Code、Codex、Cursor、Grok Build 等のツールと使い方）
- 組織での AI 利用

情報源の優先順位は、kaz_utb の X タイムライン、次にフォロー中アカウントの公開投稿、最後にテーマ検索。タイムラインを直接読めるかは PoC で確認する（[設計 doc の未確認事項](docs/design/cloud-agent-workflow.md#未確認事項と-poc-で確かめること)）。

## ナレッジベースの運用

```
[X] ──(Grok Bot / sifter CLI)──→ 一次情報
                                   ↓ 振り分け
          即蒸留 → notes.md    継続蓄積 → docs/knowledge/<topic>.md    蓄積待ち → 手を付けない
```

- **即蒸留**: 単発で意味が完結するもの。`docs/knowledge/notes.md` に日付見出しで追記
- **継続蓄積**: 同じテーマで流入が続くもの。`docs/knowledge/<topic>.md` または `docs/knowledge/<topic>/` に追記
- **蓄積待ち**: 類似サンプルが揃うまで整理しないもの。既存記事に触らず、notes.md の末尾に候補として残す
- 出典（投稿 URL、著者ハンドル、日付）を必ず残す。要約と引用を区別する
- 既存記事を書き換えるときは追記を優先し、削除は PR で理由を書く

## エージェント共通ルール

- 変更は feature ブランチ + PR。main への直接 push はしない
- PR は squash merge。merge は kaz の指示を待つ
- PR タイトルに issue 番号を入れない。issue 参照は本文に `closes #NN`
- 複数セッションが並行するときは `.worktrees/<name>` で作業する
- 日本語で記述する。commit message / PR description も日本語
- 事実と推定を分ける。外部サービスの仕様・料金は公式 URL を添える
- ナレッジ記事に投稿本文を転載するときは引用範囲を最小にし、出典 URL を付ける

## 開発ガイド（sifter CLI）

```bash
make build                                   # cmd/sifter/sifter をビルド
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
./cmd/sifter/sifter auth login               # 初回のみ。~/.sifter/credentials.json に保存
./cmd/sifter/sifter sync likes kaz_utb --max 5
./cmd/sifter/sifter likes categorize kaz_utb
./cmd/sifter/sifter likes list kaz_utb --category ai --unreviewed
```

X API は従量課金。単価と試算は [docs/sifter-cli/x-api.md](docs/sifter-cli/x-api.md) を参照し、`sync likes` は incremental + `--max 5` を既定とする。

## ディレクトリ構成

```
AGENTS.md           エージェント共通指示（このファイル）
README.md           利用者向け概要
cmd/sifter/         Go CLI エントリポイント
internal/cli/       CLI コマンド実装 (auth, sync, following, likes, history)
internal/xapi/      X API v2 クライアント（OAuth 1.0a / Bearer Token）
internal/auth/      OAuth 1.0a 資格情報の保存・読み込み
internal/store/     SQLite 永続化層
internal/domain/    diff, classifier
docs/design/        設計方針、クラウドエージェント構成、引き継ぎ
docs/sifter-cli/    sifter CLI リファレンス、X API、cron
docs/knowledge/     ナレッジベース
.claude/skills/     Claude Code 用 skill（1on1-review, sync-likes）
```

## ドキュメント索引

設計（`docs/design/`）:
- [クラウドエージェント構成と PoC 計画](docs/design/cloud-agent-workflow.md)
- [設計方針・経緯](docs/design/decisions.md)
- [引き継ぎ 2026-10](docs/design/handoff-2026-10.md) — ai-commander から Mac セッションへ

sifter CLI（`docs/sifter-cli/`）:
- [CLI リファレンス](docs/sifter-cli/README.md)
- [X API・認証・料金](docs/sifter-cli/x-api.md)
- [cron セットアップ](docs/sifter-cli/cron-setup.md)

ナレッジベース（`docs/knowledge/`）:
- [メモ・気付き](docs/knowledge/notes.md) — 流入の窓口
- [Agent Harness](docs/knowledge/agent-harness/README.md)
- [Claude Code 新機能ノート](docs/knowledge/claude-code-features.md)
- [Managed Agents](docs/knowledge/managed-agents.md) / [評価](docs/knowledge/managed-agents-evaluation.md)
