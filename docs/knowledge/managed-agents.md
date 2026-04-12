# Claude Managed Agents

Anthropic が提供する **エージェントランタイム基盤**。モデル + ツール + サンドボックス + イベント管理をフルマネージドで提供する。
このドキュメントは X 上での議論・解説・cookbook 紹介を一次情報として、Managed Agents の **入門〜実践〜知見** を継続的に蓄積する。

## 関連ドキュメント

- [Managed Agents 評価](managed-agents-evaluation.md) — 「非エンジニアの業務自動化に使えるか?」という特定の問いへの調査・評価ドキュメント (静的)

---

## 入門としての導線

### 「鍵を一切渡さない」セキュリティ設計 (理解の入口)

技術者でなくても理解できる解説として @pop_ikeda の記事が良い導入。

> AIエージェントに「鍵を一切渡さない」セキュリティ設計、知ってますか？
> Anthropicが公開したClaude Managed Agents。
> 仕組みを知ると「なるほど、こうすれば安全に任せられるのか」と腑に落ちます。
> 技術者じゃなくても読める解説記事にしました。

**ソース**: [@pop_ikeda, 2026-04-11](https://x.com/i/status/2042814195113300125)

**ポイント**:
- **Vault によるクレデンシャル分離** — トークンはアプリケーションコードを通過しない
- **サンドボックスの実行分離** — Claude 生成コードはユーザー端末に到達しない
- **Human-in-the-Loop ゲート** — 破壊的操作の前に人間の承認

→ この3本柱が「鍵を渡さない」を支える設計。非エンジニアへの説明では、この記事を最初の足場にすると入りやすい。

---

## Cookbook: 実装例

### Data Analyst Agent / Slack Data Bot

@charmaine_klee が公開した公式 cookbook 2 本。

- **Data Analyst Agent**: CSV を投入 → サンドボックス内で Python 探索 → チャート + 文章レポート出力
- **Slack Data Bot**: Slack Bot をフロントエンドにして、メンションで CSV 分析を起動 (マルチターン対話対応)

**ソース**: [@charmaine_klee, 2026-04-10](https://x.com/i/status/2042401788884558090)

**所感**:
- 非エンジニアが Slack 経由で AI Agent に CSV を投げるパターンは、組織の業務自動化テンプレートとして極めて有用
- 「エンジニアが Agent を構築 → 非エンジニアが Slack で利用」という公式が Cookbook で示されている
- 詳細は [評価ドキュメント](managed-agents-evaluation.md) 参照

---

## 周辺サービス統合

### Notion との統合

Anthropic と Notion の協業発表。「Notion のタスクボードが Claude の to-do リストになる」というメッセージング。

> @AnthropicAI runs the model and the agent harness.
> Notion is the orchestration layer: context, UI, and a shared place for ...

**ソース**: [@NotionHQ, 2026-04-08](https://x.com/i/status/2041982872698155398)

**ポイント**:
- **Anthropic** = モデル + agent harness (実行基盤)
- **Notion** = オーケストレーション層 (コンテキスト、UI、共有空間)
- → 「インフラと UX レイヤーの分業」というアーキテクチャ思想。Notion 以外にもこのパターンの統合が今後増えそう。

---

## 参考リンク

- [Managed Agents Overview (公式)](https://platform.claude.com/docs/agents/overview)
- [Managed Agents Cookbook (公式)](https://platform.claude.com/cookbook/managed-agents-data-analyst-agent)
- [Managed Agents 評価 (本リポジトリ内)](managed-agents-evaluation.md)
