//go:build !js || !wasm || classic

package domain

import "encoding/json"

// ScopaPlayer スコパのプレイヤー。
// 基底の GamePlayer (手札) に加えて、獲得した捕獲札、スコパ回数、累積得点を持つ。
type ScopaPlayer struct {
	*GamePlayer
	capturedCards []*Card
	scopaCount    int
	totalScore    int // 複数ラウンド越しの累計得点
}

// NewScopaPlayer constructs a ScopaPlayer.
func NewScopaPlayer(isHuman bool) *ScopaPlayer {
	return &ScopaPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
		scopaCount:    0,
		totalScore:    0,
	}
}

// GetCapturedCards 獲得した捕獲札を取得。
func (p *ScopaPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount 獲得した捕獲札の枚数。
func (p *ScopaPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured カードを捕獲札に追加。
func (p *ScopaPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured 捕獲札をクリア (新ラウンドの先頭で呼ぶ)。
func (p *ScopaPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// GetScopaCount スコパ回数を取得。
func (p *ScopaPlayer) GetScopaCount() int { return p.scopaCount }

// IncrementScopa スコパ回数を +1。
func (p *ScopaPlayer) IncrementScopa() { p.scopaCount++ }

// ResetScopaCount スコパ回数を 0 にする (新ラウンドの先頭)。
func (p *ScopaPlayer) ResetScopaCount() { p.scopaCount = 0 }

// GetTotalScore 累計得点を取得 (複数ラウンド通算)。
func (p *ScopaPlayer) GetTotalScore() int { return p.totalScore }

// AddScore 得点を加算。
func (p *ScopaPlayer) AddScore(n int) { p.totalScore += n }

// ResetTotalScore 累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *ScopaPlayer) ResetTotalScore() { p.totalScore = 0 }

// scopaPlayerJSON is the JSON wire format for ScopaPlayer.
type scopaPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	ScopaCount    int         `json:"sc"`
	TotalScore    int         `json:"ts"`
}

// MarshalJSON implements json.Marshaler.
func (p *ScopaPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(scopaPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		ScopaCount:    p.scopaCount,
		TotalScore:    p.totalScore,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ScopaPlayer) UnmarshalJSON(data []byte) error {
	var j scopaPlayerJSON
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
	p.scopaCount = j.ScopaCount
	p.totalScore = j.TotalScore
	return nil
}
