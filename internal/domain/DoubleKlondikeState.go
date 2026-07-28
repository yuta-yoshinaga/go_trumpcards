//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
)

var errDoubleKlondike = errors.New("doubleklondike: invalid state")

// doubleKlondikeJSON は DoubleKlondike の JSON ワイヤ形式。配り切ると trumpCards は
// 空になるため保存しない。history も永続化対象外。
type doubleKlondikeJSON struct {
	Tableau    [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard `json:"tb"`
	Stock      []*Card                                                `json:"st"`
	Waste      []*Card                                                `json:"wa"`
	Foundation [DoubleKlondikeFoundationCnt][]*Card                   `json:"fd"`
	Phase      DoubleKlondikePhase                                    `json:"ph"`
	MoveCount  int                                                    `json:"mc"`
	ActionLog  []*ActionLogEntry                                      `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *DoubleKlondike) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubleKlondikeJSON{
		Tableau:    g.tableau,
		Stock:      g.stock,
		Waste:      g.waste,
		Foundation: g.foundation,
		Phase:      g.phase,
		MoveCount:  g.moveCount,
		ActionLog:  g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *DoubleKlondike) UnmarshalJSON(data []byte) error {
	var j doubleKlondikeJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > doubleKlondikeMaxSliceLen || len(j.Waste) > doubleKlondikeMaxSliceLen || len(j.ActionLog) > doubleKlondikeMaxSliceLen {
		return errDoubleKlondike
	}
	for _, col := range j.Tableau {
		if len(col) > doubleKlondikeMaxSliceLen {
			return errDoubleKlondike
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > doubleKlondikeMaxSliceLen {
			return errDoubleKlondike
		}
	}
	if j.Phase < DoubleKlondikePhasePlaying || j.Phase > DoubleKlondikePhaseGameOver {
		return errDoubleKlondike
	}
	g.tableau = j.Tableau
	g.stock = j.Stock
	g.waste = j.Waste
	g.foundation = j.Foundation
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	g.actionLog = j.ActionLog
	g.history = nil
	return nil
}
