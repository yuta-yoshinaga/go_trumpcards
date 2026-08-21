//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

var errCurdsAndWhey = errors.New("curdsandwhey: invalid state")

// curdsAndWheyJSON は CurdsAndWhey の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。
type curdsAndWheyJSON struct {
	Columns        [CurdsAndWheyColCnt][]*Card `json:"co"`
	CompletedSuits int                         `json:"cs"`
	Phase          CurdsAndWheyPhase           `json:"ph"`
	MoveCount      int                         `json:"mc"`
	ActionLog      []*ActionLogEntry           `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*curdsAndWheySnapshot `json:"hi,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (g *CurdsAndWhey) MarshalJSON() ([]byte, error) {
	return json.Marshal(curdsAndWheyJSON{
		Columns:        g.columns,
		CompletedSuits: g.completedSuits,
		Phase:          g.phase,
		MoveCount:      g.moveCount,
		ActionLog:      g.actionLog,
		History:        g.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *CurdsAndWhey) UnmarshalJSON(data []byte) error {
	var j curdsAndWheyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > curdsAndWheyMaxSliceLen {
		return errCurdsAndWhey
	}
	if j.CompletedSuits < 0 || j.CompletedSuits > CurdsAndWheyFoundationCnt {
		return errCurdsAndWhey
	}
	if j.Phase < CurdsAndWheyPhasePlaying || j.Phase > CurdsAndWheyPhaseGameOver {
		return errCurdsAndWhey
	}
	g.columns = j.Columns
	g.completedSuits = j.CompletedSuits
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	if len(j.History) > curdsAndWheyMaxSliceLen {
		return errors.New("curdsandwhey: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// curdsAndWheySnapshotJSON is the wire format for a single undo snapshot.
// curdsAndWheySnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type curdsAndWheySnapshotJSON struct {
	Columns        [CurdsAndWheyColCnt][]*Card `json:"co"`
	CompletedSuits int                         `json:"cs"`
	Phase          CurdsAndWheyPhase           `json:"ph"`
	MoveCount      int                         `json:"mc"`
}

// MarshalJSON implements json.Marshaler for curdsAndWheySnapshot.
func (s *curdsAndWheySnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(curdsAndWheySnapshotJSON{
		Columns:        s.columns,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for curdsAndWheySnapshot.
func (s *curdsAndWheySnapshot) UnmarshalJSON(data []byte) error {
	var j curdsAndWheySnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Columns {
		if len(col) > curdsAndWheyMaxSliceLen {
			return errors.New("curdsandwhey: snapshot column exceeds maximum allowed size")
		}
	}
	s.columns = j.Columns
	s.completedSuits = j.CompletedSuits
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
