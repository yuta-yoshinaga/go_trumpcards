package domain

import "encoding/json"

// SpadesPlayer スペードプレイヤークラス
type SpadesPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid            int // 宣言したトリック数 (-1 = 未ビッド)
	bags           int // 累積バッグ数 (オーバートリック)
	scoreBreakdown SpadesScoreBreakdown
}

// SpadesScoreBreakdown はラウンド得点を構成する実際の加減点。
type SpadesScoreBreakdown struct {
	BidScore       int `json:"bidScore"`
	OvertrickScore int `json:"overtrickScore"`
	NilScore       int `json:"nilScore"`
	BagPenalty     int `json:"bagPenalty"`
}

// Total は内訳の合計を返す。
func (b SpadesScoreBreakdown) Total() int {
	return b.BidScore + b.OvertrickScore + b.NilScore - b.BagPenalty
}

// GetScoreBreakdown はラウンド得点の内訳を返す。
func (p *SpadesPlayer) GetScoreBreakdown() SpadesScoreBreakdown { return p.scoreBreakdown }

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
	p.scoreBreakdown = SpadesScoreBreakdown{}
	resetRoundWithTricks(p)
}

// spadesPlayerJSON is the JSON wire format for SpadesPlayer.
type spadesPlayerJSON struct {
	GamePlayer       *GamePlayer          `json:"gp"`
	RoundScoreHolder *RoundScoreHolder    `json:"rh"`
	TrickHolder      *TrickHolder         `json:"th"`
	Bid              int                  `json:"bd"`
	Bags             int                  `json:"bg"`
	ScoreBreakdown   SpadesScoreBreakdown `json:"sb"`
}

// MarshalJSON implements json.Marshaler.
func (p *SpadesPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(spadesPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		Bags:             p.bags,
		ScoreBreakdown:   p.scoreBreakdown,
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
	p.scoreBreakdown = j.ScoreBreakdown
	return nil
}
