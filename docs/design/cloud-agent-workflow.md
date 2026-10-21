# クラウドエージェント構成と PoC 計画

2026-10-21 時点の設計。X からのキュレーションとナレッジベースの蒸留を、kaz のローカル PC から切り離したクラウドエージェントに任せる。

## 目的と要件

kaz が 2026-10-21 に示した要件は次の 3 つ。

1. X からテーマに沿ってキュレートできること
2. スマホの Grok Bot アプリなど複数のインターフェースから依頼・確認できること
3. ローカル PC と分離したクラウドセッションで稼働すること

加えて、repo に溜めた内容を別のクラウドセッションで蒸留し、必要に応じて別リポジトリでシステム開発（例: local RAG の試作）に展開できることを望んでいる。

## 採用する構成

契約は The X Subscription の Premium（$30/月）。Cursor の料金ページがこのプランを現行プランとして掲載しており、Cursor と Grok Bot と Grok を 1 契約で使う。

| 役割 | 担当 | 根拠 |
|---|---|---|
| X からのキュレート | Grok Bot（routine で定期実行、X プラグイン） | [Grok Bot overview](https://docs.x.ai/grok-bot/overview) |
| repo への書き込み（PR 作成） | Cursor Cloud Agent。Grok Bot から起動する | [Grok Bot for engineering](https://x.ai/bot/guides/grok-bot-for-engineering)、[Cloud Agent](https://cursor.com/docs/cloud-agent) |
| 蒸留・整理の定期実行 | Cursor Automations（cron 式） | [Automations](https://cursor.com/docs/cloud-agent/automations) |
| 別リポジトリでの開発 | Cursor Cloud Agent（multi-repo、repo ごとに PR） | [Cloud Agent](https://cursor.com/docs/cloud-agent) |
| スマホからの操作 | Grok Bot アプリ、Cursor iOS アプリ、X 上の `@bot` タグ | 上記 overview / Cloud Agent |

Grok Build はローカル CLI で、クラウド実行モードとスケジュール機能の記載がないため、この構成では使わない（[Grok Build overview](https://docs.x.ai/build/overview)）。

## 公式ページで確認済みの事実（2026-10-21）

- Premium 以上に Grok Bot、Cloud Agents、Automations、Grok Build が含まれる。Plus（$8）には Grok Bot、Cloud Agents、Automations が含まれない（[Cursor pricing](https://cursor.com/docs/account/pricing)）
- 有料プランの使用量は Grok / Grok Bot / Cursor で 1 つの週次プールを共有する（同上、[docs.x.ai](https://docs.x.ai/x-subscription/grok-bot)）
- Grok Bot は常駐クラウド PC（ブラウザ / ファイルシステム / ターミナル）を持ち、routine は Bot あたり 50 本、最短 5 分間隔（[skills, routines and automations](https://docs.x.ai/grok-bot/skills-routines-and-automations)）
- Grok Bot の X プラグインは公開投稿検索なら X へのサインイン不要。DM・ブックマーク・いいねはサインインが必要（[computer and apps](https://docs.x.ai/grok-bot/computer-and-apps)）
- Grok Bot は Cursor Cloud Agent を作成し、transcript を読み、追加指示を出せる（[Grok Bot for engineering](https://x.ai/bot/guides/grok-bot-for-engineering)）
- Cursor Cloud Agent は専用 VM でブラウザ操作と MCP が使え、repo を clone して別ブランチに push し PR を開く。cron 式の Automations と GitHub / Slack / Webhook トリガーがある（[Cloud Agent](https://cursor.com/docs/cloud-agent)、[Automations](https://cursor.com/docs/cloud-agent/automations)）
- 全 Bot が 1 台のクラウド PC と認証情報を共有し、Bot 同士は別のセキュリティ境界ではない（docs.x.ai の記載）

## 未確認事項と PoC で確かめること

| 項目 | 状態 | 確認方法 |
|---|---|---|
| Grok Bot が kaz_utb のホームタイムラインを読めるか | 未確認。公式はいいね・DM・ブックマークのサインイン要件を書いているが、タイムラインの記載はない | X にサインインした Bot に「直近 24 時間のタイムラインから AI 関連を 10 件」と依頼する |
| Grok Bot がいいね一覧を読めるか | サインイン要で可能と記載。件数上限と追加課金は未確認 | 同上をいいねに対して実行し、週次プールの消費を確認する |
| Grok Bot の GitHub コネクタ単体で PR を作れるか | 未確認。読めなければ Cloud Agent 委譲に統一する | Bot に小さな docs 修正の PR 作成を依頼する |
| 週次プールの実消費量 | 未確認 | routine を 1 日 1 回で 1 週間回し、使用量画面を記録する |
| X 読み取りに PPU 相当の従量課金が別途かかるか | 未確認 | 1 週間の請求明細を確認する |
| Cloud Agent が X のページを直接読めるか | 未確認 | Cloud Agent に公開投稿 URL の要約を依頼する |

## PoC 手順

1. The X Subscription Premium を契約し、Grok Bot と Cursor の両方にサインインする。Cursor 側で GitHub の `under-the-bridge-hq/sns-sifter` への read-write を許可する
2. Grok Bot を 1 体作り、X プラグインと GitHub コネクタを接続する
3. タイムライン読み取りの可否を確認する（上表 1 行目）。読めない場合は、`sifter following list kaz_utb` のキャッシュから作ったフォロー中アカウント一覧を Bot に渡し、公開投稿検索で代替する
4. routine を 1 本作る。「1 日 1 回、テーマ（AI / Coding Agent / 組織での AI 利用）に該当する投稿を最大 N 件選び、出典 URL と 1 行要約を `docs/knowledge/inbox/YYYY-MM-DD.md` にまとめて PR を作る」。PR 作成は Cloud Agent 委譲で行う
5. Cursor Automations を 1 本作る。「週 1 回、`docs/knowledge/inbox/` の未処理ファイルを読み、AGENTS.md の振り分け原則に従って notes.md と各記事へ蒸留し、処理済みファイルを `docs/knowledge/inbox/processed/` へ移動して PR を作る」
6. 1 週間運用し、PR の品質、週次プールの消費、課金を記録して本設計 doc を更新する

N の値と `inbox/` のファイル命名は PoC 開始時に kaz が決める。

## 別リポジトリでの開発への展開

ナレッジベースから出た試作テーマ（例: local RAG）は、別リポジトリを作り Cursor Cloud Agent に実装させる。Grok Bot から Cloud Agent を起動して進捗をスマホで追う経路は公式ガイドに記載がある。sns-sifter 側には試作テーマへのリンクと結論だけを残す。

## sifter CLI ルートの位置づけ

X API を直接呼ぶ sifter CLI は、PoC が失敗したときのフォールバックと、いいね投稿を確実に取得したいときの手段として残す。実行は kaz の Mac で行い、資格情報と SQLite は Mac 側に置く。手順は [docs/sifter-cli/README.md](../sifter-cli/README.md)。

## 検討した代替案

- **ChatGPT dots + Codex cloud**: dots は Pro（$100 以上）/ Business Premium / Enterprise 向け（[learn.chatgpt.com/docs/dots](https://learn.chatgpt.com/docs/dots)）。X の読み取り経路が公式に確認できず、クラウドブラウザがブロックされる報告もあるため採用しない
- **Grok Build**: ローカル CLI のためローカル PC 分離の要件に合わない
- **X API 自前実装のみ**: いいね取得は確実だが、スマホからの操作と PC 分離を満たさない。補助ルートとして残す
