//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
)

const (
	// GoofspielPlayerCntMin は最小プレイヤー数。
	GoofspielPlayerCntMin = 2
	// GoofspielPlayerCntMax は最大プレイヤー数。
	GoofspielPlayerCntMax = 3
	// GoofspielDefaultPlayerCnt は既定のプレイヤー数。
	GoofspielDefaultPlayerCnt = 2
)

// GoofspielRounds は 1 ゲームのラウンド数 (賞札の枚数)。
const GoofspielRounds = 13

// GoofspielTieRule は同点のときの賞札の扱い。
type GoofspielTieRule int

// ゴフスピールの同点処理
const (
	// GoofspielTieDiscard 同点なら賞札は消滅する
	GoofspielTieDiscard GoofspielTieRule = 0
	// GoofspielTieCarryOver 同点なら賞札を次のラウンドへ持ち越す
	GoofspielTieCarryOver GoofspielTieRule = 1
)

// GoofspielConfig はゴフスピールのゲーム設定。
type GoofspielConfig struct {
	// PlayerCnt は参加人数。
	PlayerCnt int
	// TieRule は同点のときの賞札の扱い。
	TieRule GoofspielTieRule
}

// DefaultGoofspielConfig はデフォルト設定を返す。
func DefaultGoofspielConfig() GoofspielConfig {
	return GoofspielConfig{PlayerCnt: GoofspielDefaultPlayerCnt, TieRule: GoofspielTieDiscard}
}

// Validate は設定値の妥当性を検証する。
func (c GoofspielConfig) Validate() error {
	if c.PlayerCnt < GoofspielPlayerCntMin || c.PlayerCnt > GoofspielPlayerCntMax {
		return fmt.Errorf("player count must be between %d and %d, got %d",
			GoofspielPlayerCntMin, GoofspielPlayerCntMax, c.PlayerCnt)
	}
	return ValidateRange("tie rule", int(c.TieRule), int(GoofspielTieDiscard), int(GoofspielTieCarryOver))
}

// GoofspielBidSuit は席 i が入札に使うスートを返す。
//
// **賞札にダイヤ、入札に残りのスートを配ります。** issue は「標準52枚を 3 スートに
// 分割」と書いていますが、**3 人卓は賞札 1 + 入札 3 = 4 スート要ります**。
// ちょうど標準デッキに収まるだけで、3 スートでは足りません。
func GoofspielBidSuit(playerIdx int) int {
	suits := []int{CardDesignSpade, CardDesignClover, CardDesignHeart}
	if playerIdx < 0 || playerIdx >= len(suits) {
		return CardDesignSpade
	}
	return suits[playerIdx]
}

// GoofspielPrizeSuit は賞札のスートを返す。
func GoofspielPrizeSuit() int { return CardDesignDiamond }

// goofspielConfigJSON is the JSON wire format for GoofspielConfig.
type goofspielConfigJSON struct {
	PlayerCnt int `json:"p"`
	TieRule   int `json:"tr"`
}

// MarshalJSON implements json.Marshaler.
func (c GoofspielConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(goofspielConfigJSON{PlayerCnt: c.PlayerCnt, TieRule: int(c.TieRule)})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *GoofspielConfig) UnmarshalJSON(data []byte) error {
	var j goofspielConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.PlayerCnt = j.PlayerCnt
	c.TieRule = GoofspielTieRule(j.TieRule)
	return c.Validate()
}
