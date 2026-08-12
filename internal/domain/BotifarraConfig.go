//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"fmt"
)

// ボティファラの卓の大きさ (2 対 2 固定)。
const (
	// BotifarraPlayerCnt は参加人数。**2 対 2 のパートナー戦で固定**です。
	BotifarraPlayerCnt = 4
	// BotifarraHandSize は 1 人あたりの手札枚数 (48 / 4)。
	BotifarraHandSize = 12
	// BotifarraTrickCnt は 1 ラウンドのトリック数。
	BotifarraTrickCnt = BotifarraHandSize
)

// ボティファラの得点。
//
// **1 スート 15 点 × 4 = 60 点、それに各トリック 1 点 × 12 = 合計 72 点。**
// 36 点はちょうど半分なので、37 点以上取った側だけが得点します。
const (
	// BotifarraCardPoints は 1 ラウンドの札の点の合計。
	BotifarraCardPoints = 60
	// BotifarraTrickPoints は 1 ラウンドのトリックの点の合計 (1 トリック 1 点)。
	BotifarraTrickPoints = BotifarraTrickCnt
	// BotifarraTotalPoints は 1 ラウンドで動く点の合計。
	BotifarraTotalPoints = BotifarraCardPoints + BotifarraTrickPoints
	// BotifarraHalfPoints は折り返し。これを超えたぶんが得点になります。
	BotifarraHalfPoints = BotifarraTotalPoints / 2
)

// BotifarraTargetMin / Max / Default は上がり点の範囲。
const (
	BotifarraTargetMin     = 51
	BotifarraTargetMax     = 201
	BotifarraDefaultTarget = 101
)

// BotifarraConfig はボティファラのゲーム設定。
type BotifarraConfig struct {
	// TargetScore は上がり点。
	TargetScore int
	// AllowDoubling は倍付け (contrar / recontrar) を許すか。
	AllowDoubling bool
}

// DefaultBotifarraConfig はデフォルト設定を返す。
func DefaultBotifarraConfig() BotifarraConfig {
	return BotifarraConfig{TargetScore: BotifarraDefaultTarget, AllowDoubling: true}
}

// Validate は設定値の妥当性を検証する。
func (c BotifarraConfig) Validate() error {
	return ValidateRange("target score", c.TargetScore, BotifarraTargetMin, BotifarraTargetMax)
}

// botifarraConfigJSON is the JSON wire format for BotifarraConfig.
type botifarraConfigJSON struct {
	TargetScore   int  `json:"ts"`
	AllowDoubling bool `json:"ad"`
}

// MarshalJSON implements json.Marshaler.
func (c BotifarraConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(botifarraConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *BotifarraConfig) UnmarshalJSON(data []byte) error {
	var j botifarraConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	c.TargetScore = j.TargetScore
	c.AllowDoubling = j.AllowDoubling
	return c.Validate()
}

// BotifarraTeamOf は席 i のチームを返す (0 または 1)。
//
// **相方は向かい side。** 席 0 と 2 が組、席 1 と 3 が組です。
func BotifarraTeamOf(playerIdx int) int { return playerIdx % 2 }

// BotifarraPartnerOf は席 i の相方を返す。
func BotifarraPartnerOf(playerIdx int) int { return (playerIdx + 2) % BotifarraPlayerCnt }

// botifarraCardPoints は札の点。**マニラ (9) がいちばん高い 5 点**です。
var botifarraCardPoints = map[int]int{9: 5, 1: 4, 13: 3, 12: 2, 11: 1}

// BotifarraCardPoint は札の点を返す (点札でなければ 0)。
func BotifarraCardPoint(c *Card) int {
	if c == nil {
		return 0
	}
	return botifarraCardPoints[c.GetValue()]
}

// botifarraStrength は強さの序列。**マニラ (9) > アス (1) > 王 (13) > 騎士 (12) >
// ソタ (11) > 8 > 7 > … > 2** で、点札がそのまま上位 5 枚になります。
var botifarraStrength = map[int]int{9: 12, 1: 11, 13: 10, 12: 9, 11: 8, 8: 7, 7: 6, 6: 5, 5: 4, 4: 3, 3: 2, 2: 1}

// BotifarraRank は札の強さを返す (強いほど大きい)。
func BotifarraRank(c *Card) int {
	if c == nil {
		return 0
	}
	return botifarraStrength[c.GetValue()]
}

// botifarraValidateDeckPoints は点の合計が定数と合っていることを確かめる。
//
// **配り方や札の構成を変えたら必ずここで落ちます。** 60 点という定数は
// 「1 スート 15 点 × 4 スート」から来ていて、デッキが変わると黙って壊れます。
func botifarraValidateDeckPoints() error {
	perSuit := 0
	for _, p := range botifarraCardPoints {
		perSuit += p
	}
	if got := perSuit * 4; got != BotifarraCardPoints {
		return fmt.Errorf("botifarra: deck holds %d card points, want %d", got, BotifarraCardPoints)
	}
	return nil
}
