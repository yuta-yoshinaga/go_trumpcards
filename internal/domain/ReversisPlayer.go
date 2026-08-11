//go:build !js || !wasm || classic

package domain

import "encoding/json"

// ReversisPlayer レヴェルシのプレイヤー
type ReversisPlayer struct {
	*GamePlayer
	TrickHolder
	// chips は持ちチップ。**多いほど良い。** 勝敗はこれで決まる。
	chips int
	// roundPenalty はこのラウンドで取ったカードの失点。少ないほど良い。
	roundPenalty int
	// tookQuinola / tookDiamondAce は印付きの 2 枚を取ったか。
	tookQuinola    bool
	tookDiamondAce bool
}

// NewReversisPlayer コンストラクタ
func NewReversisPlayer(isHuman bool) *ReversisPlayer {
	return &ReversisPlayer{GamePlayer: NewGamePlayer(isHuman), chips: ReversisStartingChips}
}

// ResetGame ゲーム全体をリセットする（チップも初期値に戻す）
func (p *ReversisPlayer) ResetGame() {
	p.ResetRound()
	p.chips = ReversisStartingChips
}

// ResetRound 1 ラウンド分の状態だけを初期化する。チップは残す。
func (p *ReversisPlayer) ResetRound() {
	resetPlayerRound(p)
	p.roundPenalty = 0
	p.tookQuinola = false
	p.tookDiamondAce = false
}

// GetChips 持ちチップを返す
func (p *ReversisPlayer) GetChips() int { return p.chips }

// AddChips チップを増減する（支払いは負の値）
func (p *ReversisPlayer) AddChips(n int) { p.chips += n }

// SetChips チップを設定する（復元・テスト用）
func (p *ReversisPlayer) SetChips(n int) { p.chips = n }

// GetRoundPenalty このラウンドの失点
func (p *ReversisPlayer) GetRoundPenalty() int { return p.roundPenalty }

// AddRoundPenalty このラウンドの失点に加算する
func (p *ReversisPlayer) AddRoundPenalty(n int) { p.roundPenalty += n }

// SetRoundPenalty このラウンドの失点を設定する（テスト用）
func (p *ReversisPlayer) SetRoundPenalty(n int) { p.roundPenalty = n }

// GetTookQuinola キノラ（♥J）を取ったか
func (p *ReversisPlayer) GetTookQuinola() bool { return p.tookQuinola }

// SetTookQuinola キノラ獲得を設定する
func (p *ReversisPlayer) SetTookQuinola(b bool) { p.tookQuinola = b }

// GetTookDiamondAce ダイヤの A を取ったか
func (p *ReversisPlayer) GetTookDiamondAce() bool { return p.tookDiamondAce }

// SetTookDiamondAce ダイヤの A 獲得を設定する
func (p *ReversisPlayer) SetTookDiamondAce(b bool) { p.tookDiamondAce = b }

// reversisPlayerJSON is the JSON wire format for ReversisPlayer.
type reversisPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	// チップと失点は必ず往復させる。Worker はリクエストごとに KV から作り直すので、
	// 抜けると持ちチップがラウンド途中で初期値に戻る (#4478)。
	Chips          int  `json:"ch"`
	RoundPenalty   int  `json:"rp"`
	TookQuinola    bool `json:"tq"`
	TookDiamondAce bool `json:"td"`
}

// MarshalJSON implements json.Marshaler.
func (p *ReversisPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(reversisPlayerJSON{
		GamePlayer:     p.GamePlayer,
		TrickHolder:    &p.TrickHolder,
		Chips:          p.chips,
		RoundPenalty:   p.roundPenalty,
		TookQuinola:    p.tookQuinola,
		TookDiamondAce: p.tookDiamondAce,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *ReversisPlayer) UnmarshalJSON(data []byte) error {
	var j reversisPlayerJSON
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
	p.chips = j.Chips
	p.roundPenalty = j.RoundPenalty
	p.tookQuinola = j.TookQuinola
	p.tookDiamondAce = j.TookDiamondAce
	return nil
}
