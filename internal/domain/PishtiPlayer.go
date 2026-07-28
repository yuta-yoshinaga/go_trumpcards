//go:build !js || !wasm || extra2

package domain

import "encoding/json"

// PishtiPlayer は Pişti のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札の山と、累積した Pişti ボーナスを持つ。
type PishtiPlayer struct {
	*GamePlayer
	capturedCards []*Card
	pistiBonus    int // このゲームで獲得した Pişti ボーナス点の累計
}

// NewPishtiPlayer は PishtiPlayer を構築する。
func NewPishtiPlayer(isHuman bool) *PishtiPlayer {
	return &PishtiPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
		pistiBonus:    0,
	}
}

// GetCapturedCards は捕獲した札を取得する。
func (p *PishtiPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は捕獲した札の枚数を返す。
func (p *PishtiPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を捕獲山に追加する。
func (p *PishtiPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は捕獲山をクリアする (新規ゲームの先頭で呼ぶ)。
func (p *PishtiPlayer) ResetCaptured() {
	p.capturedCards = make([]*Card, 0)
}

// GetPistiBonus は Pişti ボーナスの累計を取得する。
func (p *PishtiPlayer) GetPistiBonus() int { return p.pistiBonus }

// AddPistiBonus は Pişti ボーナスを加算する。
func (p *PishtiPlayer) AddPistiBonus(n int) { p.pistiBonus += n }

// ResetPistiBonus は Pişti ボーナスを 0 に戻す (新規ゲーム開始時)。
func (p *PishtiPlayer) ResetPistiBonus() { p.pistiBonus = 0 }

// pishtiPlayerJSON is the JSON wire format for PishtiPlayer.
type pishtiPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	PistiBonus    int         `json:"pb"`
}

// MarshalJSON implements json.Marshaler.
func (p *PishtiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pishtiPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		PistiBonus:    p.pistiBonus,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PishtiPlayer) UnmarshalJSON(data []byte) error {
	var j pishtiPlayerJSON
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
	p.pistiBonus = j.PistiBonus
	return nil
}
