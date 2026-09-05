//go:build !js || !wasm || extra5

package domain

// MarjapussiCpuDifficulty CPU の難易度レベル
type MarjapussiCpuDifficulty int

// Marjapussi の CPU 難易度定数
const (
	// MarjapussiCpuDifficultyEasy 低難易度 (ランダムプレイ)
	MarjapussiCpuDifficultyEasy MarjapussiCpuDifficulty = iota
	// MarjapussiCpuDifficultyNormal 中難易度 (戦略プレイ)
	MarjapussiCpuDifficultyNormal
	// MarjapussiCpuDifficultyHard 高難易度 (戦略プレイ)
	MarjapussiCpuDifficultyHard
)

// MarjapussiDefaultPointLimit デフォルトの目標スコア (500点先取)
const MarjapussiDefaultPointLimit = 500

// MarjapussiConfig マルヤプッシのゲーム設定
type MarjapussiConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty MarjapussiCpuDifficulty `json:"cd"`
	// PointLimit マッチ勝利に必要な累積点 (デフォルト 500)。
	PointLimit int `json:"pl"`
	// TargetPoints マッチ勝利に必要な累積点 (PointLimit と同期)。
	TargetPoints int `json:"tp"`
}

// DefaultMarjapussiConfig デフォルト設定を返す (標準は 500 点先取)。
func DefaultMarjapussiConfig() MarjapussiConfig {
	return MarjapussiConfig{
		CpuDifficulty: MarjapussiCpuDifficultyNormal,
		PointLimit:    MarjapussiDefaultPointLimit,
		TargetPoints:  MarjapussiDefaultPointLimit,
	}
}

// Validate 設定値のドメインバリデーション
func (c MarjapussiConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(MarjapussiCpuDifficultyEasy), int(MarjapussiCpuDifficultyHard)); err != nil {
		return err
	}
	if c.PointLimit > 0 {
		if err := ValidateMin("point limit", c.PointLimit, 1); err != nil {
			return err
		}
	}
	if c.TargetPoints > 0 {
		if err := ValidateMin("target points", c.TargetPoints, 1); err != nil {
			return err
		}
	}
	if c.PointLimit <= 0 && c.TargetPoints <= 0 {
		return ValidateMin("point limit", c.PointLimit, 1)
	}
	return nil
}

func (c MarjapussiConfig) pointLimit() int {
	if c.PointLimit > 0 {
		return c.PointLimit
	}
	if c.TargetPoints > 0 {
		return c.TargetPoints
	}
	return MarjapussiDefaultPointLimit
}
