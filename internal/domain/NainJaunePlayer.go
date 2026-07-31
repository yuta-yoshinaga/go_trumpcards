//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// NainJaunePlayer はル・ナン・ジョーヌのプレイヤークラス。
type NainJaunePlayer struct {
	*GamePlayer
	// chips は通算のチップ残高。
	//
	// **マイナスに落ちても止めない。**長さは TargetDeals で決まっていて脱落の
	// 概念が無く、原典にもテーブルステークスの規定が無い。収支そのものが順位。
	chips int
}

// NewNainJaunePlayer コンストラクタ
func NewNainJaunePlayer(isHuman bool) *NainJaunePlayer {
	return &NainJaunePlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetChips はチップ残高を返す。
func (p *NainJaunePlayer) GetChips() int { return p.chips }

// AddChips はチップを増減する。
func (p *NainJaunePlayer) AddChips(n int) { p.chips += n }

// ResetDeal はディール開始時に手札を初期化する。チップは残す。
func (p *NainJaunePlayer) ResetDeal() {
	p.Reset()
	p.SetIsFinished(false)
}

// nainJaunePlayerJSON is the JSON wire format for NainJaunePlayer.
type nainJaunePlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *NainJaunePlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(nainJaunePlayerJSON{GamePlayer: p.GamePlayer, Chips: p.chips})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *NainJaunePlayer) UnmarshalJSON(data []byte) error {
	var j nainJaunePlayerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.GamePlayer != nil {
		p.GamePlayer = j.GamePlayer
	} else {
		p.GamePlayer = NewGamePlayer(false)
	}
	p.chips = j.Chips
	return nil
}
