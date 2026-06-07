//go:build !js || !wasm || solo

package domain

import "encoding/json"

// BidWhistPlayer Bid Whist プレイヤークラス
type BidWhistPlayer struct {
	*GamePlayer
	TrickHolder
	team       int          // チームインデックス (0 or 1)
	bid        *BidWhistBid // このラウンドの最終ビッド (nil = 未ビッド/パス)
	passed     bool         // このラウンドでパス済みか
	isDeclarer bool         // 落札者(宣言者)か
}

// NewBidWhistPlayer コンストラクタ
func NewBidWhistPlayer(isHuman bool, team int) *BidWhistPlayer {
	return &BidWhistPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *BidWhistPlayer) GetTeam() int { return p.team }

// GetBid このラウンドのビッドを取得 (nil = 未ビッド/パス)
func (p *BidWhistPlayer) GetBid() *BidWhistBid { return p.bid }

// SetBid ビッドを設定
func (p *BidWhistPlayer) SetBid(bid *BidWhistBid) { p.bid = bid }

// GetPassed パス済みかどうか
func (p *BidWhistPlayer) GetPassed() bool { return p.passed }

// SetPassed パス状態を設定
func (p *BidWhistPlayer) SetPassed(v bool) { p.passed = v }

// GetIsDeclarer 落札者かどうか
func (p *BidWhistPlayer) GetIsDeclarer() bool { return p.isDeclarer }

// SetIsDeclarer 落札者状態を設定
func (p *BidWhistPlayer) SetIsDeclarer(v bool) { p.isDeclarer = v }

// ResetRound ラウンドをリセット (トリック・手札・ビッド状態を初期化)
func (p *BidWhistPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.bid = nil
	p.passed = false
	p.isDeclarer = false
}

// bidWhistPlayerJSON is the JSON wire format for BidWhistPlayer.
type bidWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
	Bid         *BidWhistBid `json:"bd"`
	Passed      bool         `json:"ps"`
	IsDeclarer  bool         `json:"dc"`
}

// MarshalJSON implements json.Marshaler.
func (p *BidWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(bidWhistPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
		Bid:         p.bid,
		Passed:      p.passed,
		IsDeclarer:  p.isDeclarer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BidWhistPlayer) UnmarshalJSON(data []byte) error {
	var j bidWhistPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.team = j.Team
	p.bid = j.Bid
	p.passed = j.Passed
	p.isDeclarer = j.IsDeclarer
	return nil
}
