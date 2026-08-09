//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FortyThievesPhase フォーティシーブスゲームフェーズ
type FortyThievesPhase int

// FortyThievesのフェーズ定数
const (
	// FortyThievesPhasePlaying プレイ中
	FortyThievesPhasePlaying FortyThievesPhase = iota
	// FortyThievesPhaseGameClear ゲームクリア
	FortyThievesPhaseGameClear
	// FortyThievesPhaseGameOver ゲームオーバー
	FortyThievesPhaseGameOver
)

// FortyThievesTableauCnt タブローの列数
const FortyThievesTableauCnt = 10

// FortyThievesFoundationCnt ファンデーションの数
const FortyThievesFoundationCnt = 8

// FortyThievesTableauCard タブロー上のカード
type FortyThievesTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// FortyThievesHint ヒント
type FortyThievesHint struct {
	FromZone  string // "waste" or "tableau"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// FortyThievesConfig フォーティシーブスゲーム設定
type FortyThievesConfig struct{}

// FortyThieves フォーティシーブスゲームクラス
type FortyThieves struct {
	trumpCards *TrumpCards
	tableau    [FortyThievesTableauCnt][]*FortyThievesTableauCard
	stock      []*Card
	waste      []*Card
	foundation [FortyThievesFoundationCnt][]*Card
	phase      FortyThievesPhase
	moveCount  int
	actionLogBase
	history     []*fortyThievesSnapshot
	isStalemate bool
}

// fortyThievesSnapshot アンドゥ用スナップショット
type fortyThievesSnapshot struct {
	tableau     [FortyThievesTableauCnt][]*FortyThievesTableauCard
	stock       []*Card
	waste       []*Card
	foundation  [FortyThievesFoundationCnt][]*Card
	phase       FortyThievesPhase
	moveCount   int
	isStalemate bool
}

// NewFortyThieves コンストラクタ
func NewFortyThieves(trumpCards *TrumpCards) *FortyThieves {
	return &FortyThieves{
		trumpCards: trumpCards,
	}
}

// NewDefaultFortyThieves returns FortyThieves with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFortyThieves() *FortyThieves {
	return NewFortyThieves(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (ft *FortyThieves) Reset() {
	ft.trumpCards.Shuffle()
	ft.phase = FortyThievesPhasePlaying
	ft.moveCount = 0
	ft.actionLog = nil
	ft.history = nil
	ft.isStalemate = false

	// タブローに配る: 各列4枚、すべて表向き
	for i := range FortyThievesTableauCnt {
		ft.tableau[i] = make([]*FortyThievesTableauCard, 0, 4)
		for range 4 {
			card := ft.trumpCards.DrawCard()
			tc := &FortyThievesTableauCard{
				Card:   card,
				FaceUp: true,
			}
			ft.tableau[i] = append(ft.tableau[i], tc)
		}
	}

	// ファンデーション初期化
	for i := range FortyThievesFoundationCnt {
		ft.foundation[i] = nil
	}

	// 残りをストックへ（64枚）
	ft.stock = nil
	ft.waste = nil
	for ft.trumpCards.GetRemainingCount() > 0 {
		card := ft.trumpCards.DrawCard()
		ft.stock = append(ft.stock, card)
	}
}

// Draw ストックからウェイストにカードを1枚引く（リサイクルなし）
func (ft *FortyThieves) Draw() error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ft.stock) == 0 {
		return errors.New("no cards in stock")
	}
	ft.takeSnapshot()
	card := ft.stock[len(ft.stock)-1]
	ft.stock = ft.stock[:len(ft.stock)-1]
	ft.waste = append(ft.waste, card)
	ft.moveCount++
	ft.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	ft.checkFortyThievesStalemate()
	return nil
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ft *FortyThieves) MoveWasteToTableau(col int) error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FortyThievesTableauCnt {
		return errors.New("invalid column")
	}
	if len(ft.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ft.waste[len(ft.waste)-1]
	if !ft.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	ft.takeSnapshot()
	ft.waste = ft.waste[:len(ft.waste)-1]
	ft.tableau[col] = append(ft.tableau[col], &FortyThievesTableauCard{Card: card, FaceUp: true})
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	ft.checkFortyThievesStalemate()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ft *FortyThieves) MoveWasteToFoundation() error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ft.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ft.waste[len(ft.waste)-1]
	fIdx := ft.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ft.takeSnapshot()
	ft.waste = ft.waste[:len(ft.waste)-1]
	ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
	ft.moveCount++
	ft.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	ft.checkGameClear()
	ft.checkFortyThievesStalemate()
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (ft *FortyThieves) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= FortyThievesTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= FortyThievesTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := ft.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	// フォーティシーブスでは1枚のみ移動可能
	if cardIndex != len(fromCards)-1 {
		return errors.New("only the top card can be moved")
	}
	tc := fromCards[cardIndex]
	if !ft.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	ft.takeSnapshot()
	ft.tableau[toCol] = append(ft.tableau[toCol], tc)
	ft.tableau[fromCol] = fromCards[:cardIndex]
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	ft.checkFortyThievesStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ft *FortyThieves) MoveTableauToFoundation(col int) error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FortyThievesTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := ft.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := ft.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ft.takeSnapshot()
	ft.tableau[col] = fromCards[:len(fromCards)-1]
	ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	ft.checkGameClear()
	ft.checkFortyThievesStalemate()
	return nil
}

// GiveUp ギブアップ
func (ft *FortyThieves) GiveUp() {
	if ft.phase == FortyThievesPhasePlaying {
		ft.phase = FortyThievesPhaseGameOver
		ft.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (ft *FortyThieves) GetHint() *FortyThievesHint {
	if ft.phase != FortyThievesPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range FortyThievesTableauCnt {
		if len(ft.tableau[col]) == 0 {
			continue
		}
		tc := ft.tableau[col][len(ft.tableau[col])-1]
		fIdx := ft.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &FortyThievesHint{
				FromZone:  "tableau",
				FromCol:   col,
				CardIndex: len(ft.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: ウェイストからファンデーションへ
	if len(ft.waste) > 0 {
		card := ft.waste[len(ft.waste)-1]
		fIdx := ft.findFoundation(card)
		if fIdx >= 0 {
			return &FortyThievesHint{
				FromZone:  "waste",
				FromCol:   -1,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := range FortyThievesTableauCnt {
		fromCards := ft.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range FortyThievesTableauCnt {
			if toCol == fromCol {
				continue
			}
			// 空列への移動はヒントとして提示しない（意味のない移動）
			if len(ft.tableau[toCol]) == 0 {
				continue
			}
			if ft.canPlaceOnTableau(card, toCol) {
				return &FortyThievesHint{
					FromZone:  "tableau",
					FromCol:   fromCol,
					CardIndex: len(fromCards) - 1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	// 優先度4: ウェイストからタブローへ
	if len(ft.waste) > 0 {
		card := ft.waste[len(ft.waste)-1]
		for toCol := range FortyThievesTableauCnt {
			if ft.canPlaceOnTableau(card, toCol) {
				return &FortyThievesHint{
					FromZone:  "waste",
					FromCol:   -1,
					CardIndex: -1,
					ToZone:    "tableau",
					ToCol:     toCol,
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（ストックが空の場合に自動でファンデーションへ移動）
func (ft *FortyThieves) AutoComplete() error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !ft.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	ft.takeSnapshot()
	for {
		moved := false
		// ウェイストからファンデーションへ
		for len(ft.waste) > 0 {
			card := ft.waste[len(ft.waste)-1]
			fIdx := ft.findFoundation(card)
			if fIdx < 0 {
				break
			}
			ft.waste = ft.waste[:len(ft.waste)-1]
			ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
			ft.moveCount++
			moved = true
		}
		// タブローからファンデーションへ
		for col := range FortyThievesTableauCnt {
			if len(ft.tableau[col]) == 0 {
				continue
			}
			tc := ft.tableau[col][len(ft.tableau[col])-1]
			card := tc.Card
			fIdx := ft.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			ft.tableau[col] = ft.tableau[col][:len(ft.tableau[col])-1]
			ft.foundation[fIdx] = append(ft.foundation[fIdx], card)
			ft.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	ft.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	ft.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（ストックが空 = 全カード可視）
func (ft *FortyThieves) AllFaceUp() bool {
	return len(ft.stock) == 0
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (ft *FortyThieves) GetPhase() FortyThievesPhase { return ft.phase }

// SetPhase フェーズ設定 (テスト用)
func (ft *FortyThieves) SetPhase(phase FortyThievesPhase) { ft.phase = phase }

// GetMoveCount 移動回数取得
func (ft *FortyThieves) GetMoveCount() int { return ft.moveCount }

// GetStockCount ストック枚数取得
func (ft *FortyThieves) GetStockCount() int { return len(ft.stock) }

// GetWaste ウェイスト取得
func (ft *FortyThieves) GetWaste() []*Card { return ft.waste }

// GetTableau タブロー取得
func (ft *FortyThieves) GetTableau() [FortyThievesTableauCnt][]*FortyThievesTableauCard {
	return ft.tableau
}

// GetFoundation ファンデーション取得
func (ft *FortyThieves) GetFoundation() [FortyThievesFoundationCnt][]*Card { return ft.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (ft *FortyThieves) GetGameEndFlag() bool { return ft.phase != FortyThievesPhasePlaying }

// IsStalemate 手詰まり状態取得
func (ft *FortyThieves) IsStalemate() bool { return ft.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (ft *FortyThieves) SetIsStalemate(v bool) { ft.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (ft *FortyThieves) SetTableau(tableau [FortyThievesTableauCnt][]*FortyThievesTableauCard) {
	ft.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (ft *FortyThieves) SetStock(stock []*Card) { ft.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (ft *FortyThieves) SetWaste(waste []*Card) { ft.waste = waste }

// SetFoundation ファンデーション設定 (テスト用)
func (ft *FortyThieves) SetFoundation(foundation [FortyThievesFoundationCnt][]*Card) {
	ft.foundation = foundation
}

// Undo 直前の操作を取り消す
func (ft *FortyThieves) Undo() error {
	if ft.phase != FortyThievesPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(ft.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := ft.history[len(ft.history)-1]
	ft.history = ft.history[:len(ft.history)-1]
	ft.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (ft *FortyThieves) CanUndo() bool {
	return len(ft.history) > 0 && ft.phase == FortyThievesPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (ft *FortyThieves) UndoToEscape() int {
	return undoToEscape(ft.isStalemate, ft.history, func(s *fortyThievesSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (ft *FortyThieves) UndoN(n int) error {
	for i := range n {
		if err := ft.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定（同スート降順）
func (ft *FortyThieves) canPlaceOnTableau(card *Card, col int) bool {
	colCards := ft.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはどのカードでも置ける
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	// 同スートで降順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (ft *FortyThieves) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := ft.foundation[fIdx]
	if len(pile) == 0 {
		// 空のファンデーションにはAのみ置ける
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	// 同じスートで昇順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (ft *FortyThieves) findFoundation(card *Card) int {
	for i := range FortyThievesFoundationCnt {
		if ft.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (ft *FortyThieves) checkGameClear() {
	for i := range FortyThievesFoundationCnt {
		if len(ft.foundation[i]) != CardValueMax {
			return
		}
	}
	ft.phase = FortyThievesPhaseGameClear
}

// checkFortyThievesStalemate 手詰まり判定
func (ft *FortyThieves) checkFortyThievesStalemate() {
	if ft.phase != FortyThievesPhasePlaying {
		return
	}
	hint := ft.GetHint()
	if hint != nil {
		ft.isStalemate = false
		return
	}
	// ヒントがない場合
	if len(ft.stock) == 0 && len(ft.waste) == 0 {
		// ストックもウェイストも空で移動先なし → 手詰まり
		ft.isStalemate = true
		return
	}
	if len(ft.stock) == 0 {
		// ストック空でリサイクルなし、ウェイストのカードも移動不可 → 手詰まり
		ft.isStalemate = true
		return
	}
	// ストックにカードが残っている場合はまだ引ける
	ft.isStalemate = false
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (ft *FortyThieves) takeSnapshot() {
	snap := &fortyThievesSnapshot{
		phase:       ft.phase,
		moveCount:   ft.moveCount,
		isStalemate: ft.isStalemate,
	}
	// deep copy tableau
	for i := range FortyThievesTableauCnt {
		snap.tableau[i] = make([]*FortyThievesTableauCard, len(ft.tableau[i]))
		for j, tc := range ft.tableau[i] {
			snap.tableau[i][j] = &FortyThievesTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(ft.stock))
	copy(snap.stock, ft.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(ft.waste))
	copy(snap.waste, ft.waste)
	// deep copy foundation
	for i := range FortyThievesFoundationCnt {
		snap.foundation[i] = make([]*Card, len(ft.foundation[i]))
		copy(snap.foundation[i], ft.foundation[i])
	}
	ft.history = append(ft.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (ft *FortyThieves) restoreSnapshot(snap *fortyThievesSnapshot) {
	ft.tableau = snap.tableau
	ft.stock = snap.stock
	ft.waste = snap.waste
	ft.foundation = snap.foundation
	ft.phase = snap.phase
	ft.moveCount = snap.moveCount
	ft.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (ft *FortyThieves) appendLog(actionType, detail string, cards []*Card) {
	ft.appendLogAt(ft.moveCount, 0, actionType, detail, cards)
}

// fortyThievesJSON is the JSON wire format for FortyThieves.
type fortyThievesJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Tableau     [FortyThievesTableauCnt][]*FortyThievesTableauCard `json:"tb"`
	Stock       []*Card                                            `json:"st"`
	Waste       []*Card                                            `json:"wa"`
	Foundation  [FortyThievesFoundationCnt][]*Card                 `json:"fd"`
	Phase       FortyThievesPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	History     []*fortyThievesSnapshot                            `json:"hi,omitempty"`
}

// fortyThievesSnapshotJSON is the wire format for a single undo snapshot.
// fortyThievesSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// fortyThievesJSON's short keys to keep the KV payload compact (#1654).
type fortyThievesSnapshotJSON struct {
	Tableau     [FortyThievesTableauCnt][]*FortyThievesTableauCard `json:"tb"`
	Stock       []*Card                                            `json:"st"`
	Waste       []*Card                                            `json:"wa"`
	Foundation  [FortyThievesFoundationCnt][]*Card                 `json:"fd"`
	Phase       FortyThievesPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
}

// MarshalJSON implements json.Marshaler for fortyThievesSnapshot.
func (s *fortyThievesSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyThievesSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Waste:       s.waste,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for fortyThievesSnapshot.
func (s *fortyThievesSnapshot) UnmarshalJSON(data []byte) error {
	var j fortyThievesSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > fortyThievesMaxSliceLen || len(j.Waste) > fortyThievesMaxSliceLen {
		return fmt.Errorf("fortythieves: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > fortyThievesMaxSliceLen {
			return fmt.Errorf("fortythieves: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortyThievesMaxSliceLen {
			return fmt.Errorf("fortythieves: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.waste = j.Waste
	if s.waste == nil {
		s.waste = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ft *FortyThieves) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyThievesJSON{
		TrumpCards:  ft.trumpCards,
		Tableau:     ft.tableau,
		Stock:       ft.stock,
		Waste:       ft.waste,
		Foundation:  ft.foundation,
		Phase:       ft.phase,
		MoveCount:   ft.moveCount,
		ActionLog:   ft.actionLog,
		IsStalemate: ft.isStalemate,
		History:     ft.history,
	})
}

// fortyThievesMaxSliceLen caps slice sizes during deserialisation.
const fortyThievesMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (ft *FortyThieves) UnmarshalJSON(data []byte) error {
	var j fortyThievesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > fortyThievesMaxSliceLen || len(j.Waste) > fortyThievesMaxSliceLen ||
		len(j.ActionLog) > fortyThievesMaxSliceLen || len(j.History) > fortyThievesMaxSliceLen {
		return fmt.Errorf("fortythieves: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > fortyThievesMaxSliceLen {
			return fmt.Errorf("fortythieves: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortyThievesMaxSliceLen {
			return fmt.Errorf("fortythieves: foundation pile exceeds maximum allowed size")
		}
	}

	ft.trumpCards = j.TrumpCards
	if ft.trumpCards == nil {
		ft.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	ft.tableau = j.Tableau
	ft.stock = j.Stock
	if ft.stock == nil {
		ft.stock = make([]*Card, 0)
	}
	ft.waste = j.Waste
	if ft.waste == nil {
		ft.waste = make([]*Card, 0)
	}
	ft.foundation = j.Foundation
	ft.phase = j.Phase
	ft.moveCount = j.MoveCount
	ft.actionLog = j.ActionLog
	if ft.actionLog == nil {
		ft.actionLog = make([]*ActionLogEntry, 0)
	}
	ft.history = j.History
	if ft.history == nil {
		ft.history = make([]*fortyThievesSnapshot, 0)
	}
	ft.isStalemate = j.IsStalemate
	return nil
}
