package domain

import (
	"regexp"
	"strings"
)

// 英語の AI 関連語。word boundary 付き正規表現で誤マッチを防ぐ
// (例: "tail" や "available" の "ai" を AI と誤判定しない)。
var aiEnglishRegex = regexp.MustCompile(`(?i)\b(` +
	`claude|anthropic|openai|gpt|llm|chatgpt|` +
	`agent|agents|agentic|mcp|` +
	`cursor|copilot|codex|devin|` +
	`prompt|prompting|rag|embedding|vector|` +
	`machine learning|deep learning|neural|transformer|` +
	`ai|a\.i\.|artificial intelligence|` +
	`vibe coding|dead code|knip|` +
	`skill|skills` +
	`)\b`)

// 日本語の AI 関連語 (substring match)。日本語は word boundary を取りにくいので
// 複合語を直接列挙する。
var aiJapaneseKeywords = []string{
	"生成AI", "機械学習", "深層学習", "推論",
	"プロンプト", "エージェント", "言語モデル",
	"AI駆動", "AIエージェント", "AIメンター", "AIツール", "AI開発", "AIコーディング",
	"バイブコーディング", "コード生成",
	"Claude Code", "Skill", "スキル",
	"AI&", "AIと", "AI×",
}

// 英語の Work / 組織論関連
var workEnglishRegex = regexp.MustCompile(`(?i)\b(` +
	`management|leadership|career|hiring|recruiting|` +
	`organization|culture|team|engineering manager|sre|` +
	`productivity|workflow|process|onboarding|` +
	`startup|series [a-z]|founder|scaling` +
	`)\b`)

var workJapaneseKeywords = []string{
	"組織", "組織論", "マネジメント", "リーダーシップ", "キャリア",
	"採用", "評価", "1on1", "ワン・オン・ワン", "ピープル",
	"チーム", "チームビルディング", "文化", "カルチャー",
	"事業", "経営", "ビジネス", "シリーズB", "シリーズA",
	"領域を越境", "越境", "オンボーディング",
	"スタートアップ", "創業",
}

// ClassifyByKeyword は text を ai / work / other に分類する。
// AI と Work の両方にマッチした場合は AI を優先 (sns-sifter の主目的が AI 情報収集のため)。
func ClassifyByKeyword(text string) string {
	if aiEnglishRegex.MatchString(text) {
		return "ai"
	}
	for _, kw := range aiJapaneseKeywords {
		if strings.Contains(text, kw) {
			return "ai"
		}
	}
	if workEnglishRegex.MatchString(text) {
		return "work"
	}
	for _, kw := range workJapaneseKeywords {
		if strings.Contains(text, kw) {
			return "work"
		}
	}
	return "other"
}
