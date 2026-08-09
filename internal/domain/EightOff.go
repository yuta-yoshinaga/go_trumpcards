//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// EightOffPhase エイトオフゲームフェーズ
type EightOffPhase int

// EightOffのフェーズ定数
const (
	// EightOffPhasePlaying プレイ中
	EightOffPhasePlaying EightOffPhase = iota
	// EightOffPhaseGameClear ゲームクリア
	EightOffPhaseGameClear
	// EightOffPhaseGameOver ゲームオーバー
	EightOffPhaseGameOver
)

// EightOffTableauCnt タブローの列数
const EightOffTableauCnt = 8

// EightOffFoundationCnt ファンデーションの数
const EightOffFoundationCnt = 4

// EightOffCellCnt フリーセルの数 (FreeCellより4つ多い)
const EightOffCellCnt = 8

// EightOffTableauColCards 初期配置時に各タブロー列に配るカード枚数
const EightOffTableauColCards = 6

// EightOffHint ヒント
type EightOffHint struct {
	FromZone  string // "tableau" or "freecell"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau", "foundation", or "freecell"
	ToCol     int
}

// EightOff エイトオフゲームクラス
type EightOff struct {
	trumpCards *TrumpCards
	tableau    [EightOffTableauCnt][]*Card
	freeCells  [EightOffCellCnt]*Card
	foundation [EightOffFoundationCnt][]*Card
	phase      EightOffPhase
	moveCount  int
	actionLogBase
	history     []*eightOffSnapshot
	isStalemate bool
}

// eightOffSnapshot アンドゥ用スナップショット
type eightOffSnapshot struct {
	tableau     [EightOffTableauCnt][]*Card
	freeCells   [EightOffCellCnt]*Card
	foundation  [EightOffFoundationCnt][]*Card
	phase       EightOffPhase
	moveCount   int
	isStalemate bool
}

// NewEightOff コンストラクタ
func NewEightOff(trumpCards *TrumpCards) *EightOff {
	return &EightOff{
		trumpCards: trumpCards,
	}
}

// NewDefaultEightOff returns EightOff with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultEightOff() *EightOff {
	return NewEightOff(NewTrumpCards(0))
}

// Reset ゲームリセット
func (e *EightOff) Reset() {
	e.trumpCards.Shuffle()
	e.phase = EightOffPhasePlaying
	e.moveCount = 0
	e.actionLog = nil
	e.history = nil
	e.isStalemate = false

	// フリーセル初期化: 全てクリア後に残り4枚を前半4セルに置く。
	for i := 0; i < EightOffCellCnt; i++ {
		e.freeCells[i] = nil
	}

	// ファンデーション初期化
	for i := 0; i < EightOffFoundationCnt; i++ {
		e.foundation[i] = nil
	}

	// タブローに配る: 8列×6枚 = 48枚
	for i := 0; i < EightOffTableauCnt; i++ {
		e.tableau[i] = make([]*Card, 0, EightOffTableauColCards)
		for j := 0; j < EightOffTableauColCards; j++ {
			card := e.trumpCards.DrawCard()
			e.tableau[i] = append(e.tableau[i], card)
		}
	}

	// 残り4枚をフリーセル0〜3に配る
	for i := 0; i < 4; i++ {
		e.freeCells[i] = e.trumpCards.DrawCard()
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動（スーパームーブ対応）
func (e *EightOff) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= EightOffTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= EightOffTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := e.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}

	// 移動するカード列が有効なシーケンスか確認 (同スート降順)
	movingCards := fromCards[cardIndex:]
	if !e.isValidTableauSequence(movingCards) {
		return errors.New("cards do not form a valid sequence")
	}

	// スーパームーブ: 移動可能な最大枚数をチェック
	maxCards := e.maxMovableCards(toCol)
	if len(movingCards) > maxCards {
		return errors.New("too many cards to move")
	}

	bottomCard := movingCards[0]
	if !e.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	// 移動実行
	e.takeSnapshot()
	e.tableau[toCol] = append(e.tableau[toCol], movingCards...)
	e.tableau[fromCol] = fromCards[:cardIndex]
	e.moveCount++
	movedCards := make([]*Card, len(movingCards))
	copy(movedCards, movingCards)
	e.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	e.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (e *EightOff) MoveTableauToFoundation(col int) error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= EightOffTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := e.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= EightOffFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !e.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	e.takeSnapshot()
	e.tableau[col] = fromCards[:len(fromCards)-1]
	e.foundation[fIdx] = append(e.foundation[fIdx], card)
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	e.checkGameClear()
	e.checkStalemate()
	return nil
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (e *EightOff) MoveTableauToFreeCell(col, cell int) error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= EightOffTableauCnt {
		return errors.New("invalid column")
	}
	if cell < 0 || cell >= EightOffCellCnt {
		return errors.New("invalid cell")
	}
	fromCards := e.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if e.freeCells[cell] != nil {
		return errors.New("free cell is occupied")
	}
	e.takeSnapshot()
	card := fromCards[len(fromCards)-1]
	e.tableau[col] = fromCards[:len(fromCards)-1]
	e.freeCells[cell] = card
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("タブロー列%d→フリーセル%d", col, cell), []*Card{card})
	e.checkStalemate()
	return nil
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (e *EightOff) MoveFreeCellToTableau(cell, col int) error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= EightOffCellCnt {
		return errors.New("invalid cell")
	}
	if col < 0 || col >= EightOffTableauCnt {
		return errors.New("invalid column")
	}
	if e.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := e.freeCells[cell]
	if !e.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	e.takeSnapshot()
	e.freeCells[cell] = nil
	e.tableau[col] = append(e.tableau[col], card)
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("フリーセル%d→タブロー列%d", cell, col), []*Card{card})
	e.checkStalemate()
	return nil
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (e *EightOff) MoveFreeCellToFoundation(cell int) error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= EightOffCellCnt {
		return errors.New("invalid cell")
	}
	if e.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := e.freeCells[cell]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= EightOffFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !e.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	e.takeSnapshot()
	e.freeCells[cell] = nil
	e.foundation[fIdx] = append(e.foundation[fIdx], card)
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("フリーセル%d→ファンデーション", cell), []*Card{card})
	e.checkGameClear()
	e.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (e *EightOff) GiveUp() {
	if e.phase == EightOffPhasePlaying {
		e.phase = EightOffPhaseGameOver
		e.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (e *EightOff) GetHint() *EightOffHint {
	if e.phase != EightOffPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < EightOffTableauCnt; col++ {
		if len(e.tableau[col]) == 0 {
			continue
		}
		card := e.tableau[col][len(e.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < EightOffFoundationCnt && e.canPlaceOnFoundation(card, fIdx) {
			return &EightOffHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(e.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: フリーセルからファンデーションへ
	for cell := 0; cell < EightOffCellCnt; cell++ {
		if e.freeCells[cell] == nil {
			continue
		}
		card := e.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < EightOffFoundationCnt && e.canPlaceOnFoundation(card, fIdx) {
			return &EightOffHint{
				FromZone:  "freecell",
				FromCol:   cell,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := 0; fromCol < EightOffTableauCnt; fromCol++ {
		fromCards := e.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		// 表向きのシーケンスの先頭を探す (同スート降順)
		seqStart := len(fromCards) - 1
		for seqStart > 0 {
			if !e.isSameSuit(fromCards[seqStart], fromCards[seqStart-1]) ||
				fromCards[seqStart].GetValue() != fromCards[seqStart-1].GetValue()-1 {
				break
			}
			seqStart--
		}
		card := fromCards[seqStart]
		movingCards := fromCards[seqStart:]
		for toCol := 0; toCol < EightOffTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			// 空列にはKingしか置けない
			if len(e.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			maxCards := e.maxMovableCards(toCol)
			if len(movingCards) > maxCards {
				continue
			}
			if e.canPlaceOnTableau(card, toCol) {
				return &EightOffHint{
					FromZone:  "tableau",
					FromCol:   fromCol,
					CardIndex: seqStart,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度4: フリーセルからタブローへ
	for cell := 0; cell < EightOffCellCnt; cell++ {
		if e.freeCells[cell] == nil {
			continue
		}
		card := e.freeCells[cell]
		for toCol := 0; toCol < EightOffTableauCnt; toCol++ {
			// 空列にはKingしか置けない
			if len(e.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if e.canPlaceOnTableau(card, toCol) {
				return &EightOffHint{
					FromZone:  "freecell",
					FromCol:   cell,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度5: タブローからフリーセルへ
	for col := 0; col < EightOffTableauCnt; col++ {
		if len(e.tableau[col]) == 0 {
			continue
		}
		for cell := 0; cell < EightOffCellCnt; cell++ {
			if e.freeCells[cell] == nil {
				return &EightOffHint{
					FromZone:  "tableau",
					FromCol:   col,
					CardIndex: len(e.tableau[col]) - 1,
					ToZone:    "freecell",
					ToCol:     cell,
				}
			}
		}
		break
	}
	return nil
}

// AutoComplete オートコンプリート
func (e *EightOff) AutoComplete() error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	e.takeSnapshot()
	for {
		moved := false
		// フリーセルからファンデーションへ
		for cell := 0; cell < EightOffCellCnt; cell++ {
			if e.freeCells[cell] == nil {
				continue
			}
			card := e.freeCells[cell]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= EightOffFoundationCnt || !e.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			e.freeCells[cell] = nil
			e.foundation[fIdx] = append(e.foundation[fIdx], card)
			e.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := 0; col < EightOffTableauCnt; col++ {
			if len(e.tableau[col]) == 0 {
				continue
			}
			card := e.tableau[col][len(e.tableau[col])-1]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= EightOffFoundationCnt || !e.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			e.tableau[col] = e.tableau[col][:len(e.tableau[col])-1]
			e.foundation[fIdx] = append(e.foundation[fIdx], card)
			e.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	e.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	e.checkGameClear()
	e.checkStalemate()
	return nil
}

// Undo 直前の操作を取り消す
func (e *EightOff) Undo() error {
	if e.phase != EightOffPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(e.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := e.history[len(e.history)-1]
	e.history = e.history[:len(e.history)-1]
	e.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (e *EightOff) CanUndo() bool {
	return len(e.history) > 0 && e.phase == EightOffPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (e *EightOff) UndoToEscape() int {
	return undoToEscape(e.isStalemate, e.history, func(s *eightOffSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (e *EightOff) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := e.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (e *EightOff) GetPhase() EightOffPhase { return e.phase }

// SetPhase フェーズ設定 (テスト用)
func (e *EightOff) SetPhase(phase EightOffPhase) { e.phase = phase }

// GetMoveCount 移動回数取得
func (e *EightOff) GetMoveCount() int { return e.moveCount }

// GetTableau タブロー取得
func (e *EightOff) GetTableau() [EightOffTableauCnt][]*Card { return e.tableau }

// GetFreeCells フリーセル取得
func (e *EightOff) GetFreeCells() [EightOffCellCnt]*Card { return e.freeCells }

// GetFoundation ファンデーション取得
func (e *EightOff) GetFoundation() [EightOffFoundationCnt][]*Card { return e.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (e *EightOff) GetGameEndFlag() bool { return e.phase != EightOffPhasePlaying }

// IsStalemate 手詰まり状態取得
func (e *EightOff) IsStalemate() bool { return e.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (e *EightOff) SetIsStalemate(v bool) { e.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (e *EightOff) SetTableau(tableau [EightOffTableauCnt][]*Card) { e.tableau = tableau }

// SetFreeCells フリーセル設定 (テスト用)
func (e *EightOff) SetFreeCells(cells [EightOffCellCnt]*Card) { e.freeCells = cells }

// SetFoundation ファンデーション設定 (テスト用)
func (e *EightOff) SetFoundation(foundation [EightOffFoundationCnt][]*Card) {
	e.foundation = foundation
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定。
// EightOff固有: 空列にはKingしか置けず、ノンエンプティ列には同スート降順のみ。
func (e *EightOff) canPlaceOnTableau(card *Card, col int) bool {
	colCards := e.tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return e.isSameSuit(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (e *EightOff) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := e.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// isSameSuit 同じスートかどうか判定
func (e *EightOff) isSameSuit(card1, card2 *Card) bool {
	return card1.GetDesign() == card2.GetDesign()
}

// isValidTableauSequence 同スート降順のシーケンスか判定
func (e *EightOff) isValidTableauSequence(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if !e.isSameSuit(cards[i], cards[i-1]) || cards[i].GetValue() != cards[i-1].GetValue()-1 {
			return false
		}
	}
	return true
}

// maxMovableCards 移動可能な最大カード枚数を計算
// (1 + emptyFreeCells) * 2^(emptyTableauCols) ただしtoColが空の場合はそのcolを除外
func (e *EightOff) maxMovableCards(toCol int) int {
	emptyFreeCells := 0
	for i := 0; i < EightOffCellCnt; i++ {
		if e.freeCells[i] == nil {
			emptyFreeCells++
		}
	}
	emptyTableauCols := 0
	for i := 0; i < EightOffTableauCnt; i++ {
		if i != toCol && len(e.tableau[i]) == 0 {
			emptyTableauCols++
		}
	}
	return (1 + emptyFreeCells) << emptyTableauCols
}

// checkGameClear ゲームクリア判定
func (e *EightOff) checkGameClear() {
	for i := 0; i < EightOffFoundationCnt; i++ {
		if len(e.foundation[i]) != CardValueMax {
			return
		}
	}
	e.phase = EightOffPhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (e *EightOff) takeSnapshot() {
	snap := &eightOffSnapshot{
		phase:       e.phase,
		moveCount:   e.moveCount,
		isStalemate: e.isStalemate,
	}
	for i := 0; i < EightOffTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(e.tableau[i]))
		copy(snap.tableau[i], e.tableau[i])
	}
	snap.freeCells = e.freeCells
	for i := 0; i < EightOffFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(e.foundation[i]))
		copy(snap.foundation[i], e.foundation[i])
	}
	e.history = append(e.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (e *EightOff) restoreSnapshot(snap *eightOffSnapshot) {
	e.tableau = snap.tableau
	e.freeCells = snap.freeCells
	e.foundation = snap.foundation
	e.phase = snap.phase
	e.moveCount = snap.moveCount
	e.isStalemate = snap.isStalemate
}

// checkStalemate ソルバーで手詰まり判定
func (e *EightOff) checkStalemate() {
	if e.phase != EightOffPhasePlaying {
		return
	}
	solver := newEightOffSolver(e)
	e.isStalemate = !solver.isSolvable()
}

// appendLog 棋譜エントリを追加
func (e *EightOff) appendLog(actionType, detail string, cards []*Card) {
	e.appendLogAt(e.moveCount, 0, actionType, detail, cards)
}

// eightOffJSON is the JSON wire format for EightOff.
type eightOffJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [EightOffTableauCnt][]*Card    `json:"tb"`
	FreeCells   [EightOffCellCnt]*Card         `json:"fc"`
	Foundation  [EightOffFoundationCnt][]*Card `json:"fd"`
	Phase       EightOffPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sm"`
	History     []*eightOffSnapshot            `json:"hi,omitempty"`
}

// eightOffSnapshotJSON is the wire format for a single undo snapshot.
type eightOffSnapshotJSON struct {
	Tableau     [EightOffTableauCnt][]*Card    `json:"tb"`
	FreeCells   [EightOffCellCnt]*Card         `json:"fc"`
	Foundation  [EightOffFoundationCnt][]*Card `json:"fd"`
	Phase       EightOffPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sm"`
}

// MarshalJSON implements json.Marshaler for eightOffSnapshot.
func (s *eightOffSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(eightOffSnapshotJSON{
		Tableau:     s.tableau,
		FreeCells:   s.freeCells,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for eightOffSnapshot.
func (s *eightOffSnapshot) UnmarshalJSON(data []byte) error {
	var j eightOffSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > eightOffMaxSliceLen {
			return fmt.Errorf("eightOff: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > eightOffMaxSliceLen {
			return fmt.Errorf("eightOff: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.freeCells = j.FreeCells
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (e *EightOff) MarshalJSON() ([]byte, error) {
	return json.Marshal(eightOffJSON{
		TrumpCards:  e.trumpCards,
		Tableau:     e.tableau,
		FreeCells:   e.freeCells,
		Foundation:  e.foundation,
		Phase:       e.phase,
		MoveCount:   e.moveCount,
		ActionLog:   e.actionLog,
		IsStalemate: e.isStalemate,
		History:     e.history,
	})
}

// eightOffMaxSliceLen caps slice sizes during deserialisation.
const eightOffMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (e *EightOff) UnmarshalJSON(data []byte) error {
	var j eightOffJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > eightOffMaxSliceLen || len(j.History) > eightOffMaxSliceLen {
		return fmt.Errorf("eightOff: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > eightOffMaxSliceLen {
			return fmt.Errorf("eightOff: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > eightOffMaxSliceLen {
			return fmt.Errorf("eightOff: foundation pile exceeds maximum allowed size")
		}
	}

	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCards(0)
	}
	e.tableau = j.Tableau
	e.freeCells = j.FreeCells
	e.foundation = j.Foundation
	e.phase = j.Phase
	e.moveCount = j.MoveCount
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	e.history = j.History
	if e.history == nil {
		e.history = make([]*eightOffSnapshot, 0)
	}
	e.isStalemate = j.IsStalemate
	return nil
}
