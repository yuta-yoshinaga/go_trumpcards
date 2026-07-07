//go:build !js || !wasm || extra

package domain

import "encoding/json"

// HachiHachiPlayer は八八 (Hachi-Hachi) のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札 (取り札)、累計得点 (88 精算の符号付き
// 差分の累積) と、直近ラウンドの得点差分を保持する。
type HachiHachiPlayer struct {
	*GamePlayer
	capturedCards []*Card
	score         int // ゲームを通じた累計得点 (各ラウンドの符号付き差分の合計)
	roundDelta    int // 直近ラウンドで加算された符号付き差分 (UI 表示補助)
}

// NewHachiHachiPlayer は HachiHachiPlayer を構築する。
func NewHachiHachiPlayer(isHuman bool) *HachiHachiPlayer {
	return &HachiHachiPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
	}
}

// GetCapturedCards は取り札を取得する。
func (p *HachiHachiPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は取り札の枚数を返す。
func (p *HachiHachiPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を取り札に追加する。
func (p *HachiHachiPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は取り札をクリアする (新ラウンドの先頭で呼ぶ)。
func (p *HachiHachiPlayer) ResetCaptured() { p.capturedCards = make([]*Card, 0) }

// GetScore は累計得点を取得する。
func (p *HachiHachiPlayer) GetScore() int { return p.score }

// AddScore は累計得点を加算する (符号付き)。
func (p *HachiHachiPlayer) AddScore(n int) { p.score += n }

// ResetScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *HachiHachiPlayer) ResetScore() { p.score = 0 }

// GetRoundDelta は直近ラウンドの符号付き差分を返す。
func (p *HachiHachiPlayer) GetRoundDelta() int { return p.roundDelta }

// SetRoundDelta は直近ラウンドの符号付き差分を設定する。
func (p *HachiHachiPlayer) SetRoundDelta(n int) { p.roundDelta = n }

// ResetRound はラウンド開始時のプレイヤー状態リセット (取り札・差分)。
func (p *HachiHachiPlayer) ResetRound() {
	p.capturedCards = make([]*Card, 0)
	p.roundDelta = 0
}

// hachihachiPlayerJSON is the JSON wire format for HachiHachiPlayer.
type hachihachiPlayerJSON struct {
	GamePlayer    *GamePlayer `json:"gp"`
	CapturedCards []*Card     `json:"cc"`
	Score         int         `json:"sc"`
	RoundDelta    int         `json:"rd"`
}

// MarshalJSON implements json.Marshaler.
func (p *HachiHachiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(hachihachiPlayerJSON{
		GamePlayer:    p.GamePlayer,
		CapturedCards: p.capturedCards,
		Score:         p.score,
		RoundDelta:    p.roundDelta,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *HachiHachiPlayer) UnmarshalJSON(data []byte) error {
	var j hachihachiPlayerJSON
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
	p.score = j.Score
	p.roundDelta = j.RoundDelta
	return nil
}
