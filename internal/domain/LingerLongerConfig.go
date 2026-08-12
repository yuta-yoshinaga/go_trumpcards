//go:build !js || !wasm || extra

package domain

import "fmt"

const (
	// LingerLongerPlayerCntMin は最小プレイヤー数。
	LingerLongerPlayerCntMin = 4
	// LingerLongerPlayerCntMax は最大プレイヤー数。
	LingerLongerPlayerCntMax = 6
	// LingerLongerDefaultPlayerCnt は既定のプレイヤー数。
	LingerLongerDefaultPlayerCnt = 4
)

// LingerLongerConfig はリンガーロンガーの設定。
type LingerLongerConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int `json:"p"`
}

// DefaultLingerLongerConfig は既定設定を返す。
func DefaultLingerLongerConfig() LingerLongerConfig {
	return LingerLongerConfig{PlayerCnt: LingerLongerDefaultPlayerCnt}
}

// Validate は設定を検証する。
func (c LingerLongerConfig) Validate() error {
	if c.PlayerCnt < LingerLongerPlayerCntMin || c.PlayerCnt > LingerLongerPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			LingerLongerPlayerCntMin, LingerLongerPlayerCntMax, c.PlayerCnt)
	}
	return nil
}
