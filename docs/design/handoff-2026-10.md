# 引き継ぎ 2026-10（ai-commander から Mac セッションへ）

2026-10-21 時点で、sns-sifter の作業セッションを Linux（ai-commander）から kaz の Mac（M5Max）へ移す。以降の sifter CLI 実行と新しい設計作業は Mac で行う。

## 2026-10 に行った変更

| PR | 内容 |
|---|---|
| [#9](https://github.com/under-the-bridge-hq/sns-sifter/pull/9) | いいね同期を xmcp 経由から Go の OAuth 1.0a 直接呼び出しへ移行。`sifter auth login/status` 追加。xmcp submodule、`.mcp.json`、`internal/mcpclient/` を削除 |
| 本 PR | `CLAUDE.md` を廃止して `AGENTS.md` に集約。docs を `design/` と `sifter-cli/` に再構成。クラウドエージェント構成の設計 doc と本引き継ぎを追加 |

## ai-commander に残っているもの

git 管理外のファイル。Mac で sifter CLI を使うときに必要になれば scp で取得する。

| パス | 内容 | 扱い |
|---|---|---|
| `~/.sifter/data.db` | SQLite。liked_posts（category、reviewed_at を含む）、users、following、sync_history。最終更新 2026-04-24 | Mac で incremental sync を続けるなら取得する。無いと `--full` 再取得になり PPU 課金が増える |
| `~/.sifter/xmcp.env.bak` | 旧 xmcp の `.env`。X アプリの Consumer Key / Secret と Bearer Token。権限 0600 | `sifter auth login` に Consumer Key / Secret が必要。X Developer Console から再取得もできる |

取得コマンド（Mac から）:

```bash
mkdir -p ~/.sifter && chmod 700 ~/.sifter
scp ai-commander:~/.sifter/data.db ~/.sifter/data.db
scp ai-commander:~/.sifter/xmcp.env.bak ~/.sifter/xmcp.env.bak
chmod 600 ~/.sifter/xmcp.env.bak
```

取得後、ai-commander 側のファイルを削除するかは kaz が判断する。

## Mac で sifter CLI を使い始める手順

1. repo を clone し `make build`
2. `~/.sifter/xmcp.env.bak` の `X_OAUTH_CONSUMER_KEY` / `X_OAUTH_CONSUMER_SECRET` の値を `SIFTER_OAUTH_CONSUMER_KEY` / `SIFTER_OAUTH_CONSUMER_SECRET` として export する
3. `./cmd/sifter/sifter auth login` を実行し、表示された URL をブラウザで開いて認可する。コールバックは `http://127.0.0.1:8976/oauth/callback`（X Developer Console に登録済みの URL）。Mac 上で実行するため SSH ポートフォワードは不要
4. `./cmd/sifter/sifter auth status` で screen_name と user_id を確認する
5. `./cmd/sifter/sifter sync likes kaz_utb --max 5` で動作確認する

ブラウザ認可以降の手順は 2026-10-21 時点で未実施。request_token の取得までは実 API で成功している。

## 次にやること

1. The X Subscription Premium を契約し、[クラウドエージェント構成と PoC 計画](cloud-agent-workflow.md) の PoC 手順を進める
2. PoC の結果（タイムライン読み取り可否、週次プール消費、課金）を設計 doc に追記する
3. `.claude/skills/` の 2 つの skill を、Codex や Cloud Agent からも使える形にするか判断する。SKILL.md の形式は Claude 固有ではないが、各ツールがどのディレクトリを読むかは未確認
4. `docs/knowledge/notes.md` 末尾の蓄積待ちテーマ（Agent Harness 論、組織への AI 導入）は、新テーマ（AI / Coding Agent / 組織での AI 利用）に置き換える方針。既存記事は削除せず残す

## 環境メモ

- ai-commander では `go` が PATH にない。`/usr/local/go/bin/go` を使う
- ai-commander の `.claude/settings.local.json`（git 管理外）に `enabledMcpjsonServers: ["x-api"]` が残っている。`.mcp.json` は削除済みなので接続エラーになるだけで害はない。Mac 側では設定しない
- agmsg: Claude は `kaz-claude-utbh-sns-sifter`、Codex は `kaz-codex-utbh-sns-sifter`（team utb）
