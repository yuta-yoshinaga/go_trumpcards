//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// EasthavenPhase イーストヘイブンゲームフェーズ
type EasthavenPhase int

// Easthavenのフェーズ定数
const (
	// EasthavenPhasePlaying プレイ中
	EasthavenPhasePlaying EasthavenPhase = iota
	// EasthavenPhaseGameClear ゲームクリア
	EasthavenPhaseGameClear
	// EasthavenPhaseGameOver ゲームオーバー
	EasthavenPhaseGameOver
)

// EasthavenTableauCnt タブローの列数
const EasthavenTableauCnt = 7

// EasthavenFoundationCnt ファンデーションの数
const EasthavenFoundationCnt = 4

// EasthavenInitialColCards 各列の初期配り枚数（裏2枚＋表1枚）
const EasthavenInitialColCards = 3

// EasthavenHint ヒント
type EasthavenHint struct {
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" or "foundation"
	ToCol     int
}

// Easthaven イーストヘイブンゲームクラス
type Easthaven struct {
	trumpCards *TrumpCards
	tableau    [EasthavenTableauCnt][]*KlondikeTableauCard
	stock      []*Card
	foundation [EasthavenFoundationCnt][]*Card
	phase      EasthavenPhase
	moveCount  int
	actionLogBase
	history     []*easthavenSnapshot
	isStalemate bool
}

// easthavenSnapshot アンドゥ用スナップショット
type easthavenSnapshot struct {
	tableau     [EasthavenTableauCnt][]*KlondikeTableauCard
	stock       []*Card
	foundation  [EasthavenFoundationCnt][]*Card
	phase       EasthavenPhase
	moveCount   int
	isStalemate bool
}

// NewEasthaven コンストラクタ
func NewEasthaven(trumpCards *TrumpCards) *Easthaven {
	return &Easthaven{
		trumpCards: trumpCards,
	}
}

// NewDefaultEasthaven returns Easthaven with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultEasthaven() *Easthaven {
	return NewEasthaven(NewTrumpCards(0))
}

// Reset ゲームリセット
func (e *Easthaven) Reset() {
	e.trumpCards.Shuffle()
	e.phase = EasthavenPhasePlaying
	e.moveCount = 0
	e.actionLog = nil
	e.history = nil
	e.isStalemate = false

	// タブローに配る: 7列にそれぞれ3枚（裏2枚＋表1枚）
	for i := range EasthavenTableauCnt {
		e.tableau[i] = make([]*KlondikeTableauCard, 0, EasthavenInitialColCards)
		for j := range EasthavenInitialColCards {
			card := e.trumpCards.DrawCard()
			e.tableau[i] = append(e.tableau[i], &KlondikeTableauCard{
				Card:   card,
				FaceUp: j == EasthavenInitialColCards-1,
			})
		}
	}

	// ファンデーション初期化
	for i := range EasthavenFoundationCnt {
		e.foundation[i] = nil
	}

	// 残りをストックへ (31枚)
	e.stock = nil
	for e.trumpCards.GetRemainingCount() > 0 {
		card := e.trumpCards.DrawCard()
		e.stock = append(e.stock, card)
	}
}

// Deal ストックからタブローの各列に1枚ずつ表向きで配る（スパイダー方式）。
// 空の列がある場合は配れない。ストックが7枚未満の最後の配りは左から残り枚数分だけ配る。
func (e *Easthaven) Deal() error {
	if e.phase != EasthavenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(e.stock) == 0 {
		return errors.New("no cards in stock")
	}
	// 空列がある場合は配れない
	for i := range EasthavenTableauCnt {
		if len(e.tableau[i]) == 0 {
			return errors.New("cannot deal: empty column exists")
		}
	}
	e.takeSnapshot()
	count := EasthavenTableauCnt
	if count > len(e.stock) {
		count = len(e.stock)
	}
	for i := range count {
		card := e.stock[len(e.stock)-1]
		e.stock = e.stock[:len(e.stock)-1]
		e.tableau[i] = append(e.tableau[i], &KlondikeTableauCard{Card: card, FaceUp: true})
	}
	e.moveCount++
	e.appendLog("deal", "ストックから各列にカードを配りました", nil)
	e.checkEasthavenStalemate()
	return nil
}

// MoveTableauToTableau タブローからタブローにカードを移動
func (e *Easthaven) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if e.phase != EasthavenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= EasthavenTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= EasthavenTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := e.tableau[fromCol]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	tc := fromCards[cardIndex]
	if tc == nil || tc.Card == nil {
		return errors.New("card is nil")
	}
	if !tc.FaceUp {
		return errors.New("card is face down")
	}
	// 移動するカード列は交互色降順の連続でなければならない
	movingCards := fromCards[cardIndex:]
	if !e.isValidEasthavenSequence(movingCards) {
		return errors.New("cards are not a valid alternating-color descending sequence")
	}
	bottomCard := movingCards[0].Card
	if !e.canPlaceOnTableau(bottomCard, toCol) {
		return errors.New("cannot place card on tableau")
	}
	e.takeSnapshot()
	movedCards := make([]*Card, len(movingCards))
	for i, mc := range movingCards {
		e.tableau[toCol] = append(e.tableau[toCol], mc)
		movedCards[i] = mc.Card
	}
	e.tableau[fromCol] = fromCards[:cardIndex]
	e.autoFlipTableau(fromCol)
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), movedCards)
	e.checkEasthavenStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (e *Easthaven) MoveTableauToFoundation(col int) error {
	if e.phase != EasthavenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= EasthavenTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := e.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	if tc == nil || tc.Card == nil {
		return errors.New("card is nil")
	}
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= EasthavenFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !e.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	e.takeSnapshot()
	e.tableau[col] = fromCards[:len(fromCards)-1]
	e.foundation[fIdx] = append(e.foundation[fIdx], card)
	e.autoFlipTableau(col)
	e.moveCount++
	e.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	e.checkGameClear()
	e.checkEasthavenStalemate()
	return nil
}

// GiveUp ギブアップ
func (e *Easthaven) GiveUp() {
	if e.phase == EasthavenPhasePlaying {
		e.phase = EasthavenPhaseGameOver
		e.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (e *Easthaven) GetHint() *EasthavenHint {
	if e.phase != EasthavenPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range EasthavenTableauCnt {
		if len(e.tableau[col]) == 0 {
			continue
		}
		tc := e.tableau[col][len(e.tableau[col])-1]
		if tc == nil || tc.Card == nil {
			continue
		}
		card := tc.Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < EasthavenFoundationCnt && e.canPlaceOnFoundation(card, fIdx) {
			return &EasthavenHint{
				FromCol:   col,
				CardIndex: len(e.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: 裏カードを開けるタブロー間移動、優先度3: その他の有効な移動
	for _, exposeOnly := range []bool{true, false} {
		for fromCol := range EasthavenTableauCnt {
			fromCards := e.tableau[fromCol]
			if len(fromCards) == 0 {
				continue
			}
			firstFaceUp := -1
			for i, tc := range fromCards {
				if tc != nil && tc.FaceUp {
					firstFaceUp = i
					break
				}
			}
			if firstFaceUp < 0 {
				continue
			}
			if exposeOnly && firstFaceUp == 0 {
				continue
			}
			for startIdx := firstFaceUp; startIdx < len(fromCards); startIdx++ {
				movingCards := fromCards[startIdx:]
				if !e.isValidEasthavenSequence(movingCards) {
					continue
				}
				if exposeOnly && startIdx != firstFaceUp {
					continue
				}
				bottomCard := movingCards[0].Card
				for toCol := range EasthavenTableauCnt {
					if toCol == fromCol {
						continue
					}
					if !e.canPlaceOnTableau(bottomCard, toCol) {
						continue
					}
					// 空列への列全体移動は無意味
					if len(e.tableau[toCol]) == 0 && startIdx == 0 {
						continue
					}
					return &EasthavenHint{
						FromCol:   fromCol,
						CardIndex: startIdx,
						ToZone:    "tableau",
						ToCol:     toCol,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（全カード表向きの場合に自動でファンデーションへ移動）
func (e *Easthaven) AutoComplete() error {
	if e.phase != EasthavenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !e.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	e.takeSnapshot()
	for {
		moved := false
		for col := range EasthavenTableauCnt {
			if len(e.tableau[col]) == 0 {
				continue
			}
			tc := e.tableau[col][len(e.tableau[col])-1]
			if tc == nil || tc.Card == nil {
				continue
			}
			card := tc.Card
			fIdx := card.GetDesign() - 1
			if !e.canPlaceOnFoundation(card, fIdx) {
				continue
			}
			e.tableau[col] = e.tableau[col][:len(e.tableau[col])-1]
			e.foundation[fIdx] = append(e.foundation[fIdx], card)
			e.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	e.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	e.checkGameClear()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（ストックも含む）
func (e *Easthaven) AllFaceUp() bool {
	if len(e.stock) > 0 {
		return false
	}
	for col := range EasthavenTableauCnt {
		for _, tc := range e.tableau[col] {
			if tc == nil || !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (e *Easthaven) GetPhase() EasthavenPhase { return e.phase }

// SetPhase フェーズ設定 (テスト用)
func (e *Easthaven) SetPhase(phase EasthavenPhase) { e.phase = phase }

// GetMoveCount 移動回数取得
func (e *Easthaven) GetMoveCount() int { return e.moveCount }

// GetStockCount ストック枚数取得
func (e *Easthaven) GetStockCount() int { return len(e.stock) }

// GetTableau タブロー取得
func (e *Easthaven) GetTableau() [EasthavenTableauCnt][]*KlondikeTableauCard { return e.tableau }

// GetFoundation ファンデーション取得
func (e *Easthaven) GetFoundation() [EasthavenFoundationCnt][]*Card { return e.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (e *Easthaven) GetGameEndFlag() bool { return e.phase != EasthavenPhasePlaying }

// IsStalemate 手詰まり状態取得
func (e *Easthaven) IsStalemate() bool { return e.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (e *Easthaven) SetIsStalemate(v bool) { e.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (e *Easthaven) SetTableau(tableau [EasthavenTableauCnt][]*KlondikeTableauCard) {
	e.tableau = tableau
}

// SetStock ストック設定 (テスト用)
func (e *Easthaven) SetStock(stock []*Card) { e.stock = stock }

// SetFoundation ファンデーション設定 (テスト用)
func (e *Easthaven) SetFoundation(foundation [EasthavenFoundationCnt][]*Card) {
	e.foundation = foundation
}

// Undo 直前の操作を取り消す
func (e *Easthaven) Undo() error {
	if e.phase != EasthavenPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(e.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := e.history[len(e.history)-1]
	e.history = e.history[:len(e.history)-1]
	e.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (e *Easthaven) CanUndo() bool {
	return len(e.history) > 0 && e.phase == EasthavenPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (e *Easthaven) UndoToEscape() int {
	return undoToEscape(e.isStalemate, e.history, func(s *easthavenSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (e *Easthaven) UndoN(n int) error {
	for i := range n {
		if err := e.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
func (e *Easthaven) canPlaceOnTableau(card *Card, col int) bool {
	colCards := e.tableau[col]
	if len(colCards) == 0 {
		// 空の列にはKのみ置ける
		return card.GetValue() == CardValueMax
	}
	top := colCards[len(colCards)-1]
	if top == nil || top.Card == nil {
		return false
	}
	topCard := top.Card
	// 交互の色で降順
	return e.isAlternateColor(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (e *Easthaven) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := e.foundation[fIdx]
	if len(pile) == 0 {
		// 空のファンデーションにはAのみ置ける
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	if topCard == nil {
		return false
	}
	// 同じスートで昇順
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// isValidEasthavenSequence 交互色降順の連続かどうか判定（全て表向き）
func (e *Easthaven) isValidEasthavenSequence(cards []*KlondikeTableauCard) bool {
	if len(cards) == 0 {
		return false
	}
	if cards[0] == nil || cards[0].Card == nil || !cards[0].FaceUp {
		return false
	}
	for i := 1; i < len(cards); i++ {
		if cards[i] == nil || cards[i].Card == nil || !cards[i].FaceUp {
			return false
		}
		prev := cards[i-1].Card
		curr := cards[i].Card
		if !e.isAlternateColor(curr, prev) {
			return false
		}
		if curr.GetValue() != prev.GetValue()-1 {
			return false
		}
	}
	return true
}

// isAlternateColor 交互の色かどうか判定
func (e *Easthaven) isAlternateColor(card1, card2 *Card) bool {
	return e.isBlack(card1) != e.isBlack(card2)
}

// isBlack 黒いカードかどうか
func (e *Easthaven) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// autoFlipTableau タブローの最上部の裏カードを自動フリップ
func (e *Easthaven) autoFlipTableau(col int) {
	cards := e.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// checkGameClear ゲームクリア判定
func (e *Easthaven) checkGameClear() {
	for i := range EasthavenFoundationCnt {
		if len(e.foundation[i]) != CardValueMax {
			return
		}
	}
	e.phase = EasthavenPhaseGameClear
}

// checkEasthavenStalemate 手詰まり判定
func (e *Easthaven) checkEasthavenStalemate() {
	if e.phase != EasthavenPhasePlaying {
		return
	}
	if e.GetHint() != nil {
		e.isStalemate = false
		return
	}
	// ストックから配れるか（空列がなく、ストックがある）
	if len(e.stock) > 0 {
		hasEmpty := false
		for i := range EasthavenTableauCnt {
			if len(e.tableau[i]) == 0 {
				hasEmpty = true
				break
			}
		}
		if !hasEmpty {
			e.isStalemate = false
			return
		}
	}
	e.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (e *Easthaven) takeSnapshot() {
	snap := &easthavenSnapshot{
		phase:       e.phase,
		moveCount:   e.moveCount,
		isStalemate: e.isStalemate,
	}
	for i := range EasthavenTableauCnt {
		snap.tableau[i] = make([]*KlondikeTableauCard, len(e.tableau[i]))
		for j, tc := range e.tableau[i] {
			snap.tableau[i][j] = &KlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.stock = make([]*Card, len(e.stock))
	copy(snap.stock, e.stock)
	for i := range EasthavenFoundationCnt {
		snap.foundation[i] = make([]*Card, len(e.foundation[i]))
		copy(snap.foundation[i], e.foundation[i])
	}
	e.history = append(e.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (e *Easthaven) restoreSnapshot(snap *easthavenSnapshot) {
	e.tableau = snap.tableau
	e.stock = snap.stock
	e.foundation = snap.foundation
	e.phase = snap.phase
	e.moveCount = snap.moveCount
	e.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (e *Easthaven) appendLog(actionType, detail string, cards []*Card) {
	e.appendLogAt(e.moveCount, 0, actionType, detail, cards)
}

// easthavenMaxSliceLen caps slice sizes during deserialisation.
const easthavenMaxSliceLen = 1000

// errEasthavenTooLarge is the single sentinel returned for any oversized slice
// during deserialisation. A shared sentinel keeps the UnmarshalJSON footprint
// small across all three Worker binaries (the domain package ships in each).
var errEasthavenTooLarge = errors.New("easthaven: input array exceeds maximum allowed size")

// easthavenJSON is the JSON wire format for Easthaven.
type easthavenJSON struct {
	TrumpCards  *TrumpCards                                 `json:"tc"`
	Tableau     [EasthavenTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock       []*Card                                     `json:"st"`
	Foundation  [EasthavenFoundationCnt][]*Card             `json:"fd"`
	Phase       EasthavenPhase                              `json:"ps"`
	MoveCount   int                                         `json:"mc"`
	ActionLog   []*ActionLogEntry                           `json:"al"`
	IsStalemate bool                                        `json:"sl"`
	History     []*easthavenSnapshot                        `json:"hi,omitempty"`
}

// easthavenSnapshotJSON is the wire format for a single undo snapshot.
// easthavenSnapshot uses unexported fields, so we project to/from this shape
// with explicit Marshal/Unmarshal methods. Field names match easthavenJSON's
// short keys to keep the KV payload compact (#1654).
type easthavenSnapshotJSON struct {
	Tableau     [EasthavenTableauCnt][]*KlondikeTableauCard `json:"tb"`
	Stock       []*Card                                     `json:"st"`
	Foundation  [EasthavenFoundationCnt][]*Card             `json:"fd"`
	Phase       EasthavenPhase                              `json:"ps"`
	MoveCount   int                                         `json:"mc"`
	IsStalemate bool                                        `json:"sl"`
}

// MarshalJSON implements json.Marshaler for easthavenSnapshot.
func (s *easthavenSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(easthavenSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for easthavenSnapshot.
func (s *easthavenSnapshot) UnmarshalJSON(data []byte) error {
	var j easthavenSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > easthavenMaxSliceLen {
		return errEasthavenTooLarge
	}
	for _, col := range j.Tableau {
		if len(col) > easthavenMaxSliceLen {
			return errEasthavenTooLarge
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > easthavenMaxSliceLen {
			return errEasthavenTooLarge
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (e *Easthaven) MarshalJSON() ([]byte, error) {
	return json.Marshal(easthavenJSON{
		TrumpCards:  e.trumpCards,
		Tableau:     e.tableau,
		Stock:       e.stock,
		Foundation:  e.foundation,
		Phase:       e.phase,
		MoveCount:   e.moveCount,
		ActionLog:   e.actionLog,
		IsStalemate: e.isStalemate,
		History:     e.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (e *Easthaven) UnmarshalJSON(data []byte) error {
	var j easthavenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > easthavenMaxSliceLen || len(j.ActionLog) > easthavenMaxSliceLen ||
		len(j.History) > easthavenMaxSliceLen {
		return errEasthavenTooLarge
	}
	for _, col := range j.Tableau {
		if len(col) > easthavenMaxSliceLen {
			return errEasthavenTooLarge
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > easthavenMaxSliceLen {
			return errEasthavenTooLarge
		}
	}

	e.trumpCards = j.TrumpCards
	if e.trumpCards == nil {
		e.trumpCards = NewTrumpCards(0)
	}
	e.tableau = j.Tableau
	e.stock = j.Stock
	if e.stock == nil {
		e.stock = make([]*Card, 0)
	}
	e.foundation = j.Foundation
	e.phase = j.Phase
	e.moveCount = j.MoveCount
	e.actionLog = j.ActionLog
	if e.actionLog == nil {
		e.actionLog = make([]*ActionLogEntry, 0)
	}
	e.history = j.History
	if e.history == nil {
		e.history = make([]*easthavenSnapshot, 0)
	}
	e.isStalemate = j.IsStalemate
	return nil
}
