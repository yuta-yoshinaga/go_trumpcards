//go:build !js || !wasm || extra

package domain

// Panguingue (Pan) プレイヤー数の下限・上限・既定値
const (
	// PanPlayerCountMin プレイヤー数の下限
	PanPlayerCountMin = 3
	// PanPlayerCountMax プレイヤー数の上限
	PanPlayerCountMax = 6
	// PanDefaultPlayerCount 既定プレイヤー数（人間 1 + CPU 3）
	PanDefaultPlayerCount = 4
)

// PanCpuDifficulty CPU の難易度レベル
type PanCpuDifficulty int

// Panguingue の CPU 難易度定数
const (
	// PanCpuDifficultyEasy 低難易度
	PanCpuDifficultyEasy PanCpuDifficulty = iota
	// PanCpuDifficultyNormal 中難易度
	PanCpuDifficultyNormal
	// PanCpuDifficultyHard 高難易度
	PanCpuDifficultyHard
)

// PanConfig パングインゲの設定
type PanConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。3〜6。
	PlayerCount int `json:"pc"`
	// CpuDifficulty CPU 難易度
	CpuDifficulty PanCpuDifficulty `json:"cd"`
	// TargetRounds ゲーム終了までのラウンド数（この回数を消化した時点で累計最少が勝利）
	TargetRounds int `json:"tr"`
}

// DefaultPanConfig デフォルト設定を返す
func DefaultPanConfig() PanConfig {
	return PanConfig{
		PlayerCount:   PanDefaultPlayerCount,
		CpuDifficulty: PanCpuDifficultyNormal,
		TargetRounds:  PanDefaultTargetRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c PanConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, PanPlayerCountMin, PanPlayerCountMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(PanCpuDifficultyEasy), int(PanCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
