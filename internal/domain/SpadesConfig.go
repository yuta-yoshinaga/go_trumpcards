package domain

import "fmt"

// SpadesCpuDifficulty CPU の難易度レベル
type SpadesCpuDifficulty int

// SpadesのCPU難易度定数
const (
	// SpadesCpuDifficultyEasy 低難易度
	SpadesCpuDifficultyEasy SpadesCpuDifficulty = iota
	// SpadesCpuDifficultyNormal 中難易度
	SpadesCpuDifficultyNormal
	// SpadesCpuDifficultyHard 高難易度
	SpadesCpuDifficultyHard
)

// SpadesConfig スペードゲーム設定
type SpadesConfig struct {
	CpuDifficulty       SpadesCpuDifficulty
	PointLimit          int // ゲーム終了スコア (先に到達したプレイヤーが勝利)
	NilBonus            int // ニルビッド成功時のボーナス
	BagPenaltyThreshold int // バッグペナルティの閾値 (累積10バッグごとに-100)
}

// DefaultSpadesConfig デフォルト設定を返す
func DefaultSpadesConfig() SpadesConfig {
	return SpadesConfig{
		CpuDifficulty:       SpadesCpuDifficultyNormal,
		PointLimit:          500,
		NilBonus:            100,
		BagPenaltyThreshold: 10,
	}
}

// Validate 設定値のドメインバリデーション
func (c SpadesConfig) Validate() error {
	if c.CpuDifficulty < SpadesCpuDifficultyEasy || c.CpuDifficulty > SpadesCpuDifficultyHard {
		return fmt.Errorf("CPU difficulty must be %d-%d, got %d", int(SpadesCpuDifficultyEasy), int(SpadesCpuDifficultyHard), int(c.CpuDifficulty))
	}
	if c.PointLimit < 1 {
		return fmt.Errorf("point limit must be >= 1, got %d", c.PointLimit)
	}
	if c.NilBonus < 0 {
		return fmt.Errorf("nil bonus must be >= 0, got %d", c.NilBonus)
	}
	if c.BagPenaltyThreshold < 1 {
		return fmt.Errorf("bag penalty threshold must be >= 1, got %d", c.BagPenaltyThreshold)
	}
	return nil
}
