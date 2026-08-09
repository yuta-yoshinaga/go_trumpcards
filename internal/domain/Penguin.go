//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PenguinPhase ペンギンゲームフェーズ
type PenguinPhase int

// Penguinのフェーズ定数
const (
	// PenguinPhasePlaying プレイ中
	PenguinPhasePlaying PenguinPhase = iota
	// PenguinPhaseGameClear ゲームクリア
	PenguinPhaseGameClear
	// PenguinPhaseGameOver ゲームオーバー
	PenguinPhaseGameOver
)

// PenguinTableauCnt タブローの列数
const PenguinTableauCnt = 7

// PenguinFoundationCnt ファンデーションの数
const PenguinFoundationCnt = 4

// PenguinCellCnt フリーセル（フリッパー）の数
const PenguinCellCnt = 7

// PenguinTableauColCards 初期配置時に各タブロー列に配るカード枚数
const PenguinTableauColCards = 7

// PenguinHint ヒント
type PenguinHint struct {
	FromZone  string // "tableau" or "freecell"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau", "foundation", or "freecell"
	ToCol     int
}

// Penguin ペンギンゲームクラス
type Penguin struct {
	trumpCards *TrumpCards
	tableau    [PenguinTableauCnt][]*Card
	freeCells  [PenguinCellCnt]*Card
	foundation [PenguinFoundationCnt][]*Card
	baseRank   int
	phase      PenguinPhase
	moveCount  int
	actionLogBase
	history     []*penguinSnapshot
	isStalemate bool
}

// penguinSnapshot アンドゥ用スナップショット
type penguinSnapshot struct {
	tableau     [PenguinTableauCnt][]*Card
	freeCells   [PenguinCellCnt]*Card
	foundation  [PenguinFoundationCnt][]*Card
	baseRank    int
	phase       PenguinPhase
	moveCount   int
	isStalemate bool
}

// NewPenguin コンストラクタ
func NewPenguin(trumpCards *TrumpCards) *Penguin {
	return &Penguin{
		trumpCards: trumpCards,
	}
}

// NewDefaultPenguin returns Penguin with a standard single 52-card deck.
func NewDefaultPenguin() *Penguin {
	return NewPenguin(NewTrumpCards(0))
}

// Reset ゲームリセット
func (p *Penguin) Reset() {
	p.trumpCards.Shuffle()
	p.phase = PenguinPhasePlaying
	p.moveCount = 0
	p.actionLog = nil
	p.history = nil
	p.isStalemate = false

	// フリーセル初期化
	for i := 0; i < PenguinCellCnt; i++ {
		p.freeCells[i] = nil
	}

	// ファンデーション初期化
	for i := 0; i < PenguinFoundationCnt; i++ {
		p.foundation[i] = nil
	}

	// デッキから全カードを取り出す
	deck := make([]*Card, 0, 52)
	for p.trumpCards.GetRemainingCount() > 0 {
		deck = append(deck, p.trumpCards.DrawCard())
	}

	if len(deck) == 0 {
		return
	}

	// 1枚目のカードでベースランクを決定
	p.baseRank = deck[0].GetValue()

	// ベースランクと同じランクの残り3枚を探してフリーセルに配置
	tableauDeck := make([]*Card, 0, 49)
	tableauDeck = append(tableauDeck, deck[0])
	cellIdx := 0
	for i := 1; i < len(deck); i++ {
		if deck[i].GetValue() == p.baseRank && cellIdx < 3 {
			p.freeCells[cellIdx] = deck[i]
			cellIdx++
		} else {
			tableauDeck = append(tableauDeck, deck[i])
		}
	}

	// タブローに配る: 7列×7枚 = 49枚
	for i := 0; i < PenguinTableauCnt; i++ {
		p.tableau[i] = make([]*Card, 0, PenguinTableauColCards)
		for j := 0; j < PenguinTableauColCards; j++ {
			idx := i*PenguinTableauColCards + j
			if idx < len(tableauDeck) {
				p.tableau[i] = append(p.tableau[i], tableauDeck[idx])
			}
		}
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動（スーパームーブ対応）
func (p *Penguin) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= PenguinTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= PenguinTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := p.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}

	movingCards := fromCards[cardIndex:]
	if !p.isValidTableauSequence(movingCards) {
		return errors.New("cards do not form a valid sequence")
	}

	maxCards := p.maxMovableCards(toCol)
	if len(movingCards) > maxCards {
		return errors.New("too many cards to move")
	}

	bottomCard := movingCards[0]
	if !p.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	p.takeSnapshot()
	p.tableau[toCol] = append(p.tableau[toCol], movingCards...)
	p.tableau[fromCol] = fromCards[:cardIndex]
	p.moveCount++
	movedCards := make([]*Card, len(movingCards))
	copy(movedCards, movingCards)
	p.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	p.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (p *Penguin) MoveTableauToFoundation(col int) error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= PenguinTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := p.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= PenguinFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !p.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	p.takeSnapshot()
	p.tableau[col] = fromCards[:len(fromCards)-1]
	p.foundation[fIdx] = append(p.foundation[fIdx], card)
	p.moveCount++
	p.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// MoveTableauToFreeCell タブローからフリーセルにカードを移動
func (p *Penguin) MoveTableauToFreeCell(col, cell int) error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= PenguinTableauCnt {
		return errors.New("invalid column")
	}
	if cell < 0 || cell >= PenguinCellCnt {
		return errors.New("invalid cell")
	}
	fromCards := p.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if p.freeCells[cell] != nil {
		return errors.New("free cell is occupied")
	}
	p.takeSnapshot()
	card := fromCards[len(fromCards)-1]
	p.tableau[col] = fromCards[:len(fromCards)-1]
	p.freeCells[cell] = card
	p.moveCount++
	p.appendLog("move", fmt.Sprintf("タブロー列%d→フリーセル%d", col, cell), []*Card{card})
	p.checkStalemate()
	return nil
}

// MoveFreeCellToTableau フリーセルからタブローにカードを移動
func (p *Penguin) MoveFreeCellToTableau(cell, col int) error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= PenguinCellCnt {
		return errors.New("invalid cell")
	}
	if col < 0 || col >= PenguinTableauCnt {
		return errors.New("invalid column")
	}
	if p.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := p.freeCells[cell]
	if !p.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	p.takeSnapshot()
	p.freeCells[cell] = nil
	p.tableau[col] = append(p.tableau[col], card)
	p.moveCount++
	p.appendLog("move", fmt.Sprintf("フリーセル%d→タブロー列%d", cell, col), []*Card{card})
	p.checkStalemate()
	return nil
}

// MoveFreeCellToFoundation フリーセルからファンデーションにカードを移動
func (p *Penguin) MoveFreeCellToFoundation(cell int) error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= PenguinCellCnt {
		return errors.New("invalid cell")
	}
	if p.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := p.freeCells[cell]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= PenguinFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !p.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	p.takeSnapshot()
	p.freeCells[cell] = nil
	p.foundation[fIdx] = append(p.foundation[fIdx], card)
	p.moveCount++
	p.appendLog("move", fmt.Sprintf("フリーセル%d→ファンデーション", cell), []*Card{card})
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (p *Penguin) GiveUp() {
	if p.phase == PenguinPhasePlaying {
		p.phase = PenguinPhaseGameOver
		p.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (p *Penguin) GetHint() *PenguinHint {
	if p.phase != PenguinPhasePlaying {
		return nil
	}
	emptyColRank := p.prevRank(p.baseRank)

	// 優先度1: タブローからファンデーションへ
	for col := 0; col < PenguinTableauCnt; col++ {
		if len(p.tableau[col]) == 0 {
			continue
		}
		card := p.tableau[col][len(p.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < PenguinFoundationCnt && p.canPlaceOnFoundation(card, fIdx) {
			return &PenguinHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(p.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: フリーセルからファンデーションへ
	for cell := 0; cell < PenguinCellCnt; cell++ {
		if p.freeCells[cell] == nil {
			continue
		}
		card := p.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < PenguinFoundationCnt && p.canPlaceOnFoundation(card, fIdx) {
			return &PenguinHint{
				FromZone:  "freecell",
				FromCol:   cell,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := 0; fromCol < PenguinTableauCnt; fromCol++ {
		fromCards := p.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		seqStart := len(fromCards) - 1
		for seqStart > 0 {
			if fromCards[seqStart].GetDesign() != fromCards[seqStart-1].GetDesign() ||
				fromCards[seqStart].GetValue() != p.prevRank(fromCards[seqStart-1].GetValue()) {
				break
			}
			seqStart--
		}
		card := fromCards[seqStart]
		movingCards := fromCards[seqStart:]
		for toCol := 0; toCol < PenguinTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			if len(p.tableau[toCol]) == 0 && card.GetValue() != emptyColRank {
				continue
			}
			maxCards := p.maxMovableCards(toCol)
			if len(movingCards) > maxCards {
				continue
			}
			if p.canPlaceOnTableau(card, toCol) {
				return &PenguinHint{
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
	for cell := 0; cell < PenguinCellCnt; cell++ {
		if p.freeCells[cell] == nil {
			continue
		}
		card := p.freeCells[cell]
		for toCol := 0; toCol < PenguinTableauCnt; toCol++ {
			if len(p.tableau[toCol]) == 0 && card.GetValue() != emptyColRank {
				continue
			}
			if p.canPlaceOnTableau(card, toCol) {
				return &PenguinHint{
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
	for col := 0; col < PenguinTableauCnt; col++ {
		if len(p.tableau[col]) == 0 {
			continue
		}
		for cell := 0; cell < PenguinCellCnt; cell++ {
			if p.freeCells[cell] == nil {
				return &PenguinHint{
					FromZone:  "tableau",
					FromCol:   col,
					CardIndex: len(p.tableau[col]) - 1,
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
func (p *Penguin) AutoComplete() error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	p.takeSnapshot()
	for {
		moved := false
		for cell := 0; cell < PenguinCellCnt; cell++ {
			if p.freeCells[cell] == nil {
				continue
			}
			card := p.freeCells[cell]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= PenguinFoundationCnt || !p.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			p.freeCells[cell] = nil
			p.foundation[fIdx] = append(p.foundation[fIdx], card)
			p.moveCount++
			moved = true
		}
		for col := 0; col < PenguinTableauCnt; col++ {
			if len(p.tableau[col]) == 0 {
				continue
			}
			card := p.tableau[col][len(p.tableau[col])-1]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= PenguinFoundationCnt || !p.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			p.tableau[col] = p.tableau[col][:len(p.tableau[col])-1]
			p.foundation[fIdx] = append(p.foundation[fIdx], card)
			p.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	p.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// Undo 直前の操作を取り消す
func (p *Penguin) Undo() error {
	if p.phase != PenguinPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(p.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := p.history[len(p.history)-1]
	p.history = p.history[:len(p.history)-1]
	p.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (p *Penguin) CanUndo() bool {
	return len(p.history) > 0 && p.phase == PenguinPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す
func (p *Penguin) UndoToEscape() int {
	return undoToEscape(p.isStalemate, p.history, func(s *penguinSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する
func (p *Penguin) UndoN(n int) error {
	return undoN(p, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (p *Penguin) GetPhase() PenguinPhase { return p.phase }

// SetPhase フェーズ設定 (テスト用)
func (p *Penguin) SetPhase(phase PenguinPhase) { p.phase = phase }

// GetMoveCount 移動回数取得
func (p *Penguin) GetMoveCount() int { return p.moveCount }

// GetTableau タブロー取得
func (p *Penguin) GetTableau() [PenguinTableauCnt][]*Card { return p.tableau }

// GetFreeCells フリーセル取得
func (p *Penguin) GetFreeCells() [PenguinCellCnt]*Card { return p.freeCells }

// GetFoundation ファンデーション取得
func (p *Penguin) GetFoundation() [PenguinFoundationCnt][]*Card { return p.foundation }

// GetBaseRank ベースランク取得
func (p *Penguin) GetBaseRank() int { return p.baseRank }

// SetBaseRank ベースランク設定 (テスト用)
func (p *Penguin) SetBaseRank(r int) { p.baseRank = r }

// GetGameEndFlag returns true once the game has left the playing phase.
func (p *Penguin) GetGameEndFlag() bool { return p.phase != PenguinPhasePlaying }

// IsStalemate 手詰まり状態取得
func (p *Penguin) IsStalemate() bool { return p.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (p *Penguin) SetIsStalemate(v bool) { p.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (p *Penguin) SetTableau(tableau [PenguinTableauCnt][]*Card) { p.tableau = tableau }

// SetFreeCells フリーセル設定 (テスト用)
func (p *Penguin) SetFreeCells(cells [PenguinCellCnt]*Card) { p.freeCells = cells }

// SetFoundation ファンデーション設定 (テスト用)
func (p *Penguin) SetFoundation(foundation [PenguinFoundationCnt][]*Card) {
	p.foundation = foundation
}

// --- Private helpers ---

// nextRank ラップアラウンド付き次のランク (K→A)
func (p *Penguin) nextRank(r int) int {
	return (r % 13) + 1
}

// prevRank ラップアラウンド付き前のランク (A→K)
func (p *Penguin) prevRank(r int) int {
	return ((r + 11) % 13) + 1
}

// canPlaceOnTableau タブローにカードを置けるか判定。
// 空列にはprevRank(baseRank)のカードしか置けず、非空列には同スート・prevRankのみ。
func (p *Penguin) canPlaceOnTableau(card *Card, col int) bool {
	colCards := p.tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == p.prevRank(p.baseRank)
	}
	topCard := colCards[len(colCards)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == p.prevRank(topCard.GetValue())
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定。
// 空パイルにはbaseRankのカードのみ、非空パイルには同スート・nextRankのみ。
func (p *Penguin) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := p.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == p.baseRank
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == p.nextRank(topCard.GetValue())
}

// isValidTableauSequence 同スート・prevRankのシーケンスか判定（ラップアラウンド対応）
func (p *Penguin) isValidTableauSequence(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if cards[i].GetDesign() != cards[i-1].GetDesign() ||
			cards[i].GetValue() != p.prevRank(cards[i-1].GetValue()) {
			return false
		}
	}
	return true
}

// GetMaxMovableCards はいま一度に動かせる最大枚数を返す (#4802)。
//
// **空き列を移動先にすると上限は下がる。**その列自身を経由地に使えないため。
// 移動先を選ぶ前の一般的な上限がこちらで、空き列に置くときの上限は
// GetMaxMovableCardsToEmptyColumn。
func (p *Penguin) GetMaxMovableCards() int {
	return p.maxMovableCards(-1)
}

// GetMaxMovableCardsToEmptyColumn は空き列へ動かすときの上限を返す。
// 空き列が無いときは 0。
func (p *Penguin) GetMaxMovableCardsToEmptyColumn() int {
	for i := 0; i < PenguinTableauCnt; i++ {
		if len(p.tableau[i]) == 0 {
			return p.maxMovableCards(i)
		}
	}
	return 0
}

// maxMovableCards 移動可能な最大カード枚数を計算
func (p *Penguin) maxMovableCards(toCol int) int {
	emptyFreeCells := 0
	for i := 0; i < PenguinCellCnt; i++ {
		if p.freeCells[i] == nil {
			emptyFreeCells++
		}
	}
	emptyTableauCols := 0
	for i := 0; i < PenguinTableauCnt; i++ {
		if i != toCol && len(p.tableau[i]) == 0 {
			emptyTableauCols++
		}
	}
	return (1 + emptyFreeCells) << emptyTableauCols
}

// checkGameClear ゲームクリア判定
func (p *Penguin) checkGameClear() {
	for i := 0; i < PenguinFoundationCnt; i++ {
		if len(p.foundation[i]) != CardValueMax {
			return
		}
	}
	p.phase = PenguinPhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (p *Penguin) takeSnapshot() {
	snap := &penguinSnapshot{
		baseRank:    p.baseRank,
		phase:       p.phase,
		moveCount:   p.moveCount,
		isStalemate: p.isStalemate,
	}
	for i := 0; i < PenguinTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(p.tableau[i]))
		copy(snap.tableau[i], p.tableau[i])
	}
	snap.freeCells = p.freeCells
	for i := 0; i < PenguinFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(p.foundation[i]))
		copy(snap.foundation[i], p.foundation[i])
	}
	p.history = append(p.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (p *Penguin) restoreSnapshot(snap *penguinSnapshot) {
	p.tableau = snap.tableau
	p.freeCells = snap.freeCells
	p.foundation = snap.foundation
	p.baseRank = snap.baseRank
	p.phase = snap.phase
	p.moveCount = snap.moveCount
	p.isStalemate = snap.isStalemate
}

// checkStalemate ソルバーで手詰まり判定
func (p *Penguin) checkStalemate() {
	if p.phase != PenguinPhasePlaying {
		return
	}
	solver := newPenguinSolver(p)
	p.isStalemate = !solver.isSolvable()
}

// appendLog 棋譜エントリを追加
func (p *Penguin) appendLog(actionType, detail string, cards []*Card) {
	p.appendLogAt(p.moveCount, 0, actionType, detail, cards)
}

// penguinJSON is the JSON wire format for Penguin.
type penguinJSON struct {
	TrumpCards  *TrumpCards                   `json:"tc"`
	Tableau     [PenguinTableauCnt][]*Card    `json:"tb"`
	FreeCells   [PenguinCellCnt]*Card         `json:"fc"`
	Foundation  [PenguinFoundationCnt][]*Card `json:"fd"`
	BaseRank    int                           `json:"br"`
	Phase       PenguinPhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	ActionLog   []*ActionLogEntry             `json:"al"`
	IsStalemate bool                          `json:"sm"`
	History     []*penguinSnapshot            `json:"hi,omitempty"`
}

// penguinSnapshotJSON is the wire format for a single undo snapshot.
type penguinSnapshotJSON struct {
	Tableau     [PenguinTableauCnt][]*Card    `json:"tb"`
	FreeCells   [PenguinCellCnt]*Card         `json:"fc"`
	Foundation  [PenguinFoundationCnt][]*Card `json:"fd"`
	BaseRank    int                           `json:"br"`
	Phase       PenguinPhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	IsStalemate bool                          `json:"sm"`
}

// MarshalJSON implements json.Marshaler for penguinSnapshot.
func (s *penguinSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(penguinSnapshotJSON{
		Tableau:     s.tableau,
		FreeCells:   s.freeCells,
		Foundation:  s.foundation,
		BaseRank:    s.baseRank,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// penguinMaxSliceLen caps slice sizes during deserialisation.
const penguinMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler for penguinSnapshot.
func (s *penguinSnapshot) UnmarshalJSON(data []byte) error {
	var j penguinSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > penguinMaxSliceLen {
			return fmt.Errorf("penguin: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > penguinMaxSliceLen {
			return fmt.Errorf("penguin: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.freeCells = j.FreeCells
	s.foundation = j.Foundation
	s.baseRank = j.BaseRank
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (p *Penguin) MarshalJSON() ([]byte, error) {
	return json.Marshal(penguinJSON{
		TrumpCards:  p.trumpCards,
		Tableau:     p.tableau,
		FreeCells:   p.freeCells,
		Foundation:  p.foundation,
		BaseRank:    p.baseRank,
		Phase:       p.phase,
		MoveCount:   p.moveCount,
		ActionLog:   p.actionLog,
		IsStalemate: p.isStalemate,
		History:     p.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (p *Penguin) UnmarshalJSON(data []byte) error {
	var j penguinJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > penguinMaxSliceLen || len(j.History) > penguinMaxSliceLen {
		return fmt.Errorf("penguin: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > penguinMaxSliceLen {
			return fmt.Errorf("penguin: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > penguinMaxSliceLen {
			return fmt.Errorf("penguin: foundation pile exceeds maximum allowed size")
		}
	}

	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.tableau = j.Tableau
	p.freeCells = j.FreeCells
	p.foundation = j.Foundation
	p.baseRank = j.BaseRank
	p.phase = j.Phase
	p.moveCount = j.MoveCount
	p.actionLog = j.ActionLog
	if p.actionLog == nil {
		p.actionLog = make([]*ActionLogEntry, 0)
	}
	p.history = j.History
	if p.history == nil {
		p.history = make([]*penguinSnapshot, 0)
	}
	p.isStalemate = j.IsStalemate
	return nil
}
