package domain

import "encoding/json"

// PinochlePlayer ピノクルプレイヤークラス
type PinochlePlayer struct {
	*GamePlayer
	TrickHolder
	team        int // チームインデックス (0 or 1)
	bid         int // ビッド額 (0 = まだビッドしていない / パス済み)
	hasPassed   bool
	meldScore   int // メルドフェーズで宣言したメルド得点
	trickPoints int // トリックで獲得したカードポイント
}

// NewPinochlePlayer コンストラクタ
func NewPinochlePlayer(isHuman bool, team int) *PinochlePlayer {
	return &PinochlePlayer{
		GamePlayer: NewGamePlayer(isHuman),
		team:       team,
	}
}

// GetTeam チームインデックスを取得
func (p *PinochlePlayer) GetTeam() int { return p.team }

// GetBid ビッド額を取得
func (p *PinochlePlayer) GetBid() int { return p.bid }

// SetBid ビッド額を設定
func (p *PinochlePlayer) SetBid(bid int) { p.bid = bid }

// GetHasPassed パス済みかどうかを取得
func (p *PinochlePlayer) GetHasPassed() bool { return p.hasPassed }

// SetHasPassed パス済み状態を設定
func (p *PinochlePlayer) SetHasPassed(v bool) { p.hasPassed = v }

// GetMeldScore メルド得点を取得
func (p *PinochlePlayer) GetMeldScore() int { return p.meldScore }

// SetMeldScore メルド得点を設定
func (p *PinochlePlayer) SetMeldScore(score int) { p.meldScore = score }

// GetTrickPoints トリックポイントを取得
func (p *PinochlePlayer) GetTrickPoints() int { return p.trickPoints }

// SetTrickPoints トリックポイントを設定
func (p *PinochlePlayer) SetTrickPoints(pts int) { p.trickPoints = pts }

// AddTrickPoints トリックポイントを加算
func (p *PinochlePlayer) AddTrickPoints(pts int) { p.trickPoints += pts }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *PinochlePlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.bid = 0
	p.hasPassed = false
	p.meldScore = 0
	p.trickPoints = 0
}

// pinochlePlayerJSON is the JSON wire format for PinochlePlayer.
type pinochlePlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm"`
	Bid         int          `json:"bd"`
	HasPassed   bool         `json:"hp"`
	MeldScore   int          `json:"ms"`
	TrickPoints int          `json:"tp"`
}

// MarshalJSON implements json.Marshaler.
func (p *PinochlePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pinochlePlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Team:        p.team,
		Bid:         p.bid,
		HasPassed:   p.hasPassed,
		MeldScore:   p.meldScore,
		TrickPoints: p.trickPoints,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PinochlePlayer) UnmarshalJSON(data []byte) error {
	var j pinochlePlayerJSON
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
	p.hasPassed = j.HasPassed
	p.meldScore = j.MeldScore
	p.trickPoints = j.TrickPoints
	return nil
}
