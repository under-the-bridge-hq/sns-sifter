# ハーネスエンジニアリング: 概念と実装パターン

コア記事の「Agent = Model + Harness」テーゼを、外部情報で補強・深化させる。

---

## 概念の系譜

### コミュニティ言語の変遷

> "have you noticed the new 'community language' is 'harness', it was 'contextual prompting', before that 'prompt engineering'"
>
> — @CobusGreylingZA, [2026-04-12](https://x.com/i/status/2043363380551934461)

プロンプト → コンテキスト → ハーネスの進化をコミュニティ言語の変遷として捉える視点。コア記事の「3段階の進化」を外部から裏付け。

### harness の 12 コンポーネント

> "The 12-component breakdown is the best map I've seen of what a production harness actually contains."
>
> — @ghumare64, [2026-04-17](https://x.com/i/status/2045291112454402194)

production harness の構成要素を 12 に分解した記事。コア記事の Guides/Sensors を具体化する地図として有用。

### 各社の harness 比較

> "What does every big company think about the agent harness? Anthropic, OpenAI, CrewAI, LangChain."
>
> — @akshay_pachaar, [2026-04-10](https://x.com/i/status/2042586319390674994)

Anthropic, OpenAI, CrewAI, LangChain 各社が harness をどう構築しているか比較。同じ概念を各社が独自の用語・アーキテクチャで実装している構図。

---

## Guides の実装パターン

### Karpathy Skills

> "Karpathy 吐槽大模型写代码的毛病，编译成了大模型能看懂的约束指令。不到 70 行の一个文件，就拿了接近 6 万颗 Stars"
>
> — @Jason23818126, [2026-04-19](https://x.com/i/status/2045735027447898357)

70 行の制約指示ファイルが 6 万 Stars。Guides の本質 = モデルの癖を補正する短い制約セット。

### gh skill (GitHub CLI 統合)

> "GitHub の gh コマンドに Agent Skills のインストール、管理、公開する skill が追加。Claude Code、Cursor、Codex 等に対応"
>
> — @nukonuko, [2026-04-17](https://x.com/i/status/2044975851796893975)

Skills = Guides のパッケージ管理。npm/pip と同じ感覚で `gh skill install` できる。Guides のエコシステム化。

### APM (Agent Package Manager)

> "Microsoft が OSS で出してる APM、CLAUDE.md や Skills をチームで同期するのにかなり良さそう"
>
> — @kajikent, [2026-04-17](https://x.com/i/status/2044948850893623688)

Guides の組織的な同期問題への解。個人のハーネスを組織で共有するインフラ。

---

## Sensors の実装パターン

### /grill-me スキル

> "Claude Code に /grill-me を入れると、コードを1行も書く前に 40 個以上の質問で詰められる。'最もインパクトのあるスキル' と言われている"
>
> — @shin_sasaki19, [2026-04-18](https://x.com/i/status/2045491997017014399)

事前質問による仕様の検証。Sensors の一形態だが、行動の「前」に機能する点で Guides にも近い。仕様の曖昧さを検知するセンサー。

---

## 人間の役割: prompting as skill

> "I think 'prompting' will keep being an incredibly high-leverage skill, like writing or public speaking. It is the skill of talking to agents, mediated by the harness."
>
> — @trq212, [2026-04-09](https://x.com/i/status/2042318547519762678)

プロンプティングはハーネスに取って代わられるのではなく、ハーネスを介したエージェントとの対話スキルとして残る。コア記事の HITL→HOTL の文脈で、HOTL 側の人間に求められるスキルの一つ。
