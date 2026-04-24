# メモ・気付き

開発中の気付き、Slackでのやり取り、試したいアイデアなどを時系列で記録する。
テーマとして深掘りしたものは個別ドキュメントへ切り出す。

---

## 2026-04-10

- **UX audit を試す** — sns-sifter の CLI UX に対して実施してみる（[参考](https://speakerdeck.com/gotalab555/shi-yang-tong-ridong-kunoxian-he-claude-codede-shi-eru-wojian-zheng-suru)）
- **DeNA の AI レベル定義** — LEVEL 1〜4 の段階定義（[画像](images/dena-ai-level-definition.png)）。sns-sifter でどのレベルを目指すか参考に
- **Managed Agents 調査** → [managed-agents-evaluation.md](managed-agents-evaluation.md) に切り出し

---

## 2026-04-12

### 新規独立ドキュメント (継続蓄積)
- [Claude Code 新機能ノート](claude-code-features.md) — 新機能の継続蓄積記事
- [Managed Agents](managed-agents.md) — Managed Agents 入門/実践の継続記事

### AI 駆動開発の現場 Tips (陳腐化早いので軽くメモ)
- @m_mizutani: Claude Code + cmux + mo (md viewer) + difit (差分) でエディタ不要化
- @kenn: vibe coding の副作用 dead code → JS/TS なら `pnpm add -D knip` + 「Find all dead code using knip」呪文
- @commte: CLAUDE.md に「知識蓄積層 (思考・実装ログ)」を残すと過去文脈参照が改善。`_docs/` 運用ルール
- @cursor_ai: Cursor の cloud agents が PR にデモ動画/スクショを自動添付
- @gota_bara: uxaudit Claude Code Plugin (ユーザー体験を自動測定する OSS)
- @jacopen: dotfiles を Claude Code でモダン化
- @arceyul: よく使う Skills 一覧 (Frontend Design 等)

### 周辺ツール (陳腐化早い)
- **Microsoft markitdown** — pdf/word/excel/PowerPoint/audio/YouTube → markdown 一括変換 (LLM 前処理向け)
- **メルカリ LLM x SRE** — 次世代インシデント対応事例
- **Anthropic 16体並列 Cコンパイラ記事** — 自律長時間運転の実践 Tips の塊
- **CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1** — Claude Code エージェントチーム実験フラグ
- **ChatGPT で README 推敲** — 改善のたびに褒められる

### 単発メモ (2026-04-24)
- **t_wada『作って学ぶAIエージェント』** — @laiso 著。t_wada が買った報告 ([src](https://x.com/i/status/2046412969576423842))

### 蓄積待ち (samples 増やしてから蒸留する深掘りテーマ)
- ~~**Agent Harness 論**~~ → **独立ディレクトリに蒸留済み**: [docs/knowledge/agent-harness/](agent-harness/) (2026-04-24)
- **組織への AI 導入** (現在 11 件) — DeNA "AIにオールイン", Timee AI-DLC, LayerX 対談, 自分の登壇 (PEK2025, アーキテクチャCon)
