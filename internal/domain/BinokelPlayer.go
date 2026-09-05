package domain

import "encoding/json"

// BinokelPlayer ビノクルプレイヤークラス
type BinokelPlayer struct {
	*GamePlayer
	TrickHolder
	bid         int // ビッド額 (0 = まだビッドしていない / パス済み)
	hasPassed   bool
	meldScore   int // メルドフェーズで宣言したメルド得点
	trickPoints int // トリックで獲得したカードポイント
}

// NewBinokelPlayer コンストラクタ
// team 引数は外側の層・既存テストとのシグネチャ互換性のため可変長で受け取るが、内部では個人戦のため使用しない。
func NewBinokelPlayer(isHuman bool, _ ...int) *BinokelPlayer {
	return &BinokelPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
}

// GetTeam チームインデックスを取得 (外側の層との後方互換性のためのスタブ)
func (p *BinokelPlayer) GetTeam() int { return 0 }

// GetBid ビッド額を取得
func (p *BinokelPlayer) GetBid() int { return p.bid }

// SetBid ビッド額を設定
func (p *BinokelPlayer) SetBid(bid int) { p.bid = bid }

// GetHasPassed パス済みかどうかを取得
func (p *BinokelPlayer) GetHasPassed() bool { return p.hasPassed }

// SetHasPassed パス済み状態を設定
func (p *BinokelPlayer) SetHasPassed(v bool) { p.hasPassed = v }

// GetMeldScore メルド得点を取得
func (p *BinokelPlayer) GetMeldScore() int { return p.meldScore }

// SetMeldScore メルド得点を設定
func (p *BinokelPlayer) SetMeldScore(score int) { p.meldScore = score }

// GetTrickPoints トリックポイントを取得
func (p *BinokelPlayer) GetTrickPoints() int { return p.trickPoints }

// SetTrickPoints トリックポイントを設定
func (p *BinokelPlayer) SetTrickPoints(pts int) { p.trickPoints = pts }

// AddTrickPoints トリックポイントを加算
func (p *BinokelPlayer) AddTrickPoints(pts int) { p.trickPoints += pts }

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）
func (p *BinokelPlayer) ResetRound() {
	p.ResetTricks()
	p.Reset()
	p.SetIsFinished(false)
	p.bid = 0
	p.hasPassed = false
	p.meldScore = 0
	p.trickPoints = 0
}

// binokelPlayerJSON is the JSON wire format for BinokelPlayer.
type binokelPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Team        int          `json:"tm,omitempty"`
	Bid         int          `json:"bd"`
	HasPassed   bool         `json:"hp"`
	MeldScore   int          `json:"ms"`
	TrickPoints int          `json:"tp"`
}

// MarshalJSON implements json.Marshaler.
func (p *BinokelPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(binokelPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Bid:         p.bid,
		HasPassed:   p.hasPassed,
		MeldScore:   p.meldScore,
		TrickPoints: p.trickPoints,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BinokelPlayer) UnmarshalJSON(data []byte) error {
	var j binokelPlayerJSON
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
	p.bid = j.Bid
	p.hasPassed = j.HasPassed
	p.meldScore = j.MeldScore
	p.trickPoints = j.TrickPoints
	return nil
}
