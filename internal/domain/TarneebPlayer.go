//go:build !js || !wasm || casino

package domain

import "encoding/json"

// TarneebPlayerCnt Tarneeb プレイヤー数
const TarneebPlayerCnt = 4

// TarneebTeamCnt チーム数
const TarneebTeamCnt = 2

// TarneebPlayer Tarneeb プレイヤークラス
//
// 0+2 と 1+3 がパートナーとなる (向かい合わせのパートナーシップ)。
// bid は -1=未ビッド / 0=パス / 7-13=有効ビッド の整数値。
type TarneebPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	team int // 0 or 1
	bid  int // -1 = 未ビッド
}

// NewTarneebPlayer コンストラクタ
func NewTarneebPlayer(isHuman bool, team int) *TarneebPlayer {
	return &TarneebPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
		bid:        -1,
	}
}

// GetTeam チームインデックス取得
func (p *TarneebPlayer) GetTeam() int { return p.team }

// GetBid ビッド取得 (-1 = 未ビッド, 0 = パス)
func (p *TarneebPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *TarneebPlayer) SetBid(bid int) { p.bid = bid }

// ResetRound ラウンド単位の状態をリセット
func (p *TarneebPlayer) ResetRound() {
	p.bid = -1
	resetRoundWithTricks(p)
}

// tarneebPlayerJSON is the JSON wire format for TarneebPlayer.
type tarneebPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Team             int               `json:"tm"`
	Bid              int               `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *TarneebPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tarneebPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Team:             p.team,
		Bid:              p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TarneebPlayer) UnmarshalJSON(data []byte) error {
	var j tarneebPlayerJSON
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
	p.bid = j.Bid
	return nil
}
