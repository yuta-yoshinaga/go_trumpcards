//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BelotePlayer ベロートプレイヤークラス
type BelotePlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewBelotePlayer コンストラクタ
func NewBelotePlayer(isHuman bool, team int) *BelotePlayer {
	return &BelotePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *BelotePlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *BelotePlayer) ResetRound() {
	resetPlayerRound(p)
}

// belotePlayerJSON is the JSON wire format for BelotePlayer.
type belotePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *BelotePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(belotePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BelotePlayer) UnmarshalJSON(data []byte) error {
	var j belotePlayerJSON
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
	p.team = j.Team
	return nil
}
