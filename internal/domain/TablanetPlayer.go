//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// TablanetPlayer はタブラネット (Tablanet) のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札の山、獲得したタブラ (スイープ) 数、
// および最終得点 (ゲーム終了時に確定) を持つ。
type TablanetPlayer struct {
	*GamePlayer
	capturedCards []*Card
	tablaCount    int // このゲームで達成したタブラ (単独札での場一掃) 数
	score         int // このゲームの最終得点 (ゲーム終了時に確定)
}

// NewTablanetPlayer は TablanetPlayer を構築する。
func NewTablanetPlayer(isHuman bool) *TablanetPlayer {
	return &TablanetPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
		tablaCount:    0,
		score:         0,
	}
}

// GetCapturedCards は捕獲した札を取得する。
func (p *TablanetPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は捕獲した札の枚数を返す。
func (p *TablanetPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を捕獲山に追加する。
func (p *TablanetPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は捕獲山をクリアする (新規ゲームの先頭で呼ぶ)。
func (p *TablanetPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// GetTablaCount はタブラ達成数を取得する。
func (p *TablanetPlayer) GetTablaCount() int { return p.tablaCount }

// IncrementTabla はタブラ達成数を 1 増やす。
func (p *TablanetPlayer) IncrementTabla() { p.tablaCount++ }

// ResetTabla はタブラ達成数を 0 に戻す (新規ゲーム開始時)。
func (p *TablanetPlayer) ResetTabla() { p.tablaCount = 0 }

// GetScore は最終得点を取得する。
func (p *TablanetPlayer) GetScore() int { return p.score }

// SetScore は最終得点を設定する。
func (p *TablanetPlayer) SetScore(s int) { p.score = s }

// ResetScore は最終得点を 0 に戻す (新規ゲーム開始時)。
func (p *TablanetPlayer) ResetScore() { p.score = 0 }

// tablanetPlayerJSON is the JSON wire format for TablanetPlayer.
type tablanetPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	TablaCount    int         `json:"bc"`
	Score         int         `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *TablanetPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tablanetPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		TablaCount:    p.tablaCount,
		Score:         p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TablanetPlayer) UnmarshalJSON(data []byte) error {
	var j tablanetPlayerJSON
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
	p.tablaCount = j.TablaCount
	p.score = j.Score
	return nil
}
