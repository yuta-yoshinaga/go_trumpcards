//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
)

// エラー値。復元時の検証で使う。
var (
	errChemindeFerNegativeChips = errors.New("chemindefer: chips must not be negative")
	errChemindeFerNegativeBet   = errors.New("chemindefer: bet must not be negative")
)

// ChemindeFerPlayer はシュマン・ド・フェールの 1 席。
//
// **席は親にも子にもなる。** 誰が親かは卓の側 (ChemindeFer) が持っていて、
// 席自身は役割を覚えない。バンクが回るゲームなので、役割を席に持たせると
// ラウンドごとに 6 席ぶん書き換える羽目になるためである。
type ChemindeFerPlayer struct {
	chips   ChipHolder
	bet     int  // このラウンドで賭けている額 (子のときのみ)
	isHuman bool // 人間の席か
	name    string
}

// NewChemindeFerPlayer は ChemindeFerPlayer を構築する。
func NewChemindeFerPlayer(name string, chips int, isHuman bool) *ChemindeFerPlayer {
	p := &ChemindeFerPlayer{isHuman: isHuman, name: name}
	p.chips.SetChips(chips)
	return p
}

// GetName は席の表示名を返す。
func (p *ChemindeFerPlayer) GetName() string { return p.name }

// GetIsHuman は人間の席かを返す。
func (p *ChemindeFerPlayer) GetIsHuman() bool { return p.isHuman }

// GetChips は保有チップ数を返す。
func (p *ChemindeFerPlayer) GetChips() int { return p.chips.GetChips() }

// SetChips は保有チップ数を設定する。
func (p *ChemindeFerPlayer) SetChips(chips int) { p.chips.SetChips(chips) }

// AddChips はチップを加算する。
func (p *ChemindeFerPlayer) AddChips(amount int) { p.chips.AddChips(amount) }

// SubtractChips はチップを減算する。不足時は false を返し、額は変えない。
func (p *ChemindeFerPlayer) SubtractChips(amount int) bool {
	return p.chips.SubtractChips(amount)
}

// GetBet はこのラウンドで賭けている額を返す。
func (p *ChemindeFerPlayer) GetBet() int { return p.bet }

// SetBet はこのラウンドで賭けている額を設定する。
func (p *ChemindeFerPlayer) SetBet(bet int) { p.bet = bet }

// chemindeFerPlayerJSON is the JSON wire format for ChemindeFerPlayer.
type chemindeFerPlayerJSON struct {
	Chips   *ChipHolder `json:"ch"`
	Bet     int         `json:"bt"`
	IsHuman bool        `json:"hu"`
	Name    string      `json:"nm"`
}

// MarshalJSON implements json.Marshaler.
func (p *ChemindeFerPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(chemindeFerPlayerJSON{
		Chips:   &p.chips,
		Bet:     p.bet,
		IsHuman: p.isHuman,
		Name:    p.name,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **チップもベットも負にはならない。** ここを素通しすると、破損した保存データが
// そのまま「無から湧いたチップ」になり、卓の総額が変わってしまう。
func (p *ChemindeFerPlayer) UnmarshalJSON(data []byte) error {
	var j chemindeFerPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Chips != nil {
		p.chips = *j.Chips
	}
	if p.chips.GetChips() < 0 {
		return errChemindeFerNegativeChips
	}
	if j.Bet < 0 {
		return errChemindeFerNegativeBet
	}
	p.bet = j.Bet
	p.isHuman = j.IsHuman
	p.name = j.Name
	return nil
}
