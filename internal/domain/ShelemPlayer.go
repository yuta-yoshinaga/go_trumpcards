//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// ShelemPlayer シェレムのプレイヤー
type ShelemPlayer struct {
	*GamePlayer
	TrickHolder
	// bid は競りで出した点数 (-1: 未入札/パス)。
	bid int
	// passed は競りを降りたか。降りたら以降は入札できない。
	passed bool
	// declaredShelem は Shelem（全トリック独占）を宣言したか。
	declaredShelem bool
}

// NewShelemPlayer コンストラクタ
func NewShelemPlayer(isHuman bool) *ShelemPlayer {
	return &ShelemPlayer{GamePlayer: NewGamePlayer(isHuman), bid: -1}
}

// ResetGame ゲーム全体をリセットする
func (p *ShelemPlayer) ResetGame() { p.ResetRound() }

// ResetRound 1 ラウンド分の状態を初期化する
func (p *ShelemPlayer) ResetRound() {
	resetPlayerRound(p)
	p.bid = -1
	p.passed = false
	p.declaredShelem = false
}

// GetBid 競りで出した点数 (-1: 未入札/パス)
func (p *ShelemPlayer) GetBid() int { return p.bid }

// SetBid 入札を記録する
func (p *ShelemPlayer) SetBid(n int) { p.bid = n }

// GetPassed 競りを降りたか
func (p *ShelemPlayer) GetPassed() bool { return p.passed }

// SetPassed 競りを降りたことを記録する
func (p *ShelemPlayer) SetPassed(b bool) { p.passed = b }

// GetDeclaredShelem Shelem を宣言したか
func (p *ShelemPlayer) GetDeclaredShelem() bool { return p.declaredShelem }

// SetDeclaredShelem Shelem 宣言を記録する
func (p *ShelemPlayer) SetDeclaredShelem(b bool) { p.declaredShelem = b }

// shelemPlayerJSON is the JSON wire format for ShelemPlayer.
type shelemPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// 入札・降板・Shelem 宣言は往復させる。Worker はリクエストごとに KV から
	// 作り直すので、抜けると競りがやり直しになる (#4478)。
	Bid            int  `json:"bd"`
	Passed         bool `json:"ps"`
	DeclaredShelem bool `json:"sh"`
}

// MarshalJSON implements json.Marshaler.
func (p *ShelemPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(shelemPlayerJSON{
		GamePlayer:     p.GamePlayer,
		TrickHolder:    &p.TrickHolder,
		Bid:            p.bid,
		Passed:         p.passed,
		DeclaredShelem: p.declaredShelem,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ShelemPlayer) UnmarshalJSON(data []byte) error {
	var j shelemPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	p.passed = j.Passed
	p.declaredShelem = j.DeclaredShelem
	return nil
}
