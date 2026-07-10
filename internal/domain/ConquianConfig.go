//go:build !js || !wasm || extra

package domain

// ConquianCpuDifficulty CPU の難易度レベル
type ConquianCpuDifficulty int

// ConquianのCPU難易度定数
const (
	// ConquianCpuDifficultyEasy 低難易度
	ConquianCpuDifficultyEasy ConquianCpuDifficulty = iota
	// ConquianCpuDifficultyNormal 中難易度
	ConquianCpuDifficultyNormal
	// ConquianCpuDifficultyHard 高難易度
	ConquianCpuDifficultyHard
)

// ConquianConfig コンキャンゲーム設定
type ConquianConfig struct {
	CpuDifficulty ConquianCpuDifficulty `json:"cd"`
	TargetWins    int                   `json:"tw"` // マッチ勝利数 (先にこの数のラウンドを取ったプレイヤーが勝利)
}

// DefaultConquianConfig デフォルト設定を返す
func DefaultConquianConfig() ConquianConfig {
	return ConquianConfig{
		CpuDifficulty: ConquianCpuDifficultyNormal,
		TargetWins:    1,
	}
}

// Validate 設定値のドメインバリデーション
func (c ConquianConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ConquianCpuDifficultyEasy), int(ConquianCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("target wins", c.TargetWins, 1, 100); err != nil {
		return err
	}
	return nil
}
