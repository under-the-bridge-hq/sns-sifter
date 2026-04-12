# cron セットアップ手順 (ai-commander)

sns-sifter の `sync likes` を ai-commander 上で cron 実行するための手順。

## 前提

- ai-commander で xmcp が zellij セッション内で常駐していること (詳細: [CLAUDE.md](../CLAUDE.md))
- sifter CLI がビルド済みであること: `make build`
- xmcp の OAuth1 認証が完了していること

## アーキテクチャ

```
[cron] → sifter sync likes kaz_utb
            ↓ HTTP POST /mcp (JSON-RPC)
         [xmcp (zellij 常駐)]
            ↓ OAuth1 署名済み
         [X API]
            ↓
         SQLite (~/.sifter/data.db)
```

cron は sifter を呼ぶだけ。X API への OAuth 認証は xmcp が一手に引き受ける。

## 必要な環境変数

| 変数 | 値 | 用途 |
|------|-----|------|
| `SIFTER_DB` | `/home/kaz/.sifter/data.db` | SQLite パス |
| `XMCP_URL` | `http://127.0.0.1:8000/mcp` | xmcp HTTP エンドポイント (デフォルト値なので省略可) |
| `SIFTER_BEARER_TOKEN` | (省略可) | likes 系では使わないが、他コマンドと併用するなら設定 |

## crontab エントリ例

毎朝 6:00 に kaz_utb のいいねを incremental sync し、自動分類する例:

```crontab
# sns-sifter: いいね投稿を 1 日 1 回 sync + 分類
0 6 * * * cd /home/kaz/git/github.com/under-the-bridge-hq/sns-sifter && SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter sync likes kaz_utb --max 5 --max-pages 20 >> /home/kaz/.sifter/cron.log 2>&1 && ./cmd/sifter/sifter likes categorize kaz_utb >> /home/kaz/.sifter/cron.log 2>&1
```

ポイント:
- `--full` を付けないこと (incremental が既定)
- `--max 5` がコスト最小 (X API の最小値)。新規がない日でも $0.025 で済む
- `--max-pages 20` で 1 sync 上限 100 件 = $0.50 に頭打ち
- `>> cron.log 2>&1` でログを残す
- 連結時は `&&` で sync 成功時のみ categorize する

## Ansible での管理 (推奨)

ouchi-server の Ansible で cron job を管理する場合のロール例:

### `roles/sns_sifter/tasks/main.yml`

```yaml
---
- name: ensure sifter log dir
  ansible.builtin.file:
    path: "{{ sns_sifter_log_dir }}"
    state: directory
    owner: kaz
    group: kaz
    mode: "0755"

- name: register sifter sync likes cron job
  ansible.builtin.cron:
    name: "sns-sifter sync likes"
    user: kaz
    minute: "0"
    hour: "6"
    job: >-
      cd {{ sns_sifter_repo }} &&
      SIFTER_DB={{ sns_sifter_db }}
      {{ sns_sifter_repo }}/cmd/sifter/sifter sync likes {{ sns_sifter_target_user }}
      --max 5 --max-pages 20
      >> {{ sns_sifter_log_dir }}/cron.log 2>&1
      && {{ sns_sifter_repo }}/cmd/sifter/sifter likes categorize {{ sns_sifter_target_user }}
      >> {{ sns_sifter_log_dir }}/cron.log 2>&1

- name: ensure logrotate for sifter cron log
  ansible.builtin.copy:
    dest: /etc/logrotate.d/sns-sifter
    owner: root
    group: root
    mode: "0644"
    content: |
      {{ sns_sifter_log_dir }}/cron.log {
          weekly
          rotate 4
          compress
          missingok
          notifempty
          copytruncate
      }
```

### `roles/sns_sifter/defaults/main.yml`

```yaml
---
sns_sifter_repo: /home/kaz/git/github.com/under-the-bridge-hq/sns-sifter
sns_sifter_db: /home/kaz/.sifter/data.db
sns_sifter_log_dir: /home/kaz/.sifter
sns_sifter_target_user: kaz_utb
```

### Playbook での組み込み

```yaml
- hosts: ai_commander
  roles:
    - sns_sifter
```

## 注意事項

### xmcp の生存監視

cron は xmcp が常駐していることを前提とする。xmcp がダウンしていると `sifter sync likes` は MCP 接続エラーで失敗する。

対策案:
- **systemd user service 化** — zellij ではなく systemd で xmcp を管理する (将来的な改善)
- **ヘルスチェック cron** — 別 cron で `curl http://127.0.0.1:8000/mcp` を叩いて NG なら通知
- **ログ監視** — `/home/kaz/.sifter/cron.log` を tail/grep して MCP エラーを検出

現状は zellij + 手動再起動で運用し、必要に応じて systemd 化する。

### コスト管理

X API は読み取り $0.005/件。`--max 5` 設定で 1 日 1 回 cron する場合の試算:

| 1日の新規いいね数 | 取得件数 | 1日コスト | 月コスト | 年コスト |
|---|---|---|---|---|
| 0 件 | 5 (1ページ) | $0.025 | $0.75 | $9 |
| 10 件 | 15 (3ページ) | $0.075 | $2.25 | $27 |
| 20 件 | 25 (5ページ) | $0.125 | $3.75 | $46 |
| 50 件 | 55 (11ページ) | $0.275 | $8.25 | $100 |
| 100 件以上 | 100 (上限) | $0.50 | $15 | $182 |

ポーリング頻度を上げる場合は単純倍率で増える (15分毎 = 96倍)。日常的な収集には 1 日 1 回 で十分。

なぜ `--max 5` がベスト:
- X API の最小値 (これ以下は不可)
- 新規がぴったり収まるサイズで止められる
- max=20 だと「9件新規 + 11件既知」の場合に既知 11件分も課金される

## 動作確認

cron 設定後の動作確認:

```bash
# 手動で 1 回実行してログを確認
cd /home/kaz/git/github.com/under-the-bridge-hq/sns-sifter
SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter sync likes kaz_utb

# 統計確認
./cmd/sifter/sifter likes stats kaz_utb

# cron ログ確認
tail -f /home/kaz/.sifter/cron.log
```
