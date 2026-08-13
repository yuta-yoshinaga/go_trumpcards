//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// 卓の範囲設定。
const (
	// CincinnatiMinSeats は最小席数。
	CincinnatiMinSeats = 2
	// CincinnatiMaxSeats は最大席数。
	CincinnatiMaxSeats = 7
	// CincinnatiDefaultSeats は既定の席数。
	CincinnatiDefaultSeats = 4

	// CincinnatiMinChips は最小の初期チップ。
	CincinnatiMinChips = 100
	// CincinnatiMaxChips は最大の初期チップ。
	CincinnatiMaxChips = 100000
	// CincinnatiDefaultChips は既定の初期チップ。
	CincinnatiDefaultChips = 1000

	// CincinnatiMinAnte は最小のアンティ。
	CincinnatiMinAnte = 5
	// CincinnatiMaxAnte は最大のアンティ。
	CincinnatiMaxAnte = 100
	// CincinnatiDefaultAnte は既定のアンティ。
	CincinnatiDefaultAnte = 10
)

// 配りと進行の規則。
const (
	// CincinnatiHoleCards は各プレイヤーの手札枚数。**Holdem の 2 枚より多い。**
	CincinnatiHoleCards = 5
	// CincinnatiCommunityCards はコミュニティの枚数。
	CincinnatiCommunityCards = 5
	// CincinnatiHandSize は役を作る枚数。
	CincinnatiHandSize = 5
	// CincinnatiPoolSize は役を選ぶ母数 (手札 5 + コミュニティ 5)。
	//
	// **C(10,5) = 252 通りから最良を選ぶ。** 既存の `combinations` +
	// `evalFiveCardHand` がそのまま使えるので、探索は書き足していない。
	CincinnatiPoolSize = CincinnatiHoleCards + CincinnatiCommunityCards
	// CincinnatiBettingRounds はベットラウンドの回数。
	//
	// **1 枚ずつ 5 回めくる**のがこのゲームの形で、Holdem の 3-1-1 とは違う。
	CincinnatiBettingRounds = CincinnatiCommunityCards
)

// エラー値。設定検証で使う。
var (
	errCincinnatiSeatsRange = errors.New("cincinnati: seats out of range")
	errCincinnatiChipsRange = errors.New("cincinnati: initial chips out of range")
	errCincinnatiAnteRange  = errors.New("cincinnati: ante out of range")
	errCincinnatiDeckShort  = errors.New("cincinnati: the deck cannot serve this many seats")
)

// CincinnatiConfig はシンシナティの卓設定。
type CincinnatiConfig struct {
	Seats        int
	InitialChips int
	Ante         int
}

// DefaultCincinnatiConfig は既定の設定を返す。
func DefaultCincinnatiConfig() CincinnatiConfig {
	return CincinnatiConfig{
		Seats:        CincinnatiDefaultSeats,
		InitialChips: CincinnatiDefaultChips,
		Ante:         CincinnatiDefaultAnte,
	}
}

// Validate は設定が範囲内かを検査する。
//
// **席数は山の枚数でも縛られる。** 1 人 5 枚 + コミュニティ 5 枚なので、
// 7 人なら 40 枚要る。範囲だけ見て通すと、配っている途中で札が尽きて
// nil が手札に入り、点数計算がそこで静かに壊れる。
func (c CincinnatiConfig) Validate() error {
	switch {
	case c.Seats < CincinnatiMinSeats || c.Seats > CincinnatiMaxSeats:
		return errCincinnatiSeatsRange
	case c.InitialChips < CincinnatiMinChips || c.InitialChips > CincinnatiMaxChips:
		return errCincinnatiChipsRange
	case c.Ante < CincinnatiMinAnte || c.Ante > CincinnatiMaxAnte:
		return errCincinnatiAnteRange
	}
	if c.Seats*CincinnatiHoleCards+CincinnatiCommunityCards > 52 {
		return errCincinnatiDeckShort
	}
	return nil
}

// cincinnatiConfigJSON は CincinnatiConfig の JSON 表現。
type cincinnatiConfigJSON struct {
	Seats        int `json:"s"`
	InitialChips int `json:"c"`
	Ante         int `json:"a"`
}

// MarshalJSON implements json.Marshaler.
func (c CincinnatiConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(cincinnatiConfigJSON(c))
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **復元した設定も Validate に通す。** 席数を書き換えれば山が足りない卓を作れる。
func (c *CincinnatiConfig) UnmarshalJSON(data []byte) error {
	var j cincinnatiConfigJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	restored := CincinnatiConfig(j)
	if err := restored.Validate(); err != nil {
		return err
	}
	*c = restored
	return nil
}
