//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
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
	actionLogBase
	history     []*freeCellSnapshot
	isStalemate bool
	// sameSuit が true のときタブロー積み上げ条件を「同じスートの降順」にする
	// （Baker's Game / ベーカーズ・ゲーム）。false なら通常のフリーセル（赤黒交互の降順）。
	sameSuit bool
}

// freeCellSnapshot アンドゥ用スナップショット
type freeCellSnapshot struct {
	tableau     [FreeCellTableauCnt][]*Card
	freeCells   [FreeCellCellCnt]*Card
	foundation  [FreeCellFoundationCnt][]*Card
	phase       FreeCellPhase
	moveCount   int
	isStalemate bool
}

// NewFreeCell コンストラクタ
func NewFreeCell(trumpCards *TrumpCards) *FreeCell {
	return &FreeCell{
		trumpCards: trumpCards,
	}
}

// NewDefaultFreeCell returns FreeCell with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFreeCell() *FreeCell {
	return NewFreeCell(NewTrumpCards(0))
}

// NewBakersGame コンストラクタ。フリーセルの直接の元祖である Baker's Game
// （ベーカーズ・ゲーム）を構築する。盤面構成・配り方・スーパームーブはフリーセルと
// 同一だが、タブロー積み上げ条件が「赤黒交互の降順」ではなく「同じスートの降順」になる。
func NewBakersGame(trumpCards *TrumpCards) *FreeCell {
	f := NewFreeCell(trumpCards)
	f.sameSuit = true
	return f
}

// NewDefaultBakersGame returns a Baker's Game with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBakersGame() *FreeCell {
	return NewBakersGame(NewTrumpCards(0))
}

// Reset ゲームリセット
func (f *FreeCell) Reset() {
	f.trumpCards.Shuffle()
	f.phase = FreeCellPhasePlaying
	f.moveCount = 0
	f.actionLog = nil
	f.history = nil
	f.isStalemate = false

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
	f.checkStalemate()
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
	f.checkStalemate()
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
	f.checkStalemate()
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
	f.checkStalemate()
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
	f.checkStalemate()
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
			if !f.tableauStackable(fromCards[seqStart-1], fromCards[seqStart]) {
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
			// 優先度3ではKingの空列移動だけを評価する。King以外の空列移動は
			// 有効な手ではあるが優先度が低いので、後段のフォールバックに委ねる。
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
			// 優先度4ではKingの空列移動だけを評価する。King以外の空列移動は
			// 有効な手ではあるが優先度が低いので、後段のフォールバックに委ねる。
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
	// 優先度5: 空列への非Kingの移動（フォールバック）
	// フリーセルやタブローの非Kingカードを空列に移すことでタブロー奥のカードを
	// 露出させる手。フリーセルでは空列に任意のカードを置けるため、有効な手として
	// 候補に含める。ただし無意味な空列交換を避けるためタブロー列全体の移動は除外する。
	if hint := f.getHintToEmptyColumn(); hint != nil {
		return hint
	}
	// 優先度6: タブローからフリーセルへ
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

// getHintToEmptyColumn は非Kingカードを空列に移動するフォールバックヒントを返す。
// フリーセルから空列への移動を優先し、次にタブロー列のシーケンスを空列に移動する。
// 列全体の移動（seqStart == 0）は意味がないため除外する。
func (f *FreeCell) getHintToEmptyColumn() *FreeCellHint {
	// 空列の先頭インデックスを探す
	emptyCol := -1
	for col := 0; col < FreeCellTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			emptyCol = col
			break
		}
	}
	if emptyCol < 0 {
		return nil
	}
	// フリーセル→空列（非King）
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		card := f.freeCells[cell]
		if card == nil || card.GetValue() == CardValueMax {
			continue
		}
		return &FreeCellHint{
			FromZone:  "freecell",
			FromCol:   cell,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     emptyCol,
		}
	}
	// タブロー→空列（非King、列全体の移動を除く）
	for fromCol := 0; fromCol < FreeCellTableauCnt; fromCol++ {
		fromCards := f.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		seqStart := len(fromCards) - 1
		for seqStart > 0 {
			if !f.tableauStackable(fromCards[seqStart-1], fromCards[seqStart]) {
				break
			}
			seqStart--
		}
		// 列全体を空列に動かすのは無意味な空列交換
		if seqStart == 0 {
			continue
		}
		card := fromCards[seqStart]
		if card.GetValue() == CardValueMax {
			// Kingの空列移動は優先度3で処理済み
			continue
		}
		movingCards := fromCards[seqStart:]
		// emptyColはfromColと異なる（fromCardsが非空のため）
		if len(movingCards) <= f.maxMovableCards(emptyCol) {
			return &FreeCellHint{
				FromZone:  "tableau",
				FromCol:   fromCol,
				CardIndex: seqStart,
				ToZone:    "tableau",
				ToCol:     emptyCol,
			}
		}
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
	f.checkStalemate()
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

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (f *FreeCell) UndoToEscape() int {
	return undoToEscape(f.isStalemate, f.history, func(s *freeCellSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (f *FreeCell) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := f.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
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

// GetGameEndFlag returns true once the game has left the playing phase.
func (f *FreeCell) GetGameEndFlag() bool { return f.phase != FreeCellPhasePlaying }

// IsStalemate 手詰まり状態取得
func (f *FreeCell) IsStalemate() bool { return f.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (f *FreeCell) SetIsStalemate(v bool) { f.isStalemate = v }

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
		// フリーセルでは空列には任意のカードを置ける
		return true
	}
	topCard := colCards[len(colCards)-1]
	return f.tableauStackable(topCard, card)
}

// tableauStackable は lower を upper の上にタブローで重ねられるか判定する。
// lower は upper より 1 ランク下で、かつバリアントごとの色/スート条件を満たす必要がある。
// 通常のフリーセルでは赤黒交互、Baker's Game では同じスートを要求する。
func (f *FreeCell) tableauStackable(upper, lower *Card) bool {
	if lower.GetValue() != upper.GetValue()-1 {
		return false
	}
	if f.sameSuit {
		return upper.GetDesign() == lower.GetDesign()
	}
	return f.isAlternateColor(upper, lower)
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

// isValidTableauSequence タブロー移動可能なシーケンスか判定する。
// 通常のフリーセルでは降順かつ赤黒交互、Baker's Game では降順かつ同じスート。
func (f *FreeCell) isValidTableauSequence(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if !f.tableauStackable(cards[i-1], cards[i]) {
			return false
		}
	}
	return true
}

// GetMaxMovableCards はいま一度に動かせる最大枚数を返す (#4777)。
//
// **空き列を移動先にすると上限は下がる。**その列自身は「経由地」に使えない
// ため。移動先を選ぶ前の一般的な上限がこちらで、空き列に置くときの上限は
// GetMaxMovableCardsToEmptyColumn。
func (f *FreeCell) GetMaxMovableCards() int {
	return f.maxMovableCards(-1)
}

// GetMaxMovableCardsToEmptyColumn は空き列へ動かすときの上限を返す。
// 空き列が無いときは 0。
func (f *FreeCell) GetMaxMovableCardsToEmptyColumn() int {
	for i := 0; i < FreeCellTableauCnt; i++ {
		if len(f.tableau[i]) == 0 {
			return f.maxMovableCards(i)
		}
	}
	return 0
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
		phase:       f.phase,
		moveCount:   f.moveCount,
		isStalemate: f.isStalemate,
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
	f.isStalemate = snap.isStalemate
}

// checkStalemate ソルバーで手詰まり判定
func (f *FreeCell) checkStalemate() {
	if f.phase != FreeCellPhasePlaying {
		return
	}
	solver := newFreeCellSolver(f)
	f.isStalemate = !solver.isSolvable()
}

// appendLog 棋譜エントリを追加
func (f *FreeCell) appendLog(actionType, detail string, cards []*Card) {
	f.appendLogAt(f.moveCount, 0, actionType, detail, cards)
}

// freeCellJSON is the JSON wire format for FreeCell.
type freeCellJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [FreeCellTableauCnt][]*Card    `json:"tb"`
	FreeCells   [FreeCellCellCnt]*Card         `json:"fc"`
	Foundation  [FreeCellFoundationCnt][]*Card `json:"fd"`
	Phase       FreeCellPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sm"`
	History     []*freeCellSnapshot            `json:"hi,omitempty"`
	SameSuit    bool                           `json:"ss,omitempty"`
}

// freeCellSnapshotJSON is the wire format for a single undo snapshot.
// freeCellSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// freeCellJSON's short keys to keep the KV payload compact (#1654).
type freeCellSnapshotJSON struct {
	Tableau     [FreeCellTableauCnt][]*Card    `json:"tb"`
	FreeCells   [FreeCellCellCnt]*Card         `json:"fc"`
	Foundation  [FreeCellFoundationCnt][]*Card `json:"fd"`
	Phase       FreeCellPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sm"`
}

// MarshalJSON implements json.Marshaler for freeCellSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// FreeCell.MarshalJSON can persist the undo history (#1654).
func (s *freeCellSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(freeCellSnapshotJSON{
		Tableau:     s.tableau,
		FreeCells:   s.freeCells,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for freeCellSnapshot.
func (s *freeCellSnapshot) UnmarshalJSON(data []byte) error {
	var j freeCellSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > freeCellMaxSliceLen {
			return fmt.Errorf("freeCell: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > freeCellMaxSliceLen {
			return fmt.Errorf("freeCell: snapshot foundation pile exceeds maximum allowed size")
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
func (f *FreeCell) MarshalJSON() ([]byte, error) {
	return json.Marshal(freeCellJSON{
		TrumpCards:  f.trumpCards,
		Tableau:     f.tableau,
		FreeCells:   f.freeCells,
		Foundation:  f.foundation,
		Phase:       f.phase,
		MoveCount:   f.moveCount,
		ActionLog:   f.actionLog,
		IsStalemate: f.isStalemate,
		History:     f.history,
		SameSuit:    f.sameSuit,
	})
}

// freeCellMaxSliceLen caps slice sizes during deserialisation.
const freeCellMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (f *FreeCell) UnmarshalJSON(data []byte) error {
	var j freeCellJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > freeCellMaxSliceLen || len(j.History) > freeCellMaxSliceLen {
		return fmt.Errorf("freeCell: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > freeCellMaxSliceLen {
			return fmt.Errorf("freeCell: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > freeCellMaxSliceLen {
			return fmt.Errorf("freeCell: foundation pile exceeds maximum allowed size")
		}
	}

	f.trumpCards = j.TrumpCards
	if f.trumpCards == nil {
		f.trumpCards = NewTrumpCards(0)
	}
	f.tableau = j.Tableau
	f.freeCells = j.FreeCells
	f.foundation = j.Foundation
	f.phase = j.Phase
	f.moveCount = j.MoveCount
	f.actionLog = j.ActionLog
	if f.actionLog == nil {
		f.actionLog = make([]*ActionLogEntry, 0)
	}
	f.history = j.History
	if f.history == nil {
		f.history = make([]*freeCellSnapshot, 0)
	}
	f.isStalemate = j.IsStalemate
	f.sameSuit = j.SameSuit
	return nil
}
