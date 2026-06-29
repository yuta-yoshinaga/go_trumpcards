//go:build !js || !wasm || extra

package domain

// HandAndFootCpuDifficulty CPU の難易度レベル
type HandAndFootCpuDifficulty int

// HandAndFootのCPU難易度定数
const (
	// HandAndFootCpuDifficultyEasy 低難易度
	HandAndFootCpuDifficultyEasy HandAndFootCpuDifficulty = iota
	// HandAndFootCpuDifficultyNormal 中難易度
	HandAndFootCpuDifficultyNormal
	// HandAndFootCpuDifficultyHard 高難易度
	HandAndFootCpuDifficultyHard
)

// HandAndFootConfig ハンドアンドフットゲーム設定
type HandAndFootConfig struct {
	CpuDifficulty HandAndFootCpuDifficulty `json:"cd"`
	PointLimit    int                      `json:"pl"` // ゲーム終了スコア (先に到達したチームが勝利)
	// RedCanastasToGoOut 上がるためにチームが必要とする赤（クリーン）カナスタ数。
	// 本来のルールでは 2〜3 だが、それだと CPU 同士のラウンドが極端に長くなるため
	// デフォルト 1 とし、設定で変更可能とする（簡略化）。
	RedCanastasToGoOut int `json:"rc"`
	// BlackCanastasToGoOut 上がるためにチームが必要とする黒（ダーティ）カナスタ数。
	// 同上の理由でデフォルト 1（簡略化）。
	BlackCanastasToGoOut int `json:"bc"`
}

// HandAndFootDefaultPointLimit デフォルトの目標スコア
const HandAndFootDefaultPointLimit = 5000

// DefaultHandAndFootConfig デフォルト設定を返す
func DefaultHandAndFootConfig() HandAndFootConfig {
	return HandAndFootConfig{
		CpuDifficulty:        HandAndFootCpuDifficultyNormal,
		PointLimit:           HandAndFootDefaultPointLimit,
		RedCanastasToGoOut:   1,
		BlackCanastasToGoOut: 1,
	}
}

// Validate 設定値のドメインバリデーション
func (c HandAndFootConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(HandAndFootCpuDifficultyEasy), int(HandAndFootCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
		return err
	}
	if err := ValidateRange("red canastas to go out", c.RedCanastasToGoOut, 0, 10); err != nil {
		return err
	}
	if err := ValidateRange("black canastas to go out", c.BlackCanastasToGoOut, 0, 10); err != nil {
		return err
	}
	return nil
}
