//go:build !js || !wasm || solo

// Package domain provides core game domain models.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MonteCarloPhase はモンテカルロ・ソリティアのフェーズを表す。
type MonteCarloPhase int

// MonteCarloのフェーズ定数
const (
	// MonteCarloPhasePlaying プレイ中
	MonteCarloPhasePlaying MonteCarloPhase = iota
	// MonteCarloPhaseGameClear ゲームクリア (全 52 枚を取り除いた)
	MonteCarloPhaseGameClear
	// MonteCarloPhaseGameOver ゲームオーバー (手詰まり or ギブアップ)
	MonteCarloPhaseGameOver
)

// MonteCarloGridSize はグリッドの一辺のサイズ (5)。
const MonteCarloGridSize = 5

// MonteCarloTotalCells は総セル数 (5x5=25)。
const MonteCarloTotalCells = MonteCarloGridSize * MonteCarloGridSize

// MonteCarloDeckSize は使用デッキ枚数 (1 デッキ = 52)。
const MonteCarloDeckSize = CardCnt

// MonteCarloHintActionRemove はヒントアクション「ペア除去」。
const MonteCarloHintActionRemove = "remove"

// MonteCarloHintActionDeal はヒントアクション「山札補充」。
const MonteCarloHintActionDeal = "deal"

// MonteCarloHint はヒント情報。
type MonteCarloHint struct {
	// Action は推奨アクション ("remove" or "deal")。
	Action string
	// FromR/FromC/ToR/ToC は Action="remove" 時の対象セル座標。
	FromR int
	FromC int
	ToR   int
	ToC   int
}

// MonteCarlo はモンテカルロ・ソリティアのゲーム状態を表す。
type MonteCarlo struct {
	trumpCards   *TrumpCards
	board        [MonteCarloGridSize][MonteCarloGridSize]*Card
	phase        MonteCarloPhase
	removedCount int
	dealCount    int
	actionLog    []*ActionLogEntry
	history      []*monteCarloSnapshot
	isStalemate  bool
}

// monteCarloSnapshot はアンドゥ用のスナップショット。
type monteCarloSnapshot struct {
	board        [MonteCarloGridSize][MonteCarloGridSize]*Card
	removedCount int
	dealCount    int
	phase        MonteCarloPhase
	isStalemate  bool
	deckDrawCnt  int
	actionLogLn  int
}

// NewMonteCarlo はコンストラクタ。
func NewMonteCarlo(tc *TrumpCards) *MonteCarlo {
	return &MonteCarlo{trumpCards: tc}
}

// NewDefaultMonteCarlo は標準 1 デッキの MonteCarlo を返す。
// CUI / Web / Worker の構築箇所から共通に呼ばれる SSoT。
func NewDefaultMonteCarlo() *MonteCarlo {
	return NewMonteCarlo(NewTrumpCards(0))
}

// Reset はゲームを初期化する。デッキをシャッフルして 25 枚を盤面に配る。
func (m *MonteCarlo) Reset() {
	m.trumpCards.Shuffle()
	m.board = [MonteCarloGridSize][MonteCarloGridSize]*Card{}
	m.phase = MonteCarloPhasePlaying
	m.removedCount = 0
	m.dealCount = 0
	m.actionLog = nil
	m.history = nil
	m.isStalemate = false
	m.fillBoardFromStock()
	m.checkMonteCarloStalemate()
}

// Remove は隣接する同ランクの 2 枚のペアを盤面から取り除く。
func (m *MonteCarlo) Remove(r1, c1, r2, c2 int) error {
	if m.phase != MonteCarloPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if r1 == r2 && c1 == c2 {
		return errors.New("cannot select the same cell twice")
	}
	if !m.inBounds(r1, c1) || !m.inBounds(r2, c2) {
		return errors.New("invalid cell position")
	}
	a := m.board[r1][c1]
	b := m.board[r2][c2]
	if a == nil || b == nil {
		return errors.New("cell is empty")
	}
	if absInt(r1-r2) > 1 || absInt(c1-c2) > 1 {
		return errors.New("cells are not adjacent")
	}
	if a.GetValue() != b.GetValue() {
		return errors.New("cards do not match in rank")
	}

	m.takeSnapshot()
	m.board[r1][c1] = nil
	m.board[r2][c2] = nil
	m.removedCount += 2
	m.appendLog("remove", fmt.Sprintf("(%d,%d) と (%d,%d) を取り除いた", r1, c1, r2, c2), []*Card{a, b})
	m.checkGameClear()
	m.checkMonteCarloStalemate()
	return nil
}

// Deal は盤面を左上に詰め直し、空いたセルを山札から補充する。
func (m *MonteCarlo) Deal() error {
	if m.phase != MonteCarloPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	m.takeSnapshot()
	m.compressBoard()
	m.fillBoardFromStock()
	m.dealCount++
	m.appendLog("deal", "山札からの補充", nil)
	m.checkGameClear()
	m.checkMonteCarloStalemate()
	return nil
}

// Undo は直前の Remove または Deal を取り消す。
func (m *MonteCarlo) Undo() error {
	if len(m.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	for i := snap.deckDrawCnt; i < m.trumpCards.deckDrawCnt; i++ {
		m.trumpCards.deck[i].SetDraw(false)
	}
	m.trumpCards.deckDrawCnt = snap.deckDrawCnt
	m.board = snap.board
	m.removedCount = snap.removedCount
	m.dealCount = snap.dealCount
	m.phase = snap.phase
	m.isStalemate = snap.isStalemate
	if len(m.actionLog) > snap.actionLogLn {
		m.actionLog = m.actionLog[:snap.actionLogLn]
	}
	return nil
}

// CanUndo はアンドゥ可能かを返す。
func (m *MonteCarlo) CanUndo() bool {
	return len(m.history) > 0 && m.phase == MonteCarloPhasePlaying
}

// GiveUp はゲームを放棄する。
func (m *MonteCarlo) GiveUp() {
	if m.phase == MonteCarloPhasePlaying {
		m.phase = MonteCarloPhaseGameOver
		m.appendLog("giveup", "ギブアップしました", nil)
	}
}

// Hint は推奨手を返す。隣接同ランクのペアがあれば "remove"、
// 無ければ山札補充可能なら "deal"、いずれも不可なら nil。
func (m *MonteCarlo) Hint() *MonteCarloHint {
	if m.phase != MonteCarloPhasePlaying {
		return nil
	}
	if pair := m.findAdjacentPair(); pair != nil {
		return pair
	}
	if m.trumpCards.GetRemainingCount() > 0 || m.hasCompressionGap() {
		return &MonteCarloHint{Action: MonteCarloHintActionDeal}
	}
	return nil
}

// CheckMonteCarloStalemate は外部から手詰まり判定を再評価するための公開ラッパー。
func (m *MonteCarlo) CheckMonteCarloStalemate() {
	m.checkMonteCarloStalemate()
}

// --- Getters / setters ---

// GetPhase はフェーズを返す。
func (m *MonteCarlo) GetPhase() MonteCarloPhase { return m.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (m *MonteCarlo) SetPhase(phase MonteCarloPhase) { m.phase = phase }

// GetBoard はボードを返す。
func (m *MonteCarlo) GetBoard() [MonteCarloGridSize][MonteCarloGridSize]*Card {
	return m.board
}

// SetBoard はボードを設定する (テスト用)。
func (m *MonteCarlo) SetBoard(b [MonteCarloGridSize][MonteCarloGridSize]*Card) {
	m.board = b
}

// GetStockCount は残りの山札枚数を返す。
func (m *MonteCarlo) GetStockCount() int {
	return m.trumpCards.GetRemainingCount()
}

// GetRemovedCount は取り除いた累計枚数を返す。
func (m *MonteCarlo) GetRemovedCount() int { return m.removedCount }

// SetRemovedCount は取り除いた累計枚数を設定する (テスト用)。
func (m *MonteCarlo) SetRemovedCount(n int) { m.removedCount = n }

// GetDealCount は Deal を実行した回数を返す。
func (m *MonteCarlo) GetDealCount() int { return m.dealCount }

// GetActionLog は棋譜を返す。
func (m *MonteCarlo) GetActionLog() []*ActionLogEntry { return m.actionLog }

// GetGameEndFlag はゲームが終了フェーズに入ったかを返す。
func (m *MonteCarlo) GetGameEndFlag() bool { return m.phase != MonteCarloPhasePlaying }

// IsComplete はゲームクリア状態かを返す。
func (m *MonteCarlo) IsComplete() bool { return m.removedCount >= MonteCarloDeckSize }

// IsStalemate は手詰まり状態かを返す。
func (m *MonteCarlo) IsStalemate() bool { return m.isStalemate }

// SetIsStalemate は手詰まり状態を設定する (テスト用)。
func (m *MonteCarlo) SetIsStalemate(v bool) { m.isStalemate = v }

// --- Private helpers ---

// inBounds は座標が盤面内かを返す。
func (m *MonteCarlo) inBounds(r, c int) bool {
	return r >= 0 && r < MonteCarloGridSize && c >= 0 && c < MonteCarloGridSize
}

// fillBoardFromStock は盤面の空セルを行優先で山札から補充する。
func (m *MonteCarlo) fillBoardFromStock() {
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			if m.board[r][c] != nil {
				continue
			}
			if m.trumpCards.GetRemainingCount() == 0 {
				return
			}
			m.board[r][c] = m.trumpCards.DrawCard()
		}
	}
}

// compressBoard は非 nil カードを行優先で前詰めする。
func (m *MonteCarlo) compressBoard() {
	cards := make([]*Card, 0, MonteCarloTotalCells)
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			if m.board[r][c] != nil {
				cards = append(cards, m.board[r][c])
			}
		}
	}
	m.board = [MonteCarloGridSize][MonteCarloGridSize]*Card{}
	for i, card := range cards {
		m.board[i/MonteCarloGridSize][i%MonteCarloGridSize] = card
	}
}

// hasCompressionGap は行優先で「nil の後に非 nil が出現する」隙間があるかを返す。
func (m *MonteCarlo) hasCompressionGap() bool {
	seenNil := false
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			if m.board[r][c] == nil {
				seenNil = true
				continue
			}
			if seenNil {
				return true
			}
		}
	}
	return false
}

// findAdjacentPair は最初に見つかった隣接同ランクペアの座標を返す。
// 走査順序は行優先 (0,0)→(0,1)→...→(4,4) で、各セルに対して
// 8 方向のうち「インデックスが大きい」方のセル (右/下/右下/左下) を確認する。
// これにより各ペアが一度だけ評価される。
func (m *MonteCarlo) findAdjacentPair() *MonteCarloHint {
	var found *MonteCarloHint
	m.forEachRemovablePair(func(r, c, nr, nc int) bool {
		found = &MonteCarloHint{
			Action: MonteCarloHintActionRemove,
			FromR:  r, FromC: c,
			ToR: nr, ToC: nc,
		}
		return false // 最初の 1 組で止める
	})
	return found
}

// forEachRemovablePair は取り除ける組を 1 回ずつ訪ねる。fn が false を返すと止まる。
//
// **前向きの 4 方向だけを見る**ので、同じ組を 2 度数えない (右・左下・下・右下)。
// ヒントも手詰まり判定も件数もここを通る ── 隣接と同ランクの規則を 3 箇所に
// 置くと、片方だけ直したときに「取れる組があるのに手詰まり」になる (#5587)。
func (m *MonteCarlo) forEachRemovablePair(fn func(r, c, nr, nc int) bool) {
	dirs := []struct{ dr, dc int }{
		{0, 1},  // right
		{1, -1}, // down-left
		{1, 0},  // down
		{1, 1},  // down-right
	}
	for r := range MonteCarloGridSize {
		for c := range MonteCarloGridSize {
			a := m.board[r][c]
			if a == nil {
				continue
			}
			for _, d := range dirs {
				nr, nc := r+d.dr, c+d.dc
				if !m.inBounds(nr, nc) {
					continue
				}
				b := m.board[nr][nc]
				if b == nil || a.GetValue() != b.GetValue() {
					continue
				}
				if !fn(r, c, nr, nc) {
					return
				}
			}
		}
	}
}

// CountRemovablePairs は盤面に残っている取り除ける組の数を返す。
//
// **これがこのゲームの判断材料そのもの。**Web は常時カウンタとして出している
// のに、CUI は 25 マスを目で走査させていた (#5587)。ヒントと同じ走査を使う。
func (m *MonteCarlo) CountRemovablePairs() int {
	count := 0
	m.forEachRemovablePair(func(_, _, _, _ int) bool {
		count++
		return true
	})
	return count
}

// checkGameClear はクリア判定。
func (m *MonteCarlo) checkGameClear() {
	if m.removedCount >= MonteCarloDeckSize {
		m.phase = MonteCarloPhaseGameClear
	}
}

// checkMonteCarloStalemate は手詰まり判定。
func (m *MonteCarlo) checkMonteCarloStalemate() {
	if m.phase != MonteCarloPhasePlaying {
		return
	}
	if m.findAdjacentPair() != nil {
		m.isStalemate = false
		return
	}
	if m.trumpCards.GetRemainingCount() > 0 || m.hasCompressionGap() {
		m.isStalemate = false
		return
	}
	m.isStalemate = true
}

// takeSnapshot は現在の状態をスナップショットとして保存する。
func (m *MonteCarlo) takeSnapshot() {
	m.history = appendSnapshot(m.history, &monteCarloSnapshot{
		board:        m.board,
		removedCount: m.removedCount,
		dealCount:    m.dealCount,
		phase:        m.phase,
		isStalemate:  m.isStalemate,
		deckDrawCnt:  m.trumpCards.deckDrawCnt,
		actionLogLn:  len(m.actionLog),
	})
}

// appendLog は棋譜エントリを追加する。
func (m *MonteCarlo) appendLog(actionType, detail string, cards []*Card) {
	m.actionLog = append(m.actionLog, &ActionLogEntry{
		TurnNumber: m.removedCount/2 + m.dealCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// absInt は int の絶対値を返す。
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// monteCarloJSON はシリアライズ用のワイヤーフォーマット。
type monteCarloJSON struct {
	TrumpCards   *TrumpCards                                   `json:"tc"`
	Board        [MonteCarloGridSize][MonteCarloGridSize]*Card `json:"bd"`
	Phase        MonteCarloPhase                               `json:"ps"`
	RemovedCount int                                           `json:"rc"`
	DealCount    int                                           `json:"dc"`
	IsStalemate  bool                                          `json:"sl"`
	ActionLog    []*ActionLogEntry                             `json:"al"`
	History      []*monteCarloSnapshot                         `json:"hi,omitempty"`
}

// monteCarloSnapshotJSON is the wire format for a single undo snapshot.
// monteCarloSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names reuse
// monteCarloJSON's short keys where possible to keep the KV payload
// compact (#1654/#1860).
type monteCarloSnapshotJSON struct {
	Board        [MonteCarloGridSize][MonteCarloGridSize]*Card `json:"bd"`
	RemovedCount int                                           `json:"rc"`
	DealCount    int                                           `json:"dc"`
	Phase        MonteCarloPhase                               `json:"ps"`
	IsStalemate  bool                                          `json:"sl"`
	DeckDrawCnt  int                                           `json:"dd"`
	ActionLogLn  int                                           `json:"ll"`
}

// MarshalJSON implements json.Marshaler for monteCarloSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// MonteCarlo.MarshalJSON can persist the undo history (#1860).
func (s *monteCarloSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(monteCarloSnapshotJSON{
		Board:        s.board,
		RemovedCount: s.removedCount,
		DealCount:    s.dealCount,
		Phase:        s.phase,
		IsStalemate:  s.isStalemate,
		DeckDrawCnt:  s.deckDrawCnt,
		ActionLogLn:  s.actionLogLn,
	})
}

// UnmarshalJSON implements json.Unmarshaler for monteCarloSnapshot.
// DeckDrawCnt must be in [0, MonteCarloDeckSize]; ActionLogLn must be in
// [0, monteCarloMaxSliceLen]. Out-of-range values are rejected at the trust
// boundary so malformed KV payloads cannot crash the worker via index/slice
// panic inside Undo().
func (s *monteCarloSnapshot) UnmarshalJSON(data []byte) error {
	var j monteCarloSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.DeckDrawCnt < 0 || j.DeckDrawCnt > MonteCarloDeckSize {
		return fmt.Errorf("montecarlo: snapshot deckDrawCnt out of range")
	}
	if j.ActionLogLn < 0 || j.ActionLogLn > monteCarloMaxSliceLen {
		return fmt.Errorf("montecarlo: snapshot actionLogLn out of range")
	}
	s.board = j.Board
	s.removedCount = j.RemovedCount
	s.dealCount = j.DealCount
	s.phase = j.Phase
	s.isStalemate = j.IsStalemate
	s.deckDrawCnt = j.DeckDrawCnt
	s.actionLogLn = j.ActionLogLn
	return nil
}

// MarshalJSON implements json.Marshaler.
func (m *MonteCarlo) MarshalJSON() ([]byte, error) {
	return json.Marshal(monteCarloJSON{
		TrumpCards:   m.trumpCards,
		Board:        m.board,
		Phase:        m.phase,
		RemovedCount: m.removedCount,
		DealCount:    m.dealCount,
		IsStalemate:  m.isStalemate,
		ActionLog:    m.actionLog,
		History:      m.history,
	})
}

// monteCarloMaxSliceLen はデシリアライズ時のスライスサイズ上限。
const monteCarloMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (m *MonteCarlo) UnmarshalJSON(data []byte) error {
	var j monteCarloJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > monteCarloMaxSliceLen || len(j.History) > monteCarloMaxSliceLen {
		return fmt.Errorf("montecarlo: input array exceeds maximum allowed size")
	}
	m.trumpCards = j.TrumpCards
	if m.trumpCards == nil {
		m.trumpCards = NewTrumpCards(0)
	}
	m.board = j.Board
	m.phase = j.Phase
	m.removedCount = j.RemovedCount
	m.dealCount = j.DealCount
	m.isStalemate = j.IsStalemate
	m.actionLog = j.ActionLog
	if m.actionLog == nil {
		m.actionLog = make([]*ActionLogEntry, 0)
	}
	m.history = j.History
	if m.history == nil {
		m.history = make([]*monteCarloSnapshot, 0)
	}
	return nil
}
