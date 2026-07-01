//go:build !js || !wasm || extra

package domain

// TysiacCpuDifficulty CPU の難易度レベル
type TysiacCpuDifficulty int

// Tysiąc の CPU 難易度定数
const (
	// TysiacCpuDifficultyEasy 低難易度 (ランダムプレイ・最低ビッド)
	TysiacCpuDifficultyEasy TysiacCpuDifficulty = iota
	// TysiacCpuDifficultyNormal 中難易度 (戦略プレイ)
	TysiacCpuDifficultyNormal
	// TysiacCpuDifficultyHard 高難易度 (戦略プレイ)
	TysiacCpuDifficultyHard
)

// TysiacConfig サウザンドのゲーム設定
type TysiacConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TysiacCpuDifficulty `json:"cd"`
	// TargetPoints マッチ勝利に必要な累積点。いずれかのプレイヤーがこの値以上で勝利。
	TargetPoints int `json:"tp"`
}

// DefaultTysiacConfig デフォルト設定を返す (標準は 1000 点先取)。
func DefaultTysiacConfig() TysiacConfig {
	return TysiacConfig{CpuDifficulty: TysiacCpuDifficultyNormal, TargetPoints: TysiacWinTarget}
}

// Validate 設定値のドメインバリデーション
func (c TysiacConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(TysiacCpuDifficultyEasy), int(TysiacCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
		return err
	}
	return nil
}
