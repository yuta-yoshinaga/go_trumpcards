//go:build !js || !wasm || extra5

package domain

import "encoding/json"

// OmiPlayer オミプレイヤークラス
type OmiPlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewOmiPlayer コンストラクタ
func NewOmiPlayer(isHuman bool, team int) *OmiPlayer {
	return &OmiPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *OmiPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *OmiPlayer) ResetRound() {
	resetPlayerRound(p)
}

// omiPlayerJSON is the JSON wire format for OmiPlayer.
type omiPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *OmiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(omiPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *OmiPlayer) UnmarshalJSON(data []byte) error {
	var j omiPlayerJSON
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
