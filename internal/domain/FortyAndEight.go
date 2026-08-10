//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FortyAndEightPhase フォーティ・アンド・エイトゲームフェーズ
type FortyAndEightPhase int

// FortyAndEightのフェーズ定数
const (
	// FortyAndEightPhasePlaying プレイ中
	FortyAndEightPhasePlaying FortyAndEightPhase = iota
	// FortyAndEightPhaseGameClear ゲームクリア
	FortyAndEightPhaseGameClear
	// FortyAndEightPhaseGameOver ゲームオーバー
	FortyAndEightPhaseGameOver
)

// FortyAndEightTableauCnt タブローの列数
const FortyAndEightTableauCnt = 8

// FortyAndEightFoundationCnt ファンデーションの数
const FortyAndEightFoundationCnt = 8

// FortyAndEightTableauCard タブロー上のカード
type FortyAndEightTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// FortyAndEightHint ヒント
type FortyAndEightHint struct {
	FromZone  string // "waste" or "tableau"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// FortyAndEightConfig フォーティ・アンド・エイトゲーム設定
type FortyAndEightConfig struct{}

// FortyAndEight フォーティ・アンド・エイトゲームクラス
type FortyAndEight struct {
	trumpCards *TrumpCards
	tableau    [FortyAndEightTableauCnt][]*FortyAndEightTableauCard
	stock      []*Card
	waste      []*Card
	foundation [FortyAndEightFoundationCnt][]*Card
	phase      FortyAndEightPhase
	moveCount  int
	actionLogBase
	history     []*fortyAndEightSnapshot
	isStalemate bool
	redealUsed  bool
}

// fortyAndEightSnapshot アンドゥ用スナップショット
type fortyAndEightSnapshot struct {
	tableau     [FortyAndEightTableauCnt][]*FortyAndEightTableauCard
	stock       []*Card
	waste       []*Card
	foundation  [FortyAndEightFoundationCnt][]*Card
	phase       FortyAndEightPhase
	moveCount   int
	isStalemate bool
	redealUsed  bool
}

// NewFortyAndEight コンストラクタ
func NewFortyAndEight(trumpCards *TrumpCards) *FortyAndEight {
	return &FortyAndEight{
		trumpCards: trumpCards,
	}
}

// NewDefaultFortyAndEight returns FortyAndEight with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFortyAndEight() *FortyAndEight {
	return NewFortyAndEight(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (ft *FortyAndEight) Reset() {
	ft.trumpCards.Shuffle()
	ft.phase = FortyAndEightPhasePlaying
	ft.moveCount = 0
	ft.actionLog = nil
	ft.history = nil
	ft.isStalemate = false
	ft.redealUsed = false

	// タブローに配る: 各列5枚、すべて表向き
	for i := range FortyAndEightTableauCnt {
		ft.tableau[i] = make([]*FortyAndEightTableauCard, 0, 5)
		for range 5 {
			card := ft.trumpCards.DrawCard()
			tc := &FortyAndEightTableauCard{
				Card:   card,
				FaceUp: true,
			}
			ft.tableau[i] = append(ft.tableau[i], tc)
		}
	}

	// ファンデーション初期化
	for i := range FortyAndEightFoundationCnt {
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

// Draw ストックからウェイストにカードを1枚引く（リサイクルなし、リディールは別）
func (ft *FortyAndEight) Draw() error {
	if ft.phase != FortyAndEightPhasePlaying {
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
	ft.checkFortyAndEightStalemate()
	return nil
}

// Redeal ストックを使い切った後、ウェイストを集めて新しいストックを作る（1回限り）。
// ウェイストを反転してストックへ戻すことで、最初に引いたカードが再び最初に引かれる。
func (ft *FortyAndEight) Redeal() error {
	if ft.phase != FortyAndEightPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ft.stock) != 0 {
		return errors.New("cannot redeal: stock is not empty")
	}
	if ft.redealUsed {
		return errors.New("cannot redeal: redeal already used")
	}
	if len(ft.waste) == 0 {
		return errors.New("cannot redeal: waste is empty")
	}
	ft.takeSnapshot()
	// ウェイストを反転してストックへ（最初に引いたカードが再び最初に引かれる）
	ft.stock = make([]*Card, len(ft.waste))
	for i, c := range ft.waste {
		ft.stock[len(ft.waste)-1-i] = c
	}
	ft.waste = nil
	ft.redealUsed = true
	ft.moveCount++
	ft.appendLog("redeal", "ウェイストを集めて新しいストックを作りました", nil)
	ft.checkFortyAndEightStalemate()
	return nil
}

// MoveWasteToTableau ウェイストからタブローにカードを移動
func (ft *FortyAndEight) MoveWasteToTableau(col int) error {
	if ft.phase != FortyAndEightPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FortyAndEightTableauCnt {
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
	ft.tableau[col] = append(ft.tableau[col], &FortyAndEightTableauCard{Card: card, FaceUp: true})
	ft.moveCount++
	ft.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	ft.checkFortyAndEightStalemate()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (ft *FortyAndEight) MoveWasteToFoundation() error {
	if ft.phase != FortyAndEightPhasePlaying {
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
	ft.checkFortyAndEightStalemate()
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (ft *FortyAndEight) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if ft.phase != FortyAndEightPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= FortyAndEightTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= FortyAndEightTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := ft.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	// フォーティ・アンド・エイトでは1枚のみ移動可能
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
	ft.checkFortyAndEightStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (ft *FortyAndEight) MoveTableauToFoundation(col int) error {
	if ft.phase != FortyAndEightPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= FortyAndEightTableauCnt {
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
	ft.checkFortyAndEightStalemate()
	return nil
}

// GiveUp ギブアップ
func (ft *FortyAndEight) GiveUp() {
	if ft.phase == FortyAndEightPhasePlaying {
		ft.phase = FortyAndEightPhaseGameOver
		ft.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (ft *FortyAndEight) GetHint() *FortyAndEightHint {
	if ft.phase != FortyAndEightPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range FortyAndEightTableauCnt {
		if len(ft.tableau[col]) == 0 {
			continue
		}
		tc := ft.tableau[col][len(ft.tableau[col])-1]
		fIdx := ft.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &FortyAndEightHint{
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
			return &FortyAndEightHint{
				FromZone:  "waste",
				FromCol:   -1,
				CardIndex: -1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度3: タブローからタブローへ
	for fromCol := range FortyAndEightTableauCnt {
		fromCards := ft.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range FortyAndEightTableauCnt {
			if toCol == fromCol {
				continue
			}
			// 空列への移動はヒントとして提示しない（意味のない移動）
			if len(ft.tableau[toCol]) == 0 {
				continue
			}
			if ft.canPlaceOnTableau(card, toCol) {
				return &FortyAndEightHint{
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
		for toCol := range FortyAndEightTableauCnt {
			if ft.canPlaceOnTableau(card, toCol) {
				return &FortyAndEightHint{
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
func (ft *FortyAndEight) AutoComplete() error {
	if ft.phase != FortyAndEightPhasePlaying {
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
		for col := range FortyAndEightTableauCnt {
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
func (ft *FortyAndEight) AllFaceUp() bool {
	return len(ft.stock) == 0
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (ft *FortyAndEight) GetPhase() FortyAndEightPhase { return ft.phase }

// SetPhase フェーズ設定 (テスト用)
func (ft *FortyAndEight) SetPhase(phase FortyAndEightPhase) { ft.phase = phase }

// GetMoveCount 移動回数取得
func (ft *FortyAndEight) GetMoveCount() int { return ft.moveCount }

// GetStockCount ストック枚数取得
func (ft *FortyAndEight) GetStockCount() int { return len(ft.stock) }

// GetWaste ウェイスト取得
func (ft *FortyAndEight) GetWaste() []*Card { return ft.waste }

// GetTableau タブロー取得
func (ft *FortyAndEight) GetTableau() [FortyAndEightTableauCnt][]*FortyAndEightTableauCard {
	return ft.tableau
}

// GetFoundation ファンデーション取得
func (ft *FortyAndEight) GetFoundation() [FortyAndEightFoundationCnt][]*Card { return ft.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (ft *FortyAndEight) GetGameEndFlag() bool { return ft.phase != FortyAndEightPhasePlaying }

// IsStalemate 手詰まり状態取得
func (ft *FortyAndEight) IsStalemate() bool { return ft.isStalemate }

// GetRedealUsed リディール使用済みかどうかを返す
func (ft *FortyAndEight) GetRedealUsed() bool { return ft.redealUsed }

// CanRedeal リディール可能かどうかを返す（ストックが空 && 未使用 && ウェイストに残あり && プレイ中）
func (ft *FortyAndEight) CanRedeal() bool {
	return ft.phase == FortyAndEightPhasePlaying && len(ft.stock) == 0 && !ft.redealUsed && len(ft.waste) > 0
}

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (ft *FortyAndEight) SetIsStalemate(v bool) { ft.isStalemate = v }

// SetRedealUsed リディール使用済みフラグ設定 (テスト用)
func (ft *FortyAndEight) SetRedealUsed(v bool) { ft.redealUsed = v }

// SetTableau タブロー設定 (テスト用)
func (ft *FortyAndEight) SetTableau(tableau [FortyAndEightTableauCnt][]*FortyAndEightTableauCard) {
	ft.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (ft *FortyAndEight) SetStock(stock []*Card) { ft.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (ft *FortyAndEight) SetWaste(waste []*Card) { ft.waste = waste }

// SetFoundation ファンデーション設定 (テスト用)
func (ft *FortyAndEight) SetFoundation(foundation [FortyAndEightFoundationCnt][]*Card) {
	ft.foundation = foundation
}

// Undo 直前の操作を取り消す
func (ft *FortyAndEight) Undo() error {
	if ft.phase != FortyAndEightPhasePlaying {
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
func (ft *FortyAndEight) CanUndo() bool {
	return len(ft.history) > 0 && ft.phase == FortyAndEightPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (ft *FortyAndEight) UndoToEscape() int {
	return undoToEscape(ft.isStalemate, ft.history, func(s *fortyAndEightSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (ft *FortyAndEight) UndoN(n int) error {
	return undoN(ft, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定（同スート降順）
func (ft *FortyAndEight) canPlaceOnTableau(card *Card, col int) bool {
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
func (ft *FortyAndEight) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(ft.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (ft *FortyAndEight) findFoundation(card *Card) int {
	for i := range FortyAndEightFoundationCnt {
		if ft.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (ft *FortyAndEight) checkGameClear() {
	for i := range FortyAndEightFoundationCnt {
		if len(ft.foundation[i]) != CardValueMax {
			return
		}
	}
	ft.phase = FortyAndEightPhaseGameClear
}

// checkFortyAndEightStalemate 手詰まり判定
func (ft *FortyAndEight) checkFortyAndEightStalemate() {
	if ft.phase != FortyAndEightPhasePlaying {
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
		// ストック空、ウェイストのカードも移動不可。ただしリディール可能ならまだ手がある。
		if ft.CanRedeal() {
			ft.isStalemate = false
			return
		}
		ft.isStalemate = true
		return
	}
	// ストックにカードが残っている場合はまだ引ける
	ft.isStalemate = false
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (ft *FortyAndEight) takeSnapshot() {
	snap := &fortyAndEightSnapshot{
		phase:       ft.phase,
		moveCount:   ft.moveCount,
		isStalemate: ft.isStalemate,
		redealUsed:  ft.redealUsed,
	}
	// deep copy tableau
	for i := range FortyAndEightTableauCnt {
		snap.tableau[i] = make([]*FortyAndEightTableauCard, len(ft.tableau[i]))
		for j, tc := range ft.tableau[i] {
			snap.tableau[i][j] = &FortyAndEightTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(ft.stock))
	copy(snap.stock, ft.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(ft.waste))
	copy(snap.waste, ft.waste)
	// deep copy foundation
	for i := range FortyAndEightFoundationCnt {
		snap.foundation[i] = make([]*Card, len(ft.foundation[i]))
		copy(snap.foundation[i], ft.foundation[i])
	}
	ft.history = append(ft.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (ft *FortyAndEight) restoreSnapshot(snap *fortyAndEightSnapshot) {
	ft.tableau = snap.tableau
	ft.stock = snap.stock
	ft.waste = snap.waste
	ft.foundation = snap.foundation
	ft.phase = snap.phase
	ft.moveCount = snap.moveCount
	ft.isStalemate = snap.isStalemate
	ft.redealUsed = snap.redealUsed
}

// appendLog 棋譜エントリを追加
func (ft *FortyAndEight) appendLog(actionType, detail string, cards []*Card) {
	ft.appendLogAt(ft.moveCount, 0, actionType, detail, cards)
}

// fortyAndEightJSON is the JSON wire format for FortyAndEight.
type fortyAndEightJSON struct {
	TrumpCards  *TrumpCards                                          `json:"tc"`
	Tableau     [FortyAndEightTableauCnt][]*FortyAndEightTableauCard `json:"tb"`
	Stock       []*Card                                              `json:"st"`
	Waste       []*Card                                              `json:"wa"`
	Foundation  [FortyAndEightFoundationCnt][]*Card                  `json:"fd"`
	Phase       FortyAndEightPhase                                   `json:"ps"`
	MoveCount   int                                                  `json:"mc"`
	ActionLog   []*ActionLogEntry                                    `json:"al"`
	IsStalemate bool                                                 `json:"sl"`
	RedealUsed  bool                                                 `json:"rd"`
	History     []*fortyAndEightSnapshot                             `json:"hi,omitempty"`
}

// fortyAndEightSnapshotJSON is the wire format for a single undo snapshot.
// fortyAndEightSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// fortyAndEightJSON's short keys to keep the KV payload compact (#1654).
type fortyAndEightSnapshotJSON struct {
	Tableau     [FortyAndEightTableauCnt][]*FortyAndEightTableauCard `json:"tb"`
	Stock       []*Card                                              `json:"st"`
	Waste       []*Card                                              `json:"wa"`
	Foundation  [FortyAndEightFoundationCnt][]*Card                  `json:"fd"`
	Phase       FortyAndEightPhase                                   `json:"ps"`
	MoveCount   int                                                  `json:"mc"`
	IsStalemate bool                                                 `json:"sl"`
	RedealUsed  bool                                                 `json:"rd"`
}

// MarshalJSON implements json.Marshaler for fortyAndEightSnapshot.
func (s *fortyAndEightSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyAndEightSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Waste:       s.waste,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
		RedealUsed:  s.redealUsed,
	})
}

// UnmarshalJSON implements json.Unmarshaler for fortyAndEightSnapshot.
func (s *fortyAndEightSnapshot) UnmarshalJSON(data []byte) error {
	var j fortyAndEightSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > fortyAndEightMaxSliceLen || len(j.Waste) > fortyAndEightMaxSliceLen {
		return fmt.Errorf("fortyandeight: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > fortyAndEightMaxSliceLen {
			return fmt.Errorf("fortyandeight: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortyAndEightMaxSliceLen {
			return fmt.Errorf("fortyandeight: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for _, col := range s.tableau {
		for _, tc := range col {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("fortyandeight: snapshot tableau contains a nil card")
			}
		}
	}
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	for _, c := range s.stock {
		if c == nil {
			return fmt.Errorf("fortyandeight: snapshot stock contains a nil card")
		}
	}
	s.waste = j.Waste
	if s.waste == nil {
		s.waste = make([]*Card, 0)
	}
	for _, c := range s.waste {
		if c == nil {
			return fmt.Errorf("fortyandeight: snapshot waste contains a nil card")
		}
	}
	s.foundation = j.Foundation
	for _, pile := range s.foundation {
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("fortyandeight: snapshot foundation contains a nil card")
			}
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.redealUsed = j.RedealUsed
	return nil
}

// MarshalJSON implements json.Marshaler.
func (ft *FortyAndEight) MarshalJSON() ([]byte, error) {
	return json.Marshal(fortyAndEightJSON{
		TrumpCards:  ft.trumpCards,
		Tableau:     ft.tableau,
		Stock:       ft.stock,
		Waste:       ft.waste,
		Foundation:  ft.foundation,
		Phase:       ft.phase,
		MoveCount:   ft.moveCount,
		ActionLog:   ft.actionLog,
		IsStalemate: ft.isStalemate,
		RedealUsed:  ft.redealUsed,
		History:     ft.history,
	})
}

// fortyAndEightMaxSliceLen caps slice sizes during deserialisation.
const fortyAndEightMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (ft *FortyAndEight) UnmarshalJSON(data []byte) error {
	var j fortyAndEightJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > fortyAndEightMaxSliceLen || len(j.Waste) > fortyAndEightMaxSliceLen ||
		len(j.ActionLog) > fortyAndEightMaxSliceLen || len(j.History) > fortyAndEightMaxSliceLen {
		return fmt.Errorf("fortyandeight: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > fortyAndEightMaxSliceLen {
			return fmt.Errorf("fortyandeight: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > fortyAndEightMaxSliceLen {
			return fmt.Errorf("fortyandeight: foundation pile exceeds maximum allowed size")
		}
	}

	ft.trumpCards = j.TrumpCards
	if ft.trumpCards == nil {
		ft.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	ft.tableau = j.Tableau
	for _, col := range ft.tableau {
		for _, tc := range col {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("fortyandeight: tableau contains a nil card")
			}
		}
	}
	ft.stock = j.Stock
	if ft.stock == nil {
		ft.stock = make([]*Card, 0)
	}
	for _, c := range ft.stock {
		if c == nil {
			return fmt.Errorf("fortyandeight: stock contains a nil card")
		}
	}
	ft.waste = j.Waste
	if ft.waste == nil {
		ft.waste = make([]*Card, 0)
	}
	for _, c := range ft.waste {
		if c == nil {
			return fmt.Errorf("fortyandeight: waste contains a nil card")
		}
	}
	ft.foundation = j.Foundation
	for _, pile := range ft.foundation {
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("fortyandeight: foundation contains a nil card")
			}
		}
	}
	ft.phase = j.Phase
	ft.moveCount = j.MoveCount
	ft.actionLog = j.ActionLog
	if ft.actionLog == nil {
		ft.actionLog = make([]*ActionLogEntry, 0)
	}
	ft.history = j.History
	if ft.history == nil {
		ft.history = make([]*fortyAndEightSnapshot, 0)
	}
	ft.isStalemate = j.IsStalemate
	ft.redealUsed = j.RedealUsed
	return nil
}
