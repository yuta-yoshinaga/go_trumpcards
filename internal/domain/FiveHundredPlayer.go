package domain

import "encoding/json"

// FiveHundredPlayer 500 (Five Hundred) プレイヤークラス
type FiveHundredPlayer struct {
	*GamePlayer
	TrickHolder
	team       int             // チームインデックス (0 or 1)
	bid        *FiveHundredBid // このラウンドの最終ビッド (nil = 未ビッド)
	passed     bool            // このラウンドでパス済みか
	isDeclarer bool            // 落札者(宣言者)か
}

// NewFiveHundredPlayer コンストラクタ
func NewFiveHundredPlayer(isHuman bool, team int) *FiveHundredPlayer {
	return &FiveHundredPlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *FiveHundredPlayer) GetTeam() int { return p.team }

// GetBid このラウンドのビッドを取得 (nil = 未ビッド)
func (p *FiveHundredPlayer) GetBid() *FiveHundredBid { return p.bid }

// SetBid ビッドを設定
func (p *FiveHundredPlayer) SetBid(bid *FiveHundredBid) { p.bid = bid }

// GetPassed パス済みかどうか
func (p *FiveHundredPlayer) GetPassed() bool { return p.passed }

// SetPassed パス状態を設定
func (p *FiveHundredPlayer) SetPassed(v bool) { p.passed = v }

// GetIsDeclarer 落札者かどうか
func (p *FiveHundredPlayer) GetIsDeclarer() bool { return p.isDeclarer }

// SetIsDeclarer 落札者状態を設定
func (p *FiveHundredPlayer) SetIsDeclarer(v bool) { p.isDeclarer = v }

// ResetRound ラウンドをリセット (トリック・手札・ビッド状態を初期化)
func (p *FiveHundredPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.bid = nil
	p.passed = false
	p.isDeclarer = false
}

// fiveHundredPlayerJSON is the JSON wire format for FiveHundredPlayer.
type fiveHundredPlayerJSON struct {
	GamePlayer  *GamePlayer     `json:"gp"`
	TrickHolder *TrickHolder    `json:"th"`
	Team        int             `json:"tm"`
	Bid         *FiveHundredBid `json:"bd"`
	Passed      bool            `json:"ps"`
	IsDeclarer  bool            `json:"dc"`
}

// MarshalJSON implements json.Marshaler.
func (p *FiveHundredPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(fiveHundredPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
		Bid:         p.bid,
		Passed:      p.passed,
		IsDeclarer:  p.isDeclarer,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *FiveHundredPlayer) UnmarshalJSON(data []byte) error {
	var j fiveHundredPlayerJSON
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
