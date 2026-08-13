//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errBanLuckNegativeChips = errors.New("banluck: chips must not be negative")
	errBanLuckNegativeBet   = errors.New("banluck: bet must not be negative")
)

// BanLuckPlayer はバンラックの 1 席。
//
// **席は親にも子にもなる。** 誰が親かは卓 (BanLuck) が持っていて、席自身は
// 役割を覚えない。親が回るゲームなので、役割を席に持たせるとラウンドごとに
// 全席を書き換える羽目になり、しかも「親が 2 人いる」状態を作れてしまう。
type BanLuckPlayer struct {
	chips   ChipHolder
	bet     int  // このラウンドで賭けている額 (子のときのみ)
	isHuman bool // 人間の席か
	name    string
}

// NewBanLuckPlayer は BanLuckPlayer を構築する。
func NewBanLuckPlayer(name string, chips int, isHuman bool) *BanLuckPlayer {
	p := &BanLuckPlayer{isHuman: isHuman, name: name}
	p.chips.SetChips(chips)
	return p
}

// GetName は席の表示名を返す。
func (p *BanLuckPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *BanLuckPlayer) GetIsHuman() bool { return p.isHuman }

// GetChips は保有チップ数を返す。
func (p *BanLuckPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *BanLuckPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *BanLuckPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返し、額は変えない。
func (p *BanLuckPlayer) SubtractChips(amount int) bool { return p.chips.SubtractChips(amount) }

// GetBet はこのラウンドで賭けている額を返す。
func (p *BanLuckPlayer) GetBet() int { return p.bet }

// SetBet はこのラウンドで賭けている額を設定する。
func (p *BanLuckPlayer) SetBet(bet int) { p.bet = bet }

// banLuckPlayerJSON は BanLuckPlayer の JSON 表現。
type banLuckPlayerJSON struct {
	Chips   int    `json:"c"`
	Bet     int    `json:"b"`
	IsHuman bool   `json:"h"`
	Name    string `json:"n"`
}

// MarshalJSON implements json.Marshaler.
func (p *BanLuckPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(banLuckPlayerJSON{
		Chips: p.chips.GetChips(), Bet: p.bet, IsHuman: p.isHuman, Name: p.name,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **負の値を弾く。** チップも賭け金も減算経路が `SubtractChips` で守られて
// いるだけなので、保存を書き換えれば負の残高を持ち込める。
func (p *BanLuckPlayer) UnmarshalJSON(data []byte) error {
	var j banLuckPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips < 0 {
		return errBanLuckNegativeChips
	}
	if j.Bet < 0 {
		return errBanLuckNegativeBet
	}
	p.chips.SetChips(j.Chips)
	p.bet = j.Bet
	p.isHuman = j.IsHuman
	p.name = j.Name
	return nil
}
