//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GolfPhase ゴルフソリティアゲームフェーズ
type GolfPhase int

// Golfのフェーズ定数
const (
	// GolfPhasePlaying プレイ中
	GolfPhasePlaying GolfPhase = iota
	// GolfPhaseGameClear ゲームクリア
	GolfPhaseGameClear
	// GolfPhaseGameOver ゲームオーバー
	GolfPhaseGameOver
)

// GolfColCnt ゴルフの列数
const GolfColCnt = 7

// GolfRowCnt ゴルフの段数
const GolfRowCnt = 5

// GolfTableauCnt タブロー上のカード枚数 (7*5)
const GolfTableauCnt = 35

// GolfCard タブロー上のカード
type GolfCard struct {
	Card    *Card `json:"c"`
	Removed bool  `json:"r"`
}

// GolfHint ヒント
type GolfHint struct {
	Type string // "remove" or "draw"
	Col  int
}

// Golf ゴルフソリティアゲームクラス
type Golf struct {
	trumpCards *TrumpCards
	layout     [GolfColCnt][GolfRowCnt]*GolfCard
	stock      []*Card
	waste      []*Card
	phase      GolfPhase
	moveCount  int
	actionLogBase
	history     []*golfSnapshot
	isStalemate bool
}

// golfSnapshot アンドゥ用スナップショット
type golfSnapshot struct {
	layout      [GolfColCnt][GolfRowCnt]*GolfCard
	stock       []*Card
	waste       []*Card
	phase       GolfPhase
	moveCount   int
	isStalemate bool
}

// NewGolf コンストラクタ
func NewGolf(trumpCards *TrumpCards) *Golf {
	return &Golf{
		trumpCards: trumpCards,
	}
}

// NewDefaultGolf returns Golf with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultGolf() *Golf {
	return NewGolf(NewTrumpCards(0))
}

// Reset ゲームリセット
func (g *Golf) Reset() {
	g.trumpCards.Shuffle()
	g.phase = GolfPhasePlaying
	g.moveCount = 0
	g.actionLog = nil
	g.history = nil
	g.isStalemate = false

	// レイアウトをクリア
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			g.layout[c][r] = nil
		}
	}

	// タブロー配置: 7列×5段にカードを配る
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			card := g.trumpCards.DrawCard()
			g.layout[c][r] = &GolfCard{
				Card:    card,
				Removed: false,
			}
		}
	}

	// 残りをストックへ（最後の1枚はウェイストへ）
	g.stock = nil
	g.waste = nil
	for g.trumpCards.GetRemainingCount() > 0 {
		card := g.trumpCards.DrawCard()
		g.stock = append(g.stock, card)
	}

	// ストックから1枚をウェイストに置く（ゲーム開始の基準カード）
	if len(g.stock) > 0 {
		wasteCard := g.stock[len(g.stock)-1]
		g.stock = g.stock[:len(g.stock)-1]
		g.waste = append(g.waste, wasteCard)
	}
}

// Draw ストックからウェイストにカードを引く
func (g *Golf) Draw() error {
	if g.phase != GolfPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(g.stock) == 0 {
		return errors.New("no cards in stock")
	}
	g.takeSnapshot()
	card := g.stock[len(g.stock)-1]
	g.stock = g.stock[:len(g.stock)-1]
	g.waste = append(g.waste, card)
	g.moveCount++
	g.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	g.checkStalemate()
	return nil
}

// Remove タブローのカードを除去
func (g *Golf) Remove(col int) error {
	if g.phase != GolfPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= GolfColCnt {
		return errors.New("invalid column")
	}
	// 露出カードを探す（下から上へスキャン）
	row := g.findExposedRow(col)
	if row < 0 {
		return errors.New("no card in column")
	}
	gc := g.layout[col][row]
	if len(g.waste) == 0 {
		return errors.New("waste is empty")
	}
	wasteTop := g.waste[len(g.waste)-1]
	if !g.isAdjacentRank(gc.Card, wasteTop) {
		return errors.New("card is not adjacent to waste top")
	}
	g.takeSnapshot()
	gc.Removed = true
	// 除去したカードをウェイストの上に置く
	g.waste = append(g.waste, gc.Card)
	g.moveCount++
	g.appendLog("remove", fmt.Sprintf("カード除去: 列%d", col), []*Card{gc.Card})
	g.checkGameClear()
	g.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (g *Golf) GiveUp() {
	if g.phase == GolfPhasePlaying {
		g.phase = GolfPhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (g *Golf) GetHint() *GolfHint {
	if g.phase != GolfPhasePlaying {
		return nil
	}
	if len(g.waste) == 0 {
		return nil
	}
	// 露出カードで除去可能なものを探す
	for c := range GolfColCnt {
		row := g.findExposedRow(c)
		if row < 0 {
			continue
		}
		gc := g.layout[c][row]
		if g.isAdjacentRank(gc.Card, g.waste[len(g.waste)-1]) {
			return &GolfHint{
				Type: "remove",
				Col:  c,
			}
		}
	}
	// カード除去不可、ストックからドロー可能な場合
	if len(g.stock) > 0 {
		return &GolfHint{
			Type: "draw",
			Col:  -1,
		}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (g *Golf) Undo() error {
	if g.phase != GolfPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(g.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (g *Golf) CanUndo() bool {
	return len(g.history) > 0 && g.phase == GolfPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (g *Golf) UndoToEscape() int {
	return undoToEscape(g.isStalemate, g.history, func(s *golfSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (g *Golf) UndoN(n int) error {
	return undoN(g, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (g *Golf) GetPhase() GolfPhase { return g.phase }

// SetPhase フェーズ設定 (テスト用)
func (g *Golf) SetPhase(phase GolfPhase) { g.phase = phase }

// GetMoveCount 移動回数取得
func (g *Golf) GetMoveCount() int { return g.moveCount }

// GetStockCount ストック枚数取得
func (g *Golf) GetStockCount() int { return len(g.stock) }

// GetWaste ウェイスト取得
func (g *Golf) GetWaste() []*Card { return g.waste }

// GetLayout レイアウト取得
func (g *Golf) GetLayout() [GolfColCnt][GolfRowCnt]*GolfCard {
	return g.layout
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (g *Golf) GetGameEndFlag() bool { return g.phase != GolfPhasePlaying }

// IsStalemate 手詰まり状態取得
func (g *Golf) IsStalemate() bool { return g.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (g *Golf) SetIsStalemate(v bool) { g.isStalemate = v }

// SetLayout レイアウト設定 (テスト用)
func (g *Golf) SetLayout(layout [GolfColCnt][GolfRowCnt]*GolfCard) {
	g.layout = layout
}

// SetStock ストック設定 (テスト用)
func (g *Golf) SetStock(stock []*Card) { g.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (g *Golf) SetWaste(waste []*Card) { g.waste = waste }

// IsExposed カードが露出しているか (テスト用の公開版)
func (g *Golf) IsExposed(col, row int) bool {
	return g.isExposed(col, row)
}

// AllRemoved 全タブローカードが除去されたか
func (g *Golf) AllRemoved() bool {
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			if g.layout[c][r] != nil && !g.layout[c][r].Removed {
				return false
			}
		}
	}
	return true
}

// --- Private helpers ---

// findExposedRow 列内の露出カードの行番号を返す (-1 = なし)
func (g *Golf) findExposedRow(col int) int {
	for r := GolfRowCnt - 1; r >= 0; r-- {
		gc := g.layout[col][r]
		if gc != nil && !gc.Removed {
			return r
		}
	}
	return -1
}

// isExposed カードが露出しているか判定
func (g *Golf) isExposed(col, row int) bool {
	if col < 0 || col >= GolfColCnt || row < 0 || row >= GolfRowCnt {
		return false
	}
	gc := g.layout[col][row]
	if gc == nil || gc.Removed {
		return false
	}
	// 列内で最も下（行番号が大きい）の未除去カードが露出
	return g.findExposedRow(col) == row
}

// isAdjacentRank 2枚のカードのランクが±1か判定 (K-Aラップあり)
func (g *Golf) isAdjacentRank(card1, card2 *Card) bool {
	v1 := card1.GetValue()
	v2 := card2.GetValue()
	diff := v1 - v2
	if diff < 0 {
		diff = -diff
	}
	// 通常の±1
	if diff == 1 {
		return true
	}
	// K(13)-A(1) ラップアラウンド
	if diff == 12 {
		return true
	}
	return false
}

// checkGameClear ゲームクリア判定
func (g *Golf) checkGameClear() {
	if g.AllRemoved() {
		g.phase = GolfPhaseGameClear
	}
}

// checkStalemate 手詰まり判定
func (g *Golf) checkStalemate() {
	if g.phase != GolfPhasePlaying {
		return
	}
	// ストックにカードがあればまだドローできる
	if len(g.stock) > 0 {
		g.isStalemate = false
		return
	}
	// ヒントがあればスタルメイトではない
	hint := g.GetHint()
	if hint != nil {
		g.isStalemate = false
		return
	}
	g.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (g *Golf) takeSnapshot() {
	snap := &golfSnapshot{
		phase:       g.phase,
		moveCount:   g.moveCount,
		isStalemate: g.isStalemate,
	}
	// deep copy layout
	for c := range GolfColCnt {
		for r := range GolfRowCnt {
			if g.layout[c][r] != nil {
				snap.layout[c][r] = &GolfCard{Card: g.layout[c][r].Card, Removed: g.layout[c][r].Removed}
			}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(g.stock))
	copy(snap.stock, g.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(g.waste))
	copy(snap.waste, g.waste)
	g.history = append(g.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (g *Golf) restoreSnapshot(snap *golfSnapshot) {
	g.layout = snap.layout
	g.stock = snap.stock
	g.waste = snap.waste
	g.phase = snap.phase
	g.moveCount = snap.moveCount
	g.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (g *Golf) appendLog(actionType, detail string, cards []*Card) {
	g.appendLogAt(g.moveCount, 0, actionType, detail, cards)
}

// golfJSON is the JSON wire format for Golf.
type golfJSON struct {
	TrumpCards  *TrumpCards                       `json:"tc"`
	Layout      [GolfColCnt][GolfRowCnt]*GolfCard `json:"ly"`
	Stock       []*Card                           `json:"st"`
	Waste       []*Card                           `json:"wa"`
	Phase       GolfPhase                         `json:"ps"`
	MoveCount   int                               `json:"mc"`
	ActionLog   []*ActionLogEntry                 `json:"al"`
	IsStalemate bool                              `json:"sm"`
	History     []*golfSnapshot                   `json:"hi,omitempty"`
}

// golfSnapshotJSON is the wire format for a single undo snapshot.
// golfSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// golfJSON's short keys to keep the KV payload compact (#1654).
type golfSnapshotJSON struct {
	Layout      [GolfColCnt][GolfRowCnt]*GolfCard `json:"ly"`
	Stock       []*Card                           `json:"st"`
	Waste       []*Card                           `json:"wa"`
	Phase       GolfPhase                         `json:"ps"`
	MoveCount   int                               `json:"mc"`
	IsStalemate bool                              `json:"sm"`
}

// MarshalJSON implements json.Marshaler for golfSnapshot.
func (s *golfSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(golfSnapshotJSON{
		Layout:      s.layout,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for golfSnapshot.
func (s *golfSnapshot) UnmarshalJSON(data []byte) error {
	var j golfSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > golfMaxSliceLen {
		return fmt.Errorf("golf: snapshot stock exceeds maximum allowed size")
	}
	if len(j.Waste) > golfMaxSliceLen {
		return fmt.Errorf("golf: snapshot waste exceeds maximum allowed size")
	}
	s.layout = j.Layout
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.waste = j.Waste
	if s.waste == nil {
		s.waste = make([]*Card, 0)
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (g *Golf) MarshalJSON() ([]byte, error) {
	return json.Marshal(golfJSON{
		TrumpCards:  g.trumpCards,
		Layout:      g.layout,
		Stock:       g.stock,
		Waste:       g.waste,
		Phase:       g.phase,
		MoveCount:   g.moveCount,
		ActionLog:   g.actionLog,
		IsStalemate: g.isStalemate,
		History:     g.history,
	})
}

// golfMaxSliceLen caps slice sizes during deserialisation.
const golfMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (g *Golf) UnmarshalJSON(data []byte) error {
	var j golfJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > golfMaxSliceLen || len(j.Waste) > golfMaxSliceLen ||
		len(j.ActionLog) > golfMaxSliceLen || len(j.History) > golfMaxSliceLen {
		return fmt.Errorf("golf: input array exceeds maximum allowed size")
	}

	g.trumpCards = j.TrumpCards
	if g.trumpCards == nil {
		g.trumpCards = NewTrumpCards(0)
	}
	g.layout = j.Layout
	g.stock = j.Stock
	if g.stock == nil {
		g.stock = make([]*Card, 0)
	}
	g.waste = j.Waste
	if g.waste == nil {
		g.waste = make([]*Card, 0)
	}
	g.phase = j.Phase
	g.moveCount = j.MoveCount
	g.actionLog = j.ActionLog
	if g.actionLog == nil {
		g.actionLog = make([]*ActionLogEntry, 0)
	}
	g.history = j.History
	if g.history == nil {
		g.history = make([]*golfSnapshot, 0)
	}
	g.isStalemate = j.IsStalemate
	return nil
}
