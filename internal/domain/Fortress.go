//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// FortressPhase Fortress game phase
type FortressPhase int

// Fortress のフェーズ定数
const (
	// FortressPhasePlaying プレイ中
	FortressPhasePlaying FortressPhase = iota
	// FortressPhaseGameClear ゲームクリア
	FortressPhaseGameClear
	// FortressPhaseGameOver ゲームオーバー
	FortressPhaseGameOver
)

// FortressTableauCnt タブローの列数 (左右4列ずつ計8列)
const FortressTableauCnt = 10

// FortressColumnLen 初期配置時の各列のカード枚数
// FortressColumnLen is the longest column after the deal: 52 = 5*10 + 2, so
// two columns hold 6 and the other eight hold 5. Used only as a capacity hint.
const FortressColumnLen = 6

// FortressFoundationCnt ファンデーションの数
const FortressFoundationCnt = 4

// FortressTableauCard タブロー上のカード
type FortressTableauCard = ColumnTableauCard

// FortressHint ヒント
type FortressHint = ColumnSolitaireHint

// FortressConfig フォートレスのゲーム設定
type FortressConfig struct{}

// Fortress ゲームクラス
type Fortress struct {
	trumpCards *TrumpCards
	tableau    [FortressTableauCnt][]*FortressTableauCard
	foundation [FortressFoundationCnt][]*Card
	phase      FortressPhase
	moveCount  int
	actionLogBase
	history     []*fortressSnapshot
	isStalemate bool
}

// fortressSnapshot アンドゥ用スナップショット
type fortressSnapshot struct {
	tableau     [FortressTableauCnt][]*FortressTableauCard
	foundation  [FortressFoundationCnt][]*Card
	phase       FortressPhase
	moveCount   int
	isStalemate bool
}

// NewFortress コンストラクタ
func NewFortress(trumpCards *TrumpCards) *Fortress {
	return &Fortress{
		trumpCards: trumpCards,
	}
}

// NewDefaultFortress returns Fortress with a single 52-card deck.
func NewDefaultFortress() *Fortress {
	return NewFortress(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout:
//   - 4 Aces are pre-seeded onto the four foundations (one per suit).
//   - The remaining 48 cards are dealt face-up across 8 tableau columns
//     (6 cards each).
func (bc *Fortress) Reset() {
	bc.trumpCards.Shuffle()
	bc.phase = FortressPhasePlaying
	bc.moveCount = 0
	bc.actionLog = nil
	bc.history = nil
	bc.isStalemate = false

	for i := range FortressFoundationCnt {
		bc.foundation[i] = nil
	}
	for i := range FortressTableauCnt {
		bc.tableau[i] = make([]*FortressTableauCard, 0, FortressColumnLen)
	}

	// Fortress deals the ENTIRE deck to the tableau. Beleaguered Castle, which
	// this domain was cloned from, first pulls the four Aces onto the
	// foundations; Fortress starts with empty foundations and the player must
	// dig each Ace out. Round-robin over 10 columns: 52 = 5*10 + 2, so the
	// first two columns end up with 6 cards and the rest with 5.
	col := 0
	for {
		card := bc.trumpCards.DrawCard()
		if card == nil {
			break
		}
		bc.tableau[col] = append(bc.tableau[col], &FortressTableauCard{Card: card, FaceUp: true})
		col = (col + 1) % FortressTableauCnt
	}

	bc.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (bc *Fortress) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if bc.phase != FortressPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= FortressTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= FortressTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := bc.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	if cardIndex != len(fromCards)-1 {
		return errors.New("only the top card can be moved")
	}
	tc := fromCards[cardIndex]
	if !bc.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	bc.takeSnapshot()
	bc.tableau[toCol] = append(bc.tableau[toCol], tc)
	bc.tableau[fromCol] = fromCards[:cardIndex]
	bc.moveCount++
	bc.appendLog("move", "fortress.log.moveTableauToTableau", map[string]string{"fromCol": strconv.Itoa(fromCol), "toCol": strconv.Itoa(toCol)}, []*Card{tc.Card})
	bc.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bc *Fortress) MoveTableauToFoundation(col int) error {
	if bc.phase != FortressPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FortressTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := bc.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := bc.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	bc.takeSnapshot()
	bc.tableau[col] = fromCards[:len(fromCards)-1]
	bc.foundation[fIdx] = append(bc.foundation[fIdx], card)
	bc.moveCount++
	bc.appendLog("move", "fortress.log.moveTableauToFoundation", map[string]string{"col": strconv.Itoa(col)}, []*Card{card})
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (bc *Fortress) GiveUp() {
	if bc.phase == FortressPhasePlaying {
		bc.phase = FortressPhaseGameOver
		bc.appendLog("giveup", "fortress.log.giveUp", nil, nil)
	}
}

// GetHint ヒントを取得
func (bc *Fortress) GetHint() *FortressHint {
	return columnGetHint(bc.phase == FortressPhasePlaying, bc.tableau[:], bc.canPlaceOnTableau, bc.findFoundation)
}

// AutoComplete オートコンプリート（全ての山から可能な限りファンデーションへ）
func (bc *Fortress) AutoComplete() error {
	if bc.phase != FortressPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	bc.takeSnapshot()
	for {
		moved := false
		for col := range FortressTableauCnt {
			if len(bc.tableau[col]) == 0 {
				continue
			}
			tc := bc.tableau[col][len(bc.tableau[col])-1]
			card := tc.Card
			fIdx := bc.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			bc.tableau[col] = bc.tableau[col][:len(bc.tableau[col])-1]
			bc.foundation[fIdx] = append(bc.foundation[fIdx], card)
			bc.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	bc.appendLog("autocomplete", "fortress.log.autoComplete", nil, nil)
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Fortress は全札を表向きに配るので常にtrue）
func (bc *Fortress) AllFaceUp() bool {
	return columnAllFaceUp(bc.tableau[:])
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (bc *Fortress) GetPhase() FortressPhase { return bc.phase }

// SetPhase フェーズ設定 (テスト用)
func (bc *Fortress) SetPhase(phase FortressPhase) { bc.phase = phase }

// GetMoveCount 移動回数取得
func (bc *Fortress) GetMoveCount() int { return bc.moveCount }

// GetTableau タブロー取得
func (bc *Fortress) GetTableau() [FortressTableauCnt][]*FortressTableauCard {
	return bc.tableau
}

// GetFoundation ファンデーション取得
func (bc *Fortress) GetFoundation() [FortressFoundationCnt][]*Card {
	return bc.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (bc *Fortress) GetGameEndFlag() bool { return bc.phase != FortressPhasePlaying }

// IsStalemate 手詰まり状態取得
func (bc *Fortress) IsStalemate() bool { return bc.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (bc *Fortress) SetIsStalemate(v bool) { bc.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (bc *Fortress) SetTableau(tableau [FortressTableauCnt][]*FortressTableauCard) {
	bc.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (bc *Fortress) SetFoundation(foundation [FortressFoundationCnt][]*Card) {
	bc.foundation = foundation
}

// Undo 直前の操作を取り消す
func (bc *Fortress) Undo() error {
	if bc.phase != FortressPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(bc.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := bc.history[len(bc.history)-1]
	bc.history = bc.history[:len(bc.history)-1]
	bc.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (bc *Fortress) CanUndo() bool {
	return len(bc.history) > 0 && bc.phase == FortressPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (bc *Fortress) UndoToEscape() int {
	return undoToEscape(bc.isStalemate, bc.history, func(s *fortressSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (bc *Fortress) UndoN(n int) error {
	return undoN(bc, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
//
// Fortress: 「同スートで隣接ランク（昇順・降順どちらでも）」でタブロー間の移動可能。
// Beleaguered Castle はスート無関係の降順のみ。
// 空列には「任意のカード」を置ける (Baker's Dozen との違い)。
func (bc *Fortress) canPlaceOnTableau(card *Card, col int) bool {
	colCards := bc.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	if card.GetDesign() != topCard.GetDesign() {
		return false
	}
	// Same suit, adjacent rank in EITHER direction. Beleaguered Castle ignores
	// suit and only builds down; both halves differ here.
	diff := card.GetValue() - topCard.GetValue()
	return diff == 1 || diff == -1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (bc *Fortress) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(bc.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (bc *Fortress) findFoundation(card *Card) int {
	return columnFindFoundation(bc.foundation[:], card)
}

// checkGameClear ゲームクリア判定
func (bc *Fortress) checkGameClear() {
	if columnCheckGameClear(bc.foundation[:]) {
		bc.phase = FortressPhaseGameClear
	}
}

// checkStalemate 手詰まり判定
func (bc *Fortress) checkStalemate() {
	if bc.phase == FortressPhasePlaying {
		bc.isStalemate = columnCheckStalemate(true, bc.tableau[:], bc.canPlaceOnTableau, bc.findFoundation)
	}
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (bc *Fortress) takeSnapshot() {
	snap := &fortressSnapshot{
		phase:       bc.phase,
		moveCount:   bc.moveCount,
		isStalemate: bc.isStalemate,
	}
	for i := range FortressTableauCnt {
		snap.tableau[i] = make([]*FortressTableauCard, len(bc.tableau[i]))
		for j, tc := range bc.tableau[i] {
			snap.tableau[i][j] = &FortressTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range FortressFoundationCnt {
		snap.foundation[i] = make([]*Card, len(bc.foundation[i]))
		copy(snap.foundation[i], bc.foundation[i])
	}
	bc.history = appendSnapshot(bc.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (bc *Fortress) restoreSnapshot(snap *fortressSnapshot) {
	bc.tableau = snap.tableau
	bc.foundation = snap.foundation
	bc.phase = snap.phase
	bc.moveCount = snap.moveCount
	bc.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (bc *Fortress) appendLog(actionType, detailCode string, detailParams map[string]string, cards []*Card) {
	bc.appendLogCodeAt(bc.moveCount, 0, actionType, detailCode, detailParams, cards)
}

// fortressJSON is the JSON wire format for Fortress.
type fortressJSON struct {
	TrumpCards  *TrumpCards                                `json:"tc"`
	Tableau     [FortressTableauCnt][]*FortressTableauCard `json:"tb"`
	Foundation  [FortressFoundationCnt][]*Card             `json:"fd"`
	Phase       FortressPhase                              `json:"ps"`
	MoveCount   int                                        `json:"mc"`
	ActionLog   []*ActionLogEntry                          `json:"al"`
	IsStalemate bool                                       `json:"sl"`
	History     []*fortressSnapshot                        `json:"hi,omitempty"`
}

// fortressSnapshotJSON mirrors bakersDozenSnapshotJSON; see the
// matching wire-format note on bakersDozenSnapshot for the rationale behind
// the short keys (KV payload size, issue #1654).
type fortressSnapshotJSON struct {
	Tableau     [FortressTableauCnt][]*FortressTableauCard `json:"tb"`
	Foundation  [FortressFoundationCnt][]*Card             `json:"fd"`
	Phase       FortressPhase                              `json:"ps"`
	MoveCount   int                                        `json:"mc"`
	IsStalemate bool                                       `json:"sl"`
}

// MarshalJSON implements json.Marshaler for fortressSnapshot.
func (s *fortressSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortressSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for fortressSnapshot.
func (s *fortressSnapshot) UnmarshalJSON(data []byte) error {
	var j fortressSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > fortressMaxSliceLen {
			return fmt.Errorf("fortress: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortressMaxSliceLen {
			return fmt.Errorf("fortress: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range FortressTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*FortressTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range FortressFoundationCnt {
		if s.foundation[i] == nil {
			s.foundation[i] = make([]*Card, 0)
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (bc *Fortress) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortressJSON{
		TrumpCards:  bc.trumpCards,
		Tableau:     bc.tableau,
		Foundation:  bc.foundation,
		Phase:       bc.phase,
		MoveCount:   bc.moveCount,
		ActionLog:   bc.actionLog,
		IsStalemate: bc.isStalemate,
		History:     bc.history,
	})
}

// fortressMaxSliceLen caps slice sizes during deserialisation.
const fortressMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bc *Fortress) UnmarshalJSON(data []byte) error {
	var j fortressJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > fortressMaxSliceLen ||
		len(j.History) > fortressMaxSliceLen {
		return fmt.Errorf("fortress: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > fortressMaxSliceLen {
			return fmt.Errorf("fortress: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortressMaxSliceLen {
			return fmt.Errorf("fortress: foundation pile exceeds maximum allowed size")
		}
	}

	bc.trumpCards = j.TrumpCards
	if bc.trumpCards == nil {
		bc.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	bc.tableau = j.Tableau
	for i := range FortressTableauCnt {
		if bc.tableau[i] == nil {
			bc.tableau[i] = make([]*FortressTableauCard, 0)
		}
	}
	bc.foundation = j.Foundation
	for i := range FortressFoundationCnt {
		if bc.foundation[i] == nil {
			bc.foundation[i] = make([]*Card, 0)
		}
	}
	bc.phase = j.Phase
	bc.moveCount = j.MoveCount
	bc.actionLog = j.ActionLog
	if bc.actionLog == nil {
		bc.actionLog = make([]*ActionLogEntry, 0)
	}
	bc.history = j.History
	if bc.history == nil {
		bc.history = make([]*fortressSnapshot, 0)
	}
	bc.isStalemate = j.IsStalemate
	return nil
}
