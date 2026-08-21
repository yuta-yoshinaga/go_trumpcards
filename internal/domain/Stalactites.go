//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StalactitesPhase フリーセルゲームフェーズ
type StalactitesPhase int

// Stalactitesのフェーズ定数
const (
	// StalactitesPhasePlaying プレイ中
	StalactitesPhasePlaying StalactitesPhase = iota
	// StalactitesPhaseGameClear ゲームクリア
	StalactitesPhaseGameClear
	// StalactitesPhaseGameOver ゲームオーバー
	StalactitesPhaseGameOver
)

// StalactitesTableauCnt タブローの列数
const StalactitesTableauCnt = 8

// StalactitesFoundationCnt ファンデーションの数
const StalactitesFoundationCnt = 4

// StalactitesCellCnt フリーセルの数
const StalactitesCellCnt = 4

// StalactitesColumnLen は1列あたりの枚数。48 / 8 = 6。
const StalactitesColumnLen = 6

// StalactitesHint ヒント
type StalactitesHint struct {
	FromZone  string // "tableau" or "stalactites"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau", "foundation", or "stalactites"
	ToCol     int
}

// Stalactites フリーセルゲームクラス
type Stalactites struct {
	trumpCards *TrumpCards
	tableau    [StalactitesTableauCnt][]*Card
	cells      [StalactitesCellCnt]*Card
	foundation [StalactitesFoundationCnt][]*Card
	phase      StalactitesPhase
	moveCount  int
	actionLogBase
	history     []*stalactitesSnapshot
	isStalemate bool
	// baseRank はファンデーションの開始ランク。Ace 固定ではなく、配りの
	// 最初の stalactite のランクで毎回変わる。13 を超えたら 1 に戻る（ラップ）。
	baseRank int
}

// stalactitesSnapshot アンドゥ用スナップショット
type stalactitesSnapshot struct {
	tableau     [StalactitesTableauCnt][]*Card
	cells       [StalactitesCellCnt]*Card
	foundation  [StalactitesFoundationCnt][]*Card
	phase       StalactitesPhase
	moveCount   int
	isStalemate bool
}

// NewStalactites コンストラクタ
func NewStalactites(trumpCards *TrumpCards) *Stalactites {
	return &Stalactites{
		trumpCards: trumpCards,
	}
}

// NewDefaultStalactites returns Stalactites with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultStalactites() *Stalactites {
	return NewStalactites(NewTrumpCards(0))
}

// Reset ゲームリセット
func (f *Stalactites) Reset() {
	f.trumpCards.Shuffle()
	f.phase = StalactitesPhasePlaying
	f.moveCount = 0
	f.actionLog = nil
	f.history = nil
	f.isStalemate = false

	for i := 0; i < StalactitesFoundationCnt; i++ {
		f.foundation[i] = nil
	}

	// **セルを先に埋める。**この4枚が "stalactites" で、FreeCell と違い開始時
	// から塞がっている（＝実質フリーセル0）。最初の1枚のランクが以後すべての
	// ファンデーションの開始ランクになる。
	for i := 0; i < StalactitesCellCnt; i++ {
		f.cells[i] = f.trumpCards.DrawCard()
	}
	f.baseRank = f.cells[0].GetValue()

	// 残り48枚を8列×6枚。4 + 48 = 52。
	for i := 0; i < StalactitesTableauCnt; i++ {
		f.tableau[i] = make([]*Card, 0, StalactitesColumnLen)
		for j := 0; j < StalactitesColumnLen; j++ {
			f.tableau[i] = append(f.tableau[i], f.trumpCards.DrawCard())
		}
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動（スーパームーブ対応）
func (f *Stalactites) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= StalactitesTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= StalactitesTableauCnt {
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
func (f *Stalactites) MoveTableauToFoundation(col int) error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= StalactitesTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := f.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := f.foundationIndexFor(card)
	if fIdx < 0 {
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

// MoveTableauToStalactites タブローからフリーセルにカードを移動
func (f *Stalactites) MoveTableauToStalactites(col, cell int) error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= StalactitesTableauCnt {
		return errors.New("invalid column")
	}
	if cell < 0 || cell >= StalactitesCellCnt {
		return errors.New("invalid cell")
	}
	fromCards := f.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if f.cells[cell] != nil {
		return errors.New("free cell is occupied")
	}
	f.takeSnapshot()
	card := fromCards[len(fromCards)-1]
	f.tableau[col] = fromCards[:len(fromCards)-1]
	f.cells[cell] = card
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("タブロー列%d→フリーセル%d", col, cell), []*Card{card})
	f.checkStalemate()
	return nil
}

// MoveStalactitesToTableau フリーセルからタブローにカードを移動
func (f *Stalactites) MoveStalactitesToTableau(cell, col int) error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= StalactitesCellCnt {
		return errors.New("invalid cell")
	}
	if col < 0 || col >= StalactitesTableauCnt {
		return errors.New("invalid column")
	}
	if f.cells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := f.cells[cell]
	if !f.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	f.takeSnapshot()
	f.cells[cell] = nil
	f.tableau[col] = append(f.tableau[col], card)
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("フリーセル%d→タブロー列%d", cell, col), []*Card{card})
	f.checkStalemate()
	return nil
}

// MoveStalactitesToFoundation フリーセルからファンデーションにカードを移動
func (f *Stalactites) MoveStalactitesToFoundation(cell int) error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= StalactitesCellCnt {
		return errors.New("invalid cell")
	}
	if f.cells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := f.cells[cell]
	fIdx := f.foundationIndexFor(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	f.takeSnapshot()
	f.cells[cell] = nil
	f.foundation[fIdx] = append(f.foundation[fIdx], card)
	f.moveCount++
	f.appendLog("move", fmt.Sprintf("フリーセル%d→ファンデーション", cell), []*Card{card})
	f.checkGameClear()
	f.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (f *Stalactites) GiveUp() {
	if f.phase == StalactitesPhasePlaying {
		f.phase = StalactitesPhaseGameOver
		f.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (f *Stalactites) GetHint() *StalactitesHint {
	if f.phase != StalactitesPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < StalactitesTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			continue
		}
		card := f.tableau[col][len(f.tableau[col])-1]
		if fIdx := f.foundationIndexFor(card); fIdx >= 0 {
			return &StalactitesHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(f.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: フリーセルからファンデーションへ
	for cell := 0; cell < StalactitesCellCnt; cell++ {
		if f.cells[cell] == nil {
			continue
		}
		card := f.cells[cell]
		if fIdx := f.foundationIndexFor(card); fIdx >= 0 {
			return &StalactitesHint{
				FromZone:  "stalactites",
				FromCol:   cell,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := 0; fromCol < StalactitesTableauCnt; fromCol++ {
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
		for toCol := 0; toCol < StalactitesTableauCnt; toCol++ {
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
				return &StalactitesHint{
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
	for cell := 0; cell < StalactitesCellCnt; cell++ {
		if f.cells[cell] == nil {
			continue
		}
		card := f.cells[cell]
		for toCol := 0; toCol < StalactitesTableauCnt; toCol++ {
			// 優先度4ではKingの空列移動だけを評価する。King以外の空列移動は
			// 有効な手ではあるが優先度が低いので、後段のフォールバックに委ねる。
			if len(f.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if f.canPlaceOnTableau(card, toCol) {
				return &StalactitesHint{
					FromZone:  "stalactites",
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
	for col := 0; col < StalactitesTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			continue
		}
		for cell := 0; cell < StalactitesCellCnt; cell++ {
			if f.cells[cell] == nil {
				return &StalactitesHint{
					FromZone:  "tableau",
					FromCol:   col,
					CardIndex: len(f.tableau[col]) - 1,
					ToZone:    "stalactites",
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
func (f *Stalactites) getHintToEmptyColumn() *StalactitesHint {
	// 空列の先頭インデックスを探す
	emptyCol := -1
	for col := 0; col < StalactitesTableauCnt; col++ {
		if len(f.tableau[col]) == 0 {
			emptyCol = col
			break
		}
	}
	if emptyCol < 0 {
		return nil
	}
	// フリーセル→空列（非King）
	for cell := 0; cell < StalactitesCellCnt; cell++ {
		card := f.cells[cell]
		if card == nil || card.GetValue() == CardValueMax {
			continue
		}
		return &StalactitesHint{
			FromZone:  "stalactites",
			FromCol:   cell,
			CardIndex: -1,
			ToZone:    "tableau",
			ToCol:     emptyCol,
		}
	}
	// タブロー→空列（非King、列全体の移動を除く）
	for fromCol := 0; fromCol < StalactitesTableauCnt; fromCol++ {
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
			return &StalactitesHint{
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
func (f *Stalactites) AutoComplete() error {
	if f.phase != StalactitesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	f.takeSnapshot()
	for {
		moved := false
		// フリーセルからファンデーションへ
		for cell := 0; cell < StalactitesCellCnt; cell++ {
			if f.cells[cell] == nil {
				continue
			}
			card := f.cells[cell]
			fIdx := f.foundationIndexFor(card)
			if fIdx < 0 {
				continue
			}
			f.cells[cell] = nil
			f.foundation[fIdx] = append(f.foundation[fIdx], card)
			f.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := 0; col < StalactitesTableauCnt; col++ {
			if len(f.tableau[col]) == 0 {
				continue
			}
			card := f.tableau[col][len(f.tableau[col])-1]
			fIdx := f.foundationIndexFor(card)
			if fIdx < 0 {
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
func (f *Stalactites) Undo() error {
	if f.phase != StalactitesPhasePlaying {
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
func (f *Stalactites) CanUndo() bool {
	return len(f.history) > 0 && f.phase == StalactitesPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (f *Stalactites) UndoToEscape() int {
	return undoToEscape(f.isStalemate, f.history, func(s *stalactitesSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (f *Stalactites) UndoN(n int) error {
	return undoN(f, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (f *Stalactites) GetPhase() StalactitesPhase { return f.phase }

// SetPhase フェーズ設定 (テスト用)
func (f *Stalactites) SetPhase(phase StalactitesPhase) { f.phase = phase }

// GetMoveCount 移動回数取得
func (f *Stalactites) GetMoveCount() int { return f.moveCount }

// GetTableau タブロー取得
func (f *Stalactites) GetTableau() [StalactitesTableauCnt][]*Card { return f.tableau }

// GetCells フリーセル取得
func (f *Stalactites) GetCells() [StalactitesCellCnt]*Card { return f.cells }

// GetFoundation ファンデーション取得
func (f *Stalactites) GetFoundation() [StalactitesFoundationCnt][]*Card { return f.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (f *Stalactites) GetGameEndFlag() bool { return f.phase != StalactitesPhasePlaying }

// IsStalemate 手詰まり状態取得
func (f *Stalactites) IsStalemate() bool { return f.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (f *Stalactites) SetIsStalemate(v bool) { f.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (f *Stalactites) SetTableau(tableau [StalactitesTableauCnt][]*Card) { f.tableau = tableau }

// SetCells フリーセル設定 (テスト用)
func (f *Stalactites) SetCells(cells [StalactitesCellCnt]*Card) { f.cells = cells }

// SetFoundation ファンデーション設定 (テスト用)
func (f *Stalactites) SetFoundation(foundation [StalactitesFoundationCnt][]*Card) {
	f.foundation = foundation
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (f *Stalactites) canPlaceOnTableau(card *Card, col int) bool {
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
func (f *Stalactites) tableauStackable(upper, lower *Card) bool {
	if lower.GetValue() != upper.GetValue()-1 {
		return false
	}
	return f.isAlternateColor(upper, lower)
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (f *Stalactites) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := f.foundation[fIdx]
	// 空のパイルは基準ランクしか受け取らない。FreeCell は A 固定だった。
	if len(pile) == 0 {
		return card.GetValue() == f.baseRank
	}
	// **スートを見ない。**1つ上のランクで、13 の次は 1 に戻る。
	return card.GetValue() == stalactitesNextRank(pile[len(pile)-1].GetValue())
}

// stalactitesNextRank は 13 の次を 1 に折り返した昇順の次ランク。
func stalactitesNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// foundationIndexFor は card を受け取れるファンデーションの番号を返す。無ければ -1。
//
// **Stalactites はスートを見ないので、FreeCell のように `design - 1` で
// パイルを決められない。**継続できるパイルを先に探し、無ければ空のパイルを
// 使う（この順にしないと、継続できるのに新しいパイルを開けてしまう）。
func (f *Stalactites) foundationIndexFor(card *Card) int {
	for i := 0; i < StalactitesFoundationCnt; i++ {
		if len(f.foundation[i]) > 0 && f.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	for i := 0; i < StalactitesFoundationCnt; i++ {
		if len(f.foundation[i]) == 0 && f.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// GetBaseRank はファンデーションの開始ランク（配りごとに変わる）。
func (f *Stalactites) GetBaseRank() int { return f.baseRank }

// isAlternateColor 交互の色かどうか判定
func (f *Stalactites) isAlternateColor(card1, card2 *Card) bool {
	return f.isBlack(card1) != f.isBlack(card2)
}

// isBlack 黒いカードかどうか
func (f *Stalactites) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// isValidTableauSequence タブロー移動可能なシーケンスか判定する。
// 通常のフリーセルでは降順かつ赤黒交互、Baker's Game では降順かつ同じスート。
func (f *Stalactites) isValidTableauSequence(cards []*Card) bool {
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
func (f *Stalactites) GetMaxMovableCards() int {
	return f.maxMovableCards(-1)
}

// GetMaxMovableCardsToEmptyColumn は空き列へ動かすときの上限を返す。
// 空き列が無いときは 0。
func (f *Stalactites) GetMaxMovableCardsToEmptyColumn() int {
	for i := 0; i < StalactitesTableauCnt; i++ {
		if len(f.tableau[i]) == 0 {
			return f.maxMovableCards(i)
		}
	}
	return 0
}

// maxMovableCards 移動可能な最大カード枚数を計算
// (1 + emptyCells) * 2^(emptyTableauCols) ただしtoColが空の場合はそのcolを除外
func (f *Stalactites) maxMovableCards(toCol int) int {
	emptyCells := 0
	for i := 0; i < StalactitesCellCnt; i++ {
		if f.cells[i] == nil {
			emptyCells++
		}
	}
	emptyTableauCols := 0
	for i := 0; i < StalactitesTableauCnt; i++ {
		if i != toCol && len(f.tableau[i]) == 0 {
			emptyTableauCols++
		}
	}
	return (1 + emptyCells) << emptyTableauCols
}

// checkGameClear ゲームクリア判定
func (f *Stalactites) checkGameClear() {
	for i := 0; i < StalactitesFoundationCnt; i++ {
		if len(f.foundation[i]) != CardValueMax {
			return
		}
	}
	f.phase = StalactitesPhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (f *Stalactites) takeSnapshot() {
	snap := &stalactitesSnapshot{
		phase:       f.phase,
		moveCount:   f.moveCount,
		isStalemate: f.isStalemate,
	}
	// deep copy tableau
	for i := 0; i < StalactitesTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(snap.tableau[i], f.tableau[i])
	}
	// deep copy cells
	snap.cells = f.cells
	// deep copy foundation
	for i := 0; i < StalactitesFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(f.foundation[i]))
		copy(snap.foundation[i], f.foundation[i])
	}
	f.history = append(f.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (f *Stalactites) restoreSnapshot(snap *stalactitesSnapshot) {
	f.tableau = snap.tableau
	f.cells = snap.cells
	f.foundation = snap.foundation
	f.phase = snap.phase
	f.moveCount = snap.moveCount
	f.isStalemate = snap.isStalemate
}

// checkStalemate ソルバーで手詰まり判定
func (f *Stalactites) checkStalemate() {
	if f.phase != StalactitesPhasePlaying {
		return
	}
	solver := newStalactitesSolver(f)
	f.isStalemate = !solver.isSolvable()
}

// appendLog 棋譜エントリを追加
func (f *Stalactites) appendLog(actionType, detail string, cards []*Card) {
	f.appendLogAt(f.moveCount, 0, actionType, detail, cards)
}

// stalactitesJSON is the JSON wire format for Stalactites.
type stalactitesJSON struct {
	TrumpCards  *TrumpCards                       `json:"tc"`
	Tableau     [StalactitesTableauCnt][]*Card    `json:"tb"`
	Cells       [StalactitesCellCnt]*Card         `json:"fc"`
	Foundation  [StalactitesFoundationCnt][]*Card `json:"fd"`
	Phase       StalactitesPhase                  `json:"ps"`
	MoveCount   int                               `json:"mc"`
	ActionLog   []*ActionLogEntry                 `json:"al"`
	IsStalemate bool                              `json:"sm"`
	History     []*stalactitesSnapshot            `json:"hi,omitempty"`
	// BaseRank はファンデーションの開始ランク。配りごとに変わるので、
	// これを保存しないと復元後にファンデーションへ1枚も置けなくなる。
	BaseRank int `json:"br,omitempty"`
}

// stalactitesSnapshotJSON is the wire format for a single undo snapshot.
// stalactitesSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// stalactitesJSON's short keys to keep the KV payload compact (#1654).
type stalactitesSnapshotJSON struct {
	Tableau     [StalactitesTableauCnt][]*Card    `json:"tb"`
	Cells       [StalactitesCellCnt]*Card         `json:"fc"`
	Foundation  [StalactitesFoundationCnt][]*Card `json:"fd"`
	Phase       StalactitesPhase                  `json:"ps"`
	MoveCount   int                               `json:"mc"`
	IsStalemate bool                              `json:"sm"`
}

// MarshalJSON implements json.Marshaler for stalactitesSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// Stalactites.MarshalJSON can persist the undo history (#1654).
func (s *stalactitesSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(stalactitesSnapshotJSON{
		Tableau:     s.tableau,
		Cells:       s.cells,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for stalactitesSnapshot.
func (s *stalactitesSnapshot) UnmarshalJSON(data []byte) error {
	var j stalactitesSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > stalactitesMaxSliceLen {
			return fmt.Errorf("stalactites: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > stalactitesMaxSliceLen {
			return fmt.Errorf("stalactites: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.cells = j.Cells
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (f *Stalactites) MarshalJSON() ([]byte, error) {
	return json.Marshal(stalactitesJSON{
		TrumpCards:  f.trumpCards,
		Tableau:     f.tableau,
		Cells:       f.cells,
		Foundation:  f.foundation,
		Phase:       f.phase,
		MoveCount:   f.moveCount,
		ActionLog:   f.actionLog,
		IsStalemate: f.isStalemate,
		History:     f.history,
		BaseRank:    f.baseRank,
	})
}

// stalactitesMaxSliceLen caps slice sizes during deserialisation.
const stalactitesMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (f *Stalactites) UnmarshalJSON(data []byte) error {
	var j stalactitesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > stalactitesMaxSliceLen || len(j.History) > stalactitesMaxSliceLen {
		return fmt.Errorf("stalactites: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > stalactitesMaxSliceLen {
			return fmt.Errorf("stalactites: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > stalactitesMaxSliceLen {
			return fmt.Errorf("stalactites: foundation pile exceeds maximum allowed size")
		}
	}

	f.trumpCards = j.TrumpCards
	if f.trumpCards == nil {
		f.trumpCards = NewTrumpCards(0)
	}
	f.tableau = j.Tableau
	f.cells = j.Cells
	f.foundation = j.Foundation
	f.phase = j.Phase
	f.moveCount = j.MoveCount
	f.actionLog = j.ActionLog
	if f.actionLog == nil {
		f.actionLog = make([]*ActionLogEntry, 0)
	}
	f.history = j.History
	if f.history == nil {
		f.history = make([]*stalactitesSnapshot, 0)
	}
	f.isStalemate = j.IsStalemate
	f.baseRank = j.BaseRank
	return nil
}
