//go:build !js || !wasm || solo

package domain

import "fmt"

// SnapCpuDifficulty CPU の難易度
type SnapCpuDifficulty int

// SnapCpuDifficulty 定数
const (
	// SnapCpuEasy 易しい (反応が遅い)
	SnapCpuEasy SnapCpuDifficulty = 0
	// SnapCpuNormal 普通
	SnapCpuNormal SnapCpuDifficulty = 1
	// SnapCpuHard 難しい (反応が速い)
	SnapCpuHard SnapCpuDifficulty = 2
)

const (
	// SnapPlayerCntMin は最小プレイヤー数。
	SnapPlayerCntMin = 2
	// SnapPlayerCntMax は最大プレイヤー数。
	SnapPlayerCntMax = 4
	// SnapDefaultPlayerCnt は既定のプレイヤー数。
	SnapDefaultPlayerCnt = 2
)

// SnapMinReactionMs CPU 反応時間の下限 (人間に勝ち目を残すための床値)
const SnapMinReactionMs = 80

// SnapConfig はスナップの設定。
type SnapConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int `json:"p"`
	// CpuDifficulty は CPU の難易度。
	CpuDifficulty SnapCpuDifficulty `json:"d"`
}

// DefaultSnapConfig は既定設定を返す。
func DefaultSnapConfig() SnapConfig {
	return SnapConfig{PlayerCnt: SnapDefaultPlayerCnt, CpuDifficulty: SnapCpuNormal}
}

// Validate は設定を検証する。
func (c SnapConfig) Validate() error {
	if c.PlayerCnt < SnapPlayerCntMin || c.PlayerCnt > SnapPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			SnapPlayerCntMin, SnapPlayerCntMax, c.PlayerCnt)
	}
	if c.CpuDifficulty < SnapCpuEasy || c.CpuDifficulty > SnapCpuHard {
		return fmt.Errorf("invalid cpu difficulty: %d", c.CpuDifficulty)
	}
	return nil
}
