//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"fmt"
	"slices"
)

// SergeantMajorPlayer はサージェントメジャーのプレイヤー。
type SergeantMajorPlayer struct {
	*GamePlayer
	TrickHolder
	// target はこのラウンドのノルマ (8 / 5 / 3)。**席順で決まり、宣言しません。**
	target int
	// score は「ノルマとの差」の累計。**勝敗はこれで決まる。**
	score int
}

// NewSergeantMajorPlayer はコンストラクタ。
func NewSergeantMajorPlayer(isHuman bool) *SergeantMajorPlayer {
	return &SergeantMajorPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *SergeantMajorPlayer) ResetGame() {
	p.ResetRound()
	p.score = 0
}

// ResetRound は 1 ラウンド分の状態を初期化する。
func (p *SergeantMajorPlayer) ResetRound() {
	resetPlayerRound(p)
	p.target = 0
}

// GetTarget はこのラウンドのノルマを返す。
func (p *SergeantMajorPlayer) GetTarget() int { return p.target }

// SetTarget はノルマを設定する。
func (p *SergeantMajorPlayer) SetTarget(n int) { p.target = n }

// GetScore は累計得点（ノルマとの差の合計）を返す。
func (p *SergeantMajorPlayer) GetScore() int { return p.score }

// AddScore は得点を加算する。
func (p *SergeantMajorPlayer) AddScore(n int) { p.score += n }

// SetScore は得点を設定する（復元・テスト用）。
func (p *SergeantMajorPlayer) SetScore(n int) { p.score = n }

// sergeantMajorPlayerJSON is the JSON wire format for SergeantMajorPlayer.
type sergeantMajorPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Target      int          `json:"tg"`
	Score       int          `json:"sc"`
}

// MarshalJSON implements json.Marshaler.
func (p *SergeantMajorPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sergeantMajorPlayerJSON{
		GamePlayer: p.GamePlayer, TrickHolder: &p.TrickHolder, Target: p.target, Score: p.score,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **ノルマは勝敗そのものなので範囲を検証する** (#5302〜#5305、#5309 と同じ形)。
// target が壊れると finishRound の差分が変わり、順位がそのまま入れ替わります。
func (p *SergeantMajorPlayer) UnmarshalJSON(data []byte) error {
	var j sergeantMajorPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// 0 は「ラウンド開始前（ResetRound 直後）」の正当な値。
	if j.Target != 0 && !slices.Contains(SergeantMajorTargets[:], j.Target) {
		return fmt.Errorf("invalid target: %d", j.Target)
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
	p.score = j.Score
	return nil
}
