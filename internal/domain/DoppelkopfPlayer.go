//go:build !js || !wasm || casino

package domain

import "encoding/json"

// DoppelkopfPlayer ドッペルコップのプレイヤークラス。手札（GamePlayer）、獲得
// トリック（TrickHolder）、所持チップ（ChipHolder）を保持する。チーム（Re/Kontra）
// は Doppelkopf 本体がラウンドごとにクラブ Q の保持状況から自動割当する。
type DoppelkopfPlayer struct {
	*GamePlayer
	TrickHolder
	ChipHolder
}

// NewDoppelkopfPlayer コンストラクタ
func NewDoppelkopfPlayer(isHuman bool, startChips int) *DoppelkopfPlayer {
	p := &DoppelkopfPlayer{
		GamePlayer: NewGamePlayer(isHuman),
	}
	p.SetChips(startChips)
	return p
}

// ResetRound ラウンドをリセット（トリック・手札・終了状態を初期化）。チップは
// 累積するためリセットしない。
func (p *DoppelkopfPlayer) ResetRound() {
	resetPlayerRound(p)
}

// doppelkopfPlayerJSON is the JSON wire format for DoppelkopfPlayer.
type doppelkopfPlayerJSON struct {
	GamePlayer  *GamePlayer  `json:"gp"`
	TrickHolder *TrickHolder `json:"th"`
	ChipHolder  *ChipHolder  `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *DoppelkopfPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(doppelkopfPlayerJSON{
		GamePlayer:  p.GamePlayer,
		TrickHolder: &p.TrickHolder,
		ChipHolder:  &p.ChipHolder,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *DoppelkopfPlayer) UnmarshalJSON(data []byte) error {
	var j doppelkopfPlayerJSON
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
