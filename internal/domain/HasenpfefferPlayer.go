//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"fmt"
)

// HasenpfefferPlayer はハーゼンプフェファーのプレイヤー。
type HasenpfefferPlayer struct {
	*GamePlayer
	TrickHolder
	// bid は宣言したトリック数（-1 = 未宣言、0 = 降りた）。
	bid int
}

// NewHasenpfefferPlayer はコンストラクタ。
func NewHasenpfefferPlayer(isHuman bool) *HasenpfefferPlayer {
	p := &HasenpfefferPlayer{GamePlayer: NewGamePlayer(isHuman)}
	p.bid = -1
	return p
}

// ResetGame はゲーム全体をリセットする。
func (p *HasenpfefferPlayer) ResetGame() { p.ResetRound() }

// ResetRound は 1 ハンド分の状態を初期化する。
func (p *HasenpfefferPlayer) ResetRound() {
	resetPlayerRound(p)
	p.bid = -1
}

// GetBid は宣言したトリック数を返す（-1 = 未宣言、0 = 降りた）。
func (p *HasenpfefferPlayer) GetBid() int { return p.bid }

// SetBid は宣言を設定する。
func (p *HasenpfefferPlayer) SetBid(n int) { p.bid = n }

// HasBid は宣言済みかを返す（降りたのも宣言済み）。
func (p *HasenpfefferPlayer) HasBid() bool { return p.bid >= 0 }

// hasenpfefferPlayerJSON is the JSON wire format for HasenpfefferPlayer.
//
// **bid は非公開なので明示的に載せる。** 抜けると Worker がリクエストごとに
// 状態を作り直したときに競りの結果が消えます (#4478)。
type hasenpfefferPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	Bid         int          `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *HasenpfefferPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(hasenpfefferPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		Bid:         p.bid,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **宣言は勝敗そのものなので範囲を検証する。** 落札額が壊れると finishHand の
// 成否判定と失敗時の失点が変わります (#5302〜#5305、#5309 と同じ形)。
func (p *HasenpfefferPlayer) UnmarshalJSON(data []byte) error {
	var j hasenpfefferPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// -1 は未宣言、0 は降りた、それ以外は 3..6 の落札額。
	if j.Bid < -1 || j.Bid > HasenpfefferMaxBid || (j.Bid > 0 && j.Bid < HasenpfefferMinBid) {
		return fmt.Errorf("invalid bid: %d", j.Bid)
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if j.TrickHolder != nil {
		p.TrickHolder = *j.TrickHolder
	}
	p.bid = j.Bid
	return nil
}
