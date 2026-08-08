//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CanfieldPhase キャンフィールドゲームフェーズ
type CanfieldPhase int

// Canfieldのフェーズ定数
const (
	// CanfieldPhasePlaying プレイ中
	CanfieldPhasePlaying CanfieldPhase = iota
	// CanfieldPhaseGameClear ゲームクリア
	CanfieldPhaseGameClear
	// CanfieldPhaseGameOver ゲームオーバー
	CanfieldPhaseGameOver
)

// CanfieldTableauCnt タブローの列数
const CanfieldTableauCnt = 4

// CanfieldFoundationCnt ファンデーションの数
const CanfieldFoundationCnt = 4

// CanfieldReserveSize リザーブ枚数
const CanfieldReserveSize = 13

// CanfieldDrawCount ウェイストへの1回のドロー枚数
const CanfieldDrawCount = 3

// CanfieldTableauCard タブロー上のカード
type CanfieldTableauCard struct {
	Card *Card `json:"c"`
}

// CanfieldHint ヒント
type CanfieldHint struct {
	FromZone  string // "waste" / "tableau" / "reserve"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" / "foundation"
	ToCol     int
}

// Canfield キャンフィールドゲームクラス
type Canfield struct {
	trumpCards *TrumpCards
	tableau    [CanfieldTableauCnt][]*CanfieldTableauCard
	reserve    []*Card
	stock      []*Card
	waste      []*Card
	foundation [CanfieldFoundationCnt][]*Card
	baseRank   int
	phase      CanfieldPhase
	moveCount  int
	actionLogBase
	history []*canfieldSnapshot
}

// canfieldSnapshot アンドゥ用スナップショット
type canfieldSnapshot struct {
	tableau    [CanfieldTableauCnt][]*CanfieldTableauCard
	reserve    []*Card
	stock      []*Card
	waste      []*Card
	foundation [CanfieldFoundationCnt][]*Card
	phase      CanfieldPhase
	moveCount  int
}

// NewCanfield コンストラクタ
func NewCanfield(trumpCards *TrumpCards) *Canfield {
	return &Canfield{trumpCards: trumpCards}
}

// NewDefaultCanfield returns Canfield with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCanfield() *Canfield {
	return NewCanfield(NewTrumpCards(0))
}

// Reset ゲームリセット
func (c *Canfield) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CanfieldPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil

	// リザーブに13枚
	c.reserve = make([]*Card, 0, CanfieldReserveSize)
	for i := 0; i < CanfieldReserveSize; i++ {
		c.reserve = append(c.reserve, c.trumpCards.DrawCard())
	}

	// タブローに4列、各1枚
	for i := 0; i < CanfieldTableauCnt; i++ {
		c.tableau[i] = []*CanfieldTableauCard{{Card: c.trumpCards.DrawCard()}}
	}

	// ファンデーション初期化
	for i := 0; i < CanfieldFoundationCnt; i++ {
		c.foundation[i] = nil
	}

	// ベースランク決定: 次の1枚をファンデーションの1つに置く
	base := c.trumpCards.DrawCard()
	c.baseRank = base.GetValue()
	fIdx := base.GetDesign() - 1
	if fIdx >= 0 && fIdx < CanfieldFoundationCnt {
		c.foundation[fIdx] = []*Card{base}
	}

	// 残りをストックへ
	c.stock = nil
	c.waste = nil
	for c.trumpCards.GetRemainingCount() > 0 {
		c.stock = append(c.stock, c.trumpCards.DrawCard())
	}
}

// Draw ストックからウェイストにカードを引く
func (c *Canfield) Draw() error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(c.stock) == 0 {
		if len(c.waste) == 0 {
			return errors.New("no cards in stock or waste")
		}
		c.takeSnapshot()
		for i := len(c.waste) - 1; i >= 0; i-- {
			c.stock = append(c.stock, c.waste[i])
		}
		c.waste = nil
		c.appendLog("recycle", "ウェイストをストックに戻しました", nil)
		return nil
	}
	c.takeSnapshot()
	count := CanfieldDrawCount
	if count > len(c.stock) {
		count = len(c.stock)
	}
	drawn := make([]*Card, 0, count)
	for i := 0; i < count; i++ {
		card := c.stock[len(c.stock)-1]
		c.stock = c.stock[:len(c.stock)-1]
		c.waste = append(c.waste, card)
		drawn = append(drawn, card)
	}
	c.moveCount++
	c.appendLog("draw", "ストックからカードを引きました", drawn)
	return nil
}

// MoveWasteToTableau ウェイストからタブローに移動
func (c *Canfield) MoveWasteToTableau(col int) error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= CanfieldTableauCnt {
		return errors.New("invalid column")
	}
	if len(c.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := c.waste[len(c.waste)-1]
	if !c.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	c.takeSnapshot()
	c.waste = c.waste[:len(c.waste)-1]
	c.tableau[col] = append(c.tableau[col], &CanfieldTableauCard{Card: card})
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	c.autoFillFromReserve()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションに移動
func (c *Canfield) MoveWasteToFoundation() error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(c.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := c.waste[len(c.waste)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= CanfieldFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.waste = c.waste[:len(c.waste)-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.moveCount++
	c.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	c.checkGameClear()
	return nil
}

// MoveTableauToTableau タブロー間で移動
func (c *Canfield) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= CanfieldTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= CanfieldTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := c.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	moving := fromCards[cardIndex:]
	bottom := moving[0].Card
	if !c.canPlaceOnTableau(bottom, toCol) {
		return errors.New("cannot place card on tableau")
	}
	c.takeSnapshot()
	moved := make([]*Card, len(moving))
	for i, mc := range moving {
		c.tableau[toCol] = append(c.tableau[toCol], mc)
		moved[i] = mc.Card
	}
	c.tableau[fromCol] = fromCards[:cardIndex]
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), moved)
	c.autoFillFromReserve()
	return nil
}

// MoveTableauToFoundation タブローからファンデーションに移動
func (c *Canfield) MoveTableauToFoundation(col int) error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= CanfieldTableauCnt {
		return errors.New("invalid column")
	}
	from := c.tableau[col]
	if len(from) == 0 {
		return errors.New("tableau column is empty")
	}
	card := from[len(from)-1].Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= CanfieldFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.tableau[col] = from[:len(from)-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	c.checkGameClear()
	c.autoFillFromReserve()
	return nil
}

// MoveReserveToTableau リザーブからタブローに移動
func (c *Canfield) MoveReserveToTableau(col int) error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= CanfieldTableauCnt {
		return errors.New("invalid column")
	}
	if len(c.reserve) == 0 {
		return errors.New("reserve is empty")
	}
	card := c.reserve[len(c.reserve)-1]
	if !c.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	c.takeSnapshot()
	c.reserve = c.reserve[:len(c.reserve)-1]
	c.tableau[col] = append(c.tableau[col], &CanfieldTableauCard{Card: card})
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("リザーブ→タブロー列%d", col), []*Card{card})
	return nil
}

// MoveReserveToFoundation リザーブからファンデーションに移動
func (c *Canfield) MoveReserveToFoundation() error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(c.reserve) == 0 {
		return errors.New("reserve is empty")
	}
	card := c.reserve[len(c.reserve)-1]
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= CanfieldFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.reserve = c.reserve[:len(c.reserve)-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.moveCount++
	c.appendLog("move", "リザーブ→ファンデーション", []*Card{card})
	c.checkGameClear()
	return nil
}

// GiveUp ギブアップ
func (c *Canfield) GiveUp() {
	if c.phase == CanfieldPhasePlaying {
		c.phase = CanfieldPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (c *Canfield) GetHint() *CanfieldHint {
	if c.phase != CanfieldPhasePlaying {
		return nil
	}
	// 優先度1: リザーブからファンデーションへ
	if len(c.reserve) > 0 {
		card := c.reserve[len(c.reserve)-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < CanfieldFoundationCnt && c.canPlaceOnFoundation(card, fIdx) {
			return &CanfieldHint{FromZone: "reserve", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度2: タブローからファンデーションへ
	for col := 0; col < CanfieldTableauCnt; col++ {
		if len(c.tableau[col]) == 0 {
			continue
		}
		card := c.tableau[col][len(c.tableau[col])-1].Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < CanfieldFoundationCnt && c.canPlaceOnFoundation(card, fIdx) {
			return &CanfieldHint{FromZone: "tableau", FromCol: col, CardIndex: len(c.tableau[col]) - 1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度3: ウェイストからファンデーションへ
	if len(c.waste) > 0 {
		card := c.waste[len(c.waste)-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < CanfieldFoundationCnt && c.canPlaceOnFoundation(card, fIdx) {
			return &CanfieldHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度4: リザーブからタブローへ
	if len(c.reserve) > 0 {
		card := c.reserve[len(c.reserve)-1]
		for toCol := 0; toCol < CanfieldTableauCnt; toCol++ {
			if c.canPlaceOnTableau(card, toCol) {
				return &CanfieldHint{FromZone: "reserve", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	// 優先度5: ウェイストからタブローへ
	if len(c.waste) > 0 {
		card := c.waste[len(c.waste)-1]
		for toCol := 0; toCol < CanfieldTableauCnt; toCol++ {
			if c.canPlaceOnTableau(card, toCol) {
				return &CanfieldHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート
func (c *Canfield) AutoComplete() error {
	if c.phase != CanfieldPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(c.reserve) != 0 || len(c.stock) != 0 || len(c.waste) != 0 {
		return errors.New("reserve, stock, and waste must be empty")
	}
	c.takeSnapshot()
	for {
		moved := false
		for col := 0; col < CanfieldTableauCnt; col++ {
			if len(c.tableau[col]) == 0 {
				continue
			}
			card := c.tableau[col][len(c.tableau[col])-1].Card
			fIdx := card.GetDesign() - 1
			if fIdx < 0 || fIdx >= CanfieldFoundationCnt {
				continue
			}
			if !c.canPlaceOnFoundation(card, fIdx) {
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
	return nil
}

// --- Getters / Setters ---

// GetPhase フェーズ取得
func (c *Canfield) GetPhase() CanfieldPhase { return c.phase }

// SetPhase フェーズ設定
func (c *Canfield) SetPhase(p CanfieldPhase) { c.phase = p }

// GetMoveCount 移動回数取得
func (c *Canfield) GetMoveCount() int { return c.moveCount }

// GetStockCount ストック残枚数
func (c *Canfield) GetStockCount() int { return len(c.stock) }

// GetWaste ウェイスト取得
func (c *Canfield) GetWaste() []*Card { return c.waste }

// GetReserve リザーブ取得
func (c *Canfield) GetReserve() []*Card { return c.reserve }

// GetTableau タブロー取得
func (c *Canfield) GetTableau() [CanfieldTableauCnt][]*CanfieldTableauCard { return c.tableau }

// GetFoundation ファンデーション取得
func (c *Canfield) GetFoundation() [CanfieldFoundationCnt][]*Card { return c.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (c *Canfield) GetGameEndFlag() bool { return c.phase != CanfieldPhasePlaying }

// GetBaseRank ベースランク取得
func (c *Canfield) GetBaseRank() int { return c.baseRank }

// SetBaseRank ベースランク設定 (テスト用)
func (c *Canfield) SetBaseRank(r int) { c.baseRank = r }

// SetTableau タブロー設定 (テスト用)
func (c *Canfield) SetTableau(t [CanfieldTableauCnt][]*CanfieldTableauCard) { c.tableau = t }

// SetStock ストック設定 (テスト用)
func (c *Canfield) SetStock(s []*Card) { c.stock = s }

// SetWaste ウェイスト設定 (テスト用)
func (c *Canfield) SetWaste(w []*Card) { c.waste = w }

// SetReserve リザーブ設定 (テスト用)
func (c *Canfield) SetReserve(r []*Card) { c.reserve = r }

// SetFoundation ファンデーション設定 (テスト用)
func (c *Canfield) SetFoundation(f [CanfieldFoundationCnt][]*Card) { c.foundation = f }

// Undo 直前の操作を取り消す
func (c *Canfield) Undo() error {
	if c.phase != CanfieldPhasePlaying {
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

// CanUndo アンドゥ可能か
func (c *Canfield) CanUndo() bool {
	return len(c.history) > 0 && c.phase == CanfieldPhasePlaying
}

// UndoN n回連続アンドゥ
func (c *Canfield) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := c.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

func (c *Canfield) canPlaceOnTableau(card *Card, col int) bool {
	colCards := c.tableau[col]
	if len(colCards) == 0 {
		// リザーブが残っている間は空列に自分で置けない (自動補充)
		return len(c.reserve) == 0
	}
	top := colCards[len(colCards)-1].Card
	return c.isAlternateColor(card, top) && card.GetValue() == c.prevRank(top.GetValue())
}

func (c *Canfield) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := c.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == c.baseRank
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == c.nextRank(top.GetValue())
}

func (c *Canfield) nextRank(r int) int {
	return (r % 13) + 1
}

func (c *Canfield) prevRank(r int) int {
	return ((r + 11) % 13) + 1
}

func (c *Canfield) isAlternateColor(a, b *Card) bool {
	return c.isBlack(a) != c.isBlack(b)
}

func (c *Canfield) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

func (c *Canfield) autoFillFromReserve() {
	if len(c.reserve) == 0 {
		return
	}
	for col := 0; col < CanfieldTableauCnt; col++ {
		if len(c.tableau[col]) == 0 && len(c.reserve) > 0 {
			card := c.reserve[len(c.reserve)-1]
			c.reserve = c.reserve[:len(c.reserve)-1]
			c.tableau[col] = []*CanfieldTableauCard{{Card: card}}
		}
	}
}

func (c *Canfield) checkGameClear() {
	for i := 0; i < CanfieldFoundationCnt; i++ {
		if len(c.foundation[i]) != CardValueMax {
			return
		}
	}
	c.phase = CanfieldPhaseGameClear
}

func (c *Canfield) takeSnapshot() {
	snap := &canfieldSnapshot{
		phase:     c.phase,
		moveCount: c.moveCount,
	}
	for i := 0; i < CanfieldTableauCnt; i++ {
		snap.tableau[i] = make([]*CanfieldTableauCard, len(c.tableau[i]))
		for j, tc := range c.tableau[i] {
			snap.tableau[i][j] = &CanfieldTableauCard{Card: tc.Card}
		}
	}
	snap.reserve = make([]*Card, len(c.reserve))
	copy(snap.reserve, c.reserve)
	snap.stock = make([]*Card, len(c.stock))
	copy(snap.stock, c.stock)
	snap.waste = make([]*Card, len(c.waste))
	copy(snap.waste, c.waste)
	for i := 0; i < CanfieldFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(c.foundation[i]))
		copy(snap.foundation[i], c.foundation[i])
	}
	c.history = append(c.history, snap)
}

func (c *Canfield) restoreSnapshot(snap *canfieldSnapshot) {
	c.tableau = snap.tableau
	c.reserve = snap.reserve
	c.stock = snap.stock
	c.waste = snap.waste
	c.foundation = snap.foundation
	c.phase = snap.phase
	c.moveCount = snap.moveCount
}

func (c *Canfield) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// canfieldJSON is the JSON wire format for Canfield.
type canfieldJSON struct {
	TrumpCards *TrumpCards                                `json:"tc"`
	Tableau    [CanfieldTableauCnt][]*CanfieldTableauCard `json:"tb"`
	Reserve    []*Card                                    `json:"rv"`
	Stock      []*Card                                    `json:"st"`
	Waste      []*Card                                    `json:"wa"`
	Foundation [CanfieldFoundationCnt][]*Card             `json:"fd"`
	BaseRank   int                                        `json:"br"`
	Phase      CanfieldPhase                              `json:"ps"`
	MoveCount  int                                        `json:"mc"`
	ActionLog  []*ActionLogEntry                          `json:"al"`
	History    []*canfieldSnapshot                        `json:"hi,omitempty"`
}

// canfieldSnapshotJSON is the wire format for a single undo snapshot.
// canfieldSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// canfieldJSON's short keys to keep the KV payload compact (#1654).
type canfieldSnapshotJSON struct {
	Tableau    [CanfieldTableauCnt][]*CanfieldTableauCard `json:"tb"`
	Reserve    []*Card                                    `json:"rv"`
	Stock      []*Card                                    `json:"st"`
	Waste      []*Card                                    `json:"wa"`
	Foundation [CanfieldFoundationCnt][]*Card             `json:"fd"`
	Phase      CanfieldPhase                              `json:"ps"`
	MoveCount  int                                        `json:"mc"`
}

// MarshalJSON implements json.Marshaler for canfieldSnapshot.
func (s *canfieldSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(canfieldSnapshotJSON{
		Tableau:    s.tableau,
		Reserve:    s.reserve,
		Stock:      s.stock,
		Waste:      s.waste,
		Foundation: s.foundation,
		Phase:      s.phase,
		MoveCount:  s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for canfieldSnapshot.
func (s *canfieldSnapshot) UnmarshalJSON(data []byte) error {
	var j canfieldSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > canfieldMaxSliceLen || len(j.Stock) > canfieldMaxSliceLen ||
		len(j.Waste) > canfieldMaxSliceLen {
		return fmt.Errorf("canfield: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > canfieldMaxSliceLen {
			return fmt.Errorf("canfield: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > canfieldMaxSliceLen {
			return fmt.Errorf("canfield: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.reserve = j.Reserve
	if s.reserve == nil {
		s.reserve = make([]*Card, 0)
	}
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
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c *Canfield) MarshalJSON() ([]byte, error) {
	return json.Marshal(canfieldJSON{
		TrumpCards: c.trumpCards,
		Tableau:    c.tableau,
		Reserve:    c.reserve,
		Stock:      c.stock,
		Waste:      c.waste,
		Foundation: c.foundation,
		BaseRank:   c.baseRank,
		Phase:      c.phase,
		MoveCount:  c.moveCount,
		ActionLog:  c.actionLog,
		History:    c.history,
	})
}

// canfieldMaxSliceLen caps slice sizes during deserialisation.
const canfieldMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (c *Canfield) UnmarshalJSON(data []byte) error {
	var j canfieldJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > canfieldMaxSliceLen || len(j.Stock) > canfieldMaxSliceLen ||
		len(j.Waste) > canfieldMaxSliceLen || len(j.ActionLog) > canfieldMaxSliceLen ||
		len(j.History) > canfieldMaxSliceLen {
		return fmt.Errorf("canfield: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > canfieldMaxSliceLen {
			return fmt.Errorf("canfield: tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > canfieldMaxSliceLen {
			return fmt.Errorf("canfield: foundation pile exceeds maximum allowed size")
		}
	}
	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.tableau = j.Tableau
	c.reserve = j.Reserve
	if c.reserve == nil {
		c.reserve = make([]*Card, 0)
	}
	c.stock = j.Stock
	if c.stock == nil {
		c.stock = make([]*Card, 0)
	}
	c.waste = j.Waste
	if c.waste == nil {
		c.waste = make([]*Card, 0)
	}
	c.foundation = j.Foundation
	c.baseRank = j.BaseRank
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	c.history = j.History
	if c.history == nil {
		c.history = make([]*canfieldSnapshot, 0)
	}
	return nil
}
