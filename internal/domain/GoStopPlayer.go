//go:build !js || !wasm || extra

package domain

import "encoding/json"

// GoStopPlayer はゴーストップ (Go-Stop) のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札 (取り札)、累計得点、そのラウンドで
// ゴーを宣言した回数 (掛け金/倍率とバク判定に使う) と「直近に確定したカテゴリ点」
// (再決断検出の基準値) を持つ。
type GoStopPlayer struct {
	*GamePlayer
	capturedCards   []*Card
	score           int  // ゲームを通じた累計得点
	goCount         int  // このラウンドのゴー宣言回数
	calledGo        bool // このラウンドで一度でもゴーを宣言したか
	lastScorePoints int  // 直近のゴー/ストップ決断時点のカテゴリ点 (再決断検出の基準)
}

// NewGoStopPlayer は GoStopPlayer を構築する。
func NewGoStopPlayer(isHuman bool) *GoStopPlayer {
	return &GoStopPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
	}
}

// GetCapturedCards は取り札を取得する。
func (p *GoStopPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は取り札の枚数を返す。
func (p *GoStopPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を取り札に追加する。
func (p *GoStopPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は取り札をクリアする (新ラウンドの先頭で呼ぶ)。
func (p *GoStopPlayer) ResetCaptured() { p.capturedCards = make([]*Card, 0) }

// GetScore は累計得点を取得する。
func (p *GoStopPlayer) GetScore() int { return p.score }

// AddScore は累計得点を加算する。
func (p *GoStopPlayer) AddScore(n int) { p.score += n }

// ResetScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *GoStopPlayer) ResetScore() { p.score = 0 }

// GetGoCount はこのラウンドのゴー宣言回数を返す。
func (p *GoStopPlayer) GetGoCount() int { return p.goCount }

// IncGoCount はゴー宣言回数を 1 増やす。
func (p *GoStopPlayer) IncGoCount() { p.goCount++ }

// GetCalledGo はこのラウンドで一度でもゴーを宣言したかを返す。
func (p *GoStopPlayer) GetCalledGo() bool { return p.calledGo }

// SetCalledGo はゴー宣言フラグを設定する。
func (p *GoStopPlayer) SetCalledGo(v bool) { p.calledGo = v }

// GetLastScorePoints は再決断検出の基準となる直近確定カテゴリ点を返す。
func (p *GoStopPlayer) GetLastScorePoints() int { return p.lastScorePoints }

// SetLastScorePoints は再決断検出の基準値を設定する。
func (p *GoStopPlayer) SetLastScorePoints(n int) { p.lastScorePoints = n }

// ResetRound はラウンド開始時のプレイヤー状態リセット (取り札・ゴー・基準点)。
func (p *GoStopPlayer) ResetRound() {
	p.capturedCards = make([]*Card, 0)
	p.goCount = 0
	p.calledGo = false
	p.lastScorePoints = 0
}

// gostopPlayerJSON is the JSON wire format for GoStopPlayer.
type gostopPlayerJSON struct {
	GamePlayer      *GamePlayer `json:"gp"`
	CapturedCards   []*Card     `json:"cc"`
	Score           int         `json:"sc"`
	GoCount         int         `json:"gc"`
	CalledGo        bool        `json:"cg"`
	LastScorePoints int         `json:"ls"`
}

// MarshalJSON implements json.Marshaler.
func (p *GoStopPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(gostopPlayerJSON{
		GamePlayer:      p.GamePlayer,
		CapturedCards:   p.capturedCards,
		Score:           p.score,
		GoCount:         p.goCount,
		CalledGo:        p.calledGo,
		LastScorePoints: p.lastScorePoints,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *GoStopPlayer) UnmarshalJSON(data []byte) error {
	var j gostopPlayerJSON
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
	p.goCount = j.GoCount
	p.calledGo = j.CalledGo
	p.lastScorePoints = j.LastScorePoints
	return nil
}
