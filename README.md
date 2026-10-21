# sns-sifter

X 上の情報をふるい分けて AI・Coding Agent・組織での AI 利用のナレッジベース（`docs/knowledge/`）を育てるリポジトリ。収集はクラウドエージェント（Grok Bot + Cursor Cloud Agent）へ移行中で、X API を直接呼ぶ Go CLI `sifter` は補助ルートとして残す。エージェント向けの指示は [AGENTS.md](AGENTS.md)。

## クラウドエージェント構成

[docs/design/cloud-agent-workflow.md](docs/design/cloud-agent-workflow.md) を参照。

## sifter CLI クイックスタート（Mac で実行）

```bash
make build
export SIFTER_OAUTH_CONSUMER_KEY=<consumer-key>
export SIFTER_OAUTH_CONSUMER_SECRET=<consumer-secret>
./cmd/sifter/sifter auth login
./cmd/sifter/sifter auth status
./cmd/sifter/sifter sync following kaz_utb
./cmd/sifter/sifter sync likes kaz_utb --max 5
./cmd/sifter/sifter likes categorize kaz_utb
./cmd/sifter/sifter likes list kaz_utb --category ai --unreviewed
```

`auth login` は初回のみ実行する。認可 URL を標準出力に表示し、認可後の資格情報を `~/.sifter/credentials.json`（権限 0600）に保存する。保存済み資格情報がある場合、sifter は全 API 呼び出しに OAuth 1.0a を使用する。資格情報がない場合は `SIFTER_BEARER_TOKEN` で公開データを取得できる。`sync likes` には保存済み資格情報が必要。

コールバック URL は X Developer Console に `http://127.0.0.1:8976/oauth/callback` を登録する。

## ドキュメント

- [クラウドエージェント構成と PoC 計画](docs/design/cloud-agent-workflow.md)
- [設計方針・経緯](docs/design/decisions.md)
- [引き継ぎ 2026-10](docs/design/handoff-2026-10.md)
- [sifter CLI リファレンス](docs/sifter-cli/README.md)
- [X API・認証・料金](docs/sifter-cli/x-api.md)
- [cron セットアップ](docs/sifter-cli/cron-setup.md)
- [ナレッジベース](docs/knowledge/notes.md)
