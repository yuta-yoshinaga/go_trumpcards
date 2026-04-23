package domain

import "encoding/json"

// CassinoPlayer カシノのプレイヤー。
// 基底の GamePlayer (手札) に加えて、獲得した捕獲札、スイープ数、累積得点を持つ。
type CassinoPlayer struct {
	*GamePlayer
	capturedCards []*Card
	sweepCount    int
	totalScore    int // 複数ラウンド越しの累計得点
}

// NewCassinoPlayer constructs a CassinoPlayer.
func NewCassinoPlayer(isHuman bool) *CassinoPlayer {
	return &CassinoPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
		sweepCount:    0,
		totalScore:    0,
	}
}

// GetCapturedCards 獲得した捕獲札を取得。
func (p *CassinoPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount 獲得した捕獲札の枚数。
func (p *CassinoPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured カードを捕獲札に追加。
func (p *CassinoPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured 捕獲札をクリア (新ラウンドの先頭で呼ぶ)。
func (p *CassinoPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// GetSweepCount スイープ回数を取得。
func (p *CassinoPlayer) GetSweepCount() int { return p.sweepCount }

// IncrementSweep スイープ回数を +1。
func (p *CassinoPlayer) IncrementSweep() { p.sweepCount++ }

// ResetSweepCount スイープ回数を 0 にする (新ラウンドの先頭)。
func (p *CassinoPlayer) ResetSweepCount() { p.sweepCount = 0 }

// GetTotalScore 累計得点を取得 (複数ラウンド通算)。
func (p *CassinoPlayer) GetTotalScore() int { return p.totalScore }

// AddScore 得点を加算。
func (p *CassinoPlayer) AddScore(n int) { p.totalScore += n }

// ResetTotalScore 累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *CassinoPlayer) ResetTotalScore() { p.totalScore = 0 }

// cassinoPlayerJSON is the JSON wire format for CassinoPlayer.
type cassinoPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	SweepCount    int         `json:"sw"`
	TotalScore    int         `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *CassinoPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(cassinoPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		SweepCount:    p.sweepCount,
		TotalScore:    p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *CassinoPlayer) UnmarshalJSON(data []byte) error {
	var j cassinoPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.capturedCards = j.CapturedCards
	if p.capturedCards == nil {
		p.capturedCards = make([]*Card, 0)
	}
	p.sweepCount = j.SweepCount
	p.totalScore = j.TotalScore
	return nil
}
