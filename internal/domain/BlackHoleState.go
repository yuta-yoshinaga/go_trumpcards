//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
)

var errBlackHole = errors.New("blackhole: invalid state")

// blackHoleJSON は BlackHole の JSON ワイヤ形式。配り切ると trumpCards は空になるため
// 保存しない。history も永続化対象外。
type blackHoleJSON struct {
	Fans      [][]*Card         `json:"fn"`
	BlackHole []*Card           `json:"bh"`
	Phase     BlackHolePhase    `json:"ph"`
	MoveCount int               `json:"mc"`
	ActionLog []*ActionLogEntry `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*blackHoleSnapshot `json:"hi,omitempty"`
}

// MarshalJSON implements json.Marshaler.
func (g *BlackHole) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackHoleJSON{
		Fans:      g.fans,
		BlackHole: g.blackHole,
		Phase:     g.phase,
		MoveCount: g.moveCount,
		ActionLog: g.actionLog,
		History:   g.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *BlackHole) UnmarshalJSON(data []byte) error {
	var j blackHoleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > blackHoleMaxSliceLen || len(j.BlackHole) > blackHoleMaxSliceLen || len(j.ActionLog) > blackHoleMaxSliceLen {
		return errBlackHole
	}
	for _, fan := range j.Fans {
		if len(fan) > blackHoleMaxSliceLen {
			return errBlackHole
		}
	}
	if j.Phase < BlackHolePhasePlaying || j.Phase > BlackHolePhaseGameOver {
		return errBlackHole
	}
	g.fans = j.Fans
	g.blackHole = j.BlackHole
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	if len(j.History) > blackHoleMaxSliceLen {
		return errors.New("blackhole: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// blackHoleSnapshotJSON is the wire format for a single undo snapshot.
// blackHoleSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type blackHoleSnapshotJSON struct {
	Fans      [][]*Card      `json:"fn"`
	BlackHole []*Card        `json:"bh"`
	Phase     BlackHolePhase `json:"ph"`
	MoveCount int            `json:"mc"`
}

// MarshalJSON implements json.Marshaler for blackHoleSnapshot.
func (s *blackHoleSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackHoleSnapshotJSON{
		Fans:      s.fans,
		BlackHole: s.blackHole,
		Phase:     s.phase,
		MoveCount: s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for blackHoleSnapshot.
func (s *blackHoleSnapshot) UnmarshalJSON(data []byte) error {
	var j blackHoleSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > blackHoleMaxSliceLen || len(j.BlackHole) > blackHoleMaxSliceLen {
		return errors.New("blackhole: snapshot array exceeds maximum allowed size")
	}
	for _, fan := range j.Fans {
		if len(fan) > blackHoleMaxSliceLen {
			return errors.New("blackhole: snapshot fan exceeds maximum allowed size")
		}
	}
	s.fans = j.Fans
	s.blackHole = j.BlackHole
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
