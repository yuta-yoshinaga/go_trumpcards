package domain

import "encoding/json"

// NinetyNinePlayer ナインティナインプレイヤークラス
type NinetyNinePlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid    int     // 宣言したトリック数 (-1 = 未宣言)。埋めた3枚のスート合計で決まる (0-9)
	buried []*Card // 伏せた3枚 (このディール中は場外)
}

// NewNinetyNinePlayer コンストラクタ
func NewNinetyNinePlayer(isHuman bool) *NinetyNinePlayer {
	return &NinetyNinePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未宣言)
func (p *NinetyNinePlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *NinetyNinePlayer) SetBid(bid int) { p.bid = bid }

// GetBuried 伏せたカードを取得
func (p *NinetyNinePlayer) GetBuried() []*Card { return p.buried }

// SetBuried 伏せたカードを設定
func (p *NinetyNinePlayer) SetBuried(cards []*Card) { p.buried = cards }

// ResetRound ラウンドをリセット（宣言・トリック・伏せ札・手札・終了状態を初期化）
func (p *NinetyNinePlayer) ResetRound() {
	p.bid = -1
	p.buried = nil
	p.SetRoundScore(0)
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
}

// ninetyNinePlayerJSON is the JSON wire format for NinetyNinePlayer.
type ninetyNinePlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
	Buried           []*Card           `json:"bu"`
}

// MarshalJSON implements json.Marshaler.
func (p *NinetyNinePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(ninetyNinePlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
		Buried:           p.buried,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *NinetyNinePlayer) UnmarshalJSON(data []byte) error {
	var j ninetyNinePlayerJSON
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
	p.buried = j.Buried
	return nil
}
