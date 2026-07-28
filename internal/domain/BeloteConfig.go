//go:build !js || !wasm || extra3

package domain

// BeloteCpuDifficulty CPU の難易度レベル
type BeloteCpuDifficulty int

// BeloteのCPU難易度定数
const (
	// BeloteCpuDifficultyEasy 低難易度
	BeloteCpuDifficultyEasy BeloteCpuDifficulty = iota
	// BeloteCpuDifficultyNormal 中難易度
	BeloteCpuDifficultyNormal
	// BeloteCpuDifficultyHard 高難易度
	BeloteCpuDifficultyHard
)

// BeloteConfig ベロートゲーム設定
type BeloteConfig struct {
	CpuDifficulty        BeloteCpuDifficulty `json:"cd"`
	TargetScore          int                 `json:"ts"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト1000)
	DixDeDer             int                 `json:"dd"` // 最終トリックボーナス (デフォルト10)
	EnableBeloteRebelote bool                `json:"br"` // K+Q トランプによるベロート/レベロート (+20) を有効化
}

// DefaultBeloteConfig デフォルト設定を返す
func DefaultBeloteConfig() BeloteConfig {
	return BeloteConfig{
		CpuDifficulty:        BeloteCpuDifficultyNormal,
		TargetScore:          1000,
		DixDeDer:             10,
		EnableBeloteRebelote: true,
	}
}

// Validate 設定値のドメインバリデーション
func (c BeloteConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(BeloteCpuDifficultyEasy), int(BeloteCpuDifficultyHard)); err != nil {
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
