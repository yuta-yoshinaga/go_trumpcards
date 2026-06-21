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
}

// MarshalJSON implements json.Marshaler.
func (g *BlackHole) MarshalJSON() ([]byte, error) {
	return json.Marshal(blackHoleJSON{
		Fans:      g.fans,
		BlackHole: g.blackHole,
		Phase:     g.phase,
		MoveCount: g.moveCount,
		ActionLog: g.actionLog,
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
	g.actionLog = j.ActionLog
	g.history = nil
	return nil
}
