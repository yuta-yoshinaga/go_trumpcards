//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// JassPlayer ヤス(シーバー)プレイヤークラス
type JassPlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewJassPlayer コンストラクタ
func NewJassPlayer(isHuman bool, team int) *JassPlayer {
	return &JassPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *JassPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *JassPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// jassPlayerJSON is the JSON wire format for JassPlayer.
type jassPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *JassPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(jassPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *JassPlayer) UnmarshalJSON(data []byte) error {
	var j jassPlayerJSON
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
