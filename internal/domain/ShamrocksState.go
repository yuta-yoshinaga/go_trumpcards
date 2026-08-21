//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

var errShamrocks = errors.New("shamrocks: invalid state")

// shamrocksJSON は Shamrocks の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。
type shamrocksJSON struct {
	Fans        [][]*Card                       `json:"fn"`
	Foundation  [ShamrocksFoundationCnt][]*Card `json:"fd"`
	RedealsLeft int                             `json:"rd"`
	Phase       ShamrocksPhase                  `json:"ph"`
	MoveCount   int                             `json:"mc"`
	ActionLog   []*ActionLogEntry               `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*shamrocksSnapshot `json:"hi,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (g *Shamrocks) MarshalJSON() ([]byte, error) {
	return json.Marshal(shamrocksJSON{
		Fans:        g.fans,
		Foundation:  g.foundation,
		RedealsLeft: g.redealsLeft,
		Phase:       g.phase,
		MoveCount:   g.moveCount,
		ActionLog:   g.actionLog,
		History:     g.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Shamrocks) UnmarshalJSON(data []byte) error {
	var j shamrocksJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > shamrocksMaxSliceLen || len(j.ActionLog) > shamrocksMaxSliceLen {
		return errShamrocks
	}
	if j.RedealsLeft < 0 || j.RedealsLeft > ShamrocksMaxRedeals {
		return errShamrocks
	}
	g.fans = j.Fans
	g.foundation = j.Foundation
	g.redealsLeft = j.RedealsLeft
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	if len(j.History) > shamrocksMaxSliceLen {
		return errors.New("shamrocks: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// shamrocksSnapshotJSON is the wire format for a single undo snapshot.
// shamrocksSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type shamrocksSnapshotJSON struct {
	Fans        [][]*Card                       `json:"fn"`
	Foundation  [ShamrocksFoundationCnt][]*Card `json:"fd"`
	RedealsLeft int                             `json:"rd"`
	Phase       ShamrocksPhase                  `json:"ph"`
	MoveCount   int                             `json:"mc"`
}

// MarshalJSON implements json.Marshaler for shamrocksSnapshot.
func (s *shamrocksSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(shamrocksSnapshotJSON{
		Fans:        s.fans,
		Foundation:  s.foundation,
		RedealsLeft: s.redealsLeft,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for shamrocksSnapshot.
func (s *shamrocksSnapshot) UnmarshalJSON(data []byte) error {
	var j shamrocksSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > shamrocksMaxSliceLen {
		return errors.New("shamrocks: snapshot fan count exceeds maximum allowed size")
	}
	for _, fan := range j.Fans {
		if len(fan) > shamrocksMaxSliceLen {
			return errors.New("shamrocks: snapshot fan exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > shamrocksMaxSliceLen {
			return errors.New("shamrocks: snapshot pile exceeds maximum allowed size")
		}
	}
	s.fans = j.Fans
	s.foundation = j.Foundation
	s.redealsLeft = j.RedealsLeft
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
