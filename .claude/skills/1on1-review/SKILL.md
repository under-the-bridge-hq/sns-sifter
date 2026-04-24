---
name: 1on1-review
description: いいね投稿の 1on1 レビューセッション。未レビューの投稿を表示し、対話的に蒸留先を振り分けてドキュメントに追記する。
disable-model-invocation: true
allowed-tools: Bash(./cmd/sifter/sifter *) Read Edit Write
argument-hint: "[category: ai|work|other]"
---

1on1 review セッションを開始する。

## 対象

カテゴリ指定があれば `$ARGUMENTS`、なければ `ai` をデフォルトとする。

## 手順

1. 未レビュー投稿を表示:
   `./cmd/sifter/sifter likes list kaz_utb --category $ARGUMENTS --unreviewed`
   (引数なしなら `--category ai --unreviewed`)

2. ユーザーと対話しながら各投稿を振り分ける:
   - **即蒸留 (陳腐化早い単発情報)** → `docs/knowledge/notes.md` に追記 + reviewed マーク
   - **継続蓄積記事** → 該当する `docs/knowledge/<topic>.md` に追記 + reviewed マーク
     - Claude Code 新機能 → `docs/knowledge/claude-code-features.md`
     - Managed Agents → `docs/knowledge/managed-agents.md`
     - 新テーマなら新規ファイル作成を提案
   - **蓄積待ち (深掘り用にサンプル蓄積)** → reviewed マークしない、次回再登場

3. 投稿の詳細が必要な場合:
   `./cmd/sifter/sifter likes show kaz_utb <tweet_id>`

4. 振り分け決定後に reviewed マーク:
   `./cmd/sifter/sifter likes review kaz_utb <tweet_id>...`

5. セッション終了時に統計を表示:
   `./cmd/sifter/sifter likes stats kaz_utb`

## 振り分けの原則

- 単発で意味が完結し深い議論不要 → 即蒸留
- 同テーマで継続的に情報流入あり → 継続蓄積記事
- 類似サンプルを増やしてから整理したい → 蓄積待ち

## 蓄積待ち中のテーマ (参考)

- **Agent Harness 論** — "harness" 共通概念 / Evals / prompting as skill
- **組織への AI 導入** — DeNA / Timee / LayerX 対談 / kaz_utb 登壇
