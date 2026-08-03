//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
)

var errDoubleKlondike = errors.New("doubleklondike: invalid state")

// doubleKlondikeJSON は DoubleKlondike の JSON ワイヤ形式。配り切ると trumpCards は
// 空になるため保存しない。
type doubleKlondikeJSON struct {
	Tableau    [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard `json:"tb"`
	Stock      []*Card                                                `json:"st"`
	Waste      []*Card                                                `json:"wa"`
	Foundation [DoubleKlondikeFoundationCnt][]*Card                   `json:"fd"`
	Phase      DoubleKlondikePhase                                    `json:"ph"`
	MoveCount  int                                                    `json:"mc"`
	ActionLog  []*ActionLogEntry                                      `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*doubleKlondikeSnapshot `json:"hi,omitempty"`
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
		History:    g.history,
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
	if len(j.History) > doubleKlondikeMaxSliceLen {
		return errors.New("doubleklondike: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// doubleKlondikeSnapshotJSON is the wire format for a single undo snapshot.
// doubleKlondikeSnapshot uses unexported fields, so marshalling it directly
// would emit `[{},{}]` -- the undo depth would survive but every snapshot would
// be blank, and Undo would wipe the board instead of rewinding it (#4478).
type doubleKlondikeSnapshotJSON struct {
	Tableau    [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard `json:"tb"`
	Stock      []*Card                                                `json:"st"`
	Waste      []*Card                                                `json:"ws"`
	Foundation [DoubleKlondikeFoundationCnt][]*Card                   `json:"fd"`
	Phase      DoubleKlondikePhase                                    `json:"ph"`
	MoveCount  int                                                    `json:"mc"`
}

// MarshalJSON implements json.Marshaler for doubleKlondikeSnapshot.
func (s *doubleKlondikeSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(doubleKlondikeSnapshotJSON{
		Tableau:    s.tableau,
		Stock:      s.stock,
		Waste:      s.waste,
		Foundation: s.foundation,
		Phase:      s.phase,
		MoveCount:  s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for doubleKlondikeSnapshot.
func (s *doubleKlondikeSnapshot) UnmarshalJSON(data []byte) error {
	var j doubleKlondikeSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > doubleKlondikeMaxSliceLen || len(j.Waste) > doubleKlondikeMaxSliceLen {
		return errors.New("doubleklondike: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > doubleKlondikeMaxSliceLen {
			return errors.New("doubleklondike: snapshot column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > doubleKlondikeMaxSliceLen {
			return errors.New("doubleklondike: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	s.waste = j.Waste
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
