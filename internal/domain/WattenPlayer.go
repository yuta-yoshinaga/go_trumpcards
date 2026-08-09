//go:build !js || !wasm || extra

package domain

import "encoding/json"

// WattenPlayer ヴァッテンプレイヤークラス
type WattenPlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewWattenPlayer コンストラクタ
func NewWattenPlayer(isHuman bool, team int) *WattenPlayer {
	return &WattenPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *WattenPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *WattenPlayer) ResetRound() {
	resetPlayerRound(p)
}

// wattenPlayerJSON is the JSON wire format for WattenPlayer.
type wattenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *WattenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(wattenPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *WattenPlayer) UnmarshalJSON(data []byte) error {
	var j wattenPlayerJSON
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
	if j.Team < 0 || j.Team >= WattenTeamCnt {
		return NewDomainError(ErrInvalidPlay, "チームが範囲外です")
	}
	p.team = j.Team
	return nil
}
