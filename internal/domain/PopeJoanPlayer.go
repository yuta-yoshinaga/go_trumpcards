//go:build !js || !wasm || extra3

package domain

import "encoding/json"

// PopeJoanPlayer はポープ・ジョーンのプレイヤークラス。
type PopeJoanPlayer struct {
	*GamePlayer
	// chips は通算のチップ残高。
	//
	// **マイナスに落ちても止めない。**この game の長さは TargetDeals で決まって
	// いて脱落の概念が無く、原典にもテーブルステークスの規定が無いため、資金
	// 切れで座を外す自然な区切りが存在しない。収支そのものが順位になる。
	chips int
}

// NewPopeJoanPlayer コンストラクタ
func NewPopeJoanPlayer(isHuman bool) *PopeJoanPlayer {
	return &PopeJoanPlayer{GamePlayer: NewGamePlayer(isHuman)}
}

// GetChips はチップ残高を返す。
func (p *PopeJoanPlayer) GetChips() int { return p.chips }

// AddChips はチップを増減する。
func (p *PopeJoanPlayer) AddChips(n int) { p.chips += n }

// ResetDeal はディール開始時に手札を初期化する。チップは残す。
func (p *PopeJoanPlayer) ResetDeal() {
	p.Reset()
	p.SetIsFinished(false)
}

// popeJoanPlayerJSON is the JSON wire format for PopeJoanPlayer.
type popeJoanPlayerJSON struct {
	GamePlayer *GamePlayer `json:"gp"`
	Chips      int         `json:"ch"`
}

// MarshalJSON implements json.Marshaler.
func (p *PopeJoanPlayer) MarshalJSON() ([]byte, error) {
	return json.Marshal(popeJoanPlayerJSON{GamePlayer: p.GamePlayer, Chips: p.chips})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *PopeJoanPlayer) UnmarshalJSON(data []byte) error {
	var j popeJoanPlayerJSON
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
