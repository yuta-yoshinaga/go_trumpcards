//go:build !js || !wasm || extra

package domain

import "fmt"

// ViraCpuDifficulty CPU の難易度。
type ViraCpuDifficulty int

const (
	// ViraCpuDifficultyEasy ランダムな合法手。
	ViraCpuDifficultyEasy ViraCpuDifficulty = iota
	// ViraCpuDifficultyNormal 基本的な切り札管理。
	ViraCpuDifficultyNormal
	// ViraCpuDifficultyHard 発展的なヒューリスティック。
	ViraCpuDifficultyHard
)

// ViraConfig ヴィーラのゲーム設定。
type ViraConfig struct {
	// CpuDifficulty CPU の難易度。
	CpuDifficulty ViraCpuDifficulty `json:"cd"`
	// TargetRounds マッチを終える局数。各プレイヤーがディーラーを 1 巡する 3 の倍数。
	TargetRounds int `json:"tr"`
}

// DefaultViraConfig デフォルト設定を返す (6 局 = 各プレイヤーがディーラーを 2 回)。
func DefaultViraConfig() ViraConfig {
	return ViraConfig{CpuDifficulty: ViraCpuDifficultyNormal, TargetRounds: 6}
}

// Validate 設定値のドメインバリデーション。
func (c ViraConfig) Validate() error {
	if err := ValidateRange("CPU difficulty", int(c.CpuDifficulty), int(ViraCpuDifficultyEasy), int(ViraCpuDifficultyHard)); err != nil {
		return err
	}
	if err := ValidateMin("target rounds", c.TargetRounds, ViraPlayerCnt); err != nil {
		return err
	}
	// **局数はプレイヤー数の倍数でなければならない。**ディーラーは 1 局ごとに回るので、
	// 倍数でないと各プレイヤーがディーラーを務めた回数が揃わず、
	// 配り順の有利不利が精算に残ったままマッチが終わる。
	if c.TargetRounds%ViraPlayerCnt != 0 {
		return NewDomainError(ErrInvalidPlay,
			fmt.Sprintf("局数は %d の倍数でなければなりません: %d", ViraPlayerCnt, c.TargetRounds))
	}
	return nil
}
