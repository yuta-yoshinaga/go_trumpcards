//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// SheepsheadPlayer シープスヘッドのプレイヤークラス。手札（GamePlayer）、獲得
// トリック（TrickHolder）、所持チップ（ChipHolder）を保持する。チーム（ピッカー
// 組／ディフェンダー組）は Sheepshead 本体がラウンドごとに管理する。
type SheepsheadPlayer struct {
	*GamePlayer
	TrickHolder
	ChipHolder
}

// NewSheepsheadPlayer コンストラクタ
func NewSheepsheadPlayer(isHuman bool, startChips int) *SheepsheadPlayer {
	p := &SheepsheadPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
	p.SetChips(startChips)
	return p
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）。チップは
// 累積するためリセットしない。
func (p *SheepsheadPlayer) ResetRound() {
	resetPlayerRound(p)
}

// sheepsheadPlayerJSON is the JSON wire format for SheepsheadPlayer.
type sheepsheadPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	ChipHolder  *ChipHolder  `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *SheepsheadPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(sheepsheadPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		ChipHolder:  &p.ChipHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *SheepsheadPlayer) UnmarshalJSON(data []byte) error {
	var j sheepsheadPlayerJSON
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
	if j.ChipHolder != nil {
		p.ChipHolder = *j.ChipHolder
	}
	return nil
}
