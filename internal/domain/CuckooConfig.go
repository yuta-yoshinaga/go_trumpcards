//go:build !js || !wasm || extra2

package domain

// CuckooCpuDifficulty CPU の難易度レベル
type CuckooCpuDifficulty int

// Cuckoo の CPU 難易度定数
const (
	// CuckooCpuDifficultyEasy 低難易度
	CuckooCpuDifficultyEasy CuckooCpuDifficulty = iota
	// CuckooCpuDifficultyNormal 中難易度
	CuckooCpuDifficultyNormal
	// CuckooCpuDifficultyHard 高難易度
	CuckooCpuDifficultyHard
)

// CuckooMinLives 設定可能な初期ライフの下限
const CuckooMinLives = 1

// CuckooMaxLives 設定可能な初期ライフの上限
const CuckooMaxLives = 10

// CuckooStartLives 各プレイヤーのデフォルト初期ライフ数
const CuckooStartLives = 3

// CuckooConfig Cuckoo (カッコー) ゲーム設定
type CuckooConfig struct {
	CpuDifficulty CuckooCpuDifficulty `json:"cd"`
	InitialLives  int                 `json:"il"` // 各プレイヤーの初期ライフ数
}

// DefaultCuckooConfig デフォルト設定を返す
func DefaultCuckooConfig() CuckooConfig {
	return CuckooConfig{
		CpuDifficulty: CuckooCpuDifficultyNormal,
		InitialLives:  CuckooStartLives,
	}
}

// Validate 設定値のドメインバリデーション
func (c CuckooConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CuckooCpuDifficultyEasy), int(CuckooCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateRange("initial lives", c.InitialLives, CuckooMinLives, CuckooMaxLives); err != nil {
		return err
	}
	return nil
}
