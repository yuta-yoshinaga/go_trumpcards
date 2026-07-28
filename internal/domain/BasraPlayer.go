//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// BasraPlayer はバスラ (Basra) のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札の山、獲得したバスラ (スイープ) 数、
// および最終得点 (ゲーム終了時に確定) を持つ。
type BasraPlayer struct {
	*GamePlayer
	capturedCards []*Card
	basraCount    int // このゲームで達成したバスラ (単独札での場一掃) 数
	score         int // このゲームの最終得点 (ゲーム終了時に確定)
}

// NewBasraPlayer は BasraPlayer を構築する。
func NewBasraPlayer(isHuman bool) *BasraPlayer {
	return &BasraPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
		basraCount:    0,
		score:         0,
	}
}

// GetCapturedCards は捕獲した札を取得する。
func (p *BasraPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は捕獲した札の枚数を返す。
func (p *BasraPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を捕獲山に追加する。
func (p *BasraPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は捕獲山をクリアする (新規ゲームの先頭で呼ぶ)。
func (p *BasraPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// GetBasraCount はバスラ達成数を取得する。
func (p *BasraPlayer) GetBasraCount() int { return p.basraCount }

// IncrementBasra はバスラ達成数を 1 増やす。
func (p *BasraPlayer) IncrementBasra() { p.basraCount++ }

// ResetBasra はバスラ達成数を 0 に戻す (新規ゲーム開始時)。
func (p *BasraPlayer) ResetBasra() { p.basraCount = 0 }

// GetScore は最終得点を取得する。
func (p *BasraPlayer) GetScore() int { return p.score }

// SetScore は最終得点を設定する。
func (p *BasraPlayer) SetScore(s int) { p.score = s }

// ResetScore は最終得点を 0 に戻す (新規ゲーム開始時)。
func (p *BasraPlayer) ResetScore() { p.score = 0 }

// basraPlayerJSON is the JSON wire format for BasraPlayer.
type basraPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	BasraCount    int         `json:"bc"`
	Score         int         `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *BasraPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(basraPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		BasraCount:    p.basraCount,
		Score:         p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *BasraPlayer) UnmarshalJSON(data []byte) error {
	var j basraPlayerJSON
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
	p.basraCount = j.BasraCount
	p.score = j.Score
	return nil
}
