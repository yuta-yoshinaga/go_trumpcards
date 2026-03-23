package domain

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
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(SpadesCpuDifficultyEasy), int(SpadesCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	if err := ValidateMin("nil bonus", c.NilBonus, 0); err != nil {
		return err
	}
	if err := ValidateMin("bag penalty threshold", c.BagPenaltyThreshold, 1); err != nil {
		return err
	}
	return nil
}
