package domain

import "encoding/json"

// CallBreakPlayer Call Break プレイヤークラス
//
// roundScore / cumulativeScore は ×10 された整数で保持される。
// 例: 表示上の 4.1 点は内部値 41、-4 点は -40 として扱われる。
type CallBreakPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid int // 宣言したトリック数 (-1 = 未ビッド)
}

// NewCallBreakPlayer コンストラクタ
func NewCallBreakPlayer(isHuman bool) *CallBreakPlayer {
	return &CallBreakPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド)
func (p *CallBreakPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *CallBreakPlayer) SetBid(bid int) { p.bid = bid }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *CallBreakPlayer) ResetRound() {
	p.bid = -1
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// callBreakPlayerJSON is the JSON wire format for CallBreakPlayer.
type callBreakPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *CallBreakPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(callBreakPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CallBreakPlayer) UnmarshalJSON(data []byte) error {
	var j callBreakPlayerJSON
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
