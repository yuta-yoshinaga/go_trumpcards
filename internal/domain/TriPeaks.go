//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TriPeaksPhase トリピークスゲームフェーズ
type TriPeaksPhase int

// TriPeaksのフェーズ定数
const (
	// TriPeaksPhasePlaying プレイ中
	TriPeaksPhasePlaying TriPeaksPhase = iota
	// TriPeaksPhaseGameClear ゲームクリア
	TriPeaksPhaseGameClear
	// TriPeaksPhaseGameOver ゲームオーバー
	TriPeaksPhaseGameOver
)

// TriPeaksRowCnt トリピークスの段数
const TriPeaksRowCnt = 4

// TriPeaksColCnt トリピークスの最大列数
const TriPeaksColCnt = 10

// TriPeaksTableauCnt タブロー上のカード枚数 (3+6+9+10)
const TriPeaksTableauCnt = 28

// TriPeaksCard タブロー上のカード
type TriPeaksCard struct {
	Card    *Card `json:"c"`
	Removed bool  `json:"r"`
}

// TriPeaksHint ヒント
type TriPeaksHint struct {
	Type string // "remove" or "draw"
	Row  int
	Col  int
}

// スコアの配点。**フロント (frontend/src/utils/tripeaksScore.ts) と同じ式**で、
// これまでは得点計算そのものがフロントにしか無く、CUI からは触れなかった (#5511)。
const (
	// TriPeaksPointsPerChain は連鎖 n 手目の得点係数。n 手目は n × これ。
	TriPeaksPointsPerChain = 100
	// TriPeaksPeakBonus は 1 つの山を出し切ったときの一時金。
	TriPeaksPeakBonus = 500
)

// TriPeaks トリピークスソリティアゲームクラス
type TriPeaks struct {
	trumpCards *TrumpCards
	layout     [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard
	stock      []*Card
	waste      []*Card
	phase      TriPeaksPhase
	moveCount  int
	actionLogBase
	history     []*triPeaksSnapshot
	isStalemate bool
	// score は累計点、chain は途切れていない連続除去の長さ。
	// **山札を引く / アンドゥで chain だけが 0 に戻り、score は残る。**
	score int
	chain int
}

// triPeaksSnapshot アンドゥ用スナップショット
type triPeaksSnapshot struct {
	layout      [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard
	stock       []*Card
	waste       []*Card
	phase       TriPeaksPhase
	moveCount   int
	isStalemate bool
}

// triPeaksValidPos はタブロー上の有効なポジションを示す。
// Row 0: cols 0,3,6 (3 peak tops)
// Row 1: cols 0,1,3,4,6,7
// Row 2: cols 0,1,2,3,4,5,6,7,8
// Row 3: cols 0-9 (all)
var triPeaksValidPos [TriPeaksRowCnt][TriPeaksColCnt]bool

func init() {
	// Row 0: 3 peaks
	triPeaksValidPos[0][0] = true
	triPeaksValidPos[0][3] = true
	triPeaksValidPos[0][6] = true
	// Row 1
	for _, c := range []int{0, 1, 3, 4, 6, 7} {
		triPeaksValidPos[1][c] = true
	}
	// Row 2
	for c := 0; c <= 8; c++ {
		triPeaksValidPos[2][c] = true
	}
	// Row 3
	for c := 0; c <= 9; c++ {
		triPeaksValidPos[3][c] = true
	}
}

// triPeaksChildren は(row,col)のカードの子カード位置を返す。
// 子カードは (row+1, col) と (row+1, col+1)。
func triPeaksChildren(row, col int) [][2]int {
	if row >= TriPeaksRowCnt-1 {
		return nil
	}
	var children [][2]int
	for _, dc := range []int{0, 1} {
		nc := col + dc
		if nc < TriPeaksColCnt && triPeaksValidPos[row+1][nc] {
			children = append(children, [2]int{row + 1, nc})
		}
	}
	return children
}

// NewTriPeaks コンストラクタ
func NewTriPeaks(trumpCards *TrumpCards) *TriPeaks {
	return &TriPeaks{
		trumpCards: trumpCards,
	}
}

// NewDefaultTriPeaks returns TriPeaks with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultTriPeaks() *TriPeaks {
	return NewTriPeaks(NewTrumpCards(0))
}

// Reset ゲームリセット
func (t *TriPeaks) Reset() {
	t.trumpCards.Shuffle()
	t.phase = TriPeaksPhasePlaying
	t.moveCount = 0
	t.score, t.chain = 0, 0
	t.actionLog = nil
	t.history = nil
	t.isStalemate = false

	// レイアウトをクリア
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			t.layout[r][c] = nil
		}
	}

	// タブロー配置: row 0→3 の順に有効位置へカードを配る
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if triPeaksValidPos[r][c] {
				card := t.trumpCards.DrawCard()
				t.layout[r][c] = &TriPeaksCard{
					Card:    card,
					Removed: false,
				}
			}
		}
	}

	// 残りをストックへ（最後の1枚はウェイストへ）
	t.stock = nil
	t.waste = nil
	for t.trumpCards.GetRemainingCount() > 0 {
		card := t.trumpCards.DrawCard()
		t.stock = append(t.stock, card)
	}

	// ストックから1枚をウェイストに置く（ゲーム開始の基準カード）
	if len(t.stock) > 0 {
		wasteCard := t.stock[len(t.stock)-1]
		t.stock = t.stock[:len(t.stock)-1]
		t.waste = append(t.waste, wasteCard)
	}
}

// Draw ストックからウェイストにカードを引く
func (t *TriPeaks) Draw() error {
	if t.phase != TriPeaksPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(t.stock) == 0 {
		return errors.New("no cards in stock")
	}
	t.takeSnapshot()
	card := t.stock[len(t.stock)-1]
	t.stock = t.stock[:len(t.stock)-1]
	t.waste = append(t.waste, card)
	t.moveCount++
	// **引くと連鎖は切れるが、稼いだ点は残る。**
	t.chain = 0
	t.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	t.checkStalemate()
	return nil
}

// Remove タブローのカードを除去
func (t *TriPeaks) Remove(row, col int) error {
	if t.phase != TriPeaksPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if err := t.validatePos(row, col); err != nil {
		return err
	}
	tc := t.layout[row][col]
	if tc == nil {
		return errors.New("no card at position")
	}
	if tc.Removed {
		return errors.New("card is already removed")
	}
	if !t.isExposed(row, col) {
		return errors.New("card is not exposed")
	}
	if len(t.waste) == 0 {
		return errors.New("waste is empty")
	}
	wasteTop := t.waste[len(t.waste)-1]
	if !t.isAdjacentRank(tc.Card, wasteTop) {
		return errors.New("card is not adjacent to waste top")
	}
	t.takeSnapshot()
	peaksBefore := t.peaksCleared()
	tc.Removed = true
	// 除去したカードをウェイストの上に置く
	t.waste = append(t.waste, tc.Card)
	t.moveCount++
	t.chain++
	t.score += t.chain*TriPeaksPointsPerChain +
		(t.peaksCleared()-peaksBefore)*TriPeaksPeakBonus
	t.appendLog("remove", fmt.Sprintf("カード除去: (%d,%d)", row, col),
		[]*Card{tc.Card})
	t.checkGameClear()
	t.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (t *TriPeaks) GiveUp() {
	if t.phase == TriPeaksPhasePlaying {
		t.phase = TriPeaksPhaseGameOver
		t.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (t *TriPeaks) GetHint() *TriPeaksHint {
	if t.phase != TriPeaksPhasePlaying {
		return nil
	}
	if len(t.waste) == 0 {
		return nil
	}
	// 露出カードで除去可能なものを探す
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if !triPeaksValidPos[r][c] {
				continue
			}
			tc := t.layout[r][c]
			if tc == nil || tc.Removed {
				continue
			}
			if !t.isExposed(r, c) {
				continue
			}
			if t.isAdjacentRank(tc.Card, t.waste[len(t.waste)-1]) {
				return &TriPeaksHint{
					Type: "remove",
					Row:  r,
					Col:  c,
				}
			}
		}
	}
	// カード除去不可、ストックからドロー可能な場合
	if len(t.stock) > 0 {
		return &TriPeaksHint{
			Type: "draw",
			Row:  -1,
			Col:  -1,
		}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (t *TriPeaks) Undo() error {
	if t.phase != TriPeaksPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(t.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := t.history[len(t.history)-1]
	t.history = t.history[:len(t.history)-1]
	t.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (t *TriPeaks) CanUndo() bool {
	return len(t.history) > 0 && t.phase == TriPeaksPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (t *TriPeaks) UndoToEscape() int {
	return undoToEscape(t.isStalemate, t.history, func(s *triPeaksSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (t *TriPeaks) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := t.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (t *TriPeaks) GetPhase() TriPeaksPhase { return t.phase }

// SetPhase フェーズ設定 (テスト用)
func (t *TriPeaks) SetPhase(phase TriPeaksPhase) { t.phase = phase }

// GetMoveCount 移動回数取得
func (t *TriPeaks) GetMoveCount() int { return t.moveCount }

// GetScore は累計点を返す。
func (t *TriPeaks) GetScore() int { return t.score }

// GetCombo は途切れていない連続除去の長さを返す。山札を引くかアンドゥで 0 に戻る。
func (t *TriPeaks) GetCombo() int { return t.chain }

// triPeaksPeakOfColumn は列がどの山に属するかを返す (0=左, 1=中央, 2=右)。
func triPeaksPeakOfColumn(col int) int {
	switch {
	case col < 3:
		return 0
	case col < 6:
		return 1
	default:
		return 2
	}
}

// peaksCleared は出し切った山の数を返す。
func (t *TriPeaks) peaksCleared() int {
	var remaining [3]int
	for _, row := range t.layout {
		for col, tc := range row {
			if tc != nil && tc.Card != nil && !tc.Removed {
				remaining[triPeaksPeakOfColumn(col)]++
			}
		}
	}
	n := 0
	for _, r := range remaining {
		if r == 0 {
			n++
		}
	}
	return n
}

// GetStockCount ストック枚数取得
func (t *TriPeaks) GetStockCount() int { return len(t.stock) }

// GetWaste ウェイスト取得
func (t *TriPeaks) GetWaste() []*Card { return t.waste }

// GetLayout レイアウト取得
func (t *TriPeaks) GetLayout() [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard {
	return t.layout
}

// GetGameEndFlag returns true once the game has left the playing phase.
func (t *TriPeaks) GetGameEndFlag() bool { return t.phase != TriPeaksPhasePlaying }

// IsStalemate 手詰まり状態取得
func (t *TriPeaks) IsStalemate() bool { return t.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (t *TriPeaks) SetIsStalemate(v bool) { t.isStalemate = v }

// SetLayout レイアウト設定 (テスト用)
func (t *TriPeaks) SetLayout(layout [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard) {
	t.layout = layout
}

// SetStock ストック設定 (テスト用)
func (t *TriPeaks) SetStock(stock []*Card) { t.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (t *TriPeaks) SetWaste(waste []*Card) { t.waste = waste }

// IsExposed カードが露出しているか (テスト用の公開版)
func (t *TriPeaks) IsExposed(row, col int) bool {
	return t.isExposed(row, col)
}

// AllRemoved 全タブローカードが除去されたか
func (t *TriPeaks) AllRemoved() bool {
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if triPeaksValidPos[r][c] && t.layout[r][c] != nil && !t.layout[r][c].Removed {
				return false
			}
		}
	}
	return true
}

// --- Private helpers ---

// validatePos タブロー位置の検証
func (t *TriPeaks) validatePos(row, col int) error {
	if row < 0 || row >= TriPeaksRowCnt {
		return errors.New("invalid row")
	}
	if col < 0 || col >= TriPeaksColCnt {
		return errors.New("invalid column")
	}
	if !triPeaksValidPos[row][col] {
		return errors.New("invalid position")
	}
	return nil
}

// isExposed カードが露出しているか判定
func (t *TriPeaks) isExposed(row, col int) bool {
	if !triPeaksValidPos[row][col] {
		return false
	}
	tc := t.layout[row][col]
	if tc == nil || tc.Removed {
		return false
	}
	// 最下段は常に露出
	if row == TriPeaksRowCnt-1 {
		return true
	}
	// 全ての子カードが除去されていれば露出
	children := triPeaksChildren(row, col)
	for _, ch := range children {
		child := t.layout[ch[0]][ch[1]]
		if child != nil && !child.Removed {
			return false
		}
	}
	return true
}

// isAdjacentRank 2枚のカードのランクが±1か判定 (K-Aラップあり)
func (t *TriPeaks) isAdjacentRank(card1, card2 *Card) bool {
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
func (t *TriPeaks) checkGameClear() {
	if t.AllRemoved() {
		t.phase = TriPeaksPhaseGameClear
	}
}

// checkStalemate 手詰まり判定
func (t *TriPeaks) checkStalemate() {
	if t.phase != TriPeaksPhasePlaying {
		return
	}
	// ストックにカードがあればまだドローできる
	if len(t.stock) > 0 {
		t.isStalemate = false
		return
	}
	// ヒントがあればスタルメイトではない
	hint := t.GetHint()
	if hint != nil {
		t.isStalemate = false
		return
	}
	t.isStalemate = true
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (t *TriPeaks) takeSnapshot() {
	snap := &triPeaksSnapshot{
		phase:       t.phase,
		moveCount:   t.moveCount,
		isStalemate: t.isStalemate,
	}
	// deep copy layout
	for r := range TriPeaksRowCnt {
		for c := range TriPeaksColCnt {
			if t.layout[r][c] != nil {
				snap.layout[r][c] = &TriPeaksCard{Card: t.layout[r][c].Card, Removed: t.layout[r][c].Removed}
			}
		}
	}
	// deep copy stock
	snap.stock = make([]*Card, len(t.stock))
	copy(snap.stock, t.stock)
	// deep copy waste
	snap.waste = make([]*Card, len(t.waste))
	copy(snap.waste, t.waste)
	t.history = appendSnapshot(t.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (t *TriPeaks) restoreSnapshot(snap *triPeaksSnapshot) {
	t.layout = snap.layout
	t.stock = snap.stock
	t.waste = snap.waste
	t.phase = snap.phase
	t.moveCount = snap.moveCount
	t.isStalemate = snap.isStalemate
	// **アンドゥは連鎖だけを切る。稼いだ点は戻さない。**
	// フロントの applyTriPeaksScore も moveCount が減ったら chain のみ 0 にする。
	// ここで score を巻き戻すと、アンドゥで点を稼ぎ直す抜け道ができる。
	t.chain = 0
}

// appendLog 棋譜エントリを追加
func (t *TriPeaks) appendLog(actionType, detail string, cards []*Card) {
	t.appendLogAt(t.moveCount, 0, actionType, detail, cards)
}

// triPeaksJSON is the JSON wire format for TriPeaks.
type triPeaksJSON struct {
	TrumpCards  *TrumpCards                                   `json:"tc"`
	Layout      [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard `json:"ly"`
	Stock       []*Card                                       `json:"st"`
	Waste       []*Card                                       `json:"wa"`
	Phase       TriPeaksPhase                                 `json:"ps"`
	MoveCount   int                                           `json:"mc"`
	ActionLog   []*ActionLogEntry                             `json:"al"`
	IsStalemate bool                                          `json:"sm"`
	History     []*triPeaksSnapshot                           `json:"hi,omitempty"`
	// Score / Chain を載せないと、KV に保存して復元した時点で得点が 0 に戻る。
	Score int `json:"sc"`
	Chain int `json:"ch"`
}

// triPeaksSnapshotJSON is the wire format for a single undo snapshot.
// triPeaksSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// triPeaksJSON's short keys to keep the KV payload compact (#1654).
type triPeaksSnapshotJSON struct {
	Layout      [TriPeaksRowCnt][TriPeaksColCnt]*TriPeaksCard `json:"ly"`
	Stock       []*Card                                       `json:"st"`
	Waste       []*Card                                       `json:"wa"`
	Phase       TriPeaksPhase                                 `json:"ps"`
	MoveCount   int                                           `json:"mc"`
	IsStalemate bool                                          `json:"sm"`
}

// MarshalJSON implements json.Marshaler for triPeaksSnapshot, projecting
// the unexported fields onto an exported wire shape so that
// TriPeaks.MarshalJSON can persist the undo history (#1654).
func (s *triPeaksSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(triPeaksSnapshotJSON{
		Layout:      s.layout,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for triPeaksSnapshot.
func (s *triPeaksSnapshot) UnmarshalJSON(data []byte) error {
	var j triPeaksSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > triPeaksMaxSliceLen || len(j.Waste) > triPeaksMaxSliceLen {
		return fmt.Errorf("tripeaks: snapshot array exceeds maximum allowed size")
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
func (t *TriPeaks) MarshalJSON() ([]byte, error) {
	return json.Marshal(triPeaksJSON{
		TrumpCards:  t.trumpCards,
		Layout:      t.layout,
		Stock:       t.stock,
		Waste:       t.waste,
		Phase:       t.phase,
		MoveCount:   t.moveCount,
		ActionLog:   t.actionLog,
		IsStalemate: t.isStalemate,
		History:     t.history,
		Score:       t.score,
		Chain:       t.chain,
	})
}

// triPeaksMaxSliceLen caps slice sizes during deserialisation.
const triPeaksMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (t *TriPeaks) UnmarshalJSON(data []byte) error {
	var j triPeaksJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > triPeaksMaxSliceLen || len(j.Waste) > triPeaksMaxSliceLen ||
		len(j.ActionLog) > triPeaksMaxSliceLen || len(j.History) > triPeaksMaxSliceLen {
		return fmt.Errorf("tripeaks: input array exceeds maximum allowed size")
	}

	t.trumpCards = j.TrumpCards
	if t.trumpCards == nil {
		t.trumpCards = NewTrumpCards(0)
	}
	t.layout = j.Layout
	t.stock = j.Stock
	if t.stock == nil {
		t.stock = make([]*Card, 0)
	}
	t.waste = j.Waste
	if t.waste == nil {
		t.waste = make([]*Card, 0)
	}
	t.phase = j.Phase
	t.moveCount = j.MoveCount
	t.actionLog = j.ActionLog
	if t.actionLog == nil {
		t.actionLog = make([]*ActionLogEntry, 0)
	}
	t.history = j.History
	if t.history == nil {
		t.history = make([]*triPeaksSnapshot, 0)
	}
	t.isStalemate = j.IsStalemate
	if j.Score < 0 || j.Chain < 0 {
		return fmt.Errorf("invalid score/chain: %d/%d", j.Score, j.Chain)
	}
	t.score, t.chain = j.Score, j.Chain
	return nil
}
