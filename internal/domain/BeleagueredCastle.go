//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BeleagueredCastlePhase Beleaguered Castle game phase
type BeleagueredCastlePhase int

// BeleagueredCastle のフェーズ定数
const (
	// BeleagueredCastlePhasePlaying プレイ中
	BeleagueredCastlePhasePlaying BeleagueredCastlePhase = iota
	// BeleagueredCastlePhaseGameClear ゲームクリア
	BeleagueredCastlePhaseGameClear
	// BeleagueredCastlePhaseGameOver ゲームオーバー
	BeleagueredCastlePhaseGameOver
)

// BeleagueredCastleTableauCnt タブローの列数 (左右4列ずつ計8列)
const BeleagueredCastleTableauCnt = 8

// BeleagueredCastleColumnLen 初期配置時の各列のカード枚数
const BeleagueredCastleColumnLen = 6

// BeleagueredCastleFoundationCnt ファンデーションの数
const BeleagueredCastleFoundationCnt = 4

// BeleagueredCastleTableauCard タブロー上のカード
type BeleagueredCastleTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// BeleagueredCastleHint ヒント
type BeleagueredCastleHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// BeleagueredCastleConfig Beleaguered Castle ゲーム設定
type BeleagueredCastleConfig struct{}

// BeleagueredCastle ゲームクラス
type BeleagueredCastle struct {
	trumpCards *TrumpCards
	tableau    [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard
	foundation [BeleagueredCastleFoundationCnt][]*Card
	phase      BeleagueredCastlePhase
	moveCount  int
	actionLogBase
	history     []*beleagueredCastleSnapshot
	isStalemate bool
}

// beleagueredCastleSnapshot アンドゥ用スナップショット
type beleagueredCastleSnapshot struct {
	tableau     [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard
	foundation  [BeleagueredCastleFoundationCnt][]*Card
	phase       BeleagueredCastlePhase
	moveCount   int
	isStalemate bool
}

// NewBeleagueredCastle コンストラクタ
func NewBeleagueredCastle(trumpCards *TrumpCards) *BeleagueredCastle {
	return &BeleagueredCastle{
		trumpCards: trumpCards,
	}
}

// NewDefaultBeleagueredCastle returns BeleagueredCastle with a single 52-card deck.
func NewDefaultBeleagueredCastle() *BeleagueredCastle {
	return NewBeleagueredCastle(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout:
//   - 4 Aces are pre-seeded onto the four foundations (one per suit).
//   - The remaining 48 cards are dealt face-up across 8 tableau columns
//     (6 cards each).
func (bc *BeleagueredCastle) Reset() {
	bc.trumpCards.Shuffle()
	bc.phase = BeleagueredCastlePhasePlaying
	bc.moveCount = 0
	bc.actionLog = nil
	bc.history = nil
	bc.isStalemate = false

	for i := range BeleagueredCastleFoundationCnt {
		bc.foundation[i] = nil
	}
	for i := range BeleagueredCastleTableauCnt {
		bc.tableau[i] = make([]*BeleagueredCastleTableauCard, 0, BeleagueredCastleColumnLen)
	}

	// First pass: pull all 4 aces straight from the deck onto the
	// foundations, indexed by suit so that the same suit always
	// occupies the same foundation pile across runs.
	suitToFoundation := map[int]int{
		CardDesignSpade:   0,
		CardDesignClover:  1,
		CardDesignHeart:   2,
		CardDesignDiamond: 3,
	}
	remaining := make([]*Card, 0, CardCnt-BeleagueredCastleFoundationCnt)
	for {
		card := bc.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if card.GetValue() == 1 {
			fIdx, ok := suitToFoundation[card.GetDesign()]
			if !ok {
				// Joker or unknown suit — keep on tableau side; should not
				// happen with the default single 52-card deck.
				remaining = append(remaining, card)
				continue
			}
			bc.foundation[fIdx] = append(bc.foundation[fIdx], card)
			continue
		}
		remaining = append(remaining, card)
	}

	// Second pass: deal the remainder into the 8 columns, row by row,
	// so each column receives 6 cards.
	col := 0
	for _, card := range remaining {
		bc.tableau[col] = append(bc.tableau[col], &BeleagueredCastleTableauCard{Card: card, FaceUp: true})
		col = (col + 1) % BeleagueredCastleTableauCnt
	}

	bc.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (bc *BeleagueredCastle) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if bc.phase != BeleagueredCastlePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= BeleagueredCastleTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= BeleagueredCastleTableauCnt {
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
	bc.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	bc.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bc *BeleagueredCastle) MoveTableauToFoundation(col int) error {
	if bc.phase != BeleagueredCastlePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= BeleagueredCastleTableauCnt {
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
	bc.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (bc *BeleagueredCastle) GiveUp() {
	if bc.phase == BeleagueredCastlePhasePlaying {
		bc.phase = BeleagueredCastlePhaseGameOver
		bc.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (bc *BeleagueredCastle) GetHint() *BeleagueredCastleHint {
	if bc.phase != BeleagueredCastlePhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range BeleagueredCastleTableauCnt {
		if len(bc.tableau[col]) == 0 {
			continue
		}
		tc := bc.tableau[col][len(bc.tableau[col])-1]
		fIdx := bc.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &BeleagueredCastleHint{
				FromCol:   col,
				CardIndex: len(bc.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range BeleagueredCastleTableauCnt {
		fromCards := bc.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range BeleagueredCastleTableauCnt {
			if toCol == fromCol {
				continue
			}
			if bc.canPlaceOnTableau(card, toCol) {
				return &BeleagueredCastleHint{
					FromCol:   fromCol,
					CardIndex: len(fromCards) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全ての山から可能な限りファンデーションへ）
func (bc *BeleagueredCastle) AutoComplete() error {
	if bc.phase != BeleagueredCastlePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	bc.takeSnapshot()
	for {
		moved := false
		for col := range BeleagueredCastleTableauCnt {
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
	bc.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	bc.checkGameClear()
	bc.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Beleaguered Castle では常にtrue）
func (bc *BeleagueredCastle) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (bc *BeleagueredCastle) GetPhase() BeleagueredCastlePhase { return bc.phase }

// SetPhase フェーズ設定 (テスト用)
func (bc *BeleagueredCastle) SetPhase(phase BeleagueredCastlePhase) { bc.phase = phase }

// GetMoveCount 移動回数取得
func (bc *BeleagueredCastle) GetMoveCount() int { return bc.moveCount }

// GetTableau タブロー取得
func (bc *BeleagueredCastle) GetTableau() [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard {
	return bc.tableau
}

// GetFoundation ファンデーション取得
func (bc *BeleagueredCastle) GetFoundation() [BeleagueredCastleFoundationCnt][]*Card {
	return bc.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (bc *BeleagueredCastle) GetGameEndFlag() bool { return bc.phase != BeleagueredCastlePhasePlaying }

// IsStalemate 手詰まり状態取得
func (bc *BeleagueredCastle) IsStalemate() bool { return bc.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (bc *BeleagueredCastle) SetIsStalemate(v bool) { bc.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (bc *BeleagueredCastle) SetTableau(tableau [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard) {
	bc.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (bc *BeleagueredCastle) SetFoundation(foundation [BeleagueredCastleFoundationCnt][]*Card) {
	bc.foundation = foundation
}

// Undo 直前の操作を取り消す
func (bc *BeleagueredCastle) Undo() error {
	if bc.phase != BeleagueredCastlePhasePlaying {
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
func (bc *BeleagueredCastle) CanUndo() bool {
	return len(bc.history) > 0 && bc.phase == BeleagueredCastlePhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (bc *BeleagueredCastle) UndoToEscape() int {
	return undoToEscape(bc.isStalemate, bc.history, func(s *beleagueredCastleSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (bc *BeleagueredCastle) UndoN(n int) error {
	return undoN(bc, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
//
// Beleaguered Castle: 「スート無関係の降順」でタブロー間の移動可能。
// 空列には「任意のカード」を置ける (Baker's Dozen との違い)。
func (bc *BeleagueredCastle) canPlaceOnTableau(card *Card, col int) bool {
	colCards := bc.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (bc *BeleagueredCastle) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := bc.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (bc *BeleagueredCastle) findFoundation(card *Card) int {
	for i := range BeleagueredCastleFoundationCnt {
		if bc.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (bc *BeleagueredCastle) checkGameClear() {
	for i := range BeleagueredCastleFoundationCnt {
		if len(bc.foundation[i]) != CardValueMax {
			return
		}
	}
	bc.phase = BeleagueredCastlePhaseGameClear
}

// checkStalemate 手詰まり判定
func (bc *BeleagueredCastle) checkStalemate() {
	if bc.phase != BeleagueredCastlePhasePlaying {
		return
	}
	if bc.GetHint() != nil {
		bc.isStalemate = false
		return
	}
	bc.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (bc *BeleagueredCastle) takeSnapshot() {
	snap := &beleagueredCastleSnapshot{
		phase:       bc.phase,
		moveCount:   bc.moveCount,
		isStalemate: bc.isStalemate,
	}
	for i := range BeleagueredCastleTableauCnt {
		snap.tableau[i] = make([]*BeleagueredCastleTableauCard, len(bc.tableau[i]))
		for j, tc := range bc.tableau[i] {
			snap.tableau[i][j] = &BeleagueredCastleTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range BeleagueredCastleFoundationCnt {
		snap.foundation[i] = make([]*Card, len(bc.foundation[i]))
		copy(snap.foundation[i], bc.foundation[i])
	}
	bc.history = append(bc.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (bc *BeleagueredCastle) restoreSnapshot(snap *beleagueredCastleSnapshot) {
	bc.tableau = snap.tableau
	bc.foundation = snap.foundation
	bc.phase = snap.phase
	bc.moveCount = snap.moveCount
	bc.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (bc *BeleagueredCastle) appendLog(actionType, detail string, cards []*Card) {
	bc.appendLogAt(bc.moveCount, 0, actionType, detail, cards)
}

// beleagueredCastleJSON is the JSON wire format for BeleagueredCastle.
type beleagueredCastleJSON struct {
	TrumpCards  *TrumpCards                                                  `json:"tc"`
	Tableau     [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard `json:"tb"`
	Foundation  [BeleagueredCastleFoundationCnt][]*Card                      `json:"fd"`
	Phase       BeleagueredCastlePhase                                       `json:"ps"`
	MoveCount   int                                                          `json:"mc"`
	ActionLog   []*ActionLogEntry                                            `json:"al"`
	IsStalemate bool                                                         `json:"sl"`
	History     []*beleagueredCastleSnapshot                                 `json:"hi,omitempty"`
}

// beleagueredCastleSnapshotJSON mirrors bakersDozenSnapshotJSON; see the
// matching wire-format note on bakersDozenSnapshot for the rationale behind
// the short keys (KV payload size, issue #1654).
type beleagueredCastleSnapshotJSON struct {
	Tableau     [BeleagueredCastleTableauCnt][]*BeleagueredCastleTableauCard `json:"tb"`
	Foundation  [BeleagueredCastleFoundationCnt][]*Card                      `json:"fd"`
	Phase       BeleagueredCastlePhase                                       `json:"ps"`
	MoveCount   int                                                          `json:"mc"`
	IsStalemate bool                                                         `json:"sl"`
}

// MarshalJSON implements json.Marshaler for beleagueredCastleSnapshot.
func (s *beleagueredCastleSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(beleagueredCastleSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for beleagueredCastleSnapshot.
func (s *beleagueredCastleSnapshot) UnmarshalJSON(data []byte) error {
	var j beleagueredCastleSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > beleagueredCastleMaxSliceLen {
			return fmt.Errorf("beleagueredcastle: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > beleagueredCastleMaxSliceLen {
			return fmt.Errorf("beleagueredcastle: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range BeleagueredCastleTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*BeleagueredCastleTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range BeleagueredCastleFoundationCnt {
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
func (bc *BeleagueredCastle) MarshalJSON() ([]byte, error) {
	return json.Marshal(beleagueredCastleJSON{
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

// beleagueredCastleMaxSliceLen caps slice sizes during deserialisation.
const beleagueredCastleMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bc *BeleagueredCastle) UnmarshalJSON(data []byte) error {
	var j beleagueredCastleJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > beleagueredCastleMaxSliceLen ||
		len(j.History) > beleagueredCastleMaxSliceLen {
		return fmt.Errorf("beleagueredcastle: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > beleagueredCastleMaxSliceLen {
			return fmt.Errorf("beleagueredcastle: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > beleagueredCastleMaxSliceLen {
			return fmt.Errorf("beleagueredcastle: foundation pile exceeds maximum allowed size")
		}
	}

	bc.trumpCards = j.TrumpCards
	if bc.trumpCards == nil {
		bc.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	bc.tableau = j.Tableau
	for i := range BeleagueredCastleTableauCnt {
		if bc.tableau[i] == nil {
			bc.tableau[i] = make([]*BeleagueredCastleTableauCard, 0)
		}
	}
	bc.foundation = j.Foundation
	for i := range BeleagueredCastleFoundationCnt {
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
		bc.history = make([]*beleagueredCastleSnapshot, 0)
	}
	bc.isStalemate = j.IsStalemate
	return nil
}
