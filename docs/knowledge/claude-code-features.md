# Claude Code 新機能ノート

Claude Code は週次でリリースが入り、X 上に作者・コントリビューター・コミュニティから情報が流れてくる。
このドキュメントは liked 投稿を一次情報として、機能ごとに **「概要 / ソース / 使い方インサイト」** をまとめる継続記事。

## 更新ポリシー

- `liked_posts` に Claude Code 関連の新情報が入ったら、機能エントリを追加 or 既存エントリにソースを追記
- 1on1 review セッションで「自分はどう使うか / どんな場面で使うか」を機能ごとに育てる
- 「使い方インサイト」が薄い (TODO の) ものは、実際に使ってみてから埋める

---

## skills の "inject dynamic context" パターン

**概要**: skill 内で動的にコンテキストを差し込むパターン。`@theo` は「skills 標準化候補」として他ツール (Codex CLI, Cursor, Pi 等) にも組み込まれるべきと評価。

**ソース**: [@theo, 2026-04-11](https://x.com/i/status/2043103001121034581)

**使い方インサイト**: TODO

---

## /team-onboarding コマンド

**概要**: プロジェクト構造とセッション使用履歴を分析して `ONBOARDING.md` を生成。新メンバーに渡すと Claude Code がプロジェクト固有の文脈を持って動ける。

**ソース**: [@oikon48, 2026-04-11](https://x.com/i/status/2042796840865927288)

**使い方インサイト**: TODO — 自社プロジェクトに使えるか試したい。生成された ONBOARDING.md の品質次第。

---

## /ultraplan コマンド (preview)

**概要**: Claude が Web UI 上に実装プランを構築。読んで編集後、Web 上で実行 or ターミナルに戻して実行できる。

**ソース**: [@trq212, 2026-04-10](https://x.com/i/status/2042671370186973589)

**使い方インサイト**: TODO

---

## /loop コマンド (dynamic mode)

**概要**: `/loop` を interval 指定なしで実行すると、Claude が自律的に次の実行タイミングをスケジュールする。Monitor Tool を直接使って polling を回避することもある。

**例**: `/loop check CI on my PR`

**ソース**: [@noahzweben, 2026-04-10](https://x.com/i/status/2042670949003153647)

**使い方インサイト**: TODO — sns-sifter の cron sync 監視 (失敗時通知) に使えるか試す。

---

## /advisor コマンド (Claude Code 2.1.100)

**概要**: `/advisor [opus|sonnet|off]` で advisor モデルを設定。作業の前 / タスク完了時 / エラー解決できない時などに advisor が呼ばれる。

**ソース**: [@oikon48, 2026-04-10](https://x.com/i/status/2042550380031049878)

**使い方インサイト**: TODO — 後述の「Opus advisor + Sonnet/Haiku executor パターン」の Claude Code 版実装。

---

## Monitor Tool

**概要**: Claude が **バックグラウンド監視スクリプトを自作** できるツール。「必要なことが起きた瞬間に AI を起こす」設計で、無駄な polling を排除する。

**例プロンプト**:
> start my dev server and use the MonitorTool to observe for errors

**ソース**:
- [@claudecode_lab, 2026-04-09 — 速報](https://x.com/i/status/2042343156637974619)
- [@trq212, 2026-04-09 — プロンプト例](https://x.com/i/status/2042335178388103559)

**使い方インサイト**: TODO — sifter sync の状態監視 / ビルドエラー検知 / cron job ヘルスチェックなど、長時間タスクの「能動的待機」全般に効きそう。

---

## effortLevel: high + adaptive thinking 無効化

**概要**: Claude Code の "思慮深さ" を取り戻すための設定。作者 Boris 推奨。

```
effortLevel: high
CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING=1
```

**ソース**: [@ijin, 2026-04-08](https://x.com/i/status/2041858856817717527)

**使い方インサイト**: TODO — 実装の質が下がってると感じた時のリカバリスイッチ。複雑な設計判断を要するセッションのデフォルトにしてもいいかも。

---

## Opus advisor + Sonnet/Haiku executor パターン

**概要**: 機能というより設計パターン。Opus を「戦略立案アドバイザー」、Sonnet または Haiku を「実行役」にペアリングする。Opus 単独相当の知能を半額程度で実現できる。Claude Platform 公式機能としても提供開始 (`/advisor`)。

**ソース**:
- [@claudecode_lab, 2026-04-09 — 解説](https://x.com/i/status/2042346367058874802)
- [@claudeai, 2026-04-09 — 公式](https://x.com/i/status/2042308622181339453)

**使い方インサイト**: TODO — sns-sifter のような「設計判断は重い、実装はパターン化されている」プロジェクトに適合しそう。`/advisor opus` を試す。

---

## Claude Code 品質劣化ポストモーテム (2026-04-23)

**概要**: 2026年3〜4月にかけて Claude Code の品質低下報告が相次ぎ、Anthropic が3つの独立した原因を特定・修正した公式ポストモーテム。API は影響なし、Claude Code のみ。

**ソース**: [Anthropic Engineering Blog, 2026-04-23](https://www.anthropic.com/engineering/april-23-postmortem)

**3つの原因**:
1. **推論努力レベルの引き下げ** (3/4〜4/7): デフォルトを high → medium に変更 (レイテンシ軽減目的) → 品質低下報告 → xhigh (Opus 4.7) / high (その他) にリセット
2. **キャッシュ最適化のバグ** (3/26〜4/10): 1h アイドル後に古い推論を削除する機能のバグで、毎ターン推論が削除される状態に → 「健忘的で反復的」な挙動 + 使用制限の予期しない消費
3. **システムプロンプトの冗長性削減** (4/16〜4/20): 「ツール呼び出し間のテキストを25単語以下に」の指示追加 → Opus 4.6/4.7 で 3% の知能低下を検出 → 即座に削除

**教訓**: 複数レイヤーのレビュー・テスト・ドッグフーディングを通過したにもかかわらず、特定条件 (古いセッション) でのみ発現する問題は検出困難だった。今後はパブリックビルドの全員利用、段階的ロールアウト、システムプロンプト変更の厳格統制を実施。

**使い方インサイト**: `effortLevel: high` + `CLAUDE_CODE_DISABLE_ADAPTIVE_THINKING=1` の設定が効く背景がこのポストモーテムで裏付けられた。品質低下を感じたらまずバージョン確認・設定確認が有効。

---

## Opus 4.7 ベンチマーク

**概要**: Opus 4.7 (Thinking) が Opus 4.6 (Thinking) を主要ベンチマークで上回る。Overall #1, Expert #1 等。

**ソース**: [@arena, 2026-04-17](https://x.com/i/status/2045194638630560104)

**使い方インサイト**: TODO — advisor に Opus 4.7 を指定する価値がありそう。

---

## NotebookLM CLI 連携 (notebooklm-py)

**概要**: Google NotebookLM を Python/CLI/AI エージェントから操作できる非公式ライブラリ。Claude Code 経由で NotebookLM を動かせる。

**ソース**: [@L_go_mrk, 2026-04-15](https://x.com/i/status/2044392796749218220)

**使い方インサイト**: TODO — ナレッジベース構築の別経路として検討。liked posts → NotebookLM でポッドキャスト化とか。

---

## Claude Code × Codex 連携 (レビュー・並列処理)

**概要**: Claude Code のターミナルから OpenAI Codex を直接起動できる公式連携 (2026-03 末公開)。異なるモデルでのレビュー (アンサンブル効果)、並列処理、レビューゲート (AI 同士の自動改善ループ) が可能。

**ソース**: [@masahirochaen, 2026-04-13](https://x.com/i/status/2043660643145019619) / [動画](https://youtu.be/wp3wLI3D4EQ)

**3つのユースケース**:
1. **レビュー**: Claude Code で作ったコード・資料を Codex にレビューさせる。異なるモデルでバグの見逃しを防ぐ
2. **並列処理**: Claude Code でメイン作業しながら Codex でリサーチ・ドキュメント作成
3. **レビューゲート**: Claude Code → Codex レビュー → 修正 → 再レビューをループ。AI 同士が自動で議論・改善

**使い方**: `/codex` コマンド、または「codex レビューして」と指示。プロンプトに「作ったら Codex でレビューして」と事前に入れておくことも可能。

**使い方インサイト**: sns-sifter の PR レビューに plugin として導入したい。Claude Code で実装 → Codex でレビューのゲートを skill 化するとハーネスの Sensors 強化になる。
