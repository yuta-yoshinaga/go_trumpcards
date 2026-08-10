package domain

import "encoding/json"

// SpadesPlayer スペードプレイヤークラス
type SpadesPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid  int // 宣言したトリック数 (-1 = 未ビッド)
	bags int // 累積バッグ数 (オーバートリック)
}

// NewSpadesPlayer コンストラクタ
func NewSpadesPlayer(isHuman bool) *SpadesPlayer {
	return &SpadesPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド)
func (p *SpadesPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *SpadesPlayer) SetBid(bid int) { p.bid = bid }

// GetBags 累積バッグ数を取得
func (p *SpadesPlayer) GetBags() int { return p.bags }

// SetBags バッグ数を設定
func (p *SpadesPlayer) SetBags(bags int) { p.bags = bags }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *SpadesPlayer) ResetRound() {
	p.bid = -1
	resetRoundWithTricks(p)
}

// spadesPlayerJSON is the JSON wire format for SpadesPlayer.
type spadesPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
	Bags             int               `json:"bg"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpadesPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(spadesPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		Bags:             p.bags,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SpadesPlayer) UnmarshalJSON(data []byte) error {
	var j spadesPlayerJSON
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
	p.bags = j.Bags
	return nil
}
