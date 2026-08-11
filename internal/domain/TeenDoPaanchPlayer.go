//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"fmt"
	"slices"
)

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
//
// **ノルマと達成数は勝敗そのものなので範囲を検証する** (レビュー指摘 PR #5309)。
// 席インデックスだけ丁寧に見て、ゲーム固有の状態値を素通しするのがこの
// リポジトリの再発パターン (#5302〜#5305)。target が {3,2,5} 以外だと
// finishRound の達成判定が壊れ、met が壊れると finishGame の勝者が変わる。
func (p *TeenDoPaanchPlayer) UnmarshalJSON(data []byte) error {
	var j teenDoPaanchPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// 0 は「ラウンド開始前（ResetRound 直後）」の正当な値。
	if j.Target != 0 && !slices.Contains(TeenDoPaanchTargets[:], j.Target) {
		return fmt.Errorf("invalid target: %d", j.Target)
	}
	if j.Met < 0 || j.Met > TeenDoPaanchRoundsMax {
		return fmt.Errorf("invalid met count: %d", j.Met)
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
