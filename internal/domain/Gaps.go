//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// GapsPhase はGaps（Montana）ゲームのフェーズを表す。
type GapsPhase int

// Gapsのフェーズ定数。
const (
	// GapsPhasePlaying プレイ中
	GapsPhasePlaying GapsPhase = iota
	// GapsPhaseGameClear 全行が2..Kでソート済み
	GapsPhaseGameClear
	// GapsPhaseGameOver ギブアップまたは手詰まり+再配り不可
	GapsPhaseGameOver
)

// Gapsの盤面定数。
const (
	// GapsRowCnt 盤面の行数（=4スート）
	GapsRowCnt = 4
	// GapsColCnt 盤面の列数（2..Kの12枚分 +空き列1）
	GapsColCnt = 13
	// GapsMaxRedeals 再配りの最大回数
	GapsMaxRedeals = 3
	// GapsAnchorRank 左端列に置けるカードのランク（=2）
	GapsAnchorRank = 2
	// GapsKingRank Kのランク値
	GapsKingRank = 13
)

// GapsCell は盤面のセル。nilは「隙間（gap）」を表す。Aceは盤上に現れない。
type GapsCell = *Card

// GapsHint はヒントとして提示する「ある手」を表す。
type GapsHint struct {
	FromRow int `json:"fr"`
	FromCol int `json:"fc"`
	ToRow   int `json:"tr"`
	ToCol   int `json:"tc"`
}

// Gaps はGaps（別名Montana/Addiction）ソリティアの状態。
type Gaps struct {
	trumpCards  *TrumpCards
	grid        [GapsRowCnt][GapsColCnt]GapsCell
	phase       GapsPhase
	moveCount   int
	redealsUsed int
	actionLog   []*ActionLogEntry
	history     []*gapsSnapshot
	isStalemate bool
}

// gapsSnapshot はUndo用の状態スナップショット。
type gapsSnapshot struct {
	grid        [GapsRowCnt][GapsColCnt]GapsCell
	phase       GapsPhase
	moveCount   int
	redealsUsed int
	isStalemate bool
}

// NewGaps はTrumpCardsを注入してGapsを生成する。
func NewGaps(trumpCards *TrumpCards) *Gaps {
	return &Gaps{trumpCards: trumpCards}
}

// NewDefaultGaps は標準52枚デッキでGapsを生成する。
// CUI/Web/Workerのコンストラクトで共通利用される。
func NewDefaultGaps() *Gaps {
	return NewGaps(NewTrumpCards(0))
}

// Reset は52枚を4×13に配り、Aceの位置を隙間（nil）に置き換える。
func (g *Gaps) Reset() {
	g.trumpCards.Shuffle()
	g.phase = GapsPhasePlaying
	g.moveCount = 0
	g.redealsUsed = 0
	g.actionLog = nil
	g.history = nil
	g.isStalemate = false
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			card := g.trumpCards.DrawCard()
			if card == nil {
				g.grid[r][c] = nil
				continue
			}
			if card.GetValue() == 1 { // Aceは隙間にする
				g.grid[r][c] = nil
				continue
			}
			g.grid[r][c] = card
		}
	}
}

// --- Public moves ---

// Move は (fromRow,fromCol) のカードを (toRow,toCol) の隙間へ移動する。
func (g *Gaps) Move(fromRow, fromCol, toRow, toCol int) error {
	if g.phase != GapsPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if err := validateGapsPos(fromRow, fromCol); err != nil {
		return fmt.Errorf("from: %w", err)
	}
	if err := validateGapsPos(toRow, toCol); err != nil {
		return fmt.Errorf("to: %w", err)
	}
	if fromRow == toRow && fromCol == toCol {
		return errors.New("source and destination are the same")
	}
	src := g.grid[fromRow][fromCol]
	if src == nil {
		return errors.New("source cell is empty")
	}
	if g.grid[toRow][toCol] != nil {
		return errors.New("destination is not a gap")
	}
	if !g.isLegalMove(fromRow, fromCol, toRow, toCol) {
		return errors.New("illegal move")
	}
	g.takeSnapshot()
	g.grid[toRow][toCol] = src
	g.grid[fromRow][fromCol] = nil
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("(%d,%d)→(%d,%d)", fromRow, fromCol, toRow, toCol), []*Card{src})
	g.checkGameClear()
	g.checkStalemate()
	return nil
}

// isLegalMove は (fromRow,fromCol) のカードが (toRow,toCol) の隙間へ置けるかを判定する。
// 呼び出し側は事前に src!=nil, dst==nil および領域境界をチェックすること。
func (g *Gaps) isLegalMove(fromRow, fromCol, toRow, toCol int) bool {
	src := g.grid[fromRow][fromCol]
	if src == nil || g.grid[toRow][toCol] != nil {
		return false
	}
	if toCol == 0 {
		return src.GetValue() == GapsAnchorRank
	}
	left := g.grid[toRow][toCol-1]
	if left == nil {
		return false
	}
	if left.GetValue() == GapsKingRank {
		return false
	}
	return src.GetDesign() == left.GetDesign() && src.GetValue() == left.GetValue()+1
}

// Redeal は正しく並んでいないカードを集めて再配りする。Reshuffle後、各行の
// 「ロック済み接頭辞」の直後を隙間として残し、残りのセルにシャッフルしたカードを充てる。
func (g *Gaps) Redeal() error {
	if g.phase != GapsPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if g.redealsUsed >= GapsMaxRedeals {
		return errors.New("no redeals remaining")
	}
	g.takeSnapshot()

	locks := g.lockedPrefixLengths()
	pool := make([]*Card, 0, 48)
	var newGrid [GapsRowCnt][GapsColCnt]GapsCell
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			if c < locks[r] {
				newGrid[r][c] = g.grid[r][c] // ロック済みはそのまま
				continue
			}
			// それ以外のカードはプールへ
			if g.grid[r][c] != nil {
				pool = append(pool, g.grid[r][c])
			}
		}
	}
	// プールをシャッフル。
	rand.Shuffle(len(pool), func(i, j int) { pool[i], pool[j] = pool[j], pool[i] })
	// 各行のロック直後の列を隙間として確保し、残りのセルへプールから充てる。
	idx := 0
	for r := 0; r < GapsRowCnt; r++ {
		gapCol := locks[r]
		for c := gapCol; c < GapsColCnt; c++ {
			if c == gapCol {
				newGrid[r][c] = nil
				continue
			}
			if idx < len(pool) {
				newGrid[r][c] = pool[idx]
				idx++
			}
		}
	}
	g.grid = newGrid
	g.redealsUsed++
	g.moveCount++
	g.appendLog("redeal", fmt.Sprintf("redeal #%d", g.redealsUsed), nil)
	g.checkGameClear()
	g.checkStalemate()
	return nil
}

// lockedPrefixLengths は各行のロック済み接頭辞の長さを返す。
// 行 r の locks[r] は、左から「(c==0で値2) または (左隣と同スートで値=左隣+1)」の連続長。
func (g *Gaps) lockedPrefixLengths() [GapsRowCnt]int {
	var locks [GapsRowCnt]int
	for r := 0; r < GapsRowCnt; r++ {
		n := 0
		for c := 0; c < GapsColCnt; c++ {
			cur := g.grid[r][c]
			if cur == nil {
				break
			}
			if c == 0 {
				if cur.GetValue() != GapsAnchorRank {
					break
				}
				n++
				continue
			}
			prev := g.grid[r][c-1]
			if prev == nil {
				break
			}
			if cur.GetDesign() != prev.GetDesign() || cur.GetValue() != prev.GetValue()+1 {
				break
			}
			n++
		}
		locks[r] = n
	}
	return locks
}

// Undo は直前のスナップショットを復元する。
func (g *Gaps) Undo() error {
	if g.phase != GapsPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(g.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo はUndo可能ならtrueを返す。
func (g *Gaps) CanUndo() bool {
	return len(g.history) > 0 && g.phase == GapsPhasePlaying
}

// UndoToEscape は手詰まりから抜けるために必要なUndo回数を返す。
// 手詰まりでなければ0、脱出不能なら-1。
func (g *Gaps) UndoToEscape() int {
	if !g.isStalemate {
		return 0
	}
	for i := len(g.history) - 1; i >= 0; i-- {
		if !g.history[i].isStalemate {
			return len(g.history) - i
		}
	}
	return -1
}

// UndoN はn回連続でUndoを試みる。n<=0なら何もしない。
func (g *Gaps) UndoN(n int) error {
	if n <= 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		if err := g.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// GiveUp はゲームをギブアップして終了状態にする。
func (g *Gaps) GiveUp() {
	if g.phase == GapsPhasePlaying {
		g.phase = GapsPhaseGameOver
		g.appendLog("giveup", "give up", nil)
	}
}

// GetHint は次に打てる手を1つ返す。手詰まり/非プレイ中はnilを返す。
func (g *Gaps) GetHint() *GapsHint {
	if g.phase != GapsPhasePlaying {
		return nil
	}
	for tr := 0; tr < GapsRowCnt; tr++ {
		for tc := 0; tc < GapsColCnt; tc++ {
			if g.grid[tr][tc] != nil {
				continue
			}
			// 列0はランク2なら何でも置ける。
			if tc == 0 {
				for fr := 0; fr < GapsRowCnt; fr++ {
					for fc := 0; fc < GapsColCnt; fc++ {
						src := g.grid[fr][fc]
						if src == nil || (fr == tr && fc == tc) {
							continue
						}
						if src.GetValue() == GapsAnchorRank {
							return &GapsHint{FromRow: fr, FromCol: fc, ToRow: tr, ToCol: tc}
						}
					}
				}
				continue
			}
			left := g.grid[tr][tc-1]
			if left == nil || left.GetValue() == GapsKingRank {
				continue
			}
			// 同スート+1ランクを盤上から探す。
			for fr := 0; fr < GapsRowCnt; fr++ {
				for fc := 0; fc < GapsColCnt; fc++ {
					if fr == tr && fc == tc {
						continue
					}
					src := g.grid[fr][fc]
					if src == nil {
						continue
					}
					if src.GetDesign() == left.GetDesign() && src.GetValue() == left.GetValue()+1 {
						return &GapsHint{FromRow: fr, FromCol: fc, ToRow: tr, ToCol: tc}
					}
				}
			}
		}
	}
	return nil
}

// AllWon は各行が同スートで2..Kにソート済みかを判定する。
func (g *Gaps) AllWon() bool {
	for r := 0; r < GapsRowCnt; r++ {
		// 行は col 0..11 に 2..K の昇順、col 12 は隙間（または任意位置の隙間でも可と
		// するが、ここでは「ソート済み」の意味として 0..11 を厳密にチェック）。
		first := g.grid[r][0]
		if first == nil || first.GetValue() != GapsAnchorRank {
			return false
		}
		suit := first.GetDesign()
		for c := 0; c < 12; c++ {
			cell := g.grid[r][c]
			if cell == nil || cell.GetDesign() != suit || cell.GetValue() != c+2 {
				return false
			}
		}
	}
	return true
}

// checkGameClear は勝利判定をする。
func (g *Gaps) checkGameClear() {
	if g.AllWon() {
		g.phase = GapsPhaseGameClear
	}
}

// checkStalemate は手詰まり状態を更新する。
func (g *Gaps) checkStalemate() {
	if g.phase != GapsPhasePlaying {
		return
	}
	if g.redealsUsed < GapsMaxRedeals {
		g.isStalemate = false
		return
	}
	if g.GetHint() != nil {
		g.isStalemate = false
		return
	}
	g.isStalemate = true
}

// RecomputeStalemate はテストや復元処理のために手詰まり判定を再計算する。
func (g *Gaps) RecomputeStalemate() { g.checkStalemate() }

// --- Snapshot / undo helpers ---

func (g *Gaps) takeSnapshot() {
	snap := &gapsSnapshot{
		phase:       g.phase,
		moveCount:   g.moveCount,
		redealsUsed: g.redealsUsed,
		isStalemate: g.isStalemate,
	}
	for r := 0; r < GapsRowCnt; r++ {
		for c := 0; c < GapsColCnt; c++ {
			snap.grid[r][c] = g.grid[r][c]
		}
	}
	g.history = append(g.history, snap)
}

func (g *Gaps) restoreSnapshot(snap *gapsSnapshot) {
	g.grid = snap.grid
	g.phase = snap.phase
	g.moveCount = snap.moveCount
	g.redealsUsed = snap.redealsUsed
	g.isStalemate = snap.isStalemate
}

// appendLog はアクションログにエントリを追加する。
func (g *Gaps) appendLog(actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: g.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters / setters ---

// GetPhase はフェーズを返す。
func (g *Gaps) GetPhase() GapsPhase { return g.phase }

// SetPhase はフェーズを設定する（テスト用）。
func (g *Gaps) SetPhase(p GapsPhase) { g.phase = p }

// GetMoveCount は移動回数を返す。
func (g *Gaps) GetMoveCount() int { return g.moveCount }

// GetGrid は盤面を返す。
func (g *Gaps) GetGrid() [GapsRowCnt][GapsColCnt]GapsCell { return g.grid }

// GetLockedPrefixLengths は各行のロック済み接頭辞の長さ（再配布で保持される
// 先頭カード数）を返す。プレゼンターがロック済みカードを可視化するために使う。
func (g *Gaps) GetLockedPrefixLengths() [GapsRowCnt]int { return g.lockedPrefixLengths() }

// SetGrid は盤面を設定する（テスト用）。
func (g *Gaps) SetGrid(grid [GapsRowCnt][GapsColCnt]GapsCell) { g.grid = grid }

// GetRedealsUsed は使用済み再配り回数を返す。
func (g *Gaps) GetRedealsUsed() int { return g.redealsUsed }

// SetRedealsUsed は使用済み再配り回数を設定する（テスト用）。
func (g *Gaps) SetRedealsUsed(n int) { g.redealsUsed = n }

// GetRedealsRemaining は残りの再配り回数を返す。
func (g *Gaps) GetRedealsRemaining() int { return GapsMaxRedeals - g.redealsUsed }

// GetActionLog はアクションログを返す。
func (g *Gaps) GetActionLog() []*ActionLogEntry { return g.actionLog }

// GetGameEndFlag は終了状態かどうかを返す。
func (g *Gaps) GetGameEndFlag() bool { return g.phase != GapsPhasePlaying }

// IsStalemate は手詰まりかどうかを返す。
func (g *Gaps) IsStalemate() bool { return g.isStalemate }

// SetIsStalemate は手詰まり状態を設定する（テスト用）。
func (g *Gaps) SetIsStalemate(v bool) { g.isStalemate = v }

// --- Validation helpers ---

func validateGapsPos(row, col int) error {
	if row < 0 || row >= GapsRowCnt {
		return errors.New("row out of range")
	}
	if col < 0 || col >= GapsColCnt {
		return errors.New("col out of range")
	}
	return nil
}

// --- JSON marshalling ---

type gapsJSON struct {
	TrumpCards  *TrumpCards                      `json:"tc"`
	Grid        [GapsRowCnt][GapsColCnt]GapsCell `json:"gr"`
	Phase       GapsPhase                        `json:"ps"`
	MoveCount   int                              `json:"mc"`
	RedealsUsed int                              `json:"ru"`
	ActionLog   []*ActionLogEntry                `json:"al"`
	History     []*gapsSnapshot                  `json:"hi,omitempty"`
	IsStalemate bool                             `json:"sm"`
}

type gapsSnapshotJSON struct {
	Grid        [GapsRowCnt][GapsColCnt]GapsCell `json:"gr"`
	Phase       GapsPhase                        `json:"ps"`
	MoveCount   int                              `json:"mc"`
	RedealsUsed int                              `json:"ru"`
	IsStalemate bool                             `json:"sm"`
}

// gapsMaxSliceLen はDoSリスクを防ぐためのスライス長上限。
const gapsMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler.
func (s *gapsSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(gapsSnapshotJSON{
		Grid:        s.grid,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		RedealsUsed: s.redealsUsed,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (s *gapsSnapshot) UnmarshalJSON(data []byte) error {
	var j gapsSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	s.grid = j.Grid
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.redealsUsed = j.RedealsUsed
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (g *Gaps) MarshalJSON() ([]byte, error) {
	return json.Marshal(gapsJSON{
		TrumpCards:  g.trumpCards,
		Grid:        g.grid,
		Phase:       g.phase,
		MoveCount:   g.moveCount,
		RedealsUsed: g.redealsUsed,
		ActionLog:   g.actionLog,
		History:     g.history,
		IsStalemate: g.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *Gaps) UnmarshalJSON(data []byte) error {
	var j gapsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > gapsMaxSliceLen || len(j.History) > gapsMaxSliceLen {
		return fmt.Errorf("gaps: input array exceeds maximum allowed size")
	}
	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.grid = j.Grid
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	g.redealsUsed = j.RedealsUsed
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.history = j.History
	if g.history == nil {
		g.history = make([]*gapsSnapshot, 0)
	}
	g.isStalemate = j.IsStalemate
	return nil
}
