package domain

import "encoding/json"

// PitchPlayer ピッチプレイヤークラス
type PitchPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid int // 宣言値 (-1=未ビッド, 0=パス, 2-4=ビッド)
}

// NewPitchPlayer コンストラクタ
func NewPitchPlayer(isHuman bool) *PitchPlayer {
	return &PitchPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1=未ビッド, 0=パス, 2-4=ビッド)
func (p *PitchPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *PitchPlayer) SetBid(bid int) { p.bid = bid }

// ResetRound ラウンドをリセット (ビッド・トリック・手札・終了状態を初期化)
func (p *PitchPlayer) ResetRound() {
	p.bid = -1
	resetRoundWithTricks(p)
}

// pitchPlayerJSON is the JSON wire format for PitchPlayer.
type pitchPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *PitchPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pitchPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PitchPlayer) UnmarshalJSON(data []byte) error {
	var j pitchPlayerJSON
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
	p.bid = j.Bid
	return nil
}
