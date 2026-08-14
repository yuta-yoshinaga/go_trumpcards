//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
)

const (
	// StealingBundlesPlayerCntMin は最小プレイヤー数。
	StealingBundlesPlayerCntMin = 2
	// StealingBundlesPlayerCntMax は最大プレイヤー数。
	StealingBundlesPlayerCntMax = 4
	// StealingBundlesDefaultPlayerCnt は既定のプレイヤー数。
	StealingBundlesDefaultPlayerCnt = 4
)

// StealingBundlesHandSize は 1 回の配布で各自が受け取る枚数。
const StealingBundlesHandSize = 4

// StealingBundlesTableSize は最初に場へ公開する枚数。
const StealingBundlesTableSize = 4

// StealingBundlesDeckSize は使用するデッキ枚数 (標準 52 枚)。
const StealingBundlesDeckSize = 52

// StealingBundlesConfig はスティーリングバンドルのゲーム設定。
type StealingBundlesConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int
}

// DefaultStealingBundlesConfig はデフォルト設定を返す。
func DefaultStealingBundlesConfig() StealingBundlesConfig {
	return StealingBundlesConfig{PlayerCnt: StealingBundlesDefaultPlayerCnt}
}

// Validate は設定値の妥当性を検証する。
func (c StealingBundlesConfig) Validate() error {
	if c.PlayerCnt < StealingBundlesPlayerCntMin || c.PlayerCnt > StealingBundlesPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			StealingBundlesPlayerCntMin, StealingBundlesPlayerCntMax, c.PlayerCnt)
	}
	return nil
}

// stealingBundlesConfigJSON is the JSON wire format for StealingBundlesConfig.
type stealingBundlesConfigJSON struct {
	PlayerCnt int `json:"p"`
}

// MarshalJSON implements json.Marshaler.
func (c StealingBundlesConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(stealingBundlesConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *StealingBundlesConfig) UnmarshalJSON(data []byte) error {
	var j stealingBundlesConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.PlayerCnt = j.PlayerCnt
	return c.Validate()
}
