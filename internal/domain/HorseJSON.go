//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// horseSeatJSON は 1 席の JSON 表現。
type horseSeatJSON struct {
	Chips   int    `json:"c"`
	IsHuman bool   `json:"h"`
	Name    string `json:"n"`
}

// horseJSON は Horse の JSON 表現。
//
// **進行中のハンドは種目の JSON をそのまま抱える。** 卓は種目ごとに別の型なので、
// 正本 (席・チップ・種目・ハンド番号) と、いま打っている種目の状態を並べて持つ ──
// 抱えないと、復元のたびに打ちかけのハンドが消えて配り直しになる。
type horseJSON struct {
	Config     HorseConfig       `json:"cf"`
	Seats      []horseSeatJSON   `json:"st"`
	Phase      HorsePhase        `json:"ph"`
	Discipline HorseDiscipline   `json:"dc"`
	HandInDisc int               `json:"hd"`
	HandNumber int               `json:"hn"`
	SeatMap    []int             `json:"sm"`
	Table      json.RawMessage   `json:"tb,omitempty"`
	GameEnd    bool              `json:"ge"`
	TurnNumber int               `json:"tn"`
	ActionLog  []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *Horse) MarshalJSON() ([]byte, error) {
	seats := make([]horseSeatJSON, 0, len(g.seats))
	for _, s := range g.seats {
		seats = append(seats, horseSeatJSON{Chips: s.chips, IsHuman: s.isHuman, Name: s.name})
	}
	var table json.RawMessage
	if g.table != nil {
		b, err := json.Marshal(g.table)
		if err != nil {
			return nil, fmt.Errorf("horse: marshal table: %w", err)
		}
		table = b
	}
	return json.Marshal(horseJSON{
		Config: g.config, Seats: seats, Phase: g.phase,
		Discipline: g.discipline, HandInDisc: g.handInDiscipline, HandNumber: g.handNumber,
		SeatMap: g.seatMap, Table: table, GameEnd: g.gameEndFlag,
		TurnNumber: g.turnNumber, ActionLog: g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
//
// **種目と席数の整合まで見る。** 席数が種目の受け付ける卓サイズでない保存は、
// 復元できても次のハンドで別人数の卓が作られる ── 範囲検査だけでは通ってしまう。
func (g *Horse) UnmarshalJSON(data []byte) error {
	var j horseJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Seats) > horseMaxSliceLen || len(j.ActionLog) > horseMaxSliceLen ||
		len(j.SeatMap) > horseMaxSliceLen {
		return fmt.Errorf("horse: input array exceeds maximum allowed size")
	}
	if err := j.Config.Validate(); err != nil {
		return fmt.Errorf("horse: invalid config: %w", err)
	}
	if len(j.Seats) != j.Config.Seats {
		return fmt.Errorf("horse: seat count %d does not match config %d", len(j.Seats), j.Config.Seats)
	}
	if j.Phase < HorsePhaseHand || j.Phase > HorsePhaseMax {
		return fmt.Errorf("horse: invalid phase %d", j.Phase)
	}
	// **回さない種目を復元しない。** 種目の値は 8 つあるが、H.O.R.S.E. の卓が
	// 回すのは 5 つだけ ── 保存を書き換えれば、5 種目の卓を 2-7 Triple Draw の
	// 途中から復元できてしまい、「次のハンド」でローテーションの外に出る。
	if HorseRotationIndex(j.Config.Variant, j.Discipline) < 0 {
		return fmt.Errorf("horse: invalid discipline %d for variant %d",
			j.Discipline, j.Config.Variant)
	}
	if j.HandInDisc < 1 || j.HandInDisc > j.Config.HandsPerDiscipline {
		return fmt.Errorf("horse: hand-in-discipline %d out of range", j.HandInDisc)
	}
	if j.HandNumber < 1 {
		return fmt.Errorf("horse: hand number %d out of range", j.HandNumber)
	}
	for _, idx := range j.SeatMap {
		if idx < 0 || idx >= len(j.Seats) {
			return fmt.Errorf("horse: seat map entry %d out of range", idx)
		}
	}
	for _, s := range j.Seats {
		if s.Chips < 0 {
			return fmt.Errorf("horse: seat chips must not be negative")
		}
	}

	g.config = j.Config
	g.seats = make([]*horseSeat, 0, len(j.Seats))
	for _, s := range j.Seats {
		g.seats = append(g.seats, &horseSeat{chips: s.Chips, isHuman: s.IsHuman, name: s.Name})
	}
	g.phase = j.Phase
	g.discipline = j.Discipline
	g.handInDiscipline = j.HandInDisc
	g.handNumber = j.HandNumber
	g.seatMap = j.SeatMap
	g.gameEndFlag = j.GameEnd
	g.turnNumber = j.TurnNumber
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.table, g.endPhase, g.harvest = nil, horseEndPhase{}, nil
	if j.Phase == HorsePhaseHand {
		// **ハンド中なのに卓が無い保存は受け取らない。** 復元しても打てる手が
		// 1 つも無く、マッチは止まったまま二度と進まない ── 落ちないぶん、
		// 壊れた保存より質が悪い。
		if len(j.Table) == 0 {
			return fmt.Errorf("horse: phase is Hand but no table was saved")
		}
		if err := g.restoreTable(j.Table); err != nil {
			return err
		}
	}
	return nil
}

// restoreTable は保存された種目の状態を読み直し、回収の配線をやり直す。
//
// **配線は毎回作り直す。** 回収の関数はプレイヤーの実体を閉じ込めているので、
// 保存には残せない ── 残せるのは種目の状態だけで、繋ぎ直すのは復元側の仕事。
func (g *Horse) restoreTable(raw json.RawMessage) error {
	table, end, harvest := g.buildTable()
	if table == nil {
		return fmt.Errorf("horse: cannot rebuild the table for discipline %d", g.discipline)
	}
	if err := json.Unmarshal(raw, table); err != nil {
		return fmt.Errorf("horse: restore table: %w", err)
	}
	// **卓の人数は席数と一致していなければならない。** 種目の UnmarshalJSON は
	// プレイヤー列をまるごと差し替えるだけで人数を検めないので、少ない卓を
	// 復元すると次の 1 手で落ちる ── `players[currentTurn]` も、回収の
	// `GetPlayer(i).GetChips()` も、範囲外を見に行く。
	if n := table.GetPlayerCnt(); n != len(g.seatMap) {
		return fmt.Errorf("horse: restored table seats %d players but %d seats are funded",
			n, len(g.seatMap))
	}
	g.table, g.endPhase, g.harvest = table, end, harvest
	return nil
}
