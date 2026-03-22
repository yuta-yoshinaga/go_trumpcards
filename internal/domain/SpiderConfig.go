package domain

// SpiderDifficulty スパイダーソリティア難易度
type SpiderDifficulty int

// Spiderの難易度定数
const (
	// SpiderDifficulty1Suit 1スート（初級）
	SpiderDifficulty1Suit SpiderDifficulty = 1
	// SpiderDifficulty2Suit 2スート（中級）
	SpiderDifficulty2Suit SpiderDifficulty = 2
	// SpiderDifficulty4Suit 4スート（上級）
	SpiderDifficulty4Suit SpiderDifficulty = 4
)

// SpiderConfig スパイダーソリティア設定
type SpiderConfig struct {
	Difficulty SpiderDifficulty
}

// DefaultSpiderConfig デフォルト設定（1スート）
func DefaultSpiderConfig() SpiderConfig {
	return SpiderConfig{Difficulty: SpiderDifficulty1Suit}
}
