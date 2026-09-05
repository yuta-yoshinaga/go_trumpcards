//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CitadelPhase Citadel game phase
type CitadelPhase int

// Citadel のフェーズ定数
const (
	// CitadelPhasePlaying プレイ中
	CitadelPhasePlaying CitadelPhase = iota
	// CitadelPhaseGameClear ゲームクリア
	CitadelPhaseGameClear
	// CitadelPhaseGameOver ゲームオーバー
	CitadelPhaseGameOver
)

// CitadelTableauCnt タブローの列数 (左右4列ずつ計8列)
const CitadelTableauCnt = 8

// CitadelMaxColumnLen 初期配置時の各列の最大カード枚数 (基礎札へ流れない場合の理論上の上限)
const CitadelMaxColumnLen = 6

// CitadelFoundationCnt ファンデーションの数
const CitadelFoundationCnt = 4

// CitadelTableauCard タブロー上のカード
type CitadelTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// CitadelHint ヒント
type CitadelHint struct {
	FromCol   int    // タブロー列インデックス
	CardIndex int    // 列内のカードインデックス
	ToZone    string // "tableau" or "foundation"
	ToCol     int    // タブロー列 or ファンデーションのインデックス
}

// CitadelConfig Citadel ゲーム設定
type CitadelConfig struct{}

// Citadel ゲームクラス
type Citadel struct {
	trumpCards *TrumpCards
	tableau    [CitadelTableauCnt][]*CitadelTableauCard
	foundation [CitadelFoundationCnt][]*Card
	phase      CitadelPhase
	moveCount  int
	actionLogBase
	history     []*citadelSnapshot
	isStalemate bool
}

// citadelSnapshot アンドゥ用スナップショット
type citadelSnapshot struct {
	tableau     [CitadelTableauCnt][]*CitadelTableauCard
	foundation  [CitadelFoundationCnt][]*Card
	phase       CitadelPhase
	moveCount   int
	isStalemate bool
}

// NewCitadel コンストラクタ
func NewCitadel(trumpCards *TrumpCards) *Citadel {
	return &Citadel{
		trumpCards: trumpCards,
	}
}

// NewDefaultCitadel returns Citadel with a single 52-card deck.
func NewDefaultCitadel() *Citadel {
	return NewCitadel(NewTrumpCardsWithDecks(1, 0))
}

// Reset ゲームリセット
//
// Initial layout:
//   - 4 Aces are pre-seeded onto the four foundations (one per suit).
//   - The remaining cards are dealt one by one: if a card can be placed
//     on a foundation pile, it is immediately moved there; otherwise,
//     it is placed on the current tableau column, advancing to the next column.
//   - If the initial deal places all cards on foundations, phase transitions to GameClear.
func (c *Citadel) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CitadelPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false

	for i := range CitadelFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range CitadelTableauCnt {
		c.tableau[i] = make([]*CitadelTableauCard, 0, CitadelMaxColumnLen)
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
	remaining := make([]*Card, 0, CardCnt-CitadelFoundationCnt)
	for {
		card := c.trumpCards.DrawCard()
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
			c.foundation[fIdx] = append(c.foundation[fIdx], card)
			continue
		}
		remaining = append(remaining, card)
	}

	// Second pass: deal remaining cards one by one. If a card can immediately
	// go to a foundation pile, move it there; otherwise place it on the current
	// tableau column and advance to the next column.
	col := 0
	for _, card := range remaining {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			c.foundation[fIdx] = append(c.foundation[fIdx], card)
		} else {
			c.tableau[col] = append(c.tableau[col], &CitadelTableauCard{Card: card, FaceUp: true})
			col = (col + 1) % CitadelTableauCnt
		}
	}

	c.checkGameClear()
	c.checkStalemate()
}

// MoveTableauToTableau タブローからタブローにカードを移動（1枚のみ）
func (c *Citadel) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if c.phase != CitadelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CitadelTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= CitadelTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := c.tableau[fromCol]
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
	if !c.canPlaceOnTableau(tc.Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	c.takeSnapshot()
	c.tableau[toCol] = append(c.tableau[toCol], tc)
	c.tableau[fromCol] = fromCards[:cardIndex]
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{tc.Card})
	c.checkStalemate()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションにカードを移動
func (c *Citadel) MoveTableauToFoundation(col int) error {
	if c.phase != CitadelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= CitadelTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := c.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := fromCards[len(fromCards)-1]
	card := tc.Card
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.tableau[col] = fromCards[:len(fromCards)-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	c.checkGameClear()
	c.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (c *Citadel) GiveUp() {
	if c.phase == CitadelPhasePlaying {
		c.phase = CitadelPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (c *Citadel) GetHint() *CitadelHint {
	if c.phase != CitadelPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := range CitadelTableauCnt {
		if len(c.tableau[col]) == 0 {
			continue
		}
		tc := c.tableau[col][len(c.tableau[col])-1]
		fIdx := c.findFoundation(tc.Card)
		if fIdx >= 0 {
			return &CitadelHint{
				FromCol:   col,
				CardIndex: len(c.tableau[col]) - 1,
				ToZone:    "foundation",
				ToCol:     fIdx,
			}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := range CitadelTableauCnt {
		fromCards := c.tableau[fromCol]
		if len(fromCards) == 0 {
			continue
		}
		card := fromCards[len(fromCards)-1].Card
		for toCol := range CitadelTableauCnt {
			if toCol == fromCol {
				continue
			}
			if c.canPlaceOnTableau(card, toCol) {
				return &CitadelHint{
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
func (c *Citadel) AutoComplete() error {
	if c.phase != CitadelPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	c.takeSnapshot()
	for {
		moved := false
		for col := range CitadelTableauCnt {
			if len(c.tableau[col]) == 0 {
				continue
			}
			tc := c.tableau[col][len(c.tableau[col])-1]
			card := tc.Card
			fIdx := c.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			c.tableau[col] = c.tableau[col][:len(c.tableau[col])-1]
			c.foundation[fIdx] = append(c.foundation[fIdx], card)
			c.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	c.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	c.checkGameClear()
	c.checkStalemate()
	return nil
}

// AllFaceUp 全カードが表向きかどうか（Citadel では常にtrue）
func (c *Citadel) AllFaceUp() bool {
	return true
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (c *Citadel) GetPhase() CitadelPhase { return c.phase }

// SetPhase フェーズ設定 (テスト用)
func (c *Citadel) SetPhase(phase CitadelPhase) { c.phase = phase }

// GetMoveCount 移動回数取得
func (c *Citadel) GetMoveCount() int { return c.moveCount }

// GetTableau タブロー取得
func (c *Citadel) GetTableau() [CitadelTableauCnt][]*CitadelTableauCard {
	return c.tableau
}

// GetFoundation ファンデーション取得
func (c *Citadel) GetFoundation() [CitadelFoundationCnt][]*Card {
	return c.foundation
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (c *Citadel) GetGameEndFlag() bool { return c.phase != CitadelPhasePlaying }

// IsStalemate 手詰まり状態取得
func (c *Citadel) IsStalemate() bool { return c.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (c *Citadel) SetIsStalemate(v bool) { c.isStalemate = v }

// SetTableau タブロー設定 (テスト用)
func (c *Citadel) SetTableau(tableau [CitadelTableauCnt][]*CitadelTableauCard) {
	c.tableau = tableau
}

// SetFoundation ファンデーション設定 (テスト用)
func (c *Citadel) SetFoundation(foundation [CitadelFoundationCnt][]*Card) {
	c.foundation = foundation
}

// Undo 直前の操作を取り消す
func (c *Citadel) Undo() error {
	if c.phase != CitadelPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(c.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (c *Citadel) CanUndo() bool {
	return len(c.history) > 0 && c.phase == CitadelPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ0、脱出不可なら-1。
func (c *Citadel) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *citadelSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (c *Citadel) UndoN(n int) error {
	return undoN(c, n)
}

// --- Private helpers ---

// canPlaceOnTableau タブローにカードを置けるか判定
//
// Citadel: 「スート無関係の降順」でタブロー間の移動可能。
// 空列には「任意のカード」を置ける。
func (c *Citadel) canPlaceOnTableau(card *Card, col int) bool {
	colCards := c.tableau[col]
	if len(colCards) == 0 {
		return true
	}
	topCard := colCards[len(colCards)-1].Card
	return card.GetValue() == topCard.GetValue()-1
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (c *Citadel) canPlaceOnFoundation(card *Card, fIdx int) bool {
	return canPlaceOnFoundationPile(c.foundation[fIdx], card)
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (c *Citadel) findFoundation(card *Card) int {
	for i := range CitadelFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (c *Citadel) checkGameClear() {
	for i := range CitadelFoundationCnt {
		if len(c.foundation[i]) != CardValueMax {
			return
		}
	}
	c.phase = CitadelPhaseGameClear
}

// checkStalemate 手詰まり判定
func (c *Citadel) checkStalemate() {
	if c.phase != CitadelPhasePlaying {
		return
	}
	if c.GetHint() != nil {
		c.isStalemate = false
		return
	}
	c.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (c *Citadel) takeSnapshot() {
	snap := &citadelSnapshot{
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range CitadelTableauCnt {
		snap.tableau[i] = make([]*CitadelTableauCard, len(c.tableau[i]))
		for j, tc := range c.tableau[i] {
			snap.tableau[i][j] = &CitadelTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	for i := range CitadelFoundationCnt {
		snap.foundation[i] = make([]*Card, len(c.foundation[i]))
		copy(snap.foundation[i], c.foundation[i])
	}
	c.history = appendSnapshot(c.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (c *Citadel) restoreSnapshot(snap *citadelSnapshot) {
	c.tableau = snap.tableau
	c.foundation = snap.foundation
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (c *Citadel) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// citadelJSON is the JSON wire format for Citadel.
type citadelJSON struct {
	TrumpCards  *TrumpCards                              `json:"tc"`
	Tableau     [CitadelTableauCnt][]*CitadelTableauCard `json:"tb"`
	Foundation  [CitadelFoundationCnt][]*Card            `json:"fd"`
	Phase       CitadelPhase                             `json:"ps"`
	MoveCount   int                                      `json:"mc"`
	ActionLog   []*ActionLogEntry                        `json:"al"`
	IsStalemate bool                                     `json:"sl"`
	History     []*citadelSnapshot                       `json:"hi,omitempty"`
}

// citadelSnapshotJSON mirrors beleagueredCastleSnapshotJSON.
type citadelSnapshotJSON struct {
	Tableau     [CitadelTableauCnt][]*CitadelTableauCard `json:"tb"`
	Foundation  [CitadelFoundationCnt][]*Card            `json:"fd"`
	Phase       CitadelPhase                             `json:"ps"`
	MoveCount   int                                      `json:"mc"`
	IsStalemate bool                                     `json:"sl"`
}

// MarshalJSON implements json.Marshaler for citadelSnapshot.
func (s *citadelSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(citadelSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// citadelMaxSliceLen caps slice sizes during deserialisation.
const citadelMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler for citadelSnapshot.
func (s *citadelSnapshot) UnmarshalJSON(data []byte) error {
	var j citadelSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, col := range j.Tableau {
		if len(col) > citadelMaxSliceLen {
			return fmt.Errorf("citadel: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > citadelMaxSliceLen {
			return fmt.Errorf("citadel: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	for i := range CitadelTableauCnt {
		if s.tableau[i] == nil {
			s.tableau[i] = make([]*CitadelTableauCard, 0)
		}
	}
	s.foundation = j.Foundation
	for i := range CitadelFoundationCnt {
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
func (c *Citadel) MarshalJSON() ([]byte, error) {
	return json.Marshal(citadelJSON{
		TrumpCards:  c.trumpCards,
		Tableau:     c.tableau,
		Foundation:  c.foundation,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (c *Citadel) UnmarshalJSON(data []byte) error {
	var j citadelJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > citadelMaxSliceLen ||
		len(j.History) > citadelMaxSliceLen {
		return fmt.Errorf("citadel: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > citadelMaxSliceLen {
			return fmt.Errorf("citadel: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > citadelMaxSliceLen {
			return fmt.Errorf("citadel: foundation pile exceeds maximum allowed size")
		}
	}

	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCardsWithDecks(1, 0)
	}
	c.tableau = j.Tableau
	for i := range CitadelTableauCnt {
		if c.tableau[i] == nil {
			c.tableau[i] = make([]*CitadelTableauCard, 0)
		}
	}
	c.foundation = j.Foundation
	for i := range CitadelFoundationCnt {
		if c.foundation[i] == nil {
			c.foundation[i] = make([]*Card, 0)
		}
	}
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	c.history = j.History
	if c.history == nil {
		c.history = make([]*citadelSnapshot, 0)
	}
	c.isStalemate = j.IsStalemate
	return nil
}
