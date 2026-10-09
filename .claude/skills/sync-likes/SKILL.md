---
name: sync-likes
description: X のいいね投稿をローカル DB に同期し、カテゴリ分類して統計を表示する。日次の定期実行向け。
disable-model-invocation: true
allowed-tools: Bash(./cmd/sifter/sifter *) Bash(make build)
---

いいね投稿のローカル DB 同期を実行する。

## 手順

1. ビルド確認: `make build`
2. `./cmd/sifter/sifter auth status` で保存済み資格情報を確認する
3. incremental sync: `./cmd/sifter/sifter sync likes kaz_utb`
4. 未分類を分類: `./cmd/sifter/sifter likes categorize kaz_utb`
5. 統計表示: `./cmd/sifter/sifter likes stats kaz_utb`

## 前提

- `sifter auth login` による OAuth 1.0a 資格情報が保存済み
- 資格情報の既定の保存先は `~/.sifter/credentials.json`。変更した場合は `SIFTER_CREDENTIALS` を設定する

## エラー時

- 資格情報がない場合は `sifter auth login` の手順を案内する
- sync 失敗時はエラーメッセージをそのまま表示する
