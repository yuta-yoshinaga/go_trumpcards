//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PyramidPhase ピラミッドゲームフェーズ
type PyramidPhase int

// Pyramidのフェーズ定数
const (
	// PyramidPhasePlaying プレイ中
	PyramidPhasePlaying PyramidPhase = iota
	// PyramidPhaseGameClear ゲームクリア
	PyramidPhaseGameClear
	// PyramidPhaseGameOver ゲームオーバー
	PyramidPhaseGameOver
)

// PyramidRowCnt ピラミッドの段数
const PyramidRowCnt = 7

// PyramidCardCnt ピラミッドのカード枚数 (1+2+3+4+5+6+7)
const PyramidCardCnt = 28

// PyramidTargetSum ペア除去の合計値
const PyramidTargetSum = 13

// PyramidCard ピラミッド上のカード
type PyramidCard struct {
	Card    *Card `json:"c"`
	Removed bool  `json:"r"`
}

// PyramidHint ヒント
type PyramidHint struct {
	Type string // "pair", "king", "waste_pair", "waste_king"
	Row1 int
	Col1 int
	Row2 int // -1 for king/waste operations
	Col2 int // -1 for king/waste operations
}

// Pyramid ピラミッドソリティアゲームクラス
type Pyramid struct {
	trumpCards *TrumpCards
	pyramid    [PyramidRowCnt][]*PyramidCard
	stock      []*Card
	waste      []*Card
	phase      PyramidPhase
	moveCount  int
	actionLogBase
	history     []*pyramidSnapshot
	isStalemate bool
}

// pyramidSnapshot アンドゥ用スナップショット
type pyramidSnapshot struct {
	pyramid     [PyramidRowCnt][]*PyramidCard
	stock       []*Card
	waste       []*Card
	phase       PyramidPhase
	moveCount   int
	isStalemate bool
}

// NewPyramid コンストラクタ
func NewPyramid(trumpCards *TrumpCards) *Pyramid {
	return &Pyramid{
		trumpCards: trumpCards,
	}
}

// NewDefaultPyramid returns Pyramid with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultPyramid() *Pyramid {
	return NewPyramid(NewTrumpCards(0))
}

// Reset ゲームリセット
func (p *Pyramid) Reset() {
	p.trumpCards.Shuffle()
	p.phase = PyramidPhasePlaying
	p.moveCount = 0
	p.actionLog = nil
	p.history = nil
	p.isStalemate = false

	// ピラミッドに配る: 段iにはi+1枚（全て表向き）
	for i := range PyramidRowCnt {
		p.pyramid[i] = make([]*PyramidCard, 0, i+1)
		for range i + 1 {
			card := p.trumpCards.DrawCard()
			pc := &PyramidCard{
				Card:    card,
				Removed: false,
			}
			p.pyramid[i] = append(p.pyramid[i], pc)
		}
	}

	// 残りをストックへ
	p.stock = nil
	p.waste = nil
	for p.trumpCards.GetRemainingCount() > 0 {
		card := p.trumpCards.DrawCard()
		p.stock = append(p.stock, card)
	}
}

// Draw ストックからウェイストにカードを引く
func (p *Pyramid) Draw() error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(p.stock) == 0 {
		return errors.New("no cards in stock")
	}
	p.takeSnapshot()
	card := p.stock[len(p.stock)-1]
	p.stock = p.stock[:len(p.stock)-1]
	p.waste = append(p.waste, card)
	p.moveCount++
	p.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	p.checkStalemate()
	return nil
}

// RemovePair ピラミッド上の2枚のカードを合計13で除去
func (p *Pyramid) RemovePair(row1, col1, row2, col2 int) error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if err := p.validatePyramidPos(row1, col1); err != nil {
		return err
	}
	if err := p.validatePyramidPos(row2, col2); err != nil {
		return err
	}
	if row1 == row2 && col1 == col2 {
		return errors.New("same card selected twice")
	}
	pc1 := p.pyramid[row1][col1]
	pc2 := p.pyramid[row2][col2]
	if pc1.Removed {
		return errors.New("card 1 is already removed")
	}
	if pc2.Removed {
		return errors.New("card 2 is already removed")
	}
	if !p.isExposed(row1, col1) {
		return errors.New("card 1 is not exposed")
	}
	if !p.isExposed(row2, col2) {
		return errors.New("card 2 is not exposed")
	}
	if pc1.Card.GetValue()+pc2.Card.GetValue() != PyramidTargetSum {
		return errors.New("cards do not sum to 13")
	}
	p.takeSnapshot()
	pc1.Removed = true
	pc2.Removed = true
	p.moveCount++
	p.appendLog("remove", fmt.Sprintf("ペア除去: (%d,%d)+(%d,%d)", row1, col1, row2, col2),
		[]*Card{pc1.Card, pc2.Card})
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// IsRemovableKing は (row, col) のカードが「いま単独で除去できるキング」かを返す。
//
// **RemoveKing が通る条件と同じものを見る (#4782)。**キングは相方が要らず
// クリックだけで消せるので、Web は常時ハイライトしている。印を付ける条件と
// 実際に通る条件が別々だと、消せない札に印が付く。
func (p *Pyramid) IsRemovableKing(row, col int) bool {
	if p.phase != PyramidPhasePlaying {
		return false
	}
	if err := p.validatePyramidPos(row, col); err != nil {
		return false
	}
	// isExposed が除去済みを弾くので、ここで Removed を見直さない
	// (見ても常に同じ結果になる分岐が増えるだけ)。
	return p.isExposed(row, col) && p.pyramid[row][col].Card.GetValue() == PyramidTargetSum
}

// IsWasteKingRemovable はウェイストのトップが単独で除去できるキングかを返す。
func (p *Pyramid) IsWasteKingRemovable() bool {
	if p.phase != PyramidPhasePlaying || len(p.waste) == 0 {
		return false
	}
	return p.waste[len(p.waste)-1].GetValue() == PyramidTargetSum
}

// RemoveKing ピラミッド上のKを単独で除去
func (p *Pyramid) RemoveKing(row, col int) error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if err := p.validatePyramidPos(row, col); err != nil {
		return err
	}
	pc := p.pyramid[row][col]
	if pc.Removed {
		return errors.New("card is already removed")
	}
	if !p.isExposed(row, col) {
		return errors.New("card is not exposed")
	}
	if pc.Card.GetValue() != PyramidTargetSum {
		return errors.New("card is not a King")
	}
	p.takeSnapshot()
	pc.Removed = true
	p.moveCount++
	p.appendLog("remove", fmt.Sprintf("キング除去: (%d,%d)", row, col), []*Card{pc.Card})
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// RemoveWithWaste ウェイストのトップカードとピラミッドのカードをペアで除去
func (p *Pyramid) RemoveWithWaste(row, col int) error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if err := p.validatePyramidPos(row, col); err != nil {
		return err
	}
	if len(p.waste) == 0 {
		return errors.New("waste is empty")
	}
	pc := p.pyramid[row][col]
	if pc.Removed {
		return errors.New("card is already removed")
	}
	if !p.isExposed(row, col) {
		return errors.New("card is not exposed")
	}
	wasteCard := p.waste[len(p.waste)-1]
	if pc.Card.GetValue()+wasteCard.GetValue() != PyramidTargetSum {
		return errors.New("cards do not sum to 13")
	}
	p.takeSnapshot()
	pc.Removed = true
	p.waste = p.waste[:len(p.waste)-1]
	p.moveCount++
	p.appendLog("remove", fmt.Sprintf("ウェイスト+ピラミッド(%d,%d)除去", row, col),
		[]*Card{wasteCard, pc.Card})
	p.checkGameClear()
	p.checkStalemate()
	return nil
}

// RemoveWasteKing ウェイストのトップのKを単独で除去
func (p *Pyramid) RemoveWasteKing() error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(p.waste) == 0 {
		return errors.New("waste is empty")
	}
	wasteCard := p.waste[len(p.waste)-1]
	if wasteCard.GetValue() != PyramidTargetSum {
		return errors.New("waste top card is not a King")
	}
	p.takeSnapshot()
	p.waste = p.waste[:len(p.waste)-1]
	p.moveCount++
	p.appendLog("remove", "ウェイストのキング除去", []*Card{wasteCard})
	// checkGameClear is intentionally omitted here: AllRemoved only checks pyramid cards,
	// so removing a waste King can never trigger a game clear.
	p.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (p *Pyramid) GiveUp() {
	if p.phase == PyramidPhasePlaying {
		p.phase = PyramidPhaseGameOver
		p.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (p *Pyramid) GetHint() *PyramidHint {
	if p.phase != PyramidPhasePlaying {
		return nil
	}
	// 優先度1: ピラミッド上のキング除去
	for row := range PyramidRowCnt {
		for col := range row + 1 {
			pc := p.pyramid[row][col]
			if pc.Removed || !p.isExposed(row, col) {
				continue
			}
			if pc.Card.GetValue() == PyramidTargetSum {
				return &PyramidHint{
					Type: "king",
					Row1: row, Col1: col,
					Row2: -1, Col2: -1,
				}
			}
		}
	}
	// 優先度2: ピラミッド上のペア除去
	exposedCards := p.getExposedCards()
	for i := 0; i < len(exposedCards); i++ {
		for j := i + 1; j < len(exposedCards); j++ {
			r1, c1 := exposedCards[i][0], exposedCards[i][1]
			r2, c2 := exposedCards[j][0], exposedCards[j][1]
			v1 := p.pyramid[r1][c1].Card.GetValue()
			v2 := p.pyramid[r2][c2].Card.GetValue()
			if v1+v2 == PyramidTargetSum {
				return &PyramidHint{
					Type: "pair",
					Row1: r1, Col1: c1,
					Row2: r2, Col2: c2,
				}
			}
		}
	}
	// 優先度3: ウェイストのキング除去
	if len(p.waste) > 0 && p.waste[len(p.waste)-1].GetValue() == PyramidTargetSum {
		return &PyramidHint{
			Type: "waste_king",
			Row1: -1, Col1: -1,
			Row2: -1, Col2: -1,
		}
	}
	// 優先度4: ウェイスト+ピラミッドのペア除去
	if len(p.waste) > 0 {
		wasteVal := p.waste[len(p.waste)-1].GetValue()
		for _, ec := range exposedCards {
			row, col := ec[0], ec[1]
			if p.pyramid[row][col].Card.GetValue()+wasteVal == PyramidTargetSum {
				return &PyramidHint{
					Type: "waste_pair",
					Row1: row, Col1: col,
					Row2: -1, Col2: -1,
				}
			}
		}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (p *Pyramid) Undo() error {
	if p.phase != PyramidPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(p.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := p.history[len(p.history)-1]
	p.history = p.history[:len(p.history)-1]
	p.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (p *Pyramid) CanUndo() bool {
	return len(p.history) > 0 && p.phase == PyramidPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (p *Pyramid) UndoToEscape() int {
	return undoToEscape(p.isStalemate, p.history, func(s *pyramidSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (p *Pyramid) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := p.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (p *Pyramid) GetPhase() PyramidPhase { return p.phase }

// SetPhase フェーズ設定 (テスト用)
func (p *Pyramid) SetPhase(phase PyramidPhase) { p.phase = phase }

// GetMoveCount 移動回数取得
func (p *Pyramid) GetMoveCount() int { return p.moveCount }

// GetStockCount ストック枚数取得
func (p *Pyramid) GetStockCount() int { return len(p.stock) }

// GetWaste ウェイスト取得
func (p *Pyramid) GetWaste() []*Card { return p.waste }

// GetPyramid ピラミッド取得
func (p *Pyramid) GetPyramid() [PyramidRowCnt][]*PyramidCard { return p.pyramid }

// GetGameEndFlag returns true once the game has left the playing phase.
func (p *Pyramid) GetGameEndFlag() bool { return p.phase != PyramidPhasePlaying }

// IsStalemate 手詰まり状態取得
func (p *Pyramid) IsStalemate() bool { return p.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (p *Pyramid) SetIsStalemate(v bool) { p.isStalemate = v }

// SetPyramid ピラミッド設定 (テスト用)
func (p *Pyramid) SetPyramid(pyramid [PyramidRowCnt][]*PyramidCard) {
	p.pyramid = pyramid
}

// SetStock ストック設定 (テスト用)
func (p *Pyramid) SetStock(stock []*Card) { p.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (p *Pyramid) SetWaste(waste []*Card) { p.waste = waste }

// IsExposed カードが露出しているか (テスト用の公開版)
func (p *Pyramid) IsExposed(row, col int) bool {
	return p.isExposed(row, col)
}

// AllRemoved 全ピラミッドカードが除去されたか
func (p *Pyramid) AllRemoved() bool {
	for row := range PyramidRowCnt {
		for col := range row + 1 {
			if !p.pyramid[row][col].Removed {
				return false
			}
		}
	}
	return true
}

// --- Private helpers ---

// validatePyramidPos ピラミッド位置の検証
func (p *Pyramid) validatePyramidPos(row, col int) error {
	if row < 0 || row >= PyramidRowCnt {
		return errors.New("invalid row")
	}
	if col < 0 || col > row {
		return errors.New("invalid column")
	}
	return nil
}

// isExposed カードが露出しているか判定
func (p *Pyramid) isExposed(row, col int) bool {
	if p.pyramid[row][col].Removed {
		return false
	}
	// 最下段は常に露出
	if row == PyramidRowCnt-1 {
		return true
	}
	// 両方の子カードが除去されていれば露出
	leftChild := p.pyramid[row+1][col]
	rightChild := p.pyramid[row+1][col+1]
	return leftChild.Removed && rightChild.Removed
}

// getExposedCards 露出しているカードの位置リストを返す
func (p *Pyramid) getExposedCards() [][2]int {
	var result [][2]int
	for row := range PyramidRowCnt {
		for col := range row + 1 {
			if !p.pyramid[row][col].Removed && p.isExposed(row, col) {
				result = append(result, [2]int{row, col})
			}
		}
	}
	return result
}

// checkGameClear ゲームクリア判定
func (p *Pyramid) checkGameClear() {
	if p.AllRemoved() {
		p.phase = PyramidPhaseGameClear
	}
}

// checkStalemate 手詰まり判定
func (p *Pyramid) checkStalemate() {
	if p.phase != PyramidPhasePlaying {
		return
	}
	// ストックにカードがあればまだドローできる
	if len(p.stock) > 0 {
		p.isStalemate = false
		return
	}
	// ヒントがあればスタルメイトではない
	hint := p.GetHint()
	if hint != nil {
		p.isStalemate = false
		return
	}
	p.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (p *Pyramid) takeSnapshot() {
	snap := &pyramidSnapshot{
		phase:       p.phase,
		moveCount:   p.moveCount,
		isStalemate: p.isStalemate,
	}
	// deep copy pyramid
	for i := range PyramidRowCnt {
		snap.pyramid[i] = make([]*PyramidCard, len(p.pyramid[i]))
		for j, pc := range p.pyramid[i] {
			snap.pyramid[i][j] = &PyramidCard{Card: pc.Card, Removed: pc.Removed}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(p.stock))
	copy(snap.stock, p.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(p.waste))
	copy(snap.waste, p.waste)
	p.history = append(p.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (p *Pyramid) restoreSnapshot(snap *pyramidSnapshot) {
	p.pyramid = snap.pyramid
	p.stock = snap.stock
	p.waste = snap.waste
	p.phase = snap.phase
	p.moveCount = snap.moveCount
	p.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (p *Pyramid) appendLog(actionType, detail string, cards []*Card) {
	p.appendLogAt(p.moveCount, 0, actionType, detail, cards)
}

// pyramidJSON is the JSON wire format for Pyramid.
type pyramidJSON struct {
	TrumpCards  *TrumpCards                   `json:"tc"`
	Pyramid     [PyramidRowCnt][]*PyramidCard `json:"py"`
	Stock       []*Card                       `json:"st"`
	Waste       []*Card                       `json:"wa"`
	Phase       PyramidPhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	ActionLog   []*ActionLogEntry             `json:"al"`
	IsStalemate bool                          `json:"sm"`
	History     []*pyramidSnapshot            `json:"hi,omitempty"`
}

// pyramidSnapshotJSON is the wire format for a single undo snapshot.
// pyramidSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// pyramidJSON's short keys to keep the KV payload compact (#1654).
type pyramidSnapshotJSON struct {
	Pyramid     [PyramidRowCnt][]*PyramidCard `json:"py"`
	Stock       []*Card                       `json:"st"`
	Waste       []*Card                       `json:"wa"`
	Phase       PyramidPhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	IsStalemate bool                          `json:"sm"`
}

// MarshalJSON implements json.Marshaler for pyramidSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// Pyramid.MarshalJSON can persist the undo history (#1654).
func (s *pyramidSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(pyramidSnapshotJSON{
		Pyramid:     s.pyramid,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for pyramidSnapshot.
func (s *pyramidSnapshot) UnmarshalJSON(data []byte) error {
	var j pyramidSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > pyramidMaxSliceLen || len(j.Waste) > pyramidMaxSliceLen {
		return fmt.Errorf("pyramid: snapshot array exceeds maximum allowed size")
	}
	for _, row := range j.Pyramid {
		if len(row) > pyramidMaxSliceLen {
			return fmt.Errorf("pyramid: snapshot pyramid row exceeds maximum allowed size")
		}
	}
	s.pyramid = j.Pyramid
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
func (p *Pyramid) MarshalJSON() ([]byte, error) {
	return json.Marshal(pyramidJSON{
		TrumpCards:  p.trumpCards,
		Pyramid:     p.pyramid,
		Stock:       p.stock,
		Waste:       p.waste,
		Phase:       p.phase,
		MoveCount:   p.moveCount,
		ActionLog:   p.actionLog,
		IsStalemate: p.isStalemate,
		History:     p.history,
	})
}

// pyramidMaxSliceLen caps slice sizes during deserialisation.
const pyramidMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (p *Pyramid) UnmarshalJSON(data []byte) error {
	var j pyramidJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > pyramidMaxSliceLen || len(j.Waste) > pyramidMaxSliceLen ||
		len(j.ActionLog) > pyramidMaxSliceLen || len(j.History) > pyramidMaxSliceLen {
		return fmt.Errorf("pyramid: input array exceeds maximum allowed size")
	}
	for _, row := range j.Pyramid {
		if len(row) > pyramidMaxSliceLen {
			return fmt.Errorf("pyramid: pyramid row exceeds maximum allowed size")
		}
	}

	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.pyramid = j.Pyramid
	p.stock = j.Stock
	if p.stock == nil {
		p.stock = make([]*Card, 0)
	}
	p.waste = j.Waste
	if p.waste == nil {
		p.waste = make([]*Card, 0)
	}
	p.phase = j.Phase
	p.moveCount = j.MoveCount
	p.actionLog = j.ActionLog
	if p.actionLog == nil {
		p.actionLog = make([]*ActionLogEntry, 0)
	}
	p.history = j.History
	if p.history == nil {
		p.history = make([]*pyramidSnapshot, 0)
	}
	p.isStalemate = j.IsStalemate
	return nil
}
