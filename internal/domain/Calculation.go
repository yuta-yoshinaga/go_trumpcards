//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CalculationPhase カルキュレーションゲームフェーズ
type CalculationPhase int

// Calculationのフェーズ定数
const (
	// CalculationPhasePlaying プレイ中
	CalculationPhasePlaying CalculationPhase = iota
	// CalculationPhaseGameClear ゲームクリア
	CalculationPhaseGameClear
	// CalculationPhaseGameOver ゲームオーバー
	CalculationPhaseGameOver
)

// CalculationFoundationCnt ファンデーション数（A,2,3,4のベース）
const CalculationFoundationCnt = 4

// CalculationWasteCnt ウェイストパイルの数
const CalculationWasteCnt = 4

// CalculationHint カルキュレーションのヒント
type CalculationHint struct {
	// FromZone 移動元ゾーン "stock" または "waste"
	FromZone string
	// WasteIdx 移動元がウェイストの場合のインデックス、stockの場合は -1
	WasteIdx int
	// FoundationIdx 移動先ファンデーションのインデックス
	FoundationIdx int
}

// Calculation カルキュレーションゲームクラス
type Calculation struct {
	trumpCards  *TrumpCards
	foundations [CalculationFoundationCnt][]*Card
	wastes      [CalculationWasteCnt][]*Card
	stock       []*Card
	phase       CalculationPhase
	moveCount   int
	actionLogBase
	history     []*calculationSnapshot
	isStalemate bool
}

// calculationSnapshot アンドゥ用スナップショット
type calculationSnapshot struct {
	foundations [CalculationFoundationCnt][]*Card
	wastes      [CalculationWasteCnt][]*Card
	stock       []*Card
	phase       CalculationPhase
	moveCount   int
	isStalemate bool
}

// NewCalculation コンストラクタ
func NewCalculation(trumpCards *TrumpCards) *Calculation {
	return &Calculation{trumpCards: trumpCards}
}

// NewDefaultCalculation returns Calculation with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCalculation() *Calculation {
	return NewCalculation(NewTrumpCards(0))
}

// Reset ゲームリセット
func (c *Calculation) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CalculationPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false

	// ファンデーション初期化 (draw中に埋める)
	for i := range CalculationFoundationCnt {
		c.foundations[i] = nil
	}
	// ウェイスト初期化
	for i := range CalculationWasteCnt {
		c.wastes[i] = nil
	}

	// シャッフル済みデッキから A,2,3,4 を見つけてファンデーションのベースとし、残りをストックへ
	placed := [CalculationFoundationCnt]bool{}
	c.stock = nil
	for c.trumpCards.GetRemainingCount() > 0 {
		card := c.trumpCards.DrawCard()
		v := card.GetValue()
		if v >= 1 && v <= CalculationFoundationCnt && !placed[v-1] {
			c.foundations[v-1] = []*Card{card}
			placed[v-1] = true
			continue
		}
		c.stock = append(c.stock, card)
	}
}

// PlayStockToFoundation ストック最上段をファンデーションに置く
func (c *Calculation) PlayStockToFoundation(fIdx int) error {
	if c.phase != CalculationPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fIdx < 0 || fIdx >= CalculationFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(c.stock) == 0 {
		return errors.New("stock is empty")
	}
	card := c.stock[len(c.stock)-1]
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.stock = c.stock[:len(c.stock)-1]
	c.foundations[fIdx] = append(c.foundations[fIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("ストック→ファンデーション%d", fIdx+1), []*Card{card})
	c.checkGameClear()
	c.checkStalemate()
	return nil
}

// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
func (c *Calculation) PlayStockToWaste(wasteIdx int) error {
	if c.phase != CalculationPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if wasteIdx < 0 || wasteIdx >= CalculationWasteCnt {
		return errors.New("invalid waste index")
	}
	if len(c.stock) == 0 {
		return errors.New("stock is empty")
	}
	c.takeSnapshot()
	card := c.stock[len(c.stock)-1]
	c.stock = c.stock[:len(c.stock)-1]
	c.wastes[wasteIdx] = append(c.wastes[wasteIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("ストック→ウェイスト%d", wasteIdx+1), []*Card{card})
	c.checkStalemate()
	return nil
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (c *Calculation) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	if c.phase != CalculationPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if wasteIdx < 0 || wasteIdx >= CalculationWasteCnt {
		return errors.New("invalid waste index")
	}
	if fIdx < 0 || fIdx >= CalculationFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(c.wastes[wasteIdx]) == 0 {
		return errors.New("waste is empty")
	}
	card := c.wastes[wasteIdx][len(c.wastes[wasteIdx])-1]
	if !c.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	c.takeSnapshot()
	c.wastes[wasteIdx] = c.wastes[wasteIdx][:len(c.wastes[wasteIdx])-1]
	c.foundations[fIdx] = append(c.foundations[fIdx], card)
	c.moveCount++
	c.appendLog("move", fmt.Sprintf("ウェイスト%d→ファンデーション%d", wasteIdx+1, fIdx+1), []*Card{card})
	c.checkGameClear()
	c.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (c *Calculation) GiveUp() {
	if c.phase == CalculationPhasePlaying {
		c.phase = CalculationPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得（ファンデーションに置けるカードを優先的に提示する）
func (c *Calculation) GetHint() *CalculationHint {
	if c.phase != CalculationPhasePlaying {
		return nil
	}
	// ストック最上段がファンデーションに置けるか
	if len(c.stock) > 0 {
		card := c.stock[len(c.stock)-1]
		for fIdx := range CalculationFoundationCnt {
			if c.canPlaceOnFoundation(card, fIdx) {
				return &CalculationHint{FromZone: "stock", WasteIdx: -1, FoundationIdx: fIdx}
			}
		}
	}
	// 各ウェイスト最上段がファンデーションに置けるか
	for wIdx := range CalculationWasteCnt {
		if len(c.wastes[wIdx]) == 0 {
			continue
		}
		card := c.wastes[wIdx][len(c.wastes[wIdx])-1]
		for fIdx := range CalculationFoundationCnt {
			if c.canPlaceOnFoundation(card, fIdx) {
				return &CalculationHint{FromZone: "waste", WasteIdx: wIdx, FoundationIdx: fIdx}
			}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（ストックが空の場合、残るウェイストをファンデーションに自動で送る）。
//
// Undo 挙動: バッチ全体で 1 つだけスナップショットを取るため、Undo するとオート
// コンプリート開始前の状態にまとめて戻る（個々のカード移動単位では戻らない）。
// プレイヤー操作の「1 手」として扱う設計。
func (c *Calculation) AutoComplete() error {
	if c.phase != CalculationPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !c.AllFaceUp() {
		return errors.New("stock is not empty")
	}
	c.takeSnapshot()
	for {
		moved := false
		for wIdx := range CalculationWasteCnt {
			if len(c.wastes[wIdx]) == 0 {
				continue
			}
			card := c.wastes[wIdx][len(c.wastes[wIdx])-1]
			fIdx := c.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			c.wastes[wIdx] = c.wastes[wIdx][:len(c.wastes[wIdx])-1]
			c.foundations[fIdx] = append(c.foundations[fIdx], card)
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

// AllFaceUp 全カードが可視かどうか（ストックが空であること）
func (c *Calculation) AllFaceUp() bool { return len(c.stock) == 0 }

// Undo 直前の操作を取り消す
func (c *Calculation) Undo() error {
	if c.phase != CalculationPhasePlaying {
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
func (c *Calculation) CanUndo() bool {
	return len(c.history) > 0 && c.phase == CalculationPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (c *Calculation) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *calculationSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する
func (c *Calculation) UndoN(n int) error {
	for i := range n {
		if err := c.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Getters ---

// GetPhase フェーズ取得
func (c *Calculation) GetPhase() CalculationPhase { return c.phase }

// GetMoveCount 移動回数取得
func (c *Calculation) GetMoveCount() int { return c.moveCount }

// GetStockCount ストック枚数取得
func (c *Calculation) GetStockCount() int { return len(c.stock) }

// GetWastes ウェイスト取得
func (c *Calculation) GetWastes() [CalculationWasteCnt][]*Card { return c.wastes }

// GetFoundations ファンデーション取得
func (c *Calculation) GetFoundations() [CalculationFoundationCnt][]*Card { return c.foundations }

// GetGameEndFlag returns true once the game has left the playing phase.
func (c *Calculation) GetGameEndFlag() bool { return c.phase != CalculationPhasePlaying }

// IsStalemate 手詰まり状態取得
func (c *Calculation) IsStalemate() bool { return c.isStalemate }

// GetStockTop ストック最上段のカードを取得（空の場合 nil）
func (c *Calculation) GetStockTop() *Card {
	if len(c.stock) == 0 {
		return nil
	}
	return c.stock[len(c.stock)-1]
}

// --- Private helpers ---

// canPlaceOnFoundation ファンデーションにカードを置けるか判定
func (c *Calculation) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := c.foundations[fIdx]
	if len(pile) == 0 {
		return false
	}
	if len(pile) >= CardValueMax {
		return false
	}
	topCard := pile[len(pile)-1]
	return card.GetValue() == calculationNextValue(topCard.GetValue(), fIdx+1)
}

// GetNextFoundationRank は fIdx のファンデーションに次に置けるランクを返す。
// もう置けない (山が空 / 13枚そろった) ときは 0。
//
// **各列が +1/+2/+3/+4 ずつ 13 を法として進む (#4794)。**Web はバッジと
// 「次のカード列」で常時出しているのに、CUI は現在の一番上の札しか出さず、
// 毎手この暗算を強いていた。
//
// 判定は配置の検証 canPlaceOnFoundation が使う calculationNextValue を通す。
// **別実装にすると、案内したランクが実際には置けないことが起きる。**
func (c *Calculation) GetNextFoundationRank(fIdx int) int {
	if fIdx < 0 || fIdx >= len(c.foundations) {
		return 0
	}
	pile := c.foundations[fIdx]
	if len(pile) == 0 || len(pile) >= CardValueMax {
		return 0
	}
	return calculationNextValue(pile[len(pile)-1].GetValue(), fIdx+1)
}

// calculationNextValue ファンデーションの次に置くべき値を返す（step は 1..4、V は 1..13）。
// v+step は最大で 13+4 = 17 なので、mod 13 は一度の減算で十分（ループ不要）。
func calculationNextValue(v, step int) int {
	next := v + step
	if next > CardValueMax {
		next -= CardValueMax
	}
	return next
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）
func (c *Calculation) findFoundation(card *Card) int {
	for i := range CalculationFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (c *Calculation) checkGameClear() {
	for i := range CalculationFoundationCnt {
		if len(c.foundations[i]) != CardValueMax {
			return
		}
	}
	c.phase = CalculationPhaseGameClear
}

// checkStalemate 手詰まり判定
func (c *Calculation) checkStalemate() {
	if c.phase != CalculationPhasePlaying {
		return
	}
	// ストックが残っていればウェイストに積めるので手詰まりではない
	if len(c.stock) > 0 {
		c.isStalemate = false
		return
	}
	// ストックが空：どのウェイスト最上段もファンデーションに置けなければ手詰まり
	if c.GetHint() == nil {
		c.isStalemate = true
		return
	}
	c.isStalemate = false
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (c *Calculation) takeSnapshot() {
	snap := &calculationSnapshot{
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range CalculationFoundationCnt {
		snap.foundations[i] = make([]*Card, len(c.foundations[i]))
		copy(snap.foundations[i], c.foundations[i])
	}
	for i := range CalculationWasteCnt {
		snap.wastes[i] = make([]*Card, len(c.wastes[i]))
		copy(snap.wastes[i], c.wastes[i])
	}
	snap.stock = make([]*Card, len(c.stock))
	copy(snap.stock, c.stock)
	c.history = append(c.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (c *Calculation) restoreSnapshot(snap *calculationSnapshot) {
	c.foundations = snap.foundations
	c.wastes = snap.wastes
	c.stock = snap.stock
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (c *Calculation) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// calculationJSON is the JSON wire format for Calculation.
type calculationJSON struct {
	TrumpCards  *TrumpCards                       `json:"tc"`
	Foundations [CalculationFoundationCnt][]*Card `json:"fd"`
	Wastes      [CalculationWasteCnt][]*Card      `json:"wa"`
	Stock       []*Card                           `json:"st"`
	Phase       CalculationPhase                  `json:"ps"`
	MoveCount   int                               `json:"mc"`
	ActionLog   []*ActionLogEntry                 `json:"al"`
	IsStalemate bool                              `json:"sl"`
	History     []*calculationSnapshot            `json:"hi,omitempty"`
}

// calculationSnapshotJSON is the wire format for a single undo snapshot.
// calculationSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// calculationJSON's short keys to keep the KV payload compact (#1654).
type calculationSnapshotJSON struct {
	Foundations [CalculationFoundationCnt][]*Card `json:"fd"`
	Wastes      [CalculationWasteCnt][]*Card      `json:"wa"`
	Stock       []*Card                           `json:"st"`
	Phase       CalculationPhase                  `json:"ps"`
	MoveCount   int                               `json:"mc"`
	IsStalemate bool                              `json:"sl"`
}

// MarshalJSON implements json.Marshaler for calculationSnapshot.
func (s *calculationSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(calculationSnapshotJSON{
		Foundations: s.foundations,
		Wastes:      s.wastes,
		Stock:       s.stock,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for calculationSnapshot.
func (s *calculationSnapshot) UnmarshalJSON(data []byte) error {
	var j calculationSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > calculationMaxSliceLen {
		return fmt.Errorf("calculation: snapshot stock exceeds maximum allowed size")
	}
	for i := range CalculationFoundationCnt {
		if len(j.Foundations[i]) > calculationMaxSliceLen {
			return fmt.Errorf("calculation: snapshot foundation %d exceeds maximum allowed size", i)
		}
	}
	for i := range CalculationWasteCnt {
		if len(j.Wastes[i]) > calculationMaxSliceLen {
			return fmt.Errorf("calculation: snapshot waste %d exceeds maximum allowed size", i)
		}
	}
	s.foundations = j.Foundations
	for i := range CalculationFoundationCnt {
		if s.foundations[i] == nil {
			s.foundations[i] = make([]*Card, 0)
		}
	}
	s.wastes = j.Wastes
	for i := range CalculationWasteCnt {
		if s.wastes[i] == nil {
			s.wastes[i] = make([]*Card, 0)
		}
	}
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (c *Calculation) MarshalJSON() ([]byte, error) {
	return json.Marshal(calculationJSON{
		TrumpCards:  c.trumpCards,
		Foundations: c.foundations,
		Wastes:      c.wastes,
		Stock:       c.stock,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// calculationMaxSliceLen caps slice sizes during deserialisation.
const calculationMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (c *Calculation) UnmarshalJSON(data []byte) error {
	var j calculationJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > calculationMaxSliceLen || len(j.ActionLog) > calculationMaxSliceLen ||
		len(j.History) > calculationMaxSliceLen {
		return fmt.Errorf("calculation: input array exceeds maximum allowed size")
	}
	for i := range CalculationFoundationCnt {
		if len(j.Foundations[i]) > calculationMaxSliceLen {
			return fmt.Errorf("calculation: foundation %d exceeds maximum allowed size", i)
		}
	}
	for i := range CalculationWasteCnt {
		if len(j.Wastes[i]) > calculationMaxSliceLen {
			return fmt.Errorf("calculation: waste %d exceeds maximum allowed size", i)
		}
	}

	c.trumpCards = j.TrumpCards
	if c.trumpCards == nil {
		c.trumpCards = NewTrumpCards(0)
	}
	c.foundations = j.Foundations
	for i := range CalculationFoundationCnt {
		if c.foundations[i] == nil {
			c.foundations[i] = make([]*Card, 0)
		}
	}
	c.wastes = j.Wastes
	for i := range CalculationWasteCnt {
		if c.wastes[i] == nil {
			c.wastes[i] = make([]*Card, 0)
		}
	}
	c.stock = j.Stock
	if c.stock == nil {
		c.stock = make([]*Card, 0)
	}
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	if c.actionLog == nil {
		c.actionLog = make([]*ActionLogEntry, 0)
	}
	c.history = j.History
	if c.history == nil {
		c.history = make([]*calculationSnapshot, 0)
	}
	c.isStalemate = j.IsStalemate
	return nil
}
