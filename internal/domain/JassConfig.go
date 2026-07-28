//go:build !js || !wasm || extra3

package domain

// JassCpuDifficulty CPU の難易度レベル
type JassCpuDifficulty int

// JassのCPU難易度定数
const (
	// JassCpuDifficultyEasy 低難易度
	JassCpuDifficultyEasy JassCpuDifficulty = iota
	// JassCpuDifficultyNormal 中難易度
	JassCpuDifficultyNormal
	// JassCpuDifficultyHard 高難易度
	JassCpuDifficultyHard
)

// JassConfig ヤス(シーバー)ゲーム設定
type JassConfig struct {
	CpuDifficulty  JassCpuDifficulty `json:"cd"`
	TargetScore    int               `json:"ts"` // ゲーム終了スコア (先に到達したチームが勝利, デフォルト1000)
	LastTrickBonus int               `json:"lb"` // 最終トリックボーナス (デフォルト5)
	EnableWeis     bool              `json:"ew"` // Weis (メルド) 宣言を有効化
}

// DefaultJassConfig デフォルト設定を返す
func DefaultJassConfig() JassConfig {
	return JassConfig{
		CpuDifficulty:  JassCpuDifficultyNormal,
		TargetScore:    1000,
		LastTrickBonus: 5,
		EnableWeis:     true,
	}
}

// Validate 設定値のドメインバリデーション
func (c JassConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(JassCpuDifficultyEasy), int(JassCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	if err := ValidateMin("last trick bonus", c.LastTrickBonus, 0); err != nil {
		return err
	}
	return nil
}
