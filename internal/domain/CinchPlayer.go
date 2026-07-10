//go:build !js || !wasm || extra

package domain

import "encoding/json"

// CinchPlayer はチンチ (Cinch) のプレイヤー。基底の GamePlayer (手札) に加えて、現ディールで
// 獲得したトリック (TrickHolder)、宣言値 (bid)、全ディール通算の累計得点を持つ。
type CinchPlayer struct {
	*GamePlayer
	*TrickHolder
	bid        int // 宣言値 (-1=未ビッド, 0=パス, 1..CinchMaxBid=ビッド)
	totalScore int // 全ディール通算の累計得点
}

// NewCinchPlayer constructs a CinchPlayer.
func NewCinchPlayer(isHuman bool) *CinchPlayer {
	return &CinchPlayer{
		GamePlayer:  NewGamePlayer(isHuman),
		TrickHolder: &TrickHolder{},
		bid:         -1,
		totalScore:  0,
	}
}

// GetBid はビッド値を返す (-1=未ビッド, 0=パス, 1..=ビッド)。
func (p *CinchPlayer) GetBid() int { return p.bid }

// SetBid はビッド値を設定する。
func (p *CinchPlayer) SetBid(bid int) { p.bid = bid }

// GetTotalScore は累計得点を返す (全ディール通算)。
func (p *CinchPlayer) GetTotalScore() int { return p.totalScore }

// AddScore は得点を加算する (負の値で減点)。
func (p *CinchPlayer) AddScore(n int) { p.totalScore += n }

// ResetTotalScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *CinchPlayer) ResetTotalScore() { p.totalScore = 0 }

// ResetDeal は 1 ディール分の状態 (手札・トリック・ビッド) をクリアする。
// 累計得点は維持する。
func (p *CinchPlayer) ResetDeal() {
	p.Reset()
	p.ResetTricks()
	p.SetIsFinished(false)
	p.bid = -1
}

// cinchPlayerJSON is the JSON wire format for CinchPlayer.
type cinchPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Bid         int          `json:"bd"`
	TotalScore  int          `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *CinchPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cinchPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: p.TrickHolder,
		Bid:         p.bid,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CinchPlayer) UnmarshalJSON(data []byte) error {
	var j cinchPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = j.TrickHolder
	} else {
		p.TrickHolder = &TrickHolder{}
	}
	p.bid = j.Bid
	p.totalScore = j.TotalScore
	return nil
}
