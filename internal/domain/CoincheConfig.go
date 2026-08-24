//go:build !js || !wasm || extra3

package domain

// CoincheCpuDifficulty CPU の難易度レベル
type CoincheCpuDifficulty int

// CoincheのCPU難易度定数
const (
	// CoincheCpuDifficultyEasy 低難易度
	CoincheCpuDifficultyEasy CoincheCpuDifficulty = iota
	// CoincheCpuDifficultyNormal 中難易度
	CoincheCpuDifficultyNormal
	// CoincheCpuDifficultyHard 高難易度
	CoincheCpuDifficultyHard
)

// CoincheConfig コワンシュゲーム設定
type CoincheConfig struct {
	CpuDifficulty        CoincheCpuDifficulty `json:"cd"`
	TargetScore          int                  `json:"ts"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト1000)
	DixDeDer             int                  `json:"dd"` // 最終トリックボーナス (デフォルト10)
	EnableBeloteRebelote bool                 `json:"br"` // K+Q トランプによるベロート/ルベロート (+20) を有効化。この game で
	// coinche は倍化のことなので、メルドの側は Belote/Rebelote と呼ぶ。
}

// DefaultCoincheConfig デフォルト設定を返す
func DefaultCoincheConfig() CoincheConfig {
	return CoincheConfig{
		CpuDifficulty:        CoincheCpuDifficultyNormal,
		TargetScore:          1000,
		DixDeDer:             10,
		EnableBeloteRebelote: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c CoincheConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(CoincheCpuDifficultyEasy), int(CoincheCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	if err := ValidateMin("dix de der", c.DixDeDer, 0); err != nil {
		return err
	}
	return nil
}
