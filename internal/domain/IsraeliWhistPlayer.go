//go:build !js || !wasm || extra4

package domain

import "encoding/json"

// IsraeliWhistPlayer イスラエリホイストのプレイヤー
type IsraeliWhistPlayer struct {
	*GamePlayer
	TrickHolder
	// auctionBid は 1 段階目のオークションで出した最低ノルマ (-1: 未入札/パス)。
	auctionBid int
	// auctionSuit は 1 段階目で提示した切り札スート (0: なし)。
	auctionSuit int
	// passed は 1 段階目を降りたか。降りたら以降は入札できない。
	passed bool
	// bid は 2 段階目で宣言した目標トリック数 (-1: 未宣言)。
	bid int
	// roundScore は直前のラウンドの増減、totalScore は累計。
	roundScore int
	totalScore int
}

// NewIsraeliWhistPlayer コンストラクタ
func NewIsraeliWhistPlayer(isHuman bool) *IsraeliWhistPlayer {
	return &IsraeliWhistPlayer{GamePlayer: NewGamePlayer(isHuman), auctionBid: -1, bid: -1}
}

// ResetGame ゲーム全体をリセットする
func (p *IsraeliWhistPlayer) ResetGame() {
	p.ResetRound()
	p.totalScore = 0
}

// ResetRound 1 ラウンド分の状態を初期化する
func (p *IsraeliWhistPlayer) ResetRound() {
	resetPlayerRound(p)
	p.auctionBid = -1
	p.auctionSuit = 0
	p.passed = false
	p.bid = -1
	p.roundScore = 0
}

// GetAuctionBid オークションで出した最低ノルマ (-1: 未入札/パス)
func (p *IsraeliWhistPlayer) GetAuctionBid() int { return p.auctionBid }

// GetAuctionSuit オークションで提示した切り札スート (0: なし)
func (p *IsraeliWhistPlayer) GetAuctionSuit() int { return p.auctionSuit }

// SetAuction オークションの入札を記録する
func (p *IsraeliWhistPlayer) SetAuction(bid, suit int) { p.auctionBid, p.auctionSuit = bid, suit }

// GetPassed オークションを降りたか
func (p *IsraeliWhistPlayer) GetPassed() bool { return p.passed }

// SetPassed オークションを降りたことを記録する
func (p *IsraeliWhistPlayer) SetPassed(b bool) { p.passed = b }

// GetBid 2 段階目で宣言した目標トリック数 (-1: 未宣言)
func (p *IsraeliWhistPlayer) GetBid() int { return p.bid }

// SetBid 目標トリック数を設定する
func (p *IsraeliWhistPlayer) SetBid(n int) { p.bid = n }

// GetRoundScore 直前のラウンドの増減
func (p *IsraeliWhistPlayer) GetRoundScore() int { return p.roundScore }

// SetRoundScore 直前のラウンドの増減を設定する
func (p *IsraeliWhistPlayer) SetRoundScore(n int) { p.roundScore = n }

// GetTotalScore 累計得点
func (p *IsraeliWhistPlayer) GetTotalScore() int { return p.totalScore }

// AddTotalScore 累計得点に加算する
func (p *IsraeliWhistPlayer) AddTotalScore(n int) { p.totalScore += n }

// SetTotalScore 累計得点を設定する（復元・テスト用）
func (p *IsraeliWhistPlayer) SetTotalScore(n int) { p.totalScore = n }

// israeliWhistPlayerJSON is the JSON wire format for IsraeliWhistPlayer.
type israeliWhistPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// **2 段階ぶんの入札を両方とも往復させる。** Worker はリクエストごとに
	// KV から作り直すので、片方でも抜けると入札がやり直しになる (#4478)。
	AuctionBid  int  `json:"ab"`
	AuctionSuit int  `json:"as"`
	Passed      bool `json:"ps"`
	Bid         int  `json:"bd"`
	RoundScore  int  `json:"rs"`
	TotalScore  int  `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *IsraeliWhistPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(israeliWhistPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		AuctionBid:  p.auctionBid,
		AuctionSuit: p.auctionSuit,
		Passed:      p.passed,
		Bid:         p.bid,
		RoundScore:  p.roundScore,
		TotalScore:  p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *IsraeliWhistPlayer) UnmarshalJSON(data []byte) error {
	var j israeliWhistPlayerJSON
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
	p.auctionBid = j.AuctionBid
	p.auctionSuit = j.AuctionSuit
	p.passed = j.Passed
	p.bid = j.Bid
	p.roundScore = j.RoundScore
	p.totalScore = j.TotalScore
	return nil
}
