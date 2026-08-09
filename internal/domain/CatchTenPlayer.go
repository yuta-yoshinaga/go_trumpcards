package domain

import "encoding/json"

// CatchTenPlayerCnt Catch the Ten プレイヤー数
const CatchTenPlayerCnt = 4

// CatchTenTeamCnt チーム数
const CatchTenTeamCnt = 2

// CatchTenPlayer Catch the Ten プレイヤークラス
type CatchTenPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	team int // チームインデックス (0 or 1)
}

// NewCatchTenPlayer コンストラクタ
func NewCatchTenPlayer(isHuman bool, team int) *CatchTenPlayer {
	return &CatchTenPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *CatchTenPlayer) GetTeam() int { return p.team }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *CatchTenPlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// catchTenPlayerJSON is the JSON wire format for CatchTenPlayer.
type catchTenPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Team             int               `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *CatchTenPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(catchTenPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Team:             p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CatchTenPlayer) UnmarshalJSON(data []byte) error {
	var j catchTenPlayerJSON
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
