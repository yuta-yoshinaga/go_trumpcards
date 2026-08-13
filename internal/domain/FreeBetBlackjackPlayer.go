//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// errFreeBetNegativeChips は復元時にチップが負だったときのエラー。
var errFreeBetNegativeChips = errors.New("freebet: chips must not be negative")

// FreeBetBlackjackPlayer はフリーベット・ブラックジャックのプレイヤー。
type FreeBetBlackjackPlayer struct {
	chips ChipHolder
}

// NewFreeBetBlackjackPlayer は FreeBetBlackjackPlayer を構築する。
func NewFreeBetBlackjackPlayer(chips int) *FreeBetBlackjackPlayer {
	p := &FreeBetBlackjackPlayer{}
	p.chips.SetChips(chips)
	return p
}

// GetChips は保有チップ数を返す。
func (p *FreeBetBlackjackPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *FreeBetBlackjackPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *FreeBetBlackjackPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返す。
func (p *FreeBetBlackjackPlayer) SubtractChips(amount int) bool {
	return p.chips.SubtractChips(amount)
}

// freeBetPlayerJSON is the JSON wire format for FreeBetBlackjackPlayer.
type freeBetPlayerJSON struct {
	Chips *ChipHolder `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *FreeBetBlackjackPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(freeBetPlayerJSON{Chips: &p.chips})
}

// UnmarshalJSON implements json.Unmarshaler. **チップは負にならない。**
func (p *FreeBetBlackjackPlayer) UnmarshalJSON(data []byte) error {
	var j freeBetPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips != nil {
		p.chips = *j.Chips
	}
	if p.chips.GetChips() < 0 {
		return errFreeBetNegativeChips
	}
	return nil
}
