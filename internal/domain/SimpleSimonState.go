//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

var errSimpleSimon = errors.New("simplesimon: invalid state")

// simpleSimonJSON は SimpleSimon の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。
type simpleSimonJSON struct {
	Columns        [SimpleSimonColCnt][]*Card `json:"co"`
	CompletedSuits int                        `json:"cs"`
	Phase          SimpleSimonPhase           `json:"ph"`
	MoveCount      int                        `json:"mc"`
	ActionLog      []*ActionLogEntry          `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*simpleSimonSnapshot `json:"hi,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (g *SimpleSimon) MarshalJSON() ([]byte, error) {
	return json.Marshal(simpleSimonJSON{
		Columns:        g.columns,
		CompletedSuits: g.completedSuits,
		Phase:          g.phase,
		MoveCount:      g.moveCount,
		ActionLog:      g.actionLog,
		History:        g.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *SimpleSimon) UnmarshalJSON(data []byte) error {
	var j simpleSimonJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > simpleSimonMaxSliceLen {
		return errSimpleSimon
	}
	if j.CompletedSuits < 0 || j.CompletedSuits > SimpleSimonFoundationCnt {
		return errSimpleSimon
	}
	if j.Phase < SimpleSimonPhasePlaying || j.Phase > SimpleSimonPhaseGameOver {
		return errSimpleSimon
	}
	g.columns = j.Columns
	g.completedSuits = j.CompletedSuits
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	if len(j.History) > simpleSimonMaxSliceLen {
		return errors.New("simplesimon: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// simpleSimonSnapshotJSON is the wire format for a single undo snapshot.
// simpleSimonSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type simpleSimonSnapshotJSON struct {
	Columns        [SimpleSimonColCnt][]*Card `json:"co"`
	CompletedSuits int                        `json:"cs"`
	Phase          SimpleSimonPhase           `json:"ph"`
	MoveCount      int                        `json:"mc"`
}

// MarshalJSON implements json.Marshaler for simpleSimonSnapshot.
func (s *simpleSimonSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(simpleSimonSnapshotJSON{
		Columns:        s.columns,
		CompletedSuits: s.completedSuits,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for simpleSimonSnapshot.
func (s *simpleSimonSnapshot) UnmarshalJSON(data []byte) error {
	var j simpleSimonSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Columns {
		if len(col) > simpleSimonMaxSliceLen {
			return errors.New("simplesimon: snapshot column exceeds maximum allowed size")
		}
	}
	s.columns = j.Columns
	s.completedSuits = j.CompletedSuits
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
