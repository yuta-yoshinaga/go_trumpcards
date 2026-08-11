//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
)

// PasurPlayer はパスールのプレイヤー。
type PasurPlayer struct {
	*GamePlayer
	// captured は通常の捕獲札。
	captured []*Card
	// soorCaptured は**スール（場を空にした捕獲）で取った札**。
	//
	// **別の山に分けるのは、得点が倍になるのがその捕獲だけだから。** 混ぜると
	// あとから「どの札がスールで取られたか」を復元できない。
	soorCaptured []*Card
	// soors はスールの回数（表示用）。
	soors int
}

// NewPasurPlayer はコンストラクタ。
func NewPasurPlayer(isHuman bool) *PasurPlayer {
	return &PasurPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// ResetGame はゲーム全体をリセットする。
func (p *PasurPlayer) ResetGame() {
	p.Reset()
	p.captured = nil
	p.soorCaptured = nil
	p.soors = 0
}

// AddCaptured は通常の捕獲札を加える。
func (p *PasurPlayer) AddCaptured(cards []*Card) { p.captured = append(p.captured, cards...) }

// AddSoorCaptured はスールで取った札を加え、スール回数を 1 増やす。
func (p *PasurPlayer) AddSoorCaptured(cards []*Card) {
	p.soorCaptured = append(p.soorCaptured, cards...)
	p.soors++
}

// GetCaptured は通常の捕獲札を返す。
func (p *PasurPlayer) GetCaptured() []*Card { return p.captured }

// GetSoorCaptured はスールで取った札を返す。
func (p *PasurPlayer) GetSoorCaptured() []*Card { return p.soorCaptured }

// GetCapturedCount は捕獲した総枚数を返す。
func (p *PasurPlayer) GetCapturedCount() int { return len(p.captured) + len(p.soorCaptured) }

// GetSoors はスールの回数を返す。
func (p *PasurPlayer) GetSoors() int { return p.soors }

// pasurPlayerJSON is the JSON wire format for PasurPlayer.
type pasurPlayerJSON struct {
	GamePlayer   *GamePlayer `json:"gp"`
	Captured     []*Card     `json:"cp"`
	SoorCaptured []*Card     `json:"sc"`
	Soors        int         `json:"so"`
}

// MarshalJSON implements json.Marshaler.
func (p *PasurPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(pasurPlayerJSON{
		GamePlayer: p.GamePlayer, Captured: p.captured,
		SoorCaptured: p.soorCaptured, Soors: p.soors,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **スールの回数と札の山は対。** 回数だけ立っていると得点が倍になる根拠が
// 無いまま倍付けされ、盤面と得点が食い違います。
func (p *PasurPlayer) UnmarshalJSON(data []byte) error {
	var j pasurPlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Soors < 0 {
		return errors.New("negative soor count")
	}
	if (j.Soors == 0) != (len(j.SoorCaptured) == 0) {
		return errors.New("soor count and the soor pile disagree")
	}
	if len(j.Captured) > pasurMaxSliceLen || len(j.SoorCaptured) > pasurMaxSliceLen {
		return errors.New("pasur: input array exceeds maximum allowed size")
	}
	for _, c := range append(append([]*Card{}, j.Captured...), j.SoorCaptured...) {
		if c == nil {
			return errors.New("nil captured card")
		}
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.captured, p.soorCaptured, p.soors = j.Captured, j.SoorCaptured, j.Soors
	return nil
}
