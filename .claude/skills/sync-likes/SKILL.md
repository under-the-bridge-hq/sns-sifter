---
name: sync-likes
description: X のいいね投稿をローカル DB に同期し、カテゴリ分類して統計を表示する。日次の定期実行向け。
disable-model-invocation: true
allowed-tools: Bash(./cmd/sifter/sifter *) Bash(make build)
---

いいね投稿のローカル DB 同期を実行する。

## 手順

1. ビルド確認: `make build`
2. incremental sync: `./cmd/sifter/sifter sync likes kaz_utb`
3. 未分類を分類: `./cmd/sifter/sifter likes categorize kaz_utb`
4. 統計表示: `./cmd/sifter/sifter likes stats kaz_utb`

## 前提

- 環境変数 `SIFTER_BEARER_TOKEN` と `XMCP_URL` がセット済み
- xmcp が起動済み (ai-commander の zellij セッション 'xmcp')

## エラー時

- xmcp 未起動の場合: `http://127.0.0.1:8000/mcp` への接続失敗を報告し、起動手順を案内する
- sync 失敗時: エラーメッセージをそのまま表示する
