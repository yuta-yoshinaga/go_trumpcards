//go:build !js || !wasm || extra

package domain

// GanjifaCpuDifficulty CPU の難易度。
type GanjifaCpuDifficulty int

const (
	// GanjifaCpuDifficultyEasy ランダムな合法手。
	GanjifaCpuDifficultyEasy GanjifaCpuDifficulty = iota
	// GanjifaCpuDifficultyNormal 基本的な切り札管理。
	GanjifaCpuDifficultyNormal
	// GanjifaCpuDifficultyHard 発展的なヒューリスティック。
	GanjifaCpuDifficultyHard
)

// GanjifaConfig ガンジファのゲーム設定。
type GanjifaConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty GanjifaCpuDifficulty `json:"cd"`
	// TargetRounds マッチを終える局数。
	TargetRounds int `json:"tr"`
}

// DefaultGanjifaConfig デフォルト設定を返す (3 局 = 各プレイヤーがディーラーを 1 回)。
func DefaultGanjifaConfig() GanjifaConfig {
	return GanjifaConfig{CpuDifficulty: GanjifaCpuDifficultyNormal, TargetRounds: 3}
}

// Validate 設定値のドメインバリデーション。
func (c GanjifaConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(GanjifaCpuDifficultyEasy), int(GanjifaCpuDifficultyHard)); err != nil {
		return err
	}
	return ValidateMin("target rounds", c.TargetRounds, 1)
}
