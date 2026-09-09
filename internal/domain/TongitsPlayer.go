//go:build !js || !wasm || extra5

package domain

import "encoding/json"

// TongitsPlayer Tongitsプレイヤークラス
type TongitsPlayer struct {
	*GamePlayer
	RoundScoreHolder
	melds [][]*Card
}

// NewTongitsPlayer コンストラクタ
func NewTongitsPlayer(isHuman bool) *TongitsPlayer {
	return &TongitsPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		melds:      make([][]*Card, 0),
	}
}

// ResetRound ラウンドをリセット（手札・スコア・終了状態を初期化）
func (p *TongitsPlayer) ResetRound() {
	resetRoundScored(p)
	p.melds = nil
}

// GetMelds returns the melds publicly laid by this player.
func (p *TongitsPlayer) GetMelds() [][]*Card { return p.melds }

// GetMeld returns one public meld, or nil when the index is invalid.
func (p *TongitsPlayer) GetMeld(idx int) []*Card {
	if idx < 0 || idx >= len(p.melds) {
		return nil
	}
	return p.melds[idx]
}

// ClearMelds removes all publicly laid melds.
func (p *TongitsPlayer) ClearMelds() { p.melds = nil }

// AppendMeld lays a meld on the table.
func (p *TongitsPlayer) AppendMeld(meld []*Card) { p.melds = append(p.melds, meld) }

// AddCardToMeld adds a card to an existing public meld.
func (p *TongitsPlayer) AddCardToMeld(idx int, card *Card) { p.melds[idx] = append(p.melds[idx], card) }

// tongitsPlayerJSON is the JSON wire format for TongitsPlayer.
type tongitsPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	Melds            [][]*Card         `json:"ml"`
}

// MarshalJSON implements json.Marshaler.
func (p *TongitsPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tongitsPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		Melds:            p.melds,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TongitsPlayer) UnmarshalJSON(data []byte) error {
	var j tongitsPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.RoundScoreHolder != nil {
		p.RoundScoreHolder = *j.RoundScoreHolder
	}
	p.melds = j.Melds
	if p.melds == nil {
		p.melds = make([][]*Card, 0)
	}
	return nil
}
