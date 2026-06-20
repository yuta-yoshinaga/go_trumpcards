//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
)

var errLaBelleLucie = errors.New("labellelucie: invalid state")

// laBelleLucieJSON は LaBelleLucie の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。history も永続化対象外。
type laBelleLucieJSON struct {
	Fans        [][]*Card                          `json:"fn"`
	Foundation  [LaBelleLucieFoundationCnt][]*Card `json:"fd"`
	RedealsLeft int                                `json:"rd"`
	Phase       LaBelleLuciePhase                  `json:"ph"`
	MoveCount   int                                `json:"mc"`
	ActionLog   []*ActionLogEntry                  `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *LaBelleLucie) MarshalJSON() ([]byte, error) {
	return json.Marshal(laBelleLucieJSON{
		Fans:        g.fans,
		Foundation:  g.foundation,
		RedealsLeft: g.redealsLeft,
		Phase:       g.phase,
		MoveCount:   g.moveCount,
		ActionLog:   g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *LaBelleLucie) UnmarshalJSON(data []byte) error {
	var j laBelleLucieJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > laBelleLucieMaxSliceLen || len(j.ActionLog) > laBelleLucieMaxSliceLen {
		return errLaBelleLucie
	}
	if j.RedealsLeft < 0 || j.RedealsLeft > LaBelleLucieMaxRedeals {
		return errLaBelleLucie
	}
	g.fans = j.Fans
	g.foundation = j.Foundation
	g.redealsLeft = j.RedealsLeft
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	g.actionLog = j.ActionLog
	g.history = nil
	return nil
}
