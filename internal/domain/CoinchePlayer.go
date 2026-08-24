//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// CoinchePlayer コワンシュプレイヤークラス
type CoinchePlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewCoinchePlayer コンストラクタ
func NewCoinchePlayer(isHuman bool, team int) *CoinchePlayer {
	return &CoinchePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *CoinchePlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *CoinchePlayer) ResetRound() {
	resetPlayerRound(p)
}

// coinchePlayerJSON is the JSON wire format for CoinchePlayer.
type coinchePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *CoinchePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(coinchePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CoinchePlayer) UnmarshalJSON(data []byte) error {
	var j coinchePlayerJSON
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
