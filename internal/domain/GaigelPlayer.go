//go:build !js || !wasm || extra

package domain

import "encoding/json"

// GaigelPlayer ガイゲルプレイヤークラス
type GaigelPlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewGaigelPlayer コンストラクタ
func NewGaigelPlayer(isHuman bool, team int) *GaigelPlayer {
	return &GaigelPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *GaigelPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *GaigelPlayer) ResetRound() {
	resetPlayerRound(p)
}

// gaigelPlayerJSON is the JSON wire format for GaigelPlayer.
type gaigelPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *GaigelPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gaigelPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GaigelPlayer) UnmarshalJSON(data []byte) error {
	var j gaigelPlayerJSON
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
	if j.Team < 0 || j.Team >= GaigelTeamCnt {
		return NewDomainError(ErrInvalidPlay, "チーム番号が範囲外です")
	}
	p.team = j.Team
	return nil
}
