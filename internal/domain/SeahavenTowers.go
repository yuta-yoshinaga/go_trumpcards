//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SeahavenTowersPhase シーヘイブンタワーズゲームフェーズ
type SeahavenTowersPhase int

// SeahavenTowersのフェーズ定数
const (
	// SeahavenTowersPhasePlaying プレイ中
	SeahavenTowersPhasePlaying SeahavenTowersPhase = iota
	// SeahavenTowersPhaseGameClear ゲームクリア
	SeahavenTowersPhaseGameClear
	// SeahavenTowersPhaseGameOver ゲームオーバー
	SeahavenTowersPhaseGameOver
)

// SeahavenTowersTableauCnt タブローの列数 (Seahaven Towers は 10 列)
const SeahavenTowersTableauCnt = 10

// SeahavenTowersFoundationCnt ファンデーションの数
const SeahavenTowersFoundationCnt = 4

// SeahavenTowersCellCnt 上部リザーブセル (towers) の数 — Seahaven Towers は 2 セル
const SeahavenTowersCellCnt = 2

// SeahavenTowersTableauPerCol 初期配置でタブロー1列あたりに配るカード枚数
const SeahavenTowersTableauPerCol = 5

// SeahavenTowersHint ヒント
type SeahavenTowersHint struct {
	FromZone  string // "tableau" or "reserved"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau", "foundation", or "reserved"
	ToCol     int
}

// SeahavenTowers シーヘイブンタワーズゲームクラス。FreeCell とほぼ同じ構造を持つが
// 移動規則と空列規則が異なる:
//   - スタックは "同じスートで降順" (FreeCell は "交互の色で降順")
//   - 空列には King のみ置ける (FreeCell は任意のカードを置ける)
//   - 配り: タブロー 10 列 × 5 枚 + 上部リザーブセル 2 枚 (合計 52 枚)
type SeahavenTowers struct {
	trumpCards *TrumpCards
	tableau    [SeahavenTowersTableauCnt][]*Card
	freeCells  [SeahavenTowersCellCnt]*Card
	foundation [SeahavenTowersFoundationCnt][]*Card
	phase      SeahavenTowersPhase
	moveCount  int
	actionLogBase
	history     []*seahavenTowersSnapshot
	isStalemate bool
}

// seahavenTowersSnapshot アンドゥ用スナップショット
type seahavenTowersSnapshot struct {
	tableau     [SeahavenTowersTableauCnt][]*Card
	freeCells   [SeahavenTowersCellCnt]*Card
	foundation  [SeahavenTowersFoundationCnt][]*Card
	phase       SeahavenTowersPhase
	moveCount   int
	isStalemate bool
}

// NewSeahavenTowers コンストラクタ
func NewSeahavenTowers(trumpCards *TrumpCards) *SeahavenTowers {
	return &SeahavenTowers{
		trumpCards: trumpCards,
	}
}

// NewDefaultSeahavenTowers returns SeahavenTowers with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSeahavenTowers() *SeahavenTowers {
	return NewSeahavenTowers(NewTrumpCards(0))
}

// Reset ゲームリセット
func (s *SeahavenTowers) Reset() {
	s.trumpCards.Shuffle()
	s.phase = SeahavenTowersPhasePlaying
	s.moveCount = 0
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false

	// ファンデーション初期化
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		s.foundation[i] = nil
	}

	// タブローに配る: 10 列 × 5 枚
	for i := 0; i < SeahavenTowersTableauCnt; i++ {
		s.tableau[i] = make([]*Card, 0, SeahavenTowersTableauPerCol)
		for j := 0; j < SeahavenTowersTableauPerCol; j++ {
			card := s.trumpCards.DrawCard()
			s.tableau[i] = append(s.tableau[i], card)
		}
	}

	// 余った 2 枚をリザーブセルに配る
	for i := 0; i < SeahavenTowersCellCnt; i++ {
		s.freeCells[i] = s.trumpCards.DrawCard()
	}
}

// MoveTableauToTableau タブローからタブローにカードを移動 (スーパームーブ対応)
func (s *SeahavenTowers) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= SeahavenTowersTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= SeahavenTowersTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := s.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}

	// 移動するカード列が同スートの降順シーケンスか確認
	movingCards := fromCards[cardIndex:]
	if !s.isValidTableauSequence(movingCards) {
		return errors.New("cards do not form a valid sequence")
	}

	// 移動可能な最大枚数チェック (Seahaven は空列の倍率を適用しない)
	maxCards := s.maxMovableCards()
	if len(movingCards) > maxCards {
		return errors.New("too many cards to move")
	}

	bottomCard := movingCards[0]
	if !s.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}

	// 移動実行
	s.takeSnapshot()
	s.tableau[toCol] = append(s.tableau[toCol], movingCards...)
	s.tableau[fromCol] = fromCards[:cardIndex]
	s.moveCount++
	movedCards := make([]*Card, len(movingCards))
	copy(movedCards, movingCards)
	s.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	s.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (s *SeahavenTowers) MoveTableauToFoundation(col int) error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= SeahavenTowersTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := s.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= SeahavenTowersFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !s.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	s.takeSnapshot()
	s.tableau[col] = fromCards[:len(fromCards)-1]
	s.foundation[fIdx] = append(s.foundation[fIdx], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	s.checkGameClear()
	s.checkStalemate()
	return nil
}

// MoveTableauToFreeCell タブローからリザーブセルにカードを移動
func (s *SeahavenTowers) MoveTableauToFreeCell(col, cell int) error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= SeahavenTowersTableauCnt {
		return errors.New("invalid column")
	}
	if cell < 0 || cell >= SeahavenTowersCellCnt {
		return errors.New("invalid cell")
	}
	fromCards := s.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	if s.freeCells[cell] != nil {
		return errors.New("free cell is occupied")
	}
	s.takeSnapshot()
	card := fromCards[len(fromCards)-1]
	s.tableau[col] = fromCards[:len(fromCards)-1]
	s.freeCells[cell] = card
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("タブロー列%d→リザーブ%d", col, cell), []*Card{card})
	s.checkStalemate()
	return nil
}

// MoveFreeCellToTableau リザーブセルからタブローにカードを移動
func (s *SeahavenTowers) MoveFreeCellToTableau(cell, col int) error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= SeahavenTowersCellCnt {
		return errors.New("invalid cell")
	}
	if col < 0 || col >= SeahavenTowersTableauCnt {
		return errors.New("invalid column")
	}
	if s.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := s.freeCells[cell]
	if !s.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	s.takeSnapshot()
	s.freeCells[cell] = nil
	s.tableau[col] = append(s.tableau[col], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("リザーブ%d→タブロー列%d", cell, col), []*Card{card})
	s.checkStalemate()
	return nil
}

// MoveFreeCellToFoundation リザーブセルからファンデーションにカードを移動
func (s *SeahavenTowers) MoveFreeCellToFoundation(cell int) error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if cell < 0 || cell >= SeahavenTowersCellCnt {
		return errors.New("invalid cell")
	}
	if s.freeCells[cell] == nil {
		return errors.New("free cell is empty")
	}
	card := s.freeCells[cell]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= SeahavenTowersFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !s.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	s.takeSnapshot()
	s.freeCells[cell] = nil
	s.foundation[fIdx] = append(s.foundation[fIdx], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("リザーブ%d→ファンデーション", cell), []*Card{card})
	s.checkGameClear()
	s.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *SeahavenTowers) GiveUp() {
	if s.phase == SeahavenTowersPhasePlaying {
		s.phase = SeahavenTowersPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (s *SeahavenTowers) GetHint() *SeahavenTowersHint {
	if s.phase != SeahavenTowersPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < SeahavenTowersTableauCnt; col++ {
		if len(s.tableau[col]) == 0 {
			continue
		}
		card := s.tableau[col][len(s.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < SeahavenTowersFoundationCnt && s.canPlaceOnFoundation(card, fIdx) {
			return &SeahavenTowersHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(s.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: リザーブセルからファンデーションへ
	for cell := 0; cell < SeahavenTowersCellCnt; cell++ {
		if s.freeCells[cell] == nil {
			continue
		}
		card := s.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < SeahavenTowersFoundationCnt && s.canPlaceOnFoundation(card, fIdx) {
			return &SeahavenTowersHint{
				FromZone:  "reserved",
				FromCol:   cell,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブロー間 (同じスートの降順シーケンス)
	for fromCol := 0; fromCol < SeahavenTowersTableauCnt; fromCol++ {
		fromCards := s.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		seqStart := len(fromCards) - 1
		for seqStart > 0 {
			if !s.isSameSuitDescending(fromCards[seqStart], fromCards[seqStart-1]) {
				break
			}
			seqStart--
		}
		card := fromCards[seqStart]
		movingCards := fromCards[seqStart:]
		for toCol := 0; toCol < SeahavenTowersTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			// 空列は King のみ
			if len(s.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if len(movingCards) > s.maxMovableCards() {
				continue
			}
			if s.canPlaceOnTableau(card, toCol) {
				return &SeahavenTowersHint{
					FromZone:  "tableau",
					FromCol:   fromCol,
					CardIndex: seqStart,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度4: リザーブセルからタブローへ
	for cell := 0; cell < SeahavenTowersCellCnt; cell++ {
		if s.freeCells[cell] == nil {
			continue
		}
		card := s.freeCells[cell]
		for toCol := 0; toCol < SeahavenTowersTableauCnt; toCol++ {
			if len(s.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if s.canPlaceOnTableau(card, toCol) {
				return &SeahavenTowersHint{
					FromZone:  "reserved",
					FromCol:   cell,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度5: タブローからリザーブセルへ
	for col := 0; col < SeahavenTowersTableauCnt; col++ {
		if len(s.tableau[col]) == 0 {
			continue
		}
		for cell := 0; cell < SeahavenTowersCellCnt; cell++ {
			if s.freeCells[cell] == nil {
				return &SeahavenTowersHint{
					FromZone:  "tableau",
					FromCol:   col,
					CardIndex: len(s.tableau[col]) - 1,
					ToZone:    "reserved",
					ToCol:     cell,
				}
			}
		}
		break
	}
	return nil
}

// AutoComplete オートコンプリート
func (s *SeahavenTowers) AutoComplete() error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	s.takeSnapshot()
	for {
		moved := false
		// リザーブセル→ファンデーション
		for cell := 0; cell < SeahavenTowersCellCnt; cell++ {
			if s.freeCells[cell] == nil {
				continue
			}
			card := s.freeCells[cell]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= SeahavenTowersFoundationCnt || !s.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			s.freeCells[cell] = nil
			s.foundation[fIdx] = append(s.foundation[fIdx], card)
			s.moveCount++
			moved = true
		}
		// タブロー→ファンデーション
		for col := 0; col < SeahavenTowersTableauCnt; col++ {
			if len(s.tableau[col]) == 0 {
				continue
			}
			card := s.tableau[col][len(s.tableau[col])-1]
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= SeahavenTowersFoundationCnt || !s.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			s.tableau[col] = s.tableau[col][:len(s.tableau[col])-1]
			s.foundation[fIdx] = append(s.foundation[fIdx], card)
			s.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	s.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	s.checkGameClear()
	s.checkStalemate()
	return nil
}

// Undo 直前の操作を取り消す
func (s *SeahavenTowers) Undo() error {
	if s.phase != SeahavenTowersPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(s.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (s *SeahavenTowers) CanUndo() bool {
	return len(s.history) > 0 && s.phase == SeahavenTowersPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (s *SeahavenTowers) UndoToEscape() int {
	return undoToEscape(s.isStalemate, s.history, func(s *seahavenTowersSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (s *SeahavenTowers) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := s.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (s *SeahavenTowers) GetPhase() SeahavenTowersPhase { return s.phase }

// SetPhase フェーズ設定 (テスト用)
func (s *SeahavenTowers) SetPhase(phase SeahavenTowersPhase) { s.phase = phase }

// GetMoveCount 移動回数取得
func (s *SeahavenTowers) GetMoveCount() int { return s.moveCount }

// GetTableau タブロー取得
func (s *SeahavenTowers) GetTableau() [SeahavenTowersTableauCnt][]*Card { return s.tableau }

// GetFreeCells リザーブセル取得 (FreeCell との API 互換のため "FreeCells" 名を維持)
func (s *SeahavenTowers) GetFreeCells() [SeahavenTowersCellCnt]*Card { return s.freeCells }

// GetFoundation ファンデーション取得
func (s *SeahavenTowers) GetFoundation() [SeahavenTowersFoundationCnt][]*Card { return s.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (s *SeahavenTowers) GetGameEndFlag() bool { return s.phase != SeahavenTowersPhasePlaying }

// IsStalemate 手詰まり状態取得
func (s *SeahavenTowers) IsStalemate() bool { return s.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (s *SeahavenTowers) SetIsStalemate(v bool) { s.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (s *SeahavenTowers) SetTableau(tableau [SeahavenTowersTableauCnt][]*Card) {
	s.tableau = tableau
}

// SetFreeCells リザーブセル設定 (テスト用)
func (s *SeahavenTowers) SetFreeCells(cells [SeahavenTowersCellCnt]*Card) { s.freeCells = cells }

// SetFoundation ファンデーション設定 (テスト用)
func (s *SeahavenTowers) SetFoundation(foundation [SeahavenTowersFoundationCnt][]*Card) {
	s.foundation = foundation
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定。
// 空列には King のみ、それ以外は同じスートで 1 つ小さいランクのみ。
func (s *SeahavenTowers) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return s.isSameSuitDescending(card, topCard)
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (s *SeahavenTowers) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := s.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// isSameSuitDescending lower が upper の上に積めるか判定 (同スートで lower.value == upper.value - 1)
func (s *SeahavenTowers) isSameSuitDescending(lower, upper *Card) bool {
	return lower.GetDesign() == upper.GetDesign() && lower.GetValue() == upper.GetValue()-1
}

// isValidTableauSequence 同じスートの降順シーケンスか判定
func (s *SeahavenTowers) isValidTableauSequence(cards []*Card) bool {
	for i := 1; i < len(cards); i++ {
		if !s.isSameSuitDescending(cards[i], cards[i-1]) {
			return false
		}
	}
	return true
}

// maxMovableCards スーパームーブで移動可能な最大カード枚数を計算する。
// Seahaven Towers は空列が King のみ受け付けるため、FreeCell の "空列で枚数倍化"
// は適用せず、保守的に (1 + emptyFreeCells) を返す。
func (s *SeahavenTowers) maxMovableCards() int {
	emptyFreeCells := 0
	for i := 0; i < SeahavenTowersCellCnt; i++ {
		if s.freeCells[i] == nil {
			emptyFreeCells++
		}
	}
	return 1 + emptyFreeCells
}

// checkGameClear ゲームクリア判定
func (s *SeahavenTowers) checkGameClear() {
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		if len(s.foundation[i]) != CardValueMax {
			return
		}
	}
	s.phase = SeahavenTowersPhaseGameClear
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (s *SeahavenTowers) takeSnapshot() {
	snap := &seahavenTowersSnapshot{
		phase:       s.phase,
		moveCount:   s.moveCount,
		isStalemate: s.isStalemate,
	}
	for i := 0; i < SeahavenTowersTableauCnt; i++ {
		snap.tableau[i] = make([]*Card, len(s.tableau[i]))
		copy(snap.tableau[i], s.tableau[i])
	}
	snap.freeCells = s.freeCells
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(s.foundation[i]))
		copy(snap.foundation[i], s.foundation[i])
	}
	s.history = append(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *SeahavenTowers) restoreSnapshot(snap *seahavenTowersSnapshot) {
	s.tableau = snap.tableau
	s.freeCells = snap.freeCells
	s.foundation = snap.foundation
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.isStalemate = snap.isStalemate
}

// checkStalemate ソルバーで手詰まり判定
func (s *SeahavenTowers) checkStalemate() {
	if s.phase != SeahavenTowersPhasePlaying {
		return
	}
	solver := newSeahavenTowersSolver(s)
	s.isStalemate = !solver.isSolvable()
}

// appendLog 棋譜エントリを追加
func (s *SeahavenTowers) appendLog(actionType, detail string, cards []*Card) {
	s.appendLogAt(s.moveCount, 0, actionType, detail, cards)
}

// seahavenTowersJSON is the JSON wire format for SeahavenTowers.
type seahavenTowersJSON struct {
	TrumpCards  *TrumpCards                          `json:"tc"`
	Tableau     [SeahavenTowersTableauCnt][]*Card    `json:"tb"`
	FreeCells   [SeahavenTowersCellCnt]*Card         `json:"fc"`
	Foundation  [SeahavenTowersFoundationCnt][]*Card `json:"fd"`
	Phase       SeahavenTowersPhase                  `json:"ps"`
	MoveCount   int                                  `json:"mc"`
	ActionLog   []*ActionLogEntry                    `json:"al"`
	IsStalemate bool                                 `json:"sm"`
	History     []*seahavenTowersSnapshot            `json:"hi,omitempty"`
}

// seahavenTowersSnapshotJSON is the wire format for a single undo snapshot.
type seahavenTowersSnapshotJSON struct {
	Tableau     [SeahavenTowersTableauCnt][]*Card    `json:"tb"`
	FreeCells   [SeahavenTowersCellCnt]*Card         `json:"fc"`
	Foundation  [SeahavenTowersFoundationCnt][]*Card `json:"fd"`
	Phase       SeahavenTowersPhase                  `json:"ps"`
	MoveCount   int                                  `json:"mc"`
	IsStalemate bool                                 `json:"sm"`
}

// MarshalJSON implements json.Marshaler for seahavenTowersSnapshot.
func (s *seahavenTowersSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(seahavenTowersSnapshotJSON{
		Tableau:     s.tableau,
		FreeCells:   s.freeCells,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for seahavenTowersSnapshot.
func (s *seahavenTowersSnapshot) UnmarshalJSON(data []byte) error {
	var j seahavenTowersSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > seahavenTowersMaxSliceLen {
			return fmt.Errorf("seahavenTowers: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > seahavenTowersMaxSliceLen {
			return fmt.Errorf("seahavenTowers: snapshot foundation pile exceeds maximum allowed size")
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
func (s *SeahavenTowers) MarshalJSON() ([]byte, error) {
	return json.Marshal(seahavenTowersJSON{
		TrumpCards:  s.trumpCards,
		Tableau:     s.tableau,
		FreeCells:   s.freeCells,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		ActionLog:   s.actionLog,
		IsStalemate: s.isStalemate,
		History:     s.history,
	})
}

// seahavenTowersMaxSliceLen caps slice sizes during deserialisation.
const seahavenTowersMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (s *SeahavenTowers) UnmarshalJSON(data []byte) error {
	var j seahavenTowersJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > seahavenTowersMaxSliceLen || len(j.History) > seahavenTowersMaxSliceLen {
		return fmt.Errorf("seahavenTowers: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > seahavenTowersMaxSliceLen {
			return fmt.Errorf("seahavenTowers: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > seahavenTowersMaxSliceLen {
			return fmt.Errorf("seahavenTowers: foundation pile exceeds maximum allowed size")
		}
	}

	s.trumpCards = j.TrumpCards
	if s.trumpCards == nil {
		s.trumpCards = NewTrumpCards(0)
	}
	s.tableau = j.Tableau
	s.freeCells = j.FreeCells
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.actionLog = j.ActionLog
	if s.actionLog == nil {
		s.actionLog = make([]*ActionLogEntry, 0)
	}
	s.history = j.History
	if s.history == nil {
		s.history = make([]*seahavenTowersSnapshot, 0)
	}
	s.isStalemate = j.IsStalemate
	return nil
}
