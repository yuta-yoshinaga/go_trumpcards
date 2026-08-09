//go:build !js || !wasm || casino

package domain

import "encoding/json"

// CourtPiecePlayerCnt Court Piece プレイヤー数
const CourtPiecePlayerCnt = 4

// CourtPieceTeamCnt チーム数
const CourtPieceTeamCnt = 2

// CourtPiecePlayer Court Piece プレイヤークラス
//
// 0+2 と 1+3 がパートナーとなる (向かい合わせのパートナーシップ)。
type CourtPiecePlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	team int // 0 or 1
}

// NewCourtPiecePlayer コンストラクタ
func NewCourtPiecePlayer(isHuman bool, team int) *CourtPiecePlayer {
	return &CourtPiecePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックス取得
func (p *CourtPiecePlayer) GetTeam() int { return p.team }

// ResetRound ラウンド単位の状態をリセット
func (p *CourtPiecePlayer) ResetRound() {
	resetRoundWithTricks(p)
}

// courtPiecePlayerJSON is the JSON wire format for CourtPiecePlayer.
type courtPiecePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Team             int               `json:"tm"`
}

// MarshalJSON implements json.Marshaler.
func (p *CourtPiecePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(courtPiecePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Team:             p.team,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CourtPiecePlayer) UnmarshalJSON(data []byte) error {
	var j courtPiecePlayerJSON
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
	if j.Team < 0 || j.Team >= CourtPieceTeamCnt {
		return errCourtPieceInvalidState
	}
	p.team = j.Team
	return nil
}
