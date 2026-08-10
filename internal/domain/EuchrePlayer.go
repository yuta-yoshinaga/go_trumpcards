//go:build !js || !wasm || solo

package domain

import "encoding/json"

// EuchrePlayer ユーカープレイヤークラス
type EuchrePlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewEuchrePlayer コンストラクタ
func NewEuchrePlayer(isHuman bool, team int) *EuchrePlayer {
	return &EuchrePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *EuchrePlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *EuchrePlayer) ResetRound() {
	resetPlayerRound(p)
}

// euchrePlayerJSON is the JSON wire format for EuchrePlayer.
type euchrePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *EuchrePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(euchrePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *EuchrePlayer) UnmarshalJSON(data []byte) error {
	var j euchrePlayerJSON
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
