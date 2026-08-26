//go:build !js || !wasm || extra

package domain

import "encoding/json"

// BauernschnapsenPlayer バウエルンシュナプセンプレイヤークラス
type BauernschnapsenPlayer struct {
	*GamePlayer
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewBauernschnapsenPlayer コンストラクタ
func NewBauernschnapsenPlayer(isHuman bool, team int) *BauernschnapsenPlayer {
	return &BauernschnapsenPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *BauernschnapsenPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *BauernschnapsenPlayer) ResetRound() {
	resetPlayerRound(p)
}

// bauernschnapsenPlayerJSON is the JSON wire format for BauernschnapsenPlayer.
type bauernschnapsenPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *BauernschnapsenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bauernschnapsenPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BauernschnapsenPlayer) UnmarshalJSON(data []byte) error {
	var j bauernschnapsenPlayerJSON
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
	if j.Team < 0 || j.Team >= BauernschnapsenTeamCnt {
		return NewDomainError(ErrInvalidPlay, "チーム番号が範囲外です")
	}
	p.team = j.Team
	return nil
}
