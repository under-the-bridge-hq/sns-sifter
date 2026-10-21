# cron セットアップ手順 (ai-commander)

ai-commander 上で `sync likes` を日次実行する設定例。sifter が X API を直接呼び出すため、別の常駐プロセスは必要ない。

## 事前準備

1. `make build` で sifter CLI をビルドする。
2. `sifter auth login` で OAuth 1.0a 資格情報を保存する。ai-commander で Mac のブラウザを使う手順は [sifter CLI リファレンス](sifter-cli.md) を参照。
3. `sifter auth status` で保存先・screen_name・user_id を確認する。
4. 手動で `sifter sync likes kaz_utb --max 5 --max-pages 20` を実行して動作を確認する。

既定の資格情報保存先は `~/.sifter/credentials.json`（権限 0600）。別の場所を使う場合は、ログイン時と cron 実行時の両方に `SIFTER_CREDENTIALS` を指定する。DB は `SIFTER_DB` で指定でき、既定は `~/.sifter/data.db`。

## crontab 例

毎朝 6:00 に `kaz_utb` のいいねを incremental sync し、成功時に分類する例:

```crontab
0 6 * * * cd /home/kaz/git/github.com/under-the-bridge-hq/sns-sifter && SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter sync likes kaz_utb --max 5 --max-pages 20 >> /home/kaz/.sifter/cron.log 2>&1 && SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter likes categorize kaz_utb >> /home/kaz/.sifter/cron.log 2>&1
```

両コマンドに同じ DB パスを渡す。`--full` を指定しなければ incremental sync。`--max-pages 20` は最大 20 ページ、`--max 5` は 1 ページ最大 5 件であり、合計の取得上限は 100 件。cron は同じ OS ユーザーで実行し、そのユーザーが保存済み資格情報を読めるようにする。

## Ansible での管理例

```yaml
- name: sns-sifter sync likes
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
      && SIFTER_DB={{ sns_sifter_db }}
      {{ sns_sifter_repo }}/cmd/sifter/sifter likes categorize {{ sns_sifter_target_user }}
      >> {{ sns_sifter_log_dir }}/cron.log 2>&1
```

ログディレクトリの作成と logrotate は運用環境側で設定する。

## 読み取り料金の計算例

[X 公式料金表](https://docs.x.com/x-api/getting-started/pricing) の単価で試算する。`GET /2/users/{id}/liked_tweets` が Owned Reads になる条件は、`{id}` が認証ユーザーと一致し、そのユーザーが開発アプリの所有者であること。Owned Reads は $0.001/件、通常の Post Read は $0.005/件。表は毎日同じ件数が返る場合の単純計算で、24 時間 UTC 単位の重複課金排除などは含めない。

| 1日の取得件数 | Owned Reads / 日 | Owned Reads / 30日 | 通常 Post Read / 日 | 通常 Post Read / 30日 |
|---:|---:|---:|---:|---:|
| 5件 | $0.005 | $0.15 | $0.025 | $0.75 |
| 15件 | $0.015 | $0.45 | $0.075 | $2.25 |
| 25件 | $0.025 | $0.75 | $0.125 | $3.75 |
| 55件 | $0.055 | $1.65 | $0.275 | $8.25 |
| 100件 | $0.100 | $3.00 | $0.500 | $15.00 |

実際の返却件数と適用料金は Developer Console で確認する。

## 動作確認

```bash
SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter sync likes kaz_utb --max 5 --max-pages 20
SIFTER_DB=/home/kaz/.sifter/data.db ./cmd/sifter/sifter likes stats kaz_utb
tail -f /home/kaz/.sifter/cron.log
```
