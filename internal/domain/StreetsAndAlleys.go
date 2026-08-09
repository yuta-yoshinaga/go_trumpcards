//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// StreetsAndAlleysPhase Streets and Alleys game phase
type StreetsAndAlleysPhase int

// StreetsAndAlleys のフェーズ定数
const (
	// StreetsAndAlleysPhasePlaying プレイ中
	StreetsAndAlleysPhasePlaying StreetsAndAlleysPhase = iota
	// StreetsAndAlleysPhaseGameClear ゲームクリア
	StreetsAndAlleysPhaseGameClear
	// StreetsAndAlleysPhaseGameOver ゲームオーバー
	StreetsAndAlleysPhaseGameOver
)

// StreetsAndAlleysTableauCnt タブローの列数 (左右4列ずつ計8列)
const StreetsAndAlleysTableauCnt = 8

// StreetsAndAlleysColumnLen 初期配置時の各列の最大カード枚数。
// Streets and Alleys deals all 52 cards into 8 columns: columns 0-3 hold 7
// cards each and columns 4-7 hold 6 cards each (28 + 24 = 52). The constant is
// the longer of the two so callers sizing a column buffer never under-allocate.
const StreetsAndAlleysColumnLen = 7

// StreetsAndAlleysLongColumnCnt is the number of tableau columns that receive
// the longer 7-card deal (the first 4 columns). The remaining columns get 6.
const StreetsAndAlleysLongColumnCnt = 4

// StreetsAndAlleysLongColumnLen / StreetsAndAlleysShortColumnLen are the per-bucket
// deal sizes for the long (columns 0-3) and short (columns 4-7) columns.
const (
	StreetsAndAlleysLongColumnLen  = 7
	StreetsAndAlleysShortColumnLen = 6
)

// StreetsAndAlleysFoundationCnt ファンデーションの数
const StreetsAndAlleysFoundationCnt = 4

// StreetsAndAlleysTableauCard タブロー上のカード
type StreetsAndAlleysTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// StreetsAndAlleysHint ヒント
type StreetsAndAlleysHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// StreetsAndAlleysConfig Streets and Alleys ゲーム設定
type StreetsAndAlleysConfig struct{}

// StreetsAndAlleys ゲームクラス
type StreetsAndAlleys struct {
	trumpCards *TrumpCards
	tableau    [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard
	foundation [StreetsAndAlleysFoundationCnt][]*Card
	phase      StreetsAndAlleysPhase
	moveCount  int
	actionLogBase
	history     []*streetsAndAlleysSnapshot
	isStalemate bool
}

// streetsAndAlleysSnapshot アンドゥ用スナップショット
type streetsAndAlleysSnapshot struct {
	tableau     [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard
	foundation  [StreetsAndAlleysFoundationCnt][]*Card
	phase       StreetsAndAlleysPhase
	moveCount   int
	isStalemate bool
}

// NewStreetsAndAlleys コンストラクタ
func NewStreetsAndAlleys(trumpCards *TrumpCards) *StreetsAndAlleys {
	return &StreetsAndAlleys{
		trumpCards: trumpCards,
	}
}

// NewDefaultStreetsAndAlleys returns StreetsAndAlleys with a single 52-card deck.
func NewDefaultStreetsAndAlleys() *StreetsAndAlleys {
	return NewStreetsAndAlleys(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout (Streets and Alleys, unlike Beleaguered Castle):
//   - The four foundations start EMPTY; the player must move every Ace out of
//     the tableau themselves.
//   - All 52 cards are dealt face-up across 8 tableau columns. Columns 0-3 get
//     7 cards each and columns 4-7 get 6 cards each (28 + 24 = 52).
func (sa *StreetsAndAlleys) Reset() {
	sa.trumpCards.Shuffle()
	sa.phase = StreetsAndAlleysPhasePlaying
	sa.moveCount = 0
	sa.actionLog = nil
	sa.history = nil
	sa.isStalemate = false

	for i := range StreetsAndAlleysFoundationCnt {
		sa.foundation[i] = nil
	}
	for i := range StreetsAndAlleysTableauCnt {
		sa.tableau[i] = make([]*StreetsAndAlleysTableauCard, 0, StreetsAndAlleysColumnLen)
	}

	// Deal every card (Aces included) onto the tableau. The first
	// StreetsAndAlleysLongColumnCnt columns receive one extra card so the deal
	// splits 7/7/7/7/6/6/6/6 = 52 across the 8 columns.
	for col := range StreetsAndAlleysTableauCnt {
		colLen := StreetsAndAlleysShortColumnLen
		if col < StreetsAndAlleysLongColumnCnt {
			colLen = StreetsAndAlleysLongColumnLen
		}
		for range colLen {
			card := sa.trumpCards.DrawCard()
			if card == nil {
				break
			}
			sa.tableau[col] = append(sa.tableau[col], &StreetsAndAlleysTableauCard{Card: card, FaceUp: true})
		}
	}

	sa.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (sa *StreetsAndAlleys) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= StreetsAndAlleysTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= StreetsAndAlleysTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := sa.tableau[fromCol]
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
	if !sa.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	sa.takeSnapshot()
	sa.tableau[toCol] = append(sa.tableau[toCol], tc)
	sa.tableau[fromCol] = fromCards[:cardIndex]
	sa.moveCount++
	sa.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	sa.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (sa *StreetsAndAlleys) MoveTableauToFoundation(col int) error {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= StreetsAndAlleysTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := sa.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := sa.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	sa.takeSnapshot()
	sa.tableau[col] = fromCards[:len(fromCards)-1]
	sa.foundation[fIdx] = append(sa.foundation[fIdx], card)
	sa.moveCount++
	sa.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	sa.checkGameClear()
	sa.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (sa *StreetsAndAlleys) GiveUp() {
	if sa.phase == StreetsAndAlleysPhasePlaying {
		sa.phase = StreetsAndAlleysPhaseGameOver
		sa.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (sa *StreetsAndAlleys) GetHint() *StreetsAndAlleysHint {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range StreetsAndAlleysTableauCnt {
		if len(sa.tableau[col]) == 0 {
			continue
		}
		tc := sa.tableau[col][len(sa.tableau[col])-1]
		fIdx := sa.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &StreetsAndAlleysHint{
				FromCol:   col,
				CardIndex: len(sa.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range StreetsAndAlleysTableauCnt {
		fromCards := sa.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range StreetsAndAlleysTableauCnt {
			if toCol == fromCol {
				continue
			}
			if sa.canPlaceOnTableau(card, toCol) {
				return &StreetsAndAlleysHint{
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
func (sa *StreetsAndAlleys) AutoComplete() error {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	sa.takeSnapshot()
	for {
		moved := false
		for col := range StreetsAndAlleysTableauCnt {
			if len(sa.tableau[col]) == 0 {
				continue
			}
			tc := sa.tableau[col][len(sa.tableau[col])-1]
			card := tc.Card
			fIdx := sa.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			sa.tableau[col] = sa.tableau[col][:len(sa.tableau[col])-1]
			sa.foundation[fIdx] = append(sa.foundation[fIdx], card)
			sa.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	sa.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	sa.checkGameClear()
	sa.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Streets and Alleys では常にtrue）
func (sa *StreetsAndAlleys) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (sa *StreetsAndAlleys) GetPhase() StreetsAndAlleysPhase { return sa.phase }

// SetPhase フェーズ設定 (テスト用)
func (sa *StreetsAndAlleys) SetPhase(phase StreetsAndAlleysPhase) { sa.phase = phase }

// GetMoveCount 移動回数取得
func (sa *StreetsAndAlleys) GetMoveCount() int { return sa.moveCount }

// GetTableau タブロー取得
func (sa *StreetsAndAlleys) GetTableau() [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard {
	return sa.tableau
}

// GetFoundation ファンデーション取得
func (sa *StreetsAndAlleys) GetFoundation() [StreetsAndAlleysFoundationCnt][]*Card {
	return sa.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (sa *StreetsAndAlleys) GetGameEndFlag() bool { return sa.phase != StreetsAndAlleysPhasePlaying }

// IsStalemate 手詰まり状態取得
func (sa *StreetsAndAlleys) IsStalemate() bool { return sa.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (sa *StreetsAndAlleys) SetIsStalemate(v bool) { sa.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (sa *StreetsAndAlleys) SetTableau(tableau [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard) {
	sa.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (sa *StreetsAndAlleys) SetFoundation(foundation [StreetsAndAlleysFoundationCnt][]*Card) {
	sa.foundation = foundation
}

// Undo 直前の操作を取り消す
func (sa *StreetsAndAlleys) Undo() error {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(sa.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := sa.history[len(sa.history)-1]
	sa.history = sa.history[:len(sa.history)-1]
	sa.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (sa *StreetsAndAlleys) CanUndo() bool {
	return len(sa.history) > 0 && sa.phase == StreetsAndAlleysPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (sa *StreetsAndAlleys) UndoToEscape() int {
	return undoToEscape(sa.isStalemate, sa.history, func(s *streetsAndAlleysSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (sa *StreetsAndAlleys) UndoN(n int) error {
	return undoN(sa, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
//
// Streets and Alleys: 「スート無関係の降順」でタブロー間の移動可能。
// 空列には「任意のカード」を置ける (Baker's Dozen との違い)。
func (sa *StreetsAndAlleys) canPlaceOnTableau(card *Card, col int) bool {
	colCards := sa.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (sa *StreetsAndAlleys) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := sa.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	topCard := pile[len(pile)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (sa *StreetsAndAlleys) findFoundation(card *Card) int {
	for i := range StreetsAndAlleysFoundationCnt {
		if sa.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (sa *StreetsAndAlleys) checkGameClear() {
	for i := range StreetsAndAlleysFoundationCnt {
		if len(sa.foundation[i]) != CardValueMax {
			return
		}
	}
	sa.phase = StreetsAndAlleysPhaseGameClear
}

// checkStalemate 手詰まり判定
func (sa *StreetsAndAlleys) checkStalemate() {
	if sa.phase != StreetsAndAlleysPhasePlaying {
		return
	}
	if sa.GetHint() != nil {
		sa.isStalemate = false
		return
	}
	sa.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (sa *StreetsAndAlleys) takeSnapshot() {
	snap := &streetsAndAlleysSnapshot{
		phase:       sa.phase,
		moveCount:   sa.moveCount,
		isStalemate: sa.isStalemate,
	}
	for i := range StreetsAndAlleysTableauCnt {
		snap.tableau[i] = make([]*StreetsAndAlleysTableauCard, len(sa.tableau[i]))
		for j, tc := range sa.tableau[i] {
			snap.tableau[i][j] = &StreetsAndAlleysTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range StreetsAndAlleysFoundationCnt {
		snap.foundation[i] = make([]*Card, len(sa.foundation[i]))
		copy(snap.foundation[i], sa.foundation[i])
	}
	sa.history = append(sa.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (sa *StreetsAndAlleys) restoreSnapshot(snap *streetsAndAlleysSnapshot) {
	sa.tableau = snap.tableau
	sa.foundation = snap.foundation
	sa.phase = snap.phase
	sa.moveCount = snap.moveCount
	sa.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (sa *StreetsAndAlleys) appendLog(actionType, detail string, cards []*Card) {
	sa.appendLogAt(sa.moveCount, 0, actionType, detail, cards)
}

// streetsAndAlleysJSON is the JSON wire format for StreetsAndAlleys.
type streetsAndAlleysJSON struct {
	TrumpCards  *TrumpCards                                                `json:"tc"`
	Tableau     [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard `json:"tb"`
	Foundation  [StreetsAndAlleysFoundationCnt][]*Card                     `json:"fd"`
	Phase       StreetsAndAlleysPhase                                      `json:"ps"`
	MoveCount   int                                                        `json:"mc"`
	ActionLog   []*ActionLogEntry                                          `json:"al"`
	IsStalemate bool                                                       `json:"sl"`
	History     []*streetsAndAlleysSnapshot                                `json:"hi,omitempty"`
}

// streetsAndAlleysSnapshotJSON mirrors bakersDozenSnapshotJSON; see the
// matching wire-format note on bakersDozenSnapshot for the rationale behind
// the short keys (KV payload size, issue #1654).
type streetsAndAlleysSnapshotJSON struct {
	Tableau     [StreetsAndAlleysTableauCnt][]*StreetsAndAlleysTableauCard `json:"tb"`
	Foundation  [StreetsAndAlleysFoundationCnt][]*Card                     `json:"fd"`
	Phase       StreetsAndAlleysPhase                                      `json:"ps"`
	MoveCount   int                                                        `json:"mc"`
	IsStalemate bool                                                       `json:"sl"`
}

// MarshalJSON implements json.Marshaler for streetsAndAlleysSnapshot.
func (s *streetsAndAlleysSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(streetsAndAlleysSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for streetsAndAlleysSnapshot.
func (s *streetsAndAlleysSnapshot) UnmarshalJSON(data []byte) error {
	var j streetsAndAlleysSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > streetsAndAlleysMaxSliceLen {
			return fmt.Errorf("streetsandalleys: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > streetsAndAlleysMaxSliceLen {
			return fmt.Errorf("streetsandalleys: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range StreetsAndAlleysTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*StreetsAndAlleysTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range StreetsAndAlleysFoundationCnt {
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
func (sa *StreetsAndAlleys) MarshalJSON() ([]byte, error) {
	return json.Marshal(streetsAndAlleysJSON{
		TrumpCards:  sa.trumpCards,
		Tableau:     sa.tableau,
		Foundation:  sa.foundation,
		Phase:       sa.phase,
		MoveCount:   sa.moveCount,
		ActionLog:   sa.actionLog,
		IsStalemate: sa.isStalemate,
		History:     sa.history,
	})
}

// streetsAndAlleysMaxSliceLen caps slice sizes during deserialisation.
const streetsAndAlleysMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (sa *StreetsAndAlleys) UnmarshalJSON(data []byte) error {
	var j streetsAndAlleysJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > streetsAndAlleysMaxSliceLen ||
		len(j.History) > streetsAndAlleysMaxSliceLen {
		return fmt.Errorf("streetsandalleys: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > streetsAndAlleysMaxSliceLen {
			return fmt.Errorf("streetsandalleys: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > streetsAndAlleysMaxSliceLen {
			return fmt.Errorf("streetsandalleys: foundation pile exceeds maximum allowed size")
		}
	}

	sa.trumpCards = j.TrumpCards
	if sa.trumpCards == nil {
		sa.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	sa.tableau = j.Tableau
	for i := range StreetsAndAlleysTableauCnt {
		if sa.tableau[i] == nil {
			sa.tableau[i] = make([]*StreetsAndAlleysTableauCard, 0)
		}
	}
	sa.foundation = j.Foundation
	for i := range StreetsAndAlleysFoundationCnt {
		if sa.foundation[i] == nil {
			sa.foundation[i] = make([]*Card, 0)
		}
	}
	sa.phase = j.Phase
	sa.moveCount = j.MoveCount
	sa.actionLog = j.ActionLog
	if sa.actionLog == nil {
		sa.actionLog = make([]*ActionLogEntry, 0)
	}
	sa.history = j.History
	if sa.history == nil {
		sa.history = make([]*streetsAndAlleysSnapshot, 0)
	}
	sa.isStalemate = j.IsStalemate
	return nil
}
