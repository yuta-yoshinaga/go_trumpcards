//go:build !js || !wasm || extra4

package domain

import "fmt"

const (
	// PasurPlayerCntMin は最小プレイヤー数。
	PasurPlayerCntMin = 2
	// PasurPlayerCntMax は最大プレイヤー数。
	PasurPlayerCntMax = 4
	// PasurDefaultPlayerCnt は既定のプレイヤー数。
	PasurDefaultPlayerCnt = 4
)

// PasurConfig はパスールの設定。
type PasurConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int `json:"p"`
}

// DefaultPasurConfig は既定設定を返す。
func DefaultPasurConfig() PasurConfig {
	return PasurConfig{PlayerCnt: PasurDefaultPlayerCnt}
}

// Validate は設定を検証する。
//
// **人数は 2〜4 のどれでも配り切れる。** 場に 4 枚置いた残り 48 枚は
// 2・3・4 のいずれでも 1 人 4 枚ずつで割り切れる（48 / (n×4) = 6 / 4 / 3 パック）。
func (c PasurConfig) Validate() error {
	if c.PlayerCnt < PasurPlayerCntMin || c.PlayerCnt > PasurPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			PasurPlayerCntMin, PasurPlayerCntMax, c.PlayerCnt)
	}
	return nil
}
