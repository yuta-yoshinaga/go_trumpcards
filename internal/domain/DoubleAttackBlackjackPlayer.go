//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// errDoubleAttackNegativeChips は復元時にチップが負だったときのエラー。
var errDoubleAttackNegativeChips = errors.New("doubleattack: chips must not be negative")

// DoubleAttackBlackjackPlayer は追加ベット・ブラックジャックのプレイヤー。
//
// 卓はプレイヤー 1 人対ディーラーなので、席という概念は無くチップだけを持つ。
// 手札はスプリットで増えるため卓 (DoubleAttackBlackjack) 側が持つ。
type DoubleAttackBlackjackPlayer struct {
	chips ChipHolder
}

// NewDoubleAttackBlackjackPlayer は DoubleAttackBlackjackPlayer を構築する。
func NewDoubleAttackBlackjackPlayer(chips int) *DoubleAttackBlackjackPlayer {
	p := &DoubleAttackBlackjackPlayer{}
	p.chips.SetChips(chips)
	return p
}

// GetChips は保有チップ数を返す。
func (p *DoubleAttackBlackjackPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *DoubleAttackBlackjackPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *DoubleAttackBlackjackPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返し、額は変えない。
func (p *DoubleAttackBlackjackPlayer) SubtractChips(amount int) bool {
	return p.chips.SubtractChips(amount)
}

// doubleAttackPlayerJSON is the JSON wire format for DoubleAttackBlackjackPlayer.
type doubleAttackPlayerJSON struct {
	Chips *ChipHolder `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *DoubleAttackBlackjackPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubleAttackPlayerJSON{Chips: &p.chips})
}

// UnmarshalJSON implements json.Unmarshaler. **チップは負にならない。**
func (p *DoubleAttackBlackjackPlayer) UnmarshalJSON(data []byte) error {
	var j doubleAttackPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips != nil {
		p.chips = *j.Chips
	}
	if p.chips.GetChips() < 0 {
		return errDoubleAttackNegativeChips
	}
	return nil
}
