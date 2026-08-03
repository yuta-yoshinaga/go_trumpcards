//go:build !js || !wasm || solo

package domain

import "fmt"

// MinchiateCpuDifficulty CPU の難易度。
type MinchiateCpuDifficulty int

const (
	// MinchiateCpuDifficultyEasy ランダムな合法手。
	MinchiateCpuDifficultyEasy MinchiateCpuDifficulty = iota
	// MinchiateCpuDifficultyNormal 基本的な切り札管理。
	MinchiateCpuDifficultyNormal
	// MinchiateCpuDifficultyHard 発展的なヒューリスティック。
	MinchiateCpuDifficultyHard
)

// MinchiateConfig ミンキアーテのゲーム設定。
type MinchiateConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty MinchiateCpuDifficulty `json:"cd"`
	// TargetRounds マッチを終える局数。ディーラーが一巡する 4 の倍数。
	TargetRounds int `json:"tr"`
}

// DefaultMinchiateConfig 既定設定を返す。
func DefaultMinchiateConfig() MinchiateConfig {
	return MinchiateConfig{CpuDifficulty: MinchiateCpuDifficultyNormal, TargetRounds: 4}
}

// Validate 設定値を検証する。
func (c MinchiateConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(MinchiateCpuDifficultyEasy), int(MinchiateCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, MinchiatePlayerCnt); err != nil {
		return err
	}
	// **プレイヤー数の倍数に限る。**ディーラーは 1 局ごとに回るので、倍数でないと
	// 誰かが余分に親を務めたままマッチが終わり、スカルトの回数が不平等になる。
	if c.TargetRounds%MinchiatePlayerCnt != 0 {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("局数は %d の倍数でなければなりません: %d", MinchiatePlayerCnt, c.TargetRounds))
	}
	return nil
}
