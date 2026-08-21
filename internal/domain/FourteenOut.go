//go:build !js || !wasm || extra4

// Package domain provides core game domain models.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FourteenOutPhase はフォーティーンアウト・ソリティアのフェーズを表す。
type FourteenOutPhase int

// FourteenOutのフェーズ定数
const (
	// FourteenOutPhasePlaying プレイ中
	FourteenOutPhasePlaying FourteenOutPhase = iota
	// FourteenOutPhaseGameClear ゲームクリア (全 52 枚を取り除いた)
	FourteenOutPhaseGameClear
	// FourteenOutPhaseGameOver ゲームオーバー (手詰まり or ギブアップ)
	FourteenOutPhaseGameOver
)

// FourteenOutColumnCnt は列数 (12)。
const FourteenOutColumnCnt = 12

// FourteenOutLongColumns は 5 枚配る列の数。
//
// 52 = 12*4 + 4 なので、**左から 4 列だけが 5 枚**で残り 8 列が 4 枚になる。
// 「ほぼ均等」ではなく、どの列が長いかまで決まっている。
const FourteenOutLongColumns = 4

// FourteenOutTargetSum は取り除ける組の合計値 (14)。
//
// **K と A の組は特例ではない。**13+1=14 という一般規則そのもの。K 同士 (26) も
// K と他ランクも 14 にならないので、専用の分岐を書くと必ず一般規則と同じ答えを
// 返す枝が増えるだけになる。同ランクで組めるのは 7 と 7 だけ。
const FourteenOutTargetSum = 14

// FourteenOutDeckSize は使用デッキ枚数 (1 デッキ = 52)。
const FourteenOutDeckSize = CardCnt

// FourteenOutHintActionRemove はヒントアクション「ペア除去」。
const FourteenOutHintActionRemove = "remove"

// FourteenOutHint はヒント情報。
type FourteenOutHint struct {
	// Action は推奨アクション ("remove")。
	Action string
	// FromCol/ToCol は取り除ける 2 列の番号。動かせるのは各列の末尾だけ。
	FromCol int
	ToCol   int
}

// FourteenOut はフォーティーンアウト・ソリティアのゲーム状態を表す。
type FourteenOut struct {
	trumpCards   *TrumpCards
	columns      [][]*Card
	phase        FourteenOutPhase
	removedCount int
	actionLog    []*ActionLogEntry
	history      []*fourteenOutSnapshot
	isStalemate  bool
}

// fourteenOutSnapshot はアンドゥ用のスナップショット。
type fourteenOutSnapshot struct {
	columns      [][]*Card
	removedCount int
	phase        FourteenOutPhase
	isStalemate  bool
	actionLogLn  int
}

// NewFourteenOut はコンストラクタ。
func NewFourteenOut(tc *TrumpCards) *FourteenOut {
	return &FourteenOut{trumpCards: tc}
}

// NewDefaultFourteenOut は標準 1 デッキの FourteenOut を返す。
// CUI / Web / Worker の構築箇所から共通に呼ばれる SSoT。
func NewDefaultFourteenOut() *FourteenOut {
	return NewFourteenOut(NewTrumpCards(0))
}

// Reset はゲームを初期化する。デッキをシャッフルして 25 枚を盤面に配る。
func (m *FourteenOut) Reset() {
	m.trumpCards.Shuffle()
	m.phase = FourteenOutPhasePlaying
	m.removedCount = 0
	m.actionLog = nil
	m.history = nil
	m.isStalemate = false

	// **52 枚を配り切る。山札は残らない。**左から 4 列が 5 枚、残り 8 列が 4 枚。
	m.columns = make([][]*Card, FourteenOutColumnCnt)
	for i := range FourteenOutColumnCnt {
		size := 4
		if i < FourteenOutLongColumns {
			size = 5
		}
		col := make([]*Card, 0, size)
		for range size {
			if card := m.trumpCards.DrawCard(); card != nil {
				col = append(col, card)
			}
		}
		m.columns[i] = col
	}
	m.checkFourteenOutStalemate()
}

// Remove は 2 列の末尾札を、合計が 14 なら取り除く。
//
// **隣接は関係ない。**クローン元の Monte Carlo はグリッド上で隣り合うセルしか
// 組めないが、Fourteen Out は露出している末尾同士ならどの 2 列でも組める。
func (m *FourteenOut) Remove(c1, c2 int) error {
	if m.phase != FourteenOutPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if c1 == c2 {
		return errors.New("cannot select the same column twice")
	}
	if !m.inBounds(c1) || !m.inBounds(c2) {
		return errors.New("invalid column")
	}
	a, b := m.tail(c1), m.tail(c2)
	if a == nil || b == nil {
		return errors.New("column is empty")
	}
	if a.GetValue()+b.GetValue() != FourteenOutTargetSum {
		return errors.New("cards do not sum to 14")
	}

	m.takeSnapshot()
	m.columns[c1] = m.columns[c1][:len(m.columns[c1])-1]
	m.columns[c2] = m.columns[c2][:len(m.columns[c2])-1]
	m.removedCount += 2
	m.appendLog("remove", fmt.Sprintf("列%d と 列%d を取り除いた", c1, c2), []*Card{a, b})
	m.checkGameClear()
	m.checkFourteenOutStalemate()
	return nil
}

// Undo は直前の Remove または Deal を取り消す。
func (m *FourteenOut) Undo() error {
	if len(m.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	// **山札を巻き戻す処理は要らない。**Reset で 52 枚すべて配り切るので、
	// プレイ中に山札から引くことは一度も無い。
	m.columns = snap.columns
	m.removedCount = snap.removedCount
	m.phase = snap.phase
	m.isStalemate = snap.isStalemate
	if len(m.actionLog) > snap.actionLogLn {
		m.actionLog = m.actionLog[:snap.actionLogLn]
	}
	return nil
}

// CanUndo はアンドゥ可能かを返す。
func (m *FourteenOut) CanUndo() bool {
	return len(m.history) > 0 && m.phase == FourteenOutPhasePlaying
}

// GiveUp はゲームを放棄する。
func (m *FourteenOut) GiveUp() {
	if m.phase == FourteenOutPhasePlaying {
		m.phase = FourteenOutPhaseGameOver
		m.appendLog("giveup", "ギブアップしました", nil)
	}
}

// Hint は推奨手を返す。合計 14 の組があれば "remove"、無ければ nil。
//
// **"deal" のフォールバックは無い。**クローン元の Monte Carlo は山札から補充
// できたが、Fourteen Out は最初に配り切るので、組が無ければそれが詰み。
func (m *FourteenOut) Hint() *FourteenOutHint {
	if m.phase != FourteenOutPhasePlaying {
		return nil
	}
	return m.findRemovablePair()
}

// CheckFourteenOutStalemate は外部から手詰まり判定を再評価するための公開ラッパー。
func (m *FourteenOut) CheckFourteenOutStalemate() {
	m.checkFourteenOutStalemate()
}

// --- Getters / setters ---

// GetPhase はフェーズを返す。
func (m *FourteenOut) GetPhase() FourteenOutPhase { return m.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (m *FourteenOut) SetPhase(phase FourteenOutPhase) { m.phase = phase }

// GetColumns は各列を返す (末尾が露出している札)。
func (m *FourteenOut) GetColumns() [][]*Card { return m.columns }

// SetColumns は列を設定する (テスト用)。
func (m *FourteenOut) SetColumns(cols [][]*Card) { m.columns = cols }

// GetRemovedCount は取り除いた累計枚数を返す。
func (m *FourteenOut) GetRemovedCount() int { return m.removedCount }

// SetRemovedCount は取り除いた累計枚数を設定する (テスト用)。
func (m *FourteenOut) SetRemovedCount(n int) { m.removedCount = n }

// GetActionLog は棋譜を返す。
func (m *FourteenOut) GetActionLog() []*ActionLogEntry { return m.actionLog }

// GetGameEndFlag はゲームが終了フェーズに入ったかを返す。
func (m *FourteenOut) GetGameEndFlag() bool { return m.phase != FourteenOutPhasePlaying }

// IsComplete はゲームクリア状態かを返す。
func (m *FourteenOut) IsComplete() bool { return m.removedCount >= FourteenOutDeckSize }

// IsStalemate は手詰まり状態かを返す。
func (m *FourteenOut) IsStalemate() bool { return m.isStalemate }

// SetIsStalemate は手詰まり状態を設定する (テスト用)。
func (m *FourteenOut) SetIsStalemate(v bool) { m.isStalemate = v }

// --- Private helpers ---

// inBounds は列番号が範囲内かを返す。
func (m *FourteenOut) inBounds(c int) bool {
	return c >= 0 && c < len(m.columns)
}

// tail は列 c の露出している札 (末尾)。空列なら nil。
func (m *FourteenOut) tail(c int) *Card {
	if !m.inBounds(c) || len(m.columns[c]) == 0 {
		return nil
	}
	return m.columns[c][len(m.columns[c])-1]
}

// findRemovablePair は最初に見つかった合計 14 の組を返す。
func (m *FourteenOut) findRemovablePair() *FourteenOutHint {
	var found *FourteenOutHint
	m.forEachRemovablePair(func(c1, c2 int) bool {
		found = &FourteenOutHint{Action: FourteenOutHintActionRemove, FromCol: c1, ToCol: c2}
		return false // 最初の 1 組で止める
	})
	return found
}

// forEachRemovablePair は取り除ける組を 1 回ずつ訪ねる。fn が false を返すと止まる。
//
// **c2 > c1 の側だけを見る**ので、同じ組を 2 度数えない。ヒントも手詰まり判定も
// 件数もここを通る ── 規則を 3 箇所に置くと、片方だけ直したときに「取れる組が
// あるのに手詰まり」になる (#5587)。
//
// 隣接は見ない。露出している末尾同士なら、離れた列でも組める。
func (m *FourteenOut) forEachRemovablePair(fn func(c1, c2 int) bool) {
	for c1 := range m.columns {
		a := m.tail(c1)
		if a == nil {
			continue
		}
		for c2 := c1 + 1; c2 < len(m.columns); c2++ {
			b := m.tail(c2)
			if b == nil || a.GetValue()+b.GetValue() != FourteenOutTargetSum {
				continue
			}
			if !fn(c1, c2) {
				return
			}
		}
	}
}

// CountRemovablePairs は盤面に残っている取り除ける組の数を返す。
//
// **これがこのゲームの判断材料そのもの。**Web は常時カウンタとして出している
// のに、CUI は 25 マスを目で走査させていた (#5587)。ヒントと同じ走査を使う。
func (m *FourteenOut) CountRemovablePairs() int {
	count := 0
	m.forEachRemovablePair(func(_, _ int) bool {
		count++
		return true
	})
	return count
}

// checkGameClear はクリア判定。
func (m *FourteenOut) checkGameClear() {
	if m.removedCount >= FourteenOutDeckSize {
		m.phase = FourteenOutPhaseGameClear
	}
}

// checkFourteenOutStalemate は手詰まり判定。
func (m *FourteenOut) checkFourteenOutStalemate() {
	if m.phase != FourteenOutPhasePlaying {
		return
	}
	// **山札からの救済は無い。**組が見つからなければそれが詰み。
	m.isStalemate = m.findRemovablePair() == nil
}

// takeSnapshot は現在の状態をスナップショットとして保存する。
func (m *FourteenOut) takeSnapshot() {
	// **列は深いコピーを取る。**スライスをそのまま渡すと、Remove の再スライスは
	// 元の配列を共有したままなので、Undo しても札が戻らないことがある。
	cols := make([][]*Card, len(m.columns))
	for i, col := range m.columns {
		cols[i] = make([]*Card, len(col))
		copy(cols[i], col)
	}
	m.history = append(m.history, &fourteenOutSnapshot{
		columns:      cols,
		removedCount: m.removedCount,
		phase:        m.phase,
		isStalemate:  m.isStalemate,
		actionLogLn:  len(m.actionLog),
	})
}

func (m *FourteenOut) appendLog(actionType, detail string, cards []*Card) {
	m.actionLog = append(m.actionLog, &ActionLogEntry{
		TurnNumber: m.removedCount / 2,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// fourteenOutJSON はシリアライズ用のワイヤーフォーマット。
type fourteenOutJSON struct {
	TrumpCards   *TrumpCards            `json:"tc"`
	Columns      [][]*Card              `json:"cl"`
	Phase        FourteenOutPhase       `json:"ps"`
	RemovedCount int                    `json:"rc"`
	IsStalemate  bool                   `json:"sl"`
	ActionLog    []*ActionLogEntry      `json:"al"`
	History      []*fourteenOutSnapshot `json:"hi,omitempty"`
}

// fourteenOutSnapshotJSON is the wire format for a single undo snapshot.
// fourteenOutSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names reuse
// fourteenOutJSON's short keys where possible to keep the KV payload
// compact (#1654/#1860).
type fourteenOutSnapshotJSON struct {
	Columns      [][]*Card        `json:"cl"`
	RemovedCount int              `json:"rc"`
	Phase        FourteenOutPhase `json:"ps"`
	IsStalemate  bool             `json:"sl"`
	ActionLogLn  int              `json:"ll"`
}

// MarshalJSON implements json.Marshaler for fourteenOutSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// FourteenOut.MarshalJSON can persist the undo history (#1860).
func (s *fourteenOutSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(fourteenOutSnapshotJSON{
		Columns:      s.columns,
		RemovedCount: s.removedCount,
		Phase:        s.phase,
		IsStalemate:  s.isStalemate,
		ActionLogLn:  s.actionLogLn,
	})
}

// UnmarshalJSON implements json.Unmarshaler for fourteenOutSnapshot.
// ActionLogLn must be in [0, fourteenOutMaxSliceLen], and the column count must
// not exceed the dealt layout. Out-of-range values are rejected at the trust
// boundary so malformed KV payloads cannot crash the worker via index/slice
// panic inside Undo().
func (s *fourteenOutSnapshot) UnmarshalJSON(data []byte) error {
	var j fourteenOutSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.ActionLogLn < 0 || j.ActionLogLn > fourteenOutMaxSliceLen {
		return fmt.Errorf("fourteenout: snapshot actionLogLn out of range")
	}
	if len(j.Columns) > FourteenOutColumnCnt {
		return fmt.Errorf("fourteenout: snapshot column count out of range")
	}
	s.columns = j.Columns
	s.removedCount = j.RemovedCount
	s.phase = j.Phase
	s.isStalemate = j.IsStalemate
	s.actionLogLn = j.ActionLogLn
	return nil
}

// MarshalJSON implements json.Marshaler.
func (m *FourteenOut) MarshalJSON() ([]byte, error) {
	return json.Marshal(fourteenOutJSON{
		TrumpCards:   m.trumpCards,
		Columns:      m.columns,
		Phase:        m.phase,
		RemovedCount: m.removedCount,
		IsStalemate:  m.isStalemate,
		ActionLog:    m.actionLog,
		History:      m.history,
	})
}

// fourteenOutMaxSliceLen はデシリアライズ時のスライスサイズ上限。
const fourteenOutMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (m *FourteenOut) UnmarshalJSON(data []byte) error {
	var j fourteenOutJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > fourteenOutMaxSliceLen || len(j.History) > fourteenOutMaxSliceLen {
		return fmt.Errorf("fourteenout: input array exceeds maximum allowed size")
	}
	m.trumpCards = j.TrumpCards
	if m.trumpCards == nil {
		m.trumpCards = NewTrumpCards(0)
	}
	if len(j.Columns) > FourteenOutColumnCnt {
		return fmt.Errorf("fourteenout: column count exceeds the dealt layout")
	}
	m.columns = j.Columns
	m.phase = j.Phase
	m.removedCount = j.RemovedCount
	m.isStalemate = j.IsStalemate
	m.actionLog = j.ActionLog
	if m.actionLog == nil {
		m.actionLog = make([]*ActionLogEntry, 0)
	}
	m.history = j.History
	if m.history == nil {
		m.history = make([]*fourteenOutSnapshot, 0)
	}
	return nil
}
