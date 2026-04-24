# Agent Harness: コア記事の核心テーゼ

**ソース**: [【Timee AI Sprint Day7】Agent Harness Group 設立に寄せて](https://productpr.timee.co.jp/n/n99011b2a6947) — 橋本 和宏 (@kaz_utb), 2026-04-21

---

## 3つの課題

### 課題1: 個人は速い。でも組織は速くなっているのか？

- AI でコーディング速度は向上 (PR数・コミット数増加) しても、プロダクトリリース速度は期待ほど増加しない
- **AI 生産性パラドックス**: 制約理論 (TOC) により、コーディング高速化 → 後続工程 (レビュー、テスト、承認) にボトルネックが移動
- → 詳細は [productivity-paradox.md](productivity-paradox.md)

### 課題2: 速くなった分、ちゃんと動くものを作れているのか？

- AI 生成コードは人間のコードより 1.7倍多くの問題を含有、XSS 系脆弱性は 2.74倍 (CodeRabbit 調査)
- テスト合格 ≠ ビジネスロジック正確性。「動くけど正しくない」ギャップ
- **節約したコーディング時間の一部は、検証コストに再投資される不可避な構造**

### 課題3: 速度と安全は対立するのか？

- DORA 研究: スピードと安定性は正の相関。ただし自動化は自動では起こらない
- AI は「速度側を強力に加速するツール」だが、安全側の仕組みは自動的に付随しない
- **DORA 2025: 「AI はチームを修正しない。既にあるものを増幅する」** — 良い仕組みも悪い慣習も同じ勢いで拡大
- → 詳細は [productivity-paradox.md](productivity-paradox.md)

---

## ハーネスエンジニアリング

### Agent = Model + Harness

> ハーネスとは、AI エージェントの振る舞いを形作る、モデル以外のすべて —— ツール・ルール・センサー・ガイドの設定層の総体
>
> — Birgitta Böckeler (Thoughtworks), Martin Fowler サイト

**Karpathy のアナロジーの拡張**:
- モデル = CPU / コンテキスト = RAM / **ハーネス = OS**
- CPU 性能を高めても OS が貧弱だとアプリケーションは正常に動作しない

### 進化3段階

1. **2023-24: プロンプトエンジニアリング** — 属人的、再現性低、スケールしない
2. **2025: コンテキストエンジニアリング** — AI に渡す文脈全体の設計
3. **2026: ハーネスエンジニアリング** — AI を取り囲む環境全体の設計

置き換わりではなく上位概念による包含。モデル進化に伴いプロンプト職人芸の比重は低下し、ハーネス設計の重みが増加。

### Guides と Sensors

| | Guides (ガイド) | Sensors (センサー) |
|--|---|---|
| タイミング | 行動の**前** | 行動の**後** |
| 機能 | 進むべき方向を示す | 結果の正確性を検知 |
| 例 | AGENTS.md, コーディング規約, アーキ文書 | テスト, リンター, 型チェッカー, レビューエージェント |

両方が必要。ガイドのみ → 効果が見えない。センサーのみ → 同じ失敗を繰り返す。

→ 詳細は [harness-engineering.md](harness-engineering.md)

---

## Human in the loop → Human on the loop

| | Human in the loop (HITL) | Human on the loop (HOTL) |
|--|---|---|
| 対象 | 個々の AI 出力 | ハーネスそのもの |
| 行動 | PR レビュー、プロンプト書き直し、出力の手直し | ツール追加、ルール更新、センサー追加 |
| 特徴 | 個人の努力に依存 | 環境側で再発防止 |
| 方向 | 職人芸 | エンジニアリング |

---

## Agent Harness Group (AHG) のミッション

> Coding Agent を中心とした AI ワークフローにおいて、生産性と安全性の共進化を計測可能な形で実現するプラットフォームを構築し、Autonomous 推進を加速する

### 3つの柱

1. **Accelerate (加速)** — AI ワークフローの生産性向上。ツール自動化、エージェント能力向上、環境最適化
2. **Govern (統治・計測)** — 生産性と安全性を計測・可視化。品質メトリクス、セキュリティ監視、リスク計測
3. **Cultivate (培養・文化醸成)** — チーム文化と改善サイクル。ナレッジ共有、教育、on the loop 文化の育成

### OODA ループで3柱を共進化

Observe (Govern 計測) → Orient (Cultivate 対話) → Decide (次の施策) → Act (Accelerate 実行) → サイクル継続

計画起点の PDCA ではなく、観測起点の OODA。変化の速い状況向け。

---

## 未解決の問い

- **ハーネスカバレッジの評価**: Fowler/Böckeler 記事自身が「未解決」と指摘。業界として答えがまだない
- → [open-questions.md](open-questions.md)

---

## 参考文献 (記事内引用)

- Birgitta Böckeler, "Harness Engineering for Coding Agent Users" (Martin Fowler site, 2026)
- DORA 2025 Annual Report — https://dora.dev/research/
- CodeRabbit, "State of AI vs Human Code Generation Report" (2025-12)
- Andrej Karpathy, LLM-OS アナロジー
- a-thug, "全PRの83%をAIレビューのみでマージできるようになるまで" (Zenn, カウシェ)
