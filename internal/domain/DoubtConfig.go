package domain

// DoubtMemoryLevel CPU の記憶力レベル
type DoubtMemoryLevel int

// Doubtの記憶レベル定数
const (
	// DoubtMemoryLevelEasy 低記憶力 (約30%の確率で記憶)
	DoubtMemoryLevelEasy DoubtMemoryLevel = iota
	// DoubtMemoryLevelNormal 中記憶力 (約70%の確率で記憶)
	DoubtMemoryLevelNormal
	// DoubtMemoryLevelHard 高記憶力 (100%記憶)
	DoubtMemoryLevelHard
)

// DoubtConfig ダウトゲーム設定
type DoubtConfig struct {
	DoubtWindowSec       int
	CpuMemoryLevel       DoubtMemoryLevel
	PenaltyDrawLimit     int  // 0 = unlimited; >0 = loser draws at most N cards
	CpuHesitationEnabled bool // CPU迷い時間ディレイ
	CpuMetaAI            bool // メタAI: セッション内学習
}

// DefaultDoubtConfig デフォルト設定を返す
func DefaultDoubtConfig() DoubtConfig {
	return DoubtConfig{DoubtWindowSec: 10, CpuMemoryLevel: DoubtMemoryLevelNormal}
}
