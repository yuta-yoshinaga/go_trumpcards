//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。
var errCrazyFourPokerNegativeChips = errors.New("crazyfourpoker: chips must not be negative")

// CrazyFourPokerPlayer はクレイジー 4 ポーカーのプレイヤー。
//
// 卓はプレイヤー 1 人対ディーラーなので、席という概念は無く、チップと今ラウンドの
// 賭け金だけを持つ。
type CrazyFourPokerPlayer struct {
	chips ChipHolder
}

// NewCrazyFourPokerPlayer は CrazyFourPokerPlayer を構築する。
func NewCrazyFourPokerPlayer(chips int) *CrazyFourPokerPlayer {
	p := &CrazyFourPokerPlayer{}
	p.chips.SetChips(chips)
	return p
}

// GetChips は保有チップ数を返す。
func (p *CrazyFourPokerPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *CrazyFourPokerPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *CrazyFourPokerPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返し、額は変えない。
func (p *CrazyFourPokerPlayer) SubtractChips(amount int) bool {
	return p.chips.SubtractChips(amount)
}

// crazyFourPokerPlayerJSON is the JSON wire format for CrazyFourPokerPlayer.
type crazyFourPokerPlayerJSON struct {
	Chips *ChipHolder `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *CrazyFourPokerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyFourPokerPlayerJSON{Chips: &p.chips})
}

// UnmarshalJSON implements json.Unmarshaler. **チップは負にならない。**
func (p *CrazyFourPokerPlayer) UnmarshalJSON(data []byte) error {
	var j crazyFourPokerPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips != nil {
		p.chips = *j.Chips
	}
	if p.chips.GetChips() < 0 {
		return errCrazyFourPokerNegativeChips
	}
	return nil
}
