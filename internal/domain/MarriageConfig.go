//go:build !js || !wasm || extra5

package domain

// Marriage プレイヤー数の下限・上限・既定値
const (
	// MarriagePlayerCountMin プレイヤー数の下限
	MarriagePlayerCountMin = 2
	// MarriagePlayerCountMax プレイヤー数の上限
	MarriagePlayerCountMax = 5
	// MarriageDefaultPlayerCount 既定プレイヤー数（人間 1 + CPU 4）
	MarriageDefaultPlayerCount = 5
)

// MarriageCpuDifficulty CPU の難易度レベル
type MarriageCpuDifficulty int

// Marriage の CPU 難易度定数
const (
	// MarriageCpuDifficultyEasy 低難易度
	MarriageCpuDifficultyEasy MarriageCpuDifficulty = iota
	// MarriageCpuDifficultyNormal 中難易度
	MarriageCpuDifficultyNormal
	// MarriageCpuDifficultyHard 高難易度
	MarriageCpuDifficultyHard
)

// MarriageConfig マリッジの設定
type MarriageConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。2〜5。
	PlayerCount int `json:"pc"`
	// CpuDifficulty CPU 難易度
	CpuDifficulty MarriageCpuDifficulty `json:"cd"`
	// TargetRounds ゲーム終了までのラウンド数（この回数を消化した時点で累計最少が勝利）
	TargetRounds int `json:"tr"`
}

// DefaultMarriageConfig デフォルト設定を返す
func DefaultMarriageConfig() MarriageConfig {
	return MarriageConfig{
		PlayerCount:   MarriageDefaultPlayerCount,
		CpuDifficulty: MarriageCpuDifficultyNormal,
		TargetRounds:  MarriageDefaultTargetRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c MarriageConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, MarriagePlayerCountMin, MarriagePlayerCountMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MarriageCpuDifficultyEasy), int(MarriageCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
