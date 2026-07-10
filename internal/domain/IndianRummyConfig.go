//go:build !js || !wasm || extra

package domain

// Indian Rummy プレイヤー数の下限・上限・既定値
const (
	// IndianRummyPlayerCountMin プレイヤー数の下限
	IndianRummyPlayerCountMin = 2
	// IndianRummyPlayerCountMax プレイヤー数の上限
	IndianRummyPlayerCountMax = 4
	// IndianRummyDefaultPlayerCount 既定プレイヤー数（人間 1 + CPU 3）
	IndianRummyDefaultPlayerCount = 4
)

// IndianRummyCpuDifficulty CPU の難易度レベル
type IndianRummyCpuDifficulty int

// Indian Rummy の CPU 難易度定数
const (
	// IndianRummyCpuDifficultyEasy 低難易度
	IndianRummyCpuDifficultyEasy IndianRummyCpuDifficulty = iota
	// IndianRummyCpuDifficultyNormal 中難易度
	IndianRummyCpuDifficultyNormal
	// IndianRummyCpuDifficultyHard 高難易度
	IndianRummyCpuDifficultyHard
)

// IndianRummyConfig インドラミーの設定
type IndianRummyConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。2〜4。
	PlayerCount int `json:"pc"`
	// CpuDifficulty CPU 難易度
	CpuDifficulty IndianRummyCpuDifficulty `json:"cd"`
	// TargetRounds ゲーム終了までのラウンド数（この回数を消化した時点で累計最少が勝利）
	TargetRounds int `json:"tr"`
}

// DefaultIndianRummyConfig デフォルト設定を返す
func DefaultIndianRummyConfig() IndianRummyConfig {
	return IndianRummyConfig{
		PlayerCount:   IndianRummyDefaultPlayerCount,
		CpuDifficulty: IndianRummyCpuDifficultyNormal,
		TargetRounds:  IndianRummyDefaultTargetRounds,
	}
}

// Validate 設定値のドメインバリデーション
func (c IndianRummyConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, IndianRummyPlayerCountMin, IndianRummyPlayerCountMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(IndianRummyCpuDifficultyEasy), int(IndianRummyCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, 1); err != nil {
		return err
	}
	return nil
}
