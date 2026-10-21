package xapi

// X API 単価 (USD/件)。出典: https://docs.x.com/x-api/getting-started/pricing
const (
	// PricePostRead は一般の Post 読み取り。
	PricePostRead = 0.005
	// PriceOwnedRead は認証ユーザー自身の liked_tweets 等 (アプリ所有者が自分のデータを読む場合)。
	PriceOwnedRead = 0.001
	// PriceUserRead は User lookup。
	PriceUserRead = 0.010
	// PriceFollowingRead は Following/Followers 読み取り。
	PriceFollowingRead = 0.010
)
