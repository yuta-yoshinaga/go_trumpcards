//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

var errLaBelleLucie = errors.New("labellelucie: invalid state")

// laBelleLucieJSON は LaBelleLucie の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。
type laBelleLucieJSON struct {
	Fans        [][]*Card                          `json:"fn"`
	Foundation  [LaBelleLucieFoundationCnt][]*Card `json:"fd"`
	RedealsLeft int                                `json:"rd"`
	Phase       LaBelleLuciePhase                  `json:"ph"`
	MoveCount   int                                `json:"mc"`
	ActionLog   []*ActionLogEntry                  `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*laBelleLucieSnapshot `json:"hi,omitempty"`
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
		History:     g.history,
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
	if len(j.History) > laBelleLucieMaxSliceLen {
		return errors.New("labellelucie: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// laBelleLucieSnapshotJSON is the wire format for a single undo snapshot.
// laBelleLucieSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type laBelleLucieSnapshotJSON struct {
	Fans        [][]*Card                          `json:"fn"`
	Foundation  [LaBelleLucieFoundationCnt][]*Card `json:"fd"`
	RedealsLeft int                                `json:"rd"`
	Phase       LaBelleLuciePhase                  `json:"ph"`
	MoveCount   int                                `json:"mc"`
}

// MarshalJSON implements json.Marshaler for laBelleLucieSnapshot.
func (s *laBelleLucieSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(laBelleLucieSnapshotJSON{
		Fans:        s.fans,
		Foundation:  s.foundation,
		RedealsLeft: s.redealsLeft,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for laBelleLucieSnapshot.
func (s *laBelleLucieSnapshot) UnmarshalJSON(data []byte) error {
	var j laBelleLucieSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Fans) > laBelleLucieMaxSliceLen {
		return errors.New("labellelucie: snapshot fan count exceeds maximum allowed size")
	}
	for _, fan := range j.Fans {
		if len(fan) > laBelleLucieMaxSliceLen {
			return errors.New("labellelucie: snapshot fan exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > laBelleLucieMaxSliceLen {
			return errors.New("labellelucie: snapshot pile exceeds maximum allowed size")
		}
	}
	s.fans = j.Fans
	s.foundation = j.Foundation
	s.redealsLeft = j.RedealsLeft
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}
