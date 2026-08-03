//go:build !js || !wasm || solo

package domain

import "fmt"

// TarocchiniCpuDifficulty CPU の難易度。
type TarocchiniCpuDifficulty int

const (
	// TarocchiniCpuDifficultyEasy ランダムな合法手。
	TarocchiniCpuDifficultyEasy TarocchiniCpuDifficulty = iota
	// TarocchiniCpuDifficultyNormal 基本的な切り札管理。
	TarocchiniCpuDifficultyNormal
	// TarocchiniCpuDifficultyHard 発展的なヒューリスティック。
	TarocchiniCpuDifficultyHard
)

// TarocchiniConfig タロッキーニのゲーム設定。
type TarocchiniConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty TarocchiniCpuDifficulty `json:"cd"`
	// TargetRounds マッチを終える局数。各プレイヤーがディーラーを 1 巡する 4 の倍数。
	TargetRounds int `json:"tr"`
}

// DefaultTarocchiniConfig 既定設定を返す。
func DefaultTarocchiniConfig() TarocchiniConfig {
	return TarocchiniConfig{CpuDifficulty: TarocchiniCpuDifficultyNormal, TargetRounds: 4}
}

// Validate 設定値を検証する。
func (c TarocchiniConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty),
		int(TarocchiniCpuDifficultyEasy), int(TarocchiniCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, TarocchiniPlayerCnt); err != nil {
		return err
	}
	// **プレイヤー数の倍数に限る。**ディーラーは 1 局ごとに回るので、倍数でないと
	// 誰かが余分に親を務めたままマッチが終わり、スカルトの回数が不平等になる。
	if c.TargetRounds%TarocchiniPlayerCnt != 0 {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("局数は %d の倍数でなければなりません: %d", TarocchiniPlayerCnt, c.TargetRounds))
	}
	return nil
}
