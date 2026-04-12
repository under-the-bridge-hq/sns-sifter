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
