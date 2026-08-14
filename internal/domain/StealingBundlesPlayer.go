//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
)

// StealingBundlesPlayer はスティーリングバンドルのプレイヤー。
//
// 手札に加えて **獲得した束 (bundle)** を持つ。束は積み重ねで、末尾が一番上。
// **一番上のランクだけが外から見える**——そこが相手に狙われる場所です。
type StealingBundlesPlayer struct {
	*GamePlayer
	bundle []*Card
}

// NewStealingBundlesPlayer はコンストラクタ。
func NewStealingBundlesPlayer(isHuman bool) *StealingBundlesPlayer {
	return &StealingBundlesPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetBundle は獲得した束を返す。末尾が一番上。
func (p *StealingBundlesPlayer) GetBundle() []*Card { return p.bundle }

// GetBundleSize は束の枚数を返す。
func (p *StealingBundlesPlayer) GetBundleSize() int { return len(p.bundle) }

// GetBundleTop は束の一番上を返す (nil = 束が空)。
//
// **ここが相手の狙い所。** 同じランクを出されると束ごと持っていかれます。
func (p *StealingBundlesPlayer) GetBundleTop() *Card {
	if len(p.bundle) == 0 {
		return nil
	}
	return p.bundle[len(p.bundle)-1]
}

// AddToBundle は束の一番上に積む。
func (p *StealingBundlesPlayer) AddToBundle(cards ...*Card) {
	for _, c := range cards {
		if c != nil {
			p.bundle = append(p.bundle, c)
		}
	}
}

// TakeBundle は束を丸ごと取り出して空にする。
func (p *StealingBundlesPlayer) TakeBundle() []*Card {
	taken := p.bundle
	p.bundle = nil
	return taken
}

// SetBundle は束を設定する (主にテスト/復元用)。
func (p *StealingBundlesPlayer) SetBundle(cards []*Card) { p.bundle = cards }

// ResetGame はゲーム開始時の状態に戻す。
func (p *StealingBundlesPlayer) ResetGame() {
	p.Reset()
	p.bundle = nil
	p.SetIsFinished(false)
}

// stealingBundlesPlayerJSON is the JSON wire format for StealingBundlesPlayer.
type stealingBundlesPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Bundle     []*Card     `json:"bd"`
}

// MarshalJSON implements json.Marshaler.
func (p *StealingBundlesPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(stealingBundlesPlayerJSON{GamePlayer: p.GamePlayer, Bundle: p.bundle})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *StealingBundlesPlayer) UnmarshalJSON(data []byte) error {
	var j stealingBundlesPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer == nil {
		return errors.New("stealing bundles player is missing its base player")
	}
	// **束に nil は混ざれない。** 一番上のランクが読めないと略奪の判定ができません。
	for _, c := range j.Bundle {
		if c == nil {
			return errors.New("a bundle cannot contain a missing card")
		}
	}
	if len(j.Bundle) > StealingBundlesDeckSize {
		return errors.New("a bundle cannot hold more cards than the deck")
	}
	p.GamePlayer = j.GamePlayer
	p.bundle = j.Bundle
	return nil
}
