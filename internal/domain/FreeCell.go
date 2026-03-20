package domain

import (
	"errors"
	"fmt"
)

// FreeCellPhase フリーセルゲームフェーズ
type FreeCellPhase int

// FreeCellのフェーズ定数
const (
	// FreeCellPhasePlaying プレイ中
	FreeCellPhasePlaying FreeCellPhase = iota
	// FreeCellPhaseGameClear ゲームクリア
	FreeCellPhaseGameClear
	// FreeCellPhaseGameOver ゲームオーバー
	FreeCellPhaseGameOver
)

// FreeCellTableauCnt タブローの列数
const FreeCellTableauCnt = 8

// FreeCellFoundationCnt ファンデーションの数
const FreeCellFoundationCnt = 4

// FreeCellCellCnt フリーセルの数
const FreeCellCellCnt = 4

// FreeCellHint ヒント
type FreeCellHint struct {
	FromZone  string // "tableau" or "freecell"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau", "foundation", or "freecell"
	ToCol     int
}

// FreeCell フリーセルゲームクラス
type FreeCell struct {
	trumpCards *TrumpCards
	tableau    [FreeCellTableauCnt][]*Card
	freeCells  [FreeCellCellCnt]*Card
	foundation [FreeCellFoundationCnt][]*Card
	phase      FreeCellPhase
	moveCount  int
	actionLog  []*ActionLogEntry
	history    []*freeCellSnapshot
}

// freeCellSnapshot アンドゥ用スナップショット
type freeCellSnapshot struct {
	tableau    [FreeCellTableauCnt][]*Card
	freeCells  [FreeCellCellCnt]*Card
	foundation [FreeCellFoundationCnt][]*Card
	phase      FreeCellPhase
	moveCount  int
}

// NewFreeCell コンストラクタ
func NewFreeCell(trumpCards *TrumpCards) *FreeCell {
	return &FreeCell{
		trumpCards: trumpCards,
	}
}

// Reset ゲームリセット
func (f *FreeCell) Reset() {
	f.trumpCards.Shuffle()
	f.phase = FreeCellPhasePlaying
	f.moveCount = 0
	f.actionLog = nil
	f.history = nil

	// フリーセル初期化
	for i := 0; i < FreeCellCellCnt; i++ {
		f.freeCells[i] = nil
	}

	// ファンデーション初期化
	for i := 0; i < FreeCellFoundationCnt; i++ {
		f.foundation[i] = nil
	}

	// タブローに配る: 最初の4列に7枚、残り4列に6枚
	for i := 0; i < FreeCellTableauCnt; i++ {
		count := 7
		if i >= 4 {
			count = 6
		}
		f.tableau[i] = make([]*Card, 0, count)
		for j := 0; j < count; j++ {
			card := f.trumpCards.DrawCard()
			f.tableau[i] = append(f.tableau[i], card)
		}
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動（スーパームーブ対応）
func (f *FreeCell) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= FreeCellTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= FreeCellTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := f.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}

	// 移動するカード列が有効なシーケンスか確認
	movingCards := fromCards[cardIndex:]
	if !f.isValidTableauSequence(movingCards) {
		return errors.New("cards do not form a valid sequence")
	}

	// スーパームーブ: 移動可能な最大枚数をチェック
	maxCards := f.maxMovableCards(toCol)
	if len(movingCards) > maxCards {
		return errors.New("too many cards to move")
	}

	bottomCard := movingCards[0]
	if !f.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	// 移動実行
	f.takeSnapshot()
	f.tableau[toCol] = append(f.tableau[toCol], movingCards...)
	f.tableau[fromCol] = fromCards[:cardIndex]
	f.moveCount++
	movedCards := make([]*Card, len(movingCards))
	copy(movedCards, movingCards)
	f.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (f *FreeCell) MoveTableauToFoundation(col int) error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FreeCellTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := f.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= FreeCellFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !f.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	f.takeSnapshot()
	f.tableau[col] = fromCards[:len(fromCards)-1]
	f.foundation[fIdx] = append(f.foundation[fIdx], card)
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	f.checkGameClear()
	return nil
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (f *FreeCell) MoveTableauToFreeCell(col, cell int) error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FreeCellTableauCnt {
		return errors.New("invalid column")
	}
	if cell < 0 || cell >= FreeCellCellCnt {
		return errors.New("invalid cell")
	}
	fromCards := f.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if f.freeCells[cell] != nil {
		return errors.New("free cell is occupied")
	}
	f.takeSnapshot()
	card := fromCards[len(fromCards)-1]
	f.tableau[col] = fromCards[:len(fromCards)-1]
	f.freeCells[cell] = card
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("タブロー列%d→フリーセル%d", col, cell), []*Card{card})
	return nil
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (f *FreeCell) MoveFreeCellToTableau(cell, col int) error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= FreeCellCellCnt {
		return errors.New("invalid cell")
	}
	if col < 0 || col >= FreeCellTableauCnt {
		return errors.New("invalid column")
	}
	if f.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := f.freeCells[cell]
	if !f.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	f.takeSnapshot()
	f.freeCells[cell] = nil
	f.tableau[col] = append(f.tableau[col], card)
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("フリーセル%d→タブロー列%d", cell, col), []*Card{card})
	return nil
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (f *FreeCell) MoveFreeCellToFoundation(cell int) error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= FreeCellCellCnt {
		return errors.New("invalid cell")
	}
	if f.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := f.freeCells[cell]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= FreeCellFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !f.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	f.takeSnapshot()
	f.freeCells[cell] = nil
	f.foundation[fIdx] = append(f.foundation[fIdx], card)
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("フリーセル%d→ファンデーション", cell), []*Card{card})
	f.checkGameClear()
	return nil
}

// GiveUp ギブアップ
func (f *FreeCell) GiveUp() {
	if f.phase == FreeCellPhasePlaying {
		f.phase = FreeCellPhaseGameOver
		f.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (f *FreeCell) GetHint() *FreeCellHint {
	if f.phase != FreeCellPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < FreeCellTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			continue
		}
		card := f.tableau[col][len(f.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && f.canPlaceOnFoundation(card, fIdx) {
			return &FreeCellHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(f.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: フリーセルからファンデーションへ
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		if f.freeCells[cell] == nil {
			continue
		}
		card := f.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && f.canPlaceOnFoundation(card, fIdx) {
			return &FreeCellHint{
				FromZone:  "freecell",
				FromCol:   cell,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := 0; fromCol < FreeCellTableauCnt; fromCol++ {
		fromCards := f.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		// 表向きのシーケンスの先頭を探す
		seqStart := len(fromCards) - 1
		for seqStart > 0 {
			if !f.isAlternateColor(fromCards[seqStart], fromCards[seqStart-1]) ||
				fromCards[seqStart].GetValue() != fromCards[seqStart-1].GetValue()-1 {
				break
			}
			seqStart--
		}
		card := fromCards[seqStart]
		movingCards := fromCards[seqStart:]
		for toCol := 0; toCol < FreeCellTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			// 空列へKing以外のシーケンスを移動しても意味がない（キングのみ移動意味あり）
			if len(f.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			maxCards := f.maxMovableCards(toCol)
			if len(movingCards) > maxCards {
				continue
			}
			if f.canPlaceOnTableau(card, toCol) {
				return &FreeCellHint{
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
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		if f.freeCells[cell] == nil {
			continue
		}
		card := f.freeCells[cell]
		for toCol := 0; toCol < FreeCellTableauCnt; toCol++ {
			if len(f.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if f.canPlaceOnTableau(card, toCol) {
				return &FreeCellHint{
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
	for col := 0; col < FreeCellTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			continue
		}
		for cell := 0; cell < FreeCellCellCnt; cell++ {
			if f.freeCells[cell] == nil {
				return &FreeCellHint{
					FromZone:  "tableau",
					FromCol:   col,
					CardIndex: len(f.tableau[col]) - 1,
					ToZone:    "freecell",
					ToCol:     cell,
				}
			}
		}
		break // 最初の非空列について一つのフリーセルを見つければ十分
	}
	return nil
}

// AutoComplete オートコンプリート
func (f *FreeCell) AutoComplete() error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	f.takeSnapshot()
	for {
		moved := false
		// フリーセルからファンデーションへ
		for cell := 0; cell < FreeCellCellCnt; cell++ {
			if f.freeCells[cell] == nil {
				continue
			}
			card := f.freeCells[cell]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= FreeCellFoundationCnt || !f.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			f.freeCells[cell] = nil
			f.foundation[fIdx] = append(f.foundation[fIdx], card)
			f.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := 0; col < FreeCellTableauCnt; col++ {
			if len(f.tableau[col]) == 0 {
				continue
			}
			card := f.tableau[col][len(f.tableau[col])-1]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= FreeCellFoundationCnt || !f.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			f.tableau[col] = f.tableau[col][:len(f.tableau[col])-1]
			f.foundation[fIdx] = append(f.foundation[fIdx], card)
			f.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	f.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	f.checkGameClear()
	return nil
}

// Undo 直前の操作を取り消す
func (f *FreeCell) Undo() error {
	if f.phase != FreeCellPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(f.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := f.history[len(f.history)-1]
	f.history = f.history[:len(f.history)-1]
	f.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (f *FreeCell) CanUndo() bool {
	return len(f.history) > 0 && f.phase == FreeCellPhasePlaying
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (f *FreeCell) GetPhase() FreeCellPhase { return f.phase }

// SetPhase フェーズ設定 (テスト用)
func (f *FreeCell) SetPhase(phase FreeCellPhase) { f.phase = phase }

// GetMoveCount 移動回数取得
func (f *FreeCell) GetMoveCount() int { return f.moveCount }

// GetTableau タブロー取得
func (f *FreeCell) GetTableau() [FreeCellTableauCnt][]*Card { return f.tableau }

// GetFreeCells フリーセル取得
func (f *FreeCell) GetFreeCells() [FreeCellCellCnt]*Card { return f.freeCells }

// GetFoundation ファンデーション取得
func (f *FreeCell) GetFoundation() [FreeCellFoundationCnt][]*Card { return f.foundation }

// GetActionLog 棋譜取得
func (f *FreeCell) GetActionLog() []*ActionLogEntry { return f.actionLog }

// SetTableau タブロー設定 (テスト用)
func (f *FreeCell) SetTableau(tableau [FreeCellTableauCnt][]*Card) { f.tableau = tableau }

// SetFreeCells フリーセル設定 (テスト用)
func (f *FreeCell) SetFreeCells(cells [FreeCellCellCnt]*Card) { f.freeCells = cells }

// SetFoundation ファンデーション設定 (テスト用)
func (f *FreeCell) SetFoundation(foundation [FreeCellFoundationCnt][]*Card) {
	f.foundation = foundation
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (f *FreeCell) canPlaceOnTableau(card *Card, col int) bool {
	colCards := f.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはKのみ置ける
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return f.isAlternateColor(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (f *FreeCell) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := f.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// isAlternateColor 交互の色かどうか判定
func (f *FreeCell) isAlternateColor(card1, card2 *Card) bool {
	return f.isBlack(card1) != f.isBlack(card2)
}

// isBlack 黒いカードかどうか
func (f *FreeCell) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// isValidTableauSequence 降順交互色のシーケンスか判定
func (f *FreeCell) isValidTableauSequence(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if !f.isAlternateColor(cards[i], cards[i-1]) || cards[i].GetValue() != cards[i-1].GetValue()-1 {
			return false
		}
	}
	return true
}

// maxMovableCards 移動可能な最大カード枚数を計算
// (1 + emptyFreeCells) * 2^(emptyTableauCols) ただしtoColが空の場合はそのcolを除外
func (f *FreeCell) maxMovableCards(toCol int) int {
	emptyFreeCells := 0
	for i := 0; i < FreeCellCellCnt; i++ {
		if f.freeCells[i] == nil {
			emptyFreeCells++
		}
	}
	emptyTableauCols := 0
	for i := 0; i < FreeCellTableauCnt; i++ {
		if i != toCol && len(f.tableau[i]) == 0 {
			emptyTableauCols++
		}
	}
	return (1 + emptyFreeCells) << emptyTableauCols
}

// checkGameClear ゲームクリア判定
func (f *FreeCell) checkGameClear() {
	for i := 0; i < FreeCellFoundationCnt; i++ {
		if len(f.foundation[i]) != CardValueMax {
			return
		}
	}
	f.phase = FreeCellPhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (f *FreeCell) takeSnapshot() {
	snap := &freeCellSnapshot{
		phase:     f.phase,
		moveCount: f.moveCount,
	}
	// deep copy tableau
	for i := 0; i < FreeCellTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(snap.tableau[i], f.tableau[i])
	}
	// deep copy freeCells
	snap.freeCells = f.freeCells
	// deep copy foundation
	for i := 0; i < FreeCellFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(f.foundation[i]))
		copy(snap.foundation[i], f.foundation[i])
	}
	f.history = append(f.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (f *FreeCell) restoreSnapshot(snap *freeCellSnapshot) {
	f.tableau = snap.tableau
	f.freeCells = snap.freeCells
	f.foundation = snap.foundation
	f.phase = snap.phase
	f.moveCount = snap.moveCount
}

// appendLog 棋譜エントリを追加
func (f *FreeCell) appendLog(actionType, detail string, cards []*Card) {
	f.actionLog = append(f.actionLog, &ActionLogEntry{
		TurnNumber: f.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}
