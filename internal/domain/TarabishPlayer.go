//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// TarabishPlayer タラビッシュのプレイヤー
type TarabishPlayer struct {
	*GamePlayer
	TrickHolder
	// meldPoints はこのラウンドで申告したメルドの合計点。
	meldPoints int
	// hasBella は切り札の K+Q を持っているか（ベラ）。
	hasBella bool
	// runLen は申告したランの最長枚数（0 = ラン無し）。表示用。
	runLen int
}

// NewTarabishPlayer コンストラクタ
func NewTarabishPlayer(isHuman bool) *TarabishPlayer {
	return &TarabishPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame ゲーム全体をリセットする
func (p *TarabishPlayer) ResetGame() { p.ResetRound() }

// ResetRound 1 ラウンド分の状態を初期化する
func (p *TarabishPlayer) ResetRound() {
	resetPlayerRound(p)
	p.meldPoints = 0
	p.hasBella = false
	p.runLen = 0
}

// GetMeldPoints このラウンドのメルド点
func (p *TarabishPlayer) GetMeldPoints() int { return p.meldPoints }

// SetMeldPoints メルド点を設定する
func (p *TarabishPlayer) SetMeldPoints(n int) { p.meldPoints = n }

// GetHasBella ベラ（切り札の K+Q）を持っているか
func (p *TarabishPlayer) GetHasBella() bool { return p.hasBella }

// SetHasBella ベラの有無を設定する
func (p *TarabishPlayer) SetHasBella(b bool) { p.hasBella = b }

// GetRunLen 申告したランの最長枚数
func (p *TarabishPlayer) GetRunLen() int { return p.runLen }

// SetRunLen ランの最長枚数を設定する
func (p *TarabishPlayer) SetRunLen(n int) { p.runLen = n }

// tarabishPlayerJSON is the JSON wire format for TarabishPlayer.
type tarabishPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// メルドは配り直後に一度だけ計算するので、往復しないと
	// Worker では毎リクエスト 0 点になる (#4478)。
	MeldPoints int  `json:"mp"`
	HasBella   bool `json:"hb"`
	RunLen     int  `json:"rl"`
}

// MarshalJSON implements json.Marshaler.
func (p *TarabishPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(tarabishPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		MeldPoints:  p.meldPoints,
		HasBella:    p.hasBella,
		RunLen:      p.runLen,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TarabishPlayer) UnmarshalJSON(data []byte) error {
	var j tarabishPlayerJSON
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
	p.meldPoints = j.MeldPoints
	p.hasBella = j.HasBella
	p.runLen = j.RunLen
	return nil
}
