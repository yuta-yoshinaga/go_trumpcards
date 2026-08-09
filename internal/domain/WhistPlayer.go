package domain

import "encoding/json"

// WhistPlayerCnt ホイストプレイヤー数
const WhistPlayerCnt = 4

// WhistTeamCnt チーム数
const WhistTeamCnt = 2

// WhistPlayer ホイストプレイヤークラス
type WhistPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewWhistPlayer コンストラクタ
func NewWhistPlayer(isHuman bool, team int) *WhistPlayer {
	return &WhistPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *WhistPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *WhistPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// whistPlayerJSON is the JSON wire format for WhistPlayer.
type whistPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Team             int               `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *WhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(whistPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Team:             p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *WhistPlayer) UnmarshalJSON(data []byte) error {
	var j whistPlayerJSON
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
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.team = j.Team
	return nil
}
