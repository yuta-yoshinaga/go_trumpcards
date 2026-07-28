//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// KoiKoiPlayer はこいこい (Koi-Koi) のプレイヤー。
// 基底の GamePlayer (手札) に加えて、捕獲した札 (取り札)、累計得点、そのラウンドで
// こいこいを宣言したかのフラグ、および「直近に確定した役の点数」(新役検出の基準値) を持つ。
type KoiKoiPlayer struct {
	*GamePlayer
	capturedCards  []*Card
	score          int  // ゲームを通じた累計得点
	calledKoiKoi   bool // このラウンドでこいこいを宣言したか
	lastYakuPoints int  // 直近のこいこい決断時点で確定した役の合計点 (新役検出の基準)
}

// NewKoiKoiPlayer は KoiKoiPlayer を構築する。
func NewKoiKoiPlayer(isHuman bool) *KoiKoiPlayer {
	return &KoiKoiPlayer{
		GamePlayer:    NewGamePlayer(isHuman),
		capturedCards: make([]*Card, 0),
	}
}

// GetCapturedCards は取り札を取得する。
func (p *KoiKoiPlayer) GetCapturedCards() []*Card { return p.capturedCards }

// CapturedCount は取り札の枚数を返す。
func (p *KoiKoiPlayer) CapturedCount() int { return len(p.capturedCards) }

// AddCaptured は札を取り札に追加する。
func (p *KoiKoiPlayer) AddCaptured(cards []*Card) {
	p.capturedCards = append(p.capturedCards, cards...)
}

// ResetCaptured は取り札をクリアする (新ラウンドの先頭で呼ぶ)。
func (p *KoiKoiPlayer) ResetCaptured() { p.capturedCards = make([]*Card, 0) }

// GetScore は累計得点を取得する。
func (p *KoiKoiPlayer) GetScore() int { return p.score }

// AddScore は累計得点を加算する。
func (p *KoiKoiPlayer) AddScore(n int) { p.score += n }

// ResetScore は累計得点を 0 に戻す (新規ゲーム開始時)。
func (p *KoiKoiPlayer) ResetScore() { p.score = 0 }

// GetCalledKoiKoi はこのラウンドでこいこいを宣言したかを返す。
func (p *KoiKoiPlayer) GetCalledKoiKoi() bool { return p.calledKoiKoi }

// SetCalledKoiKoi はこいこい宣言フラグを設定する。
func (p *KoiKoiPlayer) SetCalledKoiKoi(v bool) { p.calledKoiKoi = v }

// GetLastYakuPoints は新役検出の基準となる直近確定役点を返す。
func (p *KoiKoiPlayer) GetLastYakuPoints() int { return p.lastYakuPoints }

// SetLastYakuPoints は新役検出の基準値を設定する。
func (p *KoiKoiPlayer) SetLastYakuPoints(n int) { p.lastYakuPoints = n }

// ResetRound はラウンド開始時のプレイヤー状態リセット (取り札・こいこい・基準役点)。
func (p *KoiKoiPlayer) ResetRound() {
	p.capturedCards = make([]*Card, 0)
	p.calledKoiKoi = false
	p.lastYakuPoints = 0
}

// koikoiPlayerJSON is the JSON wire format for KoiKoiPlayer.
type koikoiPlayerJSON struct {
	GamePlayer     *GamePlayer `json:"gp"`
	CapturedCards  []*Card     `json:"cc"`
	Score          int         `json:"sc"`
	CalledKoiKoi   bool        `json:"kk"`
	LastYakuPoints int         `json:"ly"`
}

// MarshalJSON implements json.Marshaler.
func (p *KoiKoiPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(koikoiPlayerJSON{
		GamePlayer:     p.GamePlayer,
		CapturedCards:  p.capturedCards,
		Score:          p.score,
		CalledKoiKoi:   p.calledKoiKoi,
		LastYakuPoints: p.lastYakuPoints,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *KoiKoiPlayer) UnmarshalJSON(data []byte) error {
	var j koikoiPlayerJSON
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
	p.calledKoiKoi = j.CalledKoiKoi
	p.lastYakuPoints = j.LastYakuPoints
	return nil
}
