//go:build !js || !wasm || extra

package domain

// ContractRummyCpuDifficulty CPU の難易度レベル
type ContractRummyCpuDifficulty int

// ContractRummy の CPU 難易度定数
const (
	// ContractRummyCpuDifficultyEasy 低難易度
	ContractRummyCpuDifficultyEasy ContractRummyCpuDifficulty = iota
	// ContractRummyCpuDifficultyNormal 中難易度
	ContractRummyCpuDifficultyNormal
	// ContractRummyCpuDifficultyHard 高難易度
	ContractRummyCpuDifficultyHard
)

// ContractRummyConfig コントラクトラミーの設定
type ContractRummyConfig struct {
	CpuDifficulty ContractRummyCpuDifficulty `json:"cd"`
	// FailContractPenalty コントラクト未達でラウンド終了したプレイヤーに加算する追加ペナルティ
	FailContractPenalty int `json:"fp"`
}

// DefaultContractRummyConfig デフォルト設定を返す
func DefaultContractRummyConfig() ContractRummyConfig {
	return ContractRummyConfig{
		CpuDifficulty:       ContractRummyCpuDifficultyNormal,
		FailContractPenalty: 25,
	}
}

// Validate 設定値のドメインバリデーション
func (c ContractRummyConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ContractRummyCpuDifficultyEasy), int(ContractRummyCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("fail contract penalty", c.FailContractPenalty, 0); err != nil {
		return err
	}
	return nil
}
