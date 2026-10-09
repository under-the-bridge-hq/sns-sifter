# sns-sifter

X 上のフォロー一覧といいね投稿を収集し、SQLite に保存・分析する Go CLI ツール。

## クイックスタート

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

コールバック URL は X Developer Console に `http://127.0.0.1:8976/oauth/callback` を登録する。ai-commander 上で Mac のブラウザを使う場合は、Mac から `ssh -L 8976:127.0.0.1:8976 ai-commander` で接続してからログインする。

## ドキュメント

- [sifter CLI リファレンス](docs/operations/sifter-cli.md)
- [X API・認証・料金](docs/operations/x-api.md)
- [cron セットアップ](docs/operations/cron-setup.md)
- [設計方針・経緯](docs/operations/design-decisions.md)
- [ナレッジベース](docs/knowledge/notes.md)
