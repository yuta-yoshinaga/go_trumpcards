//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
)

// RollingStonePlayer はローリングストーンのプレイヤー。
//
// **獲得置き場はありません。** フォローできなかったトリックは手札に戻るので、
// 持ち札がそのまま順位です——少ないほど良い。
type RollingStonePlayer struct {
	*GamePlayer
	// pickups は引き取った回数（表示用）。
	pickups int
	// finishedAt は上がった順位（0 = まだ上がっていない）。
	finishedAt int
}

// NewRollingStonePlayer はコンストラクタ。
func NewRollingStonePlayer(isHuman bool) *RollingStonePlayer {
	return &RollingStonePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *RollingStonePlayer) ResetGame() {
	p.Reset()
	p.pickups = 0
	p.finishedAt = 0
}

// GetPickups は引き取った回数を返す。
func (p *RollingStonePlayer) GetPickups() int { return p.pickups }

// AddPickup は引き取り回数を 1 増やす。
func (p *RollingStonePlayer) AddPickup() { p.pickups++ }

// GetFinishedAt は上がった順位を返す（0 = まだ）。
func (p *RollingStonePlayer) GetFinishedAt() int { return p.finishedAt }

// SetFinishedAt は上がった順位を設定する。
func (p *RollingStonePlayer) SetFinishedAt(rank int) { p.finishedAt = rank }

// HasFinished は上がったかを返す。
func (p *RollingStonePlayer) HasFinished() bool { return p.finishedAt > 0 }

// rollingStonePlayerJSON is the JSON wire format for RollingStonePlayer.
type rollingStonePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Pickups    int         `json:"pu"`
	FinishedAt int         `json:"fa"`
}

// MarshalJSON implements json.Marshaler.
func (p *RollingStonePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(rollingStonePlayerJSON{
		GamePlayer: p.GamePlayer, Pickups: p.pickups, FinishedAt: p.finishedAt,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **上がった順位と手札は対。** 上がっているのに手札が残っている、あるいは
// 手札が空なのに順位が付いていない盤面は存在しません。
func (p *RollingStonePlayer) UnmarshalJSON(data []byte) error {
	var j rollingStonePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Pickups < 0 {
		return errors.New("negative pickup count")
	}
	if j.FinishedAt < 0 || j.FinishedAt > RollingStonePlayerCntMax {
		return errors.New("invalid finishing rank")
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	if (j.FinishedAt > 0) != (p.GetCardsSize() == 0) {
		return errors.New("finishing rank and hand size disagree")
	}
	p.pickups, p.finishedAt = j.Pickups, j.FinishedAt
	return nil
}
