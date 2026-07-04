//go:build !js || !wasm || extra

package domain

// WattenCpuDifficulty CPU の難易度レベル
type WattenCpuDifficulty int

// WattenのCPU難易度定数
const (
	// WattenCpuDifficultyEasy 低難易度
	WattenCpuDifficultyEasy WattenCpuDifficulty = iota
	// WattenCpuDifficultyNormal 中難易度
	WattenCpuDifficultyNormal
	// WattenCpuDifficultyHard 高難易度
	WattenCpuDifficultyHard
)

// WattenConfig ヴァッテンゲーム設定
type WattenConfig struct {
	CpuDifficulty WattenCpuDifficulty `json:"cd"`
	TargetScore   int                 `json:"ts"` // マッチ終了スコア (先に到達したチームが勝利, デフォルト15)
	MaxRaises     int                 `json:"mr"` // 1ディールで許可する最大レイズ回数 (デフォルト5 → 最大ステーク7)
}

// DefaultWattenConfig デフォルト設定を返す
func DefaultWattenConfig() WattenConfig {
	return WattenConfig{
		CpuDifficulty: WattenCpuDifficultyNormal,
		TargetScore:   15,
		MaxRaises:     5,
	}
}

// Validate 設定値のドメインバリデーション
func (c WattenConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(WattenCpuDifficultyEasy), int(WattenCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target score", c.TargetScore, 1); err != nil {
		return err
	}
	if err := ValidateMin("max raises", c.MaxRaises, 0); err != nil {
		return err
	}
	return nil
}
