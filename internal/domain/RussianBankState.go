//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"math/rand"
)

// --- Undo (人間の単一/多段ステップ) ---

// CanUndo Undo 可能なスナップショットがあるか。
func (g *RussianBank) CanUndo() bool { return len(g.history) > 0 }

// takeSnapshot 移動前の状態を保存する。人間 (seat 0) の手番中のみ記録する。
func (g *RussianBank) takeSnapshot() {
	if g.current != 0 || g.phase != RussianBankPhasePlaying {
		return
	}
	// boardJSON, NOT MarshalJSON. MarshalJSON now carries the undo history, so
	// snapshotting through it would embed every earlier snapshot inside the new
	// one and double the payload on every move: measured 5.6 KB after move 1,
	// 1.45 MB after move 9. Past ~1 MiB the snapshot codec rejects it and the
	// whole session stops restoring from KV.
	b, err := json.Marshal(g.boardJSON())
	if err != nil {
		return
	}
	g.history = appendSnapshot(g.history, &russianBankSnapshot{stateJSON: b})
}

// Undo 直近の人間の 1 手を取り消す。
func (g *RussianBank) Undo() error {
	if len(g.history) == 0 {
		return errors.New("russianbank: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	rest := g.history[:len(g.history)-1]
	if err := json.Unmarshal(snap.stateJSON, g); err != nil {
		return err
	}
	g.history = rest
	return nil
}

// --- CPU 手番 ---

// RunCpuTurn CPU の手番を自動で進める。current が CPU かつ Playing のときのみ動作。
// 取りこぼし (強制ファウンデーション手の見逃し) は難易度に応じて確率的に発生し、
// 人間の stop を誘発する。
func (g *RussianBank) RunCpuTurn() {
	if g.phase != RussianBankPhasePlaying || !g.players[g.current].IsCPU() {
		return
	}
	miss := g.cpuMissChance()
	for g.phase == RussianBankPhasePlaying {
		if miss > 0 && g.hasForcedFoundationMove(g.current) && rand.Float64() < miss {
			break // わざと強制手を残して手番を終える
		}
		m, ok := g.bestCpuMove()
		if !ok {
			break
		}
		g.applyCpuMove(m)
		if g.phase != RussianBankPhasePlaying {
			return // 勝利 / 停滞で決着
		}
	}
	_ = g.Discard() // 手札を 1 枚捨てて (または詰みならパスして) 手番終了
}

// cpuMissChance 難易度ごとの取りこぼし発生確率。
func (g *RussianBank) cpuMissChance() float64 {
	switch g.config.CpuDifficulty {
	case RussianBankCpuDifficultyEasy:
		return 0.45
	case RussianBankCpuDifficultyHard:
		return 0
	default:
		return 0.15
	}
}

// bestCpuMove CPU が指す最善手を返す。自分のリザーブ/廃札を減らす手と
// ファウンデーション手のみを候補とし (タブロー間の振動や相手を利する手は除外)、
// 単調にリザーブ/廃札が減るため必ず有限手で終わる。
func (g *RussianBank) bestCpuMove() (rbMove, bool) {
	best := rbMove{}
	bestScore := 0
	found := false
	for _, m := range g.enumerateMoves() {
		if !g.cpuBeneficial(m) {
			continue
		}
		if s := g.scoreMove(m); !found || s > bestScore {
			best, bestScore, found = m, s, true
		}
	}
	return best, found
}

// cpuBeneficial CPU が指してよい (進展する) 手か。
func (g *RussianBank) cpuBeneficial(m rbMove) bool {
	if m.src.FromOpponent {
		return false // 相手のリザーブ/廃札を減らす手は利敵
	}
	if m.toFnd {
		return true // ファウンデーション手は常に進展
	}
	// タブロー手は自分のリザーブ/廃札を減らす場合のみ採用 (タブロー間振動を除外)。
	return m.src.Zone == RussianBankZoneReserve || m.src.Zone == RussianBankZoneWaste
}

// applyCpuMove 列挙済みの手を適用する。
func (g *RussianBank) applyCpuMove(m rbMove) {
	if m.toFnd {
		_ = g.MoveToFoundation(m.src)
		return
	}
	_ = g.MoveToTableau(m.src, m.toCol)
}

// --- JSON (状態の永続化 / KV ワーカー復元) ---

// russianBankJSON は RussianBank の JSON ワイヤ形式。decks は配り切ると空になるため保存しない。
type russianBankJSON struct {
	Players     []*RussianBankPlayer              `json:"pl"`
	Tableau     [RussianBankTableauCnt][]*Card    `json:"tb"`
	Foundations [RussianBankFoundationCnt][]*Card `json:"fd"`
	Config      RussianBankConfig                 `json:"cf"`
	Phase       RussianBankPhase                  `json:"ph"`
	Current     int                               `json:"cu"`
	Winner      int                               `json:"wn"`
	MoveCount   int                               `json:"mc"`
	PassStreak  int                               `json:"ks"`
	StopPoints  [RussianBankPlayerCnt]int         `json:"sp"`
	ActionLog   []*ActionLogEntry                 `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*russianBankSnapshot `json:"hi,omitempty"`
}

// boardJSON projects the board into the wire struct WITHOUT the undo history.
// takeSnapshot marshals this; MarshalJSON adds the history on top.
func (g *RussianBank) boardJSON() russianBankJSON {
	return russianBankJSON{
		Players:     g.players,
		Tableau:     g.tableau,
		Foundations: g.foundations,
		Config:      g.config,
		Phase:       g.phase,
		Current:     g.current,
		Winner:      g.winner,
		MoveCount:   g.moveCount,
		PassStreak:  g.passStreak,
		StopPoints:  g.stopPoints,
		ActionLog:   g.actionLog,
	}
}

// MarshalJSON implements json.Marshaler.
func (g *RussianBank) MarshalJSON() ([]byte, error) {
	j := g.boardJSON()
	j.History = g.history
	return json.Marshal(j)
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *RussianBank) UnmarshalJSON(data []byte) error {
	var j russianBankJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Players) != RussianBankPlayerCnt {
		return errRussianBank
	}
	for _, p := range j.Players {
		if p == nil {
			return errRussianBank // 改竄された null プレイヤーで nil 参照を防ぐ
		}
	}
	if len(j.ActionLog) > russianBankMaxSliceLen {
		return errRussianBank
	}
	if j.Current < 0 || j.Current >= RussianBankPlayerCnt {
		return errRussianBank
	}
	if err := j.Config.Validate(); err != nil {
		return errRussianBank // 不正な設定値の読み込みを拒否する
	}
	g.players = j.Players
	g.tableau = j.Tableau
	g.foundations = j.Foundations
	g.config = j.Config
	g.phase = j.Phase
	g.current = j.Current
	g.winner = j.Winner
	g.moveCount = j.MoveCount
	g.passStreak = j.PassStreak
	g.stopPoints = j.StopPoints
	if len(j.History) > russianBankMaxSliceLen {
		return errors.New("russianbank: history exceeds maximum allowed size")
	}
	g.actionLog = j.ActionLog
	g.history = j.History
	return nil
}

// russianBankSnapshotMaxBytes bounds one embedded snapshot document. A full
// two-player board is a few KB, so 1 MiB is far above any legitimate value and
// exists only to stop a hostile KV entry from being expanded.
const russianBankSnapshotMaxBytes = 1 << 20

// russianBankSnapshotJSON is the wire format for a single undo snapshot.
// russianBankSnapshot holds one unexported field, so marshalling it directly
// would emit `[{},{}]` -- the undo depth would survive but every snapshot would
// be blank, and Undo would wipe the board instead of rewinding it (#4478).
//
// The field is already a JSON document, so it rides as json.RawMessage rather
// than a []byte: a []byte would be re-encoded as base64 on every save, which
// costs a third more bytes per snapshot in KV for no benefit.
type russianBankSnapshotJSON struct {
	State json.RawMessage `json:"st"`
}

// MarshalJSON implements json.Marshaler for russianBankSnapshot.
func (s *russianBankSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(russianBankSnapshotJSON{State: s.stateJSON})
}

// UnmarshalJSON implements json.Unmarshaler for russianBankSnapshot.
func (s *russianBankSnapshot) UnmarshalJSON(data []byte) error {
	var j russianBankSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.State) > russianBankSnapshotMaxBytes {
		return errors.New("russianbank: snapshot state exceeds maximum allowed size")
	}
	s.stateJSON = j.State
	return nil
}
