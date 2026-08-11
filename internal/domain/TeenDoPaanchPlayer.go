//go:build !js || !wasm || solo

package domain

import "encoding/json"

// TeenDoPaanchPlayer は 3-2-5 のプレイヤー。
type TeenDoPaanchPlayer struct {
	*GamePlayer
	TrickHolder
	// target はこのラウンドのノルマ (3 / 2 / 5)。**宣言ではなく割り当て。**
	target int
	// met はノルマを達成したラウンド数。**勝敗はこれで決まる。**
	met int
}

// NewTeenDoPaanchPlayer はコンストラクタ。
func NewTeenDoPaanchPlayer(isHuman bool) *TeenDoPaanchPlayer {
	return &TeenDoPaanchPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *TeenDoPaanchPlayer) ResetGame() {
	p.ResetRound()
	p.met = 0
}

// ResetRound は 1 ラウンド分の状態を初期化する。
func (p *TeenDoPaanchPlayer) ResetRound() {
	resetPlayerRound(p)
	p.target = 0
}

// GetTarget はこのラウンドのノルマを返す。
func (p *TeenDoPaanchPlayer) GetTarget() int { return p.target }

// SetTarget はノルマを設定する。
func (p *TeenDoPaanchPlayer) SetTarget(n int) { p.target = n }

// GetMet はノルマを達成したラウンド数を返す。
func (p *TeenDoPaanchPlayer) GetMet() int { return p.met }

// AddMet は達成数を 1 増やす。
func (p *TeenDoPaanchPlayer) AddMet() { p.met++ }

// SetMet は達成数を設定する（復元・テスト用）。
func (p *TeenDoPaanchPlayer) SetMet(n int) { p.met = n }

// teenDoPaanchPlayerJSON is the JSON wire format for TeenDoPaanchPlayer.
//
// **target と met は非公開なので明示的に載せる。** 抜けると Worker が
// リクエストごとに状態を作り直したときにノルマも達成数も消えます (#4478)。
type teenDoPaanchPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Target      int          `json:"tg"`
	Met         int          `json:"mt"`
}

// MarshalJSON implements json.Marshaler.
func (p *TeenDoPaanchPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(teenDoPaanchPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Target:      p.target,
		Met:         p.met,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *TeenDoPaanchPlayer) UnmarshalJSON(data []byte) error {
	var j teenDoPaanchPlayerJSON
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
	p.target = j.Target
	p.met = j.Met
	return nil
}
