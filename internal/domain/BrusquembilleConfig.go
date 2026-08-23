//go:build !js || !wasm || classic

package domain

// BrusquembilleCpuDifficulty CPU の難易度レベル
type BrusquembilleCpuDifficulty int

// BrusquembilleのCPU難易度定数
const (
	// BrusquembilleCpuDifficultyNormal 標準難易度 (v1で唯一サポート)
	BrusquembilleCpuDifficultyNormal BrusquembilleCpuDifficulty = iota
)

// BrusquembilleConfig ブリュスカンビーユゲーム設定
type BrusquembilleConfig struct {
	CpuDifficulty BrusquembilleCpuDifficulty `json:"cd"`
	// PlayerCnt 席数 (2-5)。席 0 が人間。
	PlayerCnt int `json:"pc"`
}

// DefaultBrusquembilleConfig デフォルト設定を返す
func DefaultBrusquembilleConfig() BrusquembilleConfig {
	return BrusquembilleConfig{
		CpuDifficulty: BrusquembilleCpuDifficultyNormal,
		PlayerCnt:     BrusquembilleDefaultPlayerCnt,
	}
}

// Validate 設定値のドメインバリデーション
func (c BrusquembilleConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(BrusquembilleCpuDifficultyNormal), int(BrusquembilleCpuDifficultyNormal)); err != nil {
		return err
	}
	// **席数も検査する。** ここを素通しにすると、範囲外の席数がそのまま
	// 設定に入り、卓が組めないまま Reset される。
	return ValidateRange("player count", c.PlayerCnt,
		BrusquembilleMinPlayerCnt, BrusquembilleMaxPlayerCnt)
}
