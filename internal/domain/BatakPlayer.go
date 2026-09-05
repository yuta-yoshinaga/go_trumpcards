package domain

import "encoding/json"

// BatakPlayer Batak プレイヤークラス
//
// スコアは素の整数で保持される。
// bid の意味: -1 = 未発言 / 0 = パス (BatakPassBid) / 5..13 = 宣言
type BatakPlayer struct {
	*GamePlayer
	RoundScoreHolder
	TrickHolder
	bid int // 宣言したトリック数 (-1 = 未ビッド/未発言, 0 = パス)
}

// NewBatakPlayer コンストラクタ
func NewBatakPlayer(isHuman bool) *BatakPlayer {
	return &BatakPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		bid:        -1,
	}
}

// GetBid ビッド取得 (-1 = 未ビッド, 0 = パス)
func (p *BatakPlayer) GetBid() int { return p.bid }

// SetBid ビッド設定
func (p *BatakPlayer) SetBid(bid int) { p.bid = bid }

// ResetRound ラウンドをリセット（ビッド・トリック・手札・終了状態を初期化）
func (p *BatakPlayer) ResetRound() {
	p.bid = -1
	resetRoundWithTricks(p)
}

// batakPlayerJSON is the JSON wire format for BatakPlayer.
type batakPlayerJSON struct {
	GamePlayer       *GamePlayer       `json:"gp"`
	RoundScoreHolder *RoundScoreHolder `json:"rh"`
	TrickHolder      *TrickHolder      `json:"th"`
	Bid              int               `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *BatakPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(batakPlayerJSON{
		GamePlayer:       p.GamePlayer,
		RoundScoreHolder: &p.RoundScoreHolder,
		TrickHolder:      &p.TrickHolder,
		Bid:              p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BatakPlayer) UnmarshalJSON(data []byte) error {
	var j batakPlayerJSON
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
