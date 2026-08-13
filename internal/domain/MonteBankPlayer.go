//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var errMonteBankNegativeChips = errors.New("montebank: chips must not be negative")

// MonteBankPlayer はモンテバンクのプレイヤー。
//
// **胴元は席ではなく卓そのもの。** バンクが回るゲームではないので、席は 1 つ
// しか無く、持っているのはチップだけである。
type MonteBankPlayer struct {
	chips ChipHolder
}

// NewMonteBankPlayer は MonteBankPlayer を構築する。
func NewMonteBankPlayer(chips int) *MonteBankPlayer {
	p := new(MonteBankPlayer)
	p.chips.SetChips(chips)
	return p
}

// GetChips は保有チップ数を返す。
func (p *MonteBankPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *MonteBankPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *MonteBankPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返し、額は変えない。
func (p *MonteBankPlayer) SubtractChips(amount int) bool { return p.chips.SubtractChips(amount) }

// monteBankPlayerJSON は MonteBankPlayer の JSON 表現。
type monteBankPlayerJSON struct {
	Chips int `json:"c"`
}

// MarshalJSON implements json.Marshaler.
func (p *MonteBankPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(monteBankPlayerJSON{Chips: p.chips.GetChips()})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **負の残高を弾く。** 減算経路は `SubtractChips` が守っているだけなので、
// 保存を書き換えれば負の値を持ち込める。
func (p *MonteBankPlayer) UnmarshalJSON(data []byte) error {
	var j monteBankPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errMonteBankNegativeChips
	}
	p.chips.SetChips(j.Chips)
	return nil
}
