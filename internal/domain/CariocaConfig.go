//go:build !js || !wasm || extra

package domain

// Carioca プレイヤー数の下限・上限・既定値
const (
	// CariocaPlayerCountMin プレイヤー数の下限
	CariocaPlayerCountMin = 3
	// CariocaPlayerCountMax プレイヤー数の上限
	CariocaPlayerCountMax = 6
	// CariocaDefaultPlayerCount 既定プレイヤー数（人間 1 + CPU 3）
	CariocaDefaultPlayerCount = 4
)

// CariocaCpuDifficulty CPU の難易度レベル
type CariocaCpuDifficulty int

// Carioca の CPU 難易度定数
const (
	// CariocaCpuDifficultyEasy 低難易度
	CariocaCpuDifficultyEasy CariocaCpuDifficulty = iota
	// CariocaCpuDifficultyNormal 中難易度
	CariocaCpuDifficultyNormal
	// CariocaCpuDifficultyHard 高難易度
	CariocaCpuDifficultyHard
)

// CariocaConfig カリオカの設定
type CariocaConfig struct {
	// PlayerCount 参加プレイヤー数（人間 1 + CPU）。3〜6。
	PlayerCount int `json:"pc"`
	// CpuDifficulty CPU 難易度
	CpuDifficulty CariocaCpuDifficulty `json:"cd"`
	// FailContractPenalty コントラクト未達でラウンド終了したプレイヤーに加算する追加ペナルティ
	FailContractPenalty int `json:"fp"`
}

// DefaultCariocaConfig デフォルト設定を返す
func DefaultCariocaConfig() CariocaConfig {
	return CariocaConfig{
		PlayerCount:         CariocaDefaultPlayerCount,
		CpuDifficulty:       CariocaCpuDifficultyNormal,
		FailContractPenalty: 25,
	}
}

// Validate 設定値のドメインバリデーション
func (c CariocaConfig) Validate() error {
	if err := ValidateRange("player count", c.PlayerCount, CariocaPlayerCountMin, CariocaPlayerCountMax); err != nil {
		return err
	}
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CariocaCpuDifficultyEasy), int(CariocaCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("fail contract penalty", c.FailContractPenalty, 0); err != nil {
		return err
	}
	return nil
}
