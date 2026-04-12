# Managed Agents による非エンジニア向け業務自動化の評価

> 調査日: 2026-04-10

## 背景・課題

組織内の非エンジニアが Google Workspace、Notion、Slack 等に分散した情報の「パッチワーク」（つぎはぎ・転記・集約）に忙殺されている。

- GeminiやNotionAI、SlackAI等の個別AI機能では根本解決にならない
- 情報が複数サービスに分散していること自体が問題
- 非エンジニアが Claude Code や Cursor で自動化したいと言い出しているが、エンジニアリング知識なしでの運用はリスクが高い
- エンジニアが非エンジニアのローカル端末で動く AI Agent にハーネスをかけるのも労力が大きい

## Managed Agents とは

Anthropic が提供するフルマネージドのエージェント実行基盤（現在ベータ版）。

4つのコア概念:

| 概念 | 説明 |
|------|------|
| **Agent** | モデル・システムプロンプト・ツール・MCPサーバーの定義。再利用可能 |
| **Environment** | コンテナテンプレート（pip パッケージ、ネットワーク設定） |
| **Session** | Agent + Environment の実行インスタンス。ファイルマウント・イベント管理 |
| **Events** | アプリとエージェント間のメッセージ交換 |

### Devin 等との違い

| | Managed Agents | Devin | Claude Code |
|---|---|---|---|
| 位置づけ | エージェント構築用 API | 完成された AI 開発者 | ローカル開発ツール |
| 実行環境 | Anthropic クラウド（サンドボックス） | Devin クラウド | ユーザーのローカル端末 |
| カスタマイズ | 自由（プロンプト、ツール、環境） | 限定的 | 自由だがハーネスが難しい |
| 非エンジニア利用 | フロントエンド経由で可能 | Slack/Web UI 経由 | 困難（リスク大） |

## 課題に対する適合性

### 合致度が高い理由

**1. 「エンジニアが構築 → 非エンジニアが利用」が公式に想定されたパターン**

エンジニアが Agent/Environment を構築し、非エンジニアは Slack Bot 等のフロントエンド経由で利用する。公式 Cookbook に Slack Data Analyst Bot の実装例がある。

**2. クレデンシャルの安全な分離（Vault）**

- ユーザーごとに Vault を作成し、各サービスのトークンを格納
- トークンはアプリケーションコードを通過しない（Anthropic がプロキシ）
- 非エンジニアのローカル端末に秘密情報を置く必要がない
- 監査証跡付き

**3. Human-in-the-Loop（承認ゲート）**

- 破壊的操作（メール送信、Notion 更新等）の前に人間の承認を挟める
- Slack ボタン UI と組み合わせた実装例あり
- 「勝手にメール送っちゃった」事故を防止できる

**4. サンドボックスによる実行分離**

- コード実行は Anthropic クラウド上のコンテナ内
- 非エンジニアのローカル端末でコードが動くリスクがゼロ
- クレデンシャルもサンドボックスに到達しない設計

**5. MCP で主要 SaaS に接続可能**

- 確認済み: Slack, Notion, GitHub, Linear, Stripe, Salesforce, Asana
- Google Workspace: 公式 MCP サーバーは未確認。カスタムツールまたはサードパーティ MCP で対応可能

### 想定アーキテクチャ

```
非エンジニア
    ↓ Slack でメンション
Slack Bot (フロントエンド)
    ↓ Managed Agents API
Agent (Anthropic クラウド)
    ├── MCP → Slack API
    ├── MCP → Notion API
    ├── MCP → Google Workspace (カスタムツール)
    └── Vault (ユーザーごとのクレデンシャル)
    ↓ 承認が必要な操作
Human-in-the-Loop (Slack ボタン)
    ↓ 承認後に実行
結果を Slack に返却
```

### 注意点・リスク

- **現在ベータ版** — 動作変更のリスクあり。本番投入は慎重に
- **構築にはエンジニアリングが必要** — Agent/Environment 設定、Webhook 処理、フロントエンド構築
- **Google Workspace 連携が未検証** — 組織の主要ツールだけに、ここがボトルネックになる可能性
- **トークンコスト** — 情報集約タスクは入力トークンが膨らみがち
- **Anthropic API 直接のみ** — AWS Bedrock / Google Vertex AI では利用不可

## コスト

| 項目 | 料金 |
|------|------|
| セッションランタイム | $0.08/時間（`running` 状態のみ課金） |
| Sonnet 入力 | $3/MTok |
| Sonnet 出力 | $15/MTok |
| Haiku 入力 | $1/MTok |
| Haiku 出力 | $5/MTok |
| Web 検索 | $10/1,000検索 |

シンプルなタスクには Haiku を使い、プロンプトキャッシュを活用することでコスト最適化が可能。

## 次のアクション

- [ ] Managed Agents API でデータ分析エージェントを週末に PoC 実施
- [ ] Google Workspace MCP サーバーの有無を調査
- [ ] Slack Bot フロントエンドの構築コストを見積もる
- [ ] 組織の非エンジニアにヒアリング: 最も自動化したい「パッチワーク」業務は何か

## 参考

- [Managed Agents Overview](https://platform.claude.com/docs/agents/overview)
- [Managed Agents Environments](https://platform.claude.com/docs/agents/environments)
- [Managed Agents Tools](https://platform.claude.com/docs/agents/tools)
- [Managed Agents Sessions](https://platform.claude.com/docs/agents/sessions)
- [Cookbook: Data Analyst Agent](https://platform.claude.com/cookbook/managed-agents-data-analyst-agent)
- [Cookbook: Slack Data Bot](https://platform.claude.com/cookbook/managed-agents-slack-data-bot)
- [Cookbook: SRE Incident Responder](https://platform.claude.com/cookbook/managed-agents-sre-incident-responder)
- [Cookbook: Production Setup](https://platform.claude.com/cookbook/managed-agents-cma-operate-in-production)
- [Scaling Managed Agents (Engineering Blog)](https://www.anthropic.com/engineering/managed-agents)
