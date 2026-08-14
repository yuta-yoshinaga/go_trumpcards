//go:build !js || !wasm || extra3

package domain

import "fmt"

const (
	// RollingStonePlayerCntMin は最小プレイヤー数。
	RollingStonePlayerCntMin = 4
	// RollingStonePlayerCntMax は最大プレイヤー数。
	RollingStonePlayerCntMax = 6
	// RollingStoneDefaultPlayerCnt は既定のプレイヤー数。
	RollingStoneDefaultPlayerCnt = 4
)

// RollingStoneConfig はローリングストーンの設定。
type RollingStoneConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int `json:"p"`
}

// DefaultRollingStoneConfig は既定設定を返す。
func DefaultRollingStoneConfig() RollingStoneConfig {
	return RollingStoneConfig{PlayerCnt: RollingStoneDefaultPlayerCnt}
}

// Validate は設定を検証する。
func (c RollingStoneConfig) Validate() error {
	if c.PlayerCnt < RollingStonePlayerCntMin || c.PlayerCnt > RollingStonePlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			RollingStonePlayerCntMin, RollingStonePlayerCntMax, c.PlayerCnt)
	}
	return nil
}
