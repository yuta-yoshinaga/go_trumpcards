//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
)

var errSimpleSimon = errors.New("simplesimon: invalid state")

// simpleSimonJSON は SimpleSimon の JSON ワイヤ形式。配り切ると trumpCards は空に
// なるため保存しない。history も永続化対象外。
type simpleSimonJSON struct {
	Columns        [SimpleSimonColCnt][]*Card `json:"co"`
	CompletedSuits int                        `json:"cs"`
	Phase          SimpleSimonPhase           `json:"ph"`
	MoveCount      int                        `json:"mc"`
	ActionLog      []*ActionLogEntry          `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *SimpleSimon) MarshalJSON() ([]byte, error) {
	return json.Marshal(simpleSimonJSON{
		Columns:        g.columns,
		CompletedSuits: g.completedSuits,
		Phase:          g.phase,
		MoveCount:      g.moveCount,
		ActionLog:      g.actionLog,
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
	g.actionLog = j.ActionLog
	g.history = nil
	return nil
}
