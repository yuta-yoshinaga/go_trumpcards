//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BakersDozenPhase ベーカーズダズンゲームフェーズ
type BakersDozenPhase int

// BakersDozenのフェーズ定数
const (
	// BakersDozenPhasePlaying プレイ中
	BakersDozenPhasePlaying BakersDozenPhase = iota
	// BakersDozenPhaseGameClear ゲームクリア
	BakersDozenPhaseGameClear
	// BakersDozenPhaseGameOver ゲームオーバー
	BakersDozenPhaseGameOver
)

// BakersDozenTableauCnt タブローの列数
const BakersDozenTableauCnt = 13

// BakersDozenFoundationCnt ファンデーションの数
const BakersDozenFoundationCnt = 4

// BakersDozenTableauCard タブロー上のカード
type BakersDozenTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// BakersDozenHint ヒント
type BakersDozenHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// BakersDozenConfig ベーカーズダズンゲーム設定
type BakersDozenConfig struct{}

// BakersDozen ベーカーズダズンゲームクラス
type BakersDozen struct {
	trumpCards *TrumpCards
	tableau    [BakersDozenTableauCnt][]*BakersDozenTableauCard
	foundation [BakersDozenFoundationCnt][]*Card
	phase      BakersDozenPhase
	moveCount  int
	actionLogBase
	history     []*bakersDozenSnapshot
	isStalemate bool
}

// bakersDozenSnapshot アンドゥ用スナップショット
type bakersDozenSnapshot struct {
	tableau     [BakersDozenTableauCnt][]*BakersDozenTableauCard
	foundation  [BakersDozenFoundationCnt][]*Card
	phase       BakersDozenPhase
	moveCount   int
	isStalemate bool
}

// NewBakersDozen コンストラクタ
func NewBakersDozen(trumpCards *TrumpCards) *BakersDozen {
	return &BakersDozen{
		trumpCards: trumpCards,
	}
}

// NewDefaultBakersDozen returns BakersDozen with a single 52-card deck.
func NewDefaultBakersDozen() *BakersDozen {
	return NewBakersDozen(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
func (bd *BakersDozen) Reset() {
	bd.trumpCards.Shuffle()
	bd.phase = BakersDozenPhasePlaying
	bd.moveCount = 0
	bd.actionLog = nil
	bd.history = nil
	bd.isStalemate = false

	for i := range BakersDozenFoundationCnt {
		bd.foundation[i] = nil
	}

	// Deal 4 cards to each of 13 columns
	for i := range BakersDozenTableauCnt {
		bd.tableau[i] = make([]*BakersDozenTableauCard, 0, 4)
		for range 4 {
			card := bd.trumpCards.DrawCard()
			bd.tableau[i] = append(bd.tableau[i], &BakersDozenTableauCard{Card: card, FaceUp: true})
		}
	}

	// Move kings to the bottom of each column (signature rule of Baker's Dozen).
	// The bottom is index 0 (the buried card); the top is len-1 (the playable card).
	for i := range BakersDozenTableauCnt {
		col := bd.tableau[i]
		// Sort kings to the bottom while preserving relative order of other cards.
		kings := make([]*BakersDozenTableauCard, 0)
		others := make([]*BakersDozenTableauCard, 0, len(col))
		for _, tc := range col {
			if tc.Card.GetValue() == CardValueMax {
				kings = append(kings, tc)
			} else {
				others = append(others, tc)
			}
		}
		bd.tableau[i] = append(kings, others...)
	}

	// Detect a dead deal so the UI can offer Undo-to-Escape from move 0.
	bd.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (bd *BakersDozen) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if bd.phase != BakersDozenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= BakersDozenTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= BakersDozenTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := bd.tableau[fromCol]
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
	if !bd.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	bd.takeSnapshot()
	bd.tableau[toCol] = append(bd.tableau[toCol], tc)
	bd.tableau[fromCol] = fromCards[:cardIndex]
	bd.moveCount++
	bd.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	bd.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (bd *BakersDozen) MoveTableauToFoundation(col int) error {
	if bd.phase != BakersDozenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= BakersDozenTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := bd.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := bd.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	bd.takeSnapshot()
	bd.tableau[col] = fromCards[:len(fromCards)-1]
	bd.foundation[fIdx] = append(bd.foundation[fIdx], card)
	bd.moveCount++
	bd.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	bd.checkGameClear()
	bd.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (bd *BakersDozen) GiveUp() {
	if bd.phase == BakersDozenPhasePlaying {
		bd.phase = BakersDozenPhaseGameOver
		bd.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (bd *BakersDozen) GetHint() *BakersDozenHint {
	if bd.phase != BakersDozenPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range BakersDozenTableauCnt {
		if len(bd.tableau[col]) == 0 {
			continue
		}
		tc := bd.tableau[col][len(bd.tableau[col])-1]
		fIdx := bd.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &BakersDozenHint{
				FromCol:   col,
				CardIndex: len(bd.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range BakersDozenTableauCnt {
		fromCards := bd.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range BakersDozenTableauCnt {
			if toCol == fromCol {
				continue
			}
			if bd.canPlaceOnTableau(card, toCol) {
				return &BakersDozenHint{
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
func (bd *BakersDozen) AutoComplete() error {
	if bd.phase != BakersDozenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	bd.takeSnapshot()
	for {
		moved := false
		for col := range BakersDozenTableauCnt {
			if len(bd.tableau[col]) == 0 {
				continue
			}
			tc := bd.tableau[col][len(bd.tableau[col])-1]
			card := tc.Card
			fIdx := bd.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			bd.tableau[col] = bd.tableau[col][:len(bd.tableau[col])-1]
			bd.foundation[fIdx] = append(bd.foundation[fIdx], card)
			bd.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	bd.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	bd.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Baker's Dozenでは常にtrue）
func (bd *BakersDozen) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (bd *BakersDozen) GetPhase() BakersDozenPhase { return bd.phase }

// SetPhase フェーズ設定 (テスト用)
func (bd *BakersDozen) SetPhase(phase BakersDozenPhase) { bd.phase = phase }

// GetMoveCount 移動回数取得
func (bd *BakersDozen) GetMoveCount() int { return bd.moveCount }

// GetTableau タブロー取得
func (bd *BakersDozen) GetTableau() [BakersDozenTableauCnt][]*BakersDozenTableauCard {
	return bd.tableau
}

// GetFoundation ファンデーション取得
func (bd *BakersDozen) GetFoundation() [BakersDozenFoundationCnt][]*Card { return bd.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (bd *BakersDozen) GetGameEndFlag() bool { return bd.phase != BakersDozenPhasePlaying }

// IsStalemate 手詰まり状態取得
func (bd *BakersDozen) IsStalemate() bool { return bd.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (bd *BakersDozen) SetIsStalemate(v bool) { bd.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (bd *BakersDozen) SetTableau(tableau [BakersDozenTableauCnt][]*BakersDozenTableauCard) {
	bd.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (bd *BakersDozen) SetFoundation(foundation [BakersDozenFoundationCnt][]*Card) {
	bd.foundation = foundation
}

// Undo 直前の操作を取り消す
func (bd *BakersDozen) Undo() error {
	if bd.phase != BakersDozenPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(bd.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := bd.history[len(bd.history)-1]
	bd.history = bd.history[:len(bd.history)-1]
	bd.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (bd *BakersDozen) CanUndo() bool {
	return len(bd.history) > 0 && bd.phase == BakersDozenPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (bd *BakersDozen) UndoToEscape() int {
	if !bd.isStalemate {
		return 0
	}
	for i := len(bd.history) - 1; i >= 0; i-- {
		if !bd.history[i].isStalemate {
			return len(bd.history) - i
		}
	}
	return -1
}

// UndoN n回連続でアンドゥを実行する。
func (bd *BakersDozen) UndoN(n int) error {
	for i := range n {
		if err := bd.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
// Baker's Dozen: rank-only descending; empty columns cannot be filled.
func (bd *BakersDozen) canPlaceOnTableau(card *Card, col int) bool {
	colCards := bd.tableau[col]
	if len(colCards) == 0 {
		return false
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (bd *BakersDozen) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := bd.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (bd *BakersDozen) findFoundation(card *Card) int {
	for i := range BakersDozenFoundationCnt {
		if bd.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (bd *BakersDozen) checkGameClear() {
	for i := range BakersDozenFoundationCnt {
		if len(bd.foundation[i]) != CardValueMax {
			return
		}
	}
	bd.phase = BakersDozenPhaseGameClear
}

// checkStalemate 手詰まり判定
func (bd *BakersDozen) checkStalemate() {
	if bd.phase != BakersDozenPhasePlaying {
		return
	}
	if bd.GetHint() != nil {
		bd.isStalemate = false
		return
	}
	bd.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (bd *BakersDozen) takeSnapshot() {
	snap := &bakersDozenSnapshot{
		phase:       bd.phase,
		moveCount:   bd.moveCount,
		isStalemate: bd.isStalemate,
	}
	for i := range BakersDozenTableauCnt {
		snap.tableau[i] = make([]*BakersDozenTableauCard, len(bd.tableau[i]))
		for j, tc := range bd.tableau[i] {
			snap.tableau[i][j] = &BakersDozenTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range BakersDozenFoundationCnt {
		snap.foundation[i] = make([]*Card, len(bd.foundation[i]))
		copy(snap.foundation[i], bd.foundation[i])
	}
	bd.history = append(bd.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (bd *BakersDozen) restoreSnapshot(snap *bakersDozenSnapshot) {
	bd.tableau = snap.tableau
	bd.foundation = snap.foundation
	bd.phase = snap.phase
	bd.moveCount = snap.moveCount
	bd.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (bd *BakersDozen) appendLog(actionType, detail string, cards []*Card) {
	bd.appendLogAt(bd.moveCount, 0, actionType, detail, cards)
}

// bakersDozenJSON is the JSON wire format for BakersDozen.
type bakersDozenJSON struct {
	TrumpCards  *TrumpCards                                      `json:"tc"`
	Tableau     [BakersDozenTableauCnt][]*BakersDozenTableauCard `json:"tb"`
	Foundation  [BakersDozenFoundationCnt][]*Card                `json:"fd"`
	Phase       BakersDozenPhase                                 `json:"ps"`
	MoveCount   int                                              `json:"mc"`
	ActionLog   []*ActionLogEntry                                `json:"al"`
	IsStalemate bool                                             `json:"sl"`
	History     []*bakersDozenSnapshot                           `json:"hi,omitempty"`
}

// bakersDozenSnapshotJSON is the wire format for a single undo snapshot.
// bakersDozenSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// bakersDozenJSON's short keys to keep the KV payload compact (#1654).
type bakersDozenSnapshotJSON struct {
	Tableau     [BakersDozenTableauCnt][]*BakersDozenTableauCard `json:"tb"`
	Foundation  [BakersDozenFoundationCnt][]*Card                `json:"fd"`
	Phase       BakersDozenPhase                                 `json:"ps"`
	MoveCount   int                                              `json:"mc"`
	IsStalemate bool                                             `json:"sl"`
}

// MarshalJSON implements json.Marshaler for bakersDozenSnapshot.
func (s *bakersDozenSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(bakersDozenSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for bakersDozenSnapshot.
func (s *bakersDozenSnapshot) UnmarshalJSON(data []byte) error {
	var j bakersDozenSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > bakersDozenMaxSliceLen {
			return fmt.Errorf("bakersdozen: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > bakersDozenMaxSliceLen {
			return fmt.Errorf("bakersdozen: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range BakersDozenTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*BakersDozenTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range BakersDozenFoundationCnt {
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
func (bd *BakersDozen) MarshalJSON() ([]byte, error) {
	return json.Marshal(bakersDozenJSON{
		TrumpCards:  bd.trumpCards,
		Tableau:     bd.tableau,
		Foundation:  bd.foundation,
		Phase:       bd.phase,
		MoveCount:   bd.moveCount,
		ActionLog:   bd.actionLog,
		IsStalemate: bd.isStalemate,
		History:     bd.history,
	})
}

// bakersDozenMaxSliceLen caps slice sizes during deserialisation.
const bakersDozenMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (bd *BakersDozen) UnmarshalJSON(data []byte) error {
	var j bakersDozenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > bakersDozenMaxSliceLen ||
		len(j.History) > bakersDozenMaxSliceLen {
		return fmt.Errorf("bakersdozen: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > bakersDozenMaxSliceLen {
			return fmt.Errorf("bakersdozen: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > bakersDozenMaxSliceLen {
			return fmt.Errorf("bakersdozen: foundation pile exceeds maximum allowed size")
		}
	}

	bd.trumpCards = j.TrumpCards
	if bd.trumpCards == nil {
		bd.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	bd.tableau = j.Tableau
	for i := range BakersDozenTableauCnt {
		if bd.tableau[i] == nil {
			bd.tableau[i] = make([]*BakersDozenTableauCard, 0)
		}
	}
	bd.foundation = j.Foundation
	for i := range BakersDozenFoundationCnt {
		if bd.foundation[i] == nil {
			bd.foundation[i] = make([]*Card, 0)
		}
	}
	bd.phase = j.Phase
	bd.moveCount = j.MoveCount
	bd.actionLog = j.ActionLog
	if bd.actionLog == nil {
		bd.actionLog = make([]*ActionLogEntry, 0)
	}
	bd.history = j.History
	if bd.history == nil {
		bd.history = make([]*bakersDozenSnapshot, 0)
	}
	bd.isStalemate = j.IsStalemate
	return nil
}
