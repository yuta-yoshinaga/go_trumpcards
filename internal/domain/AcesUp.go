//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AcesUpPhase エースアップ（四つ葉のクローバー）ゲームフェーズ
type AcesUpPhase int

// AcesUpのフェーズ定数
const (
	// AcesUpPhasePlaying プレイ中
	AcesUpPhasePlaying AcesUpPhase = iota
	// AcesUpPhaseGameClear ゲームクリア
	AcesUpPhaseGameClear
	// AcesUpPhaseGameOver ゲームオーバー
	AcesUpPhaseGameOver
)

// AcesUpColCnt 場札の列数
const AcesUpColCnt = 4

// acesUpAceRank エースを最強として扱う際のランク値（A=14）
const acesUpAceRank = 14

// acesUpMaxSliceLen デシリアライズ時のスライス長上限
const acesUpMaxSliceLen = 1000

// AcesUpHint ヒント
type AcesUpHint struct {
	// Type は "remove" / "move" / "draw" のいずれか
	Type string
	// Col は対象の列番号（draw のときは -1）
	Col int
}

// AcesUp エースアップ（四つ葉のクローバー）ゲームクラス
type AcesUp struct {
	trumpCards *TrumpCards
	columns    [AcesUpColCnt][]*Card
	stock      []*Card
	discard    []*Card
	phase      AcesUpPhase
	moveCount  int
	actionLogBase
	history     []*acesUpSnapshot
	isStalemate bool
}

// acesUpSnapshot アンドゥ用スナップショット
type acesUpSnapshot struct {
	columns     [AcesUpColCnt][]*Card
	stock       []*Card
	discard     []*Card
	phase       AcesUpPhase
	moveCount   int
	isStalemate bool
}

// NewAcesUp コンストラクタ
func NewAcesUp(trumpCards *TrumpCards) *AcesUp {
	return &AcesUp{
		trumpCards: trumpCards,
	}
}

// NewDefaultAcesUp returns AcesUp with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultAcesUp() *AcesUp {
	return NewAcesUp(NewTrumpCards(0))
}

// acesUpRank はエースを最強(14)として扱うランク値を返す。
func acesUpRank(c *Card) int {
	if c.GetValue() == 1 {
		return acesUpAceRank
	}
	return c.GetValue()
}

// Reset ゲームリセット
func (a *AcesUp) Reset() {
	a.trumpCards.Shuffle()
	a.phase = AcesUpPhasePlaying
	a.moveCount = 0
	a.actionLog = nil
	a.history = nil
	a.isStalemate = false
	a.discard = nil

	// 4列を空に初期化し、各列に1枚ずつ配る
	for c := range AcesUpColCnt {
		a.columns[c] = nil
	}
	for c := range AcesUpColCnt {
		if a.trumpCards.GetRemainingCount() == 0 {
			break
		}
		a.columns[c] = append(a.columns[c], a.trumpCards.DrawCard())
	}

	// 残りをストックへ
	a.stock = nil
	for a.trumpCards.GetRemainingCount() > 0 {
		a.stock = append(a.stock, a.trumpCards.DrawCard())
	}

	a.checkStalemate()
}

// Draw ストックから各列に1枚ずつ配る
func (a *AcesUp) Draw() error {
	if a.phase != AcesUpPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(a.stock) == 0 {
		return errors.New("no cards in stock")
	}
	a.takeSnapshot()
	dealt := make([]*Card, 0, AcesUpColCnt)
	for c := range AcesUpColCnt {
		if len(a.stock) == 0 {
			break
		}
		card := a.stock[0]
		a.stock = a.stock[1:]
		a.columns[c] = append(a.columns[c], card)
		dealt = append(dealt, card)
	}
	a.moveCount++
	a.appendLog("draw", "各列にカードを配りました", dealt)
	a.checkGameClear()
	a.checkStalemate()
	return nil
}

// Remove 列の一番上のカードを除去する
func (a *AcesUp) Remove(col int) error {
	if a.phase != AcesUpPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= AcesUpColCnt {
		return errors.New("invalid column")
	}
	if len(a.columns[col]) == 0 {
		return errors.New("no card in column")
	}
	if !a.canRemove(col) {
		return errors.New("card is not removable")
	}
	a.takeSnapshot()
	top := a.popTop(col)
	a.discard = append(a.discard, top)
	a.moveCount++
	a.appendLog("remove", fmt.Sprintf("カード除去: 列%d", col), []*Card{top})
	a.checkGameClear()
	a.checkStalemate()
	return nil
}

// Move 列の一番上のカードを空き列へ移動する
func (a *AcesUp) Move(col int) error {
	if a.phase != AcesUpPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= AcesUpColCnt {
		return errors.New("invalid column")
	}
	if len(a.columns[col]) == 0 {
		return errors.New("no card in column")
	}
	dest := a.firstEmptyColumn()
	if dest < 0 {
		return errors.New("no empty column")
	}
	a.takeSnapshot()
	top := a.popTop(col)
	a.columns[dest] = append(a.columns[dest], top)
	a.moveCount++
	a.appendLog("move", fmt.Sprintf("カード移動: 列%d→列%d", col, dest), []*Card{top})
	a.checkGameClear()
	a.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (a *AcesUp) GiveUp() {
	if a.phase == AcesUpPhasePlaying {
		a.phase = AcesUpPhaseGameOver
		a.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (a *AcesUp) GetHint() *AcesUpHint {
	if a.phase != AcesUpPhasePlaying {
		return nil
	}
	// 除去可能な列を探す
	if col := a.firstRemovableCol(); col >= 0 {
		return &AcesUpHint{Type: "remove", Col: col}
	}
	// ストックから配れる
	if len(a.stock) > 0 {
		return &AcesUpHint{Type: "draw", Col: -1}
	}
	// 有効な移動を探す
	if col := a.firstProductiveMoveCol(); col >= 0 {
		return &AcesUpHint{Type: "move", Col: col}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (a *AcesUp) Undo() error {
	if a.phase != AcesUpPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(a.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (a *AcesUp) CanUndo() bool {
	return len(a.history) > 0 && a.phase == AcesUpPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (a *AcesUp) UndoToEscape() int {
	return undoToEscape(a.isStalemate, a.history, func(s *acesUpSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (a *AcesUp) UndoN(n int) error {
	return undoN(a, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (a *AcesUp) GetPhase() AcesUpPhase { return a.phase }

// SetPhase フェーズ設定 (テスト用)
func (a *AcesUp) SetPhase(phase AcesUpPhase) { a.phase = phase }

// GetMoveCount 移動回数取得
func (a *AcesUp) GetMoveCount() int { return a.moveCount }

// GetStockCount ストック枚数取得
func (a *AcesUp) GetStockCount() int { return len(a.stock) }

// GetDiscardCount 除去済み枚数取得
func (a *AcesUp) GetDiscardCount() int { return len(a.discard) }

// GetDiscardTop 捨て札の一番上（直近に除去した札）を返す。捨て札が空なら nil。
func (a *AcesUp) GetDiscardTop() *Card {
	return discardTop(a.discard)
}

// GetColumns 場札の列を取得
func (a *AcesUp) GetColumns() [AcesUpColCnt][]*Card { return a.columns }

// GetGameEndFlag returns true once the game has left the playing phase.
func (a *AcesUp) GetGameEndFlag() bool { return a.phase != AcesUpPhasePlaying }

// IsStalemate 手詰まり状態取得
func (a *AcesUp) IsStalemate() bool { return a.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (a *AcesUp) SetIsStalemate(v bool) { a.isStalemate = v }

// SetColumns 場札の列設定 (テスト用)
func (a *AcesUp) SetColumns(columns [AcesUpColCnt][]*Card) { a.columns = columns }

// SetStock ストック設定 (テスト用)
func (a *AcesUp) SetStock(stock []*Card) { a.stock = stock }

// CanRemove 指定列の一番上のカードが除去可能か (公開版)
func (a *AcesUp) CanRemove(col int) bool { return a.canRemove(col) }

// CanMove 指定列の一番上のカードが空き列へ移動可能か
func (a *AcesUp) CanMove(col int) bool {
	if col < 0 || col >= AcesUpColCnt || len(a.columns[col]) == 0 {
		return false
	}
	return a.firstEmptyColumn() >= 0
}

// --- Private helpers ---

// topCard 指定列の一番上のカードを返す (空なら nil)
func (a *AcesUp) topCard(col int) *Card {
	if col < 0 || col >= AcesUpColCnt || len(a.columns[col]) == 0 {
		return nil
	}
	return a.columns[col][len(a.columns[col])-1]
}

// popTop 指定列の一番上のカードを取り出して返す。
func (a *AcesUp) popTop(col int) *Card {
	top := a.columns[col][len(a.columns[col])-1]
	a.columns[col] = a.columns[col][:len(a.columns[col])-1]
	return top
}

// canRemove 指定列の一番上のカードが除去可能か判定する。
// 他の列の一番上に「同じスートでランクが上のカード」があれば除去可能。
func (a *AcesUp) canRemove(col int) bool {
	top := a.topCard(col)
	if top == nil {
		return false
	}
	for c := range AcesUpColCnt {
		if c == col {
			continue
		}
		other := a.topCard(c)
		if other == nil {
			continue
		}
		if other.GetDesign() == top.GetDesign() && acesUpRank(other) > acesUpRank(top) {
			return true
		}
	}
	return false
}

// firstRemovableCol 除去可能な最初の列を返す (-1 = なし)
func (a *AcesUp) firstRemovableCol() int {
	for c := range AcesUpColCnt {
		if a.canRemove(c) {
			return c
		}
	}
	return -1
}

// firstEmptyColumn 最初の空き列を返す (-1 = なし)
func (a *AcesUp) firstEmptyColumn() int {
	for c := range AcesUpColCnt {
		if len(a.columns[c]) == 0 {
			return c
		}
	}
	return -1
}

// firstProductiveMoveCol 有効な移動が可能な最初の列を返す (-1 = なし)。
// 2枚以上ある列の一番上を空き列へ動かすと下のカードが露出する。空き列は
// すべて等価なので、各列について「最初の空き列へ動かした盤面」から除去可能
// な状態へ到達できるかを深さ優先で探索する。空き列が複数あれば連続して移動
// できるため、1手だけ先読みする実装では「偽の手詰まり」を起こす（#2092 review）。
func (a *AcesUp) firstProductiveMoveCol() int {
	if a.firstEmptyColumn() < 0 {
		return -1
	}
	for c := range AcesUpColCnt {
		if len(a.columns[c]) < 2 {
			// 1枚だけの列を動かしても露出するカードがなく無意味。
			continue
		}
		next := cloneColumns(a.columns)
		e := firstEmptyColumnOf(next)
		top := next[c][len(next[c])-1]
		next[c] = next[c][:len(next[c])-1]
		next[e] = append(next[e], top)
		if acesUpReachableRemoval(next) {
			return c
		}
	}
	return -1
}

// cloneColumns は列スライス配列のディープコピーを返す。
func cloneColumns(cols [AcesUpColCnt][]*Card) [AcesUpColCnt][]*Card {
	var out [AcesUpColCnt][]*Card
	for c := range AcesUpColCnt {
		out[c] = cloneCards(cols[c])
	}
	return out
}

// firstEmptyColumnOf は任意の盤面に対する最初の空き列を返す (-1 = なし)。
func firstEmptyColumnOf(cols [AcesUpColCnt][]*Card) int {
	for c := range AcesUpColCnt {
		if len(cols[c]) == 0 {
			return c
		}
	}
	return -1
}

// acesUpHasRemovable は盤面に除去可能な一番上カードが存在するか判定する。
// 2枚の一番上カードが同じスートなら、ランクの低い方を除去できる。
func acesUpHasRemovable(cols [AcesUpColCnt][]*Card) bool {
	for i := range AcesUpColCnt {
		ti := topOf(cols[i])
		if ti == nil {
			continue
		}
		for j := range AcesUpColCnt {
			if i == j {
				continue
			}
			tj := topOf(cols[j])
			if tj == nil {
				continue
			}
			if ti.GetDesign() == tj.GetDesign() && acesUpRank(tj) > acesUpRank(ti) {
				return true
			}
		}
	}
	return false
}

// topOf は列の一番上のカードを返す (空なら nil)。
func topOf(col []*Card) *Card {
	if len(col) == 0 {
		return nil
	}
	return col[len(col)-1]
}

// acesUpReachableRemoval は「移動のみ」で除去可能な状態へ到達できるか判定する。
// 各再帰では2枚以上ある列の一番上を最初の空き列へ動かす。移動するたびに空き列
// が1つ減るため、再帰の深さは空き列数で上限が決まり（盤面は4列なので状態数は
// ごく小さい）必ず停止する。
func acesUpReachableRemoval(cols [AcesUpColCnt][]*Card) bool {
	if acesUpHasRemovable(cols) {
		return true
	}
	e := firstEmptyColumnOf(cols)
	if e < 0 {
		return false
	}
	for c := range AcesUpColCnt {
		if len(cols[c]) < 2 {
			continue
		}
		next := cloneColumns(cols)
		top := next[c][len(next[c])-1]
		next[c] = next[c][:len(next[c])-1]
		next[e] = append(next[e], top)
		if acesUpReachableRemoval(next) {
			return true
		}
	}
	return false
}

// isWon 4枚のエースだけが残っているか判定する。
func (a *AcesUp) isWon() bool {
	if len(a.stock) != 0 {
		return false
	}
	for c := range AcesUpColCnt {
		if len(a.columns[c]) != 1 {
			return false
		}
		if a.columns[c][0].GetValue() != 1 {
			return false
		}
	}
	return true
}

// checkGameClear ゲームクリア判定
func (a *AcesUp) checkGameClear() {
	if a.isWon() {
		a.phase = AcesUpPhaseGameClear
	}
}

// checkStalemate 手詰まり判定
func (a *AcesUp) checkStalemate() {
	if a.phase != AcesUpPhasePlaying {
		return
	}
	a.isStalemate = a.GetHint() == nil
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (a *AcesUp) takeSnapshot() {
	snap := &acesUpSnapshot{
		stock:       cloneCards(a.stock),
		discard:     cloneCards(a.discard),
		phase:       a.phase,
		moveCount:   a.moveCount,
		isStalemate: a.isStalemate,
	}
	for c := range AcesUpColCnt {
		snap.columns[c] = cloneCards(a.columns[c])
	}
	a.history = appendSnapshot(a.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (a *AcesUp) restoreSnapshot(snap *acesUpSnapshot) {
	a.columns = snap.columns
	a.stock = snap.stock
	a.discard = snap.discard
	a.phase = snap.phase
	a.moveCount = snap.moveCount
	a.isStalemate = snap.isStalemate
}

// cloneCards はカードスライスの浅いコピーを返す。
func cloneCards(src []*Card) []*Card {
	dst := make([]*Card, len(src))
	copy(dst, src)
	return dst
}

// appendLog 棋譜エントリを追加
func (a *AcesUp) appendLog(actionType, detail string, cards []*Card) {
	a.appendLogAt(a.moveCount, 0, actionType, detail, cards)
}

// acesUpJSON is the JSON wire format for AcesUp.
type acesUpJSON struct {
	TrumpCards  *TrumpCards           `json:"tc"`
	Columns     [AcesUpColCnt][]*Card `json:"co"`
	Stock       []*Card               `json:"st"`
	Discard     []*Card               `json:"di"`
	Phase       AcesUpPhase           `json:"ps"`
	MoveCount   int                   `json:"mc"`
	ActionLog   []*ActionLogEntry     `json:"al"`
	IsStalemate bool                  `json:"sm"`
	History     []*acesUpSnapshot     `json:"hi,omitempty"`
}

// acesUpSnapshotJSON is the wire format for a single undo snapshot.
// acesUpSnapshot uses unexported fields, so we project to/from this shape
// with explicit Marshal/Unmarshal methods. Field names match acesUpJSON's
// short keys to keep the KV payload compact (#1654).
type acesUpSnapshotJSON struct {
	Columns     [AcesUpColCnt][]*Card `json:"co"`
	Stock       []*Card               `json:"st"`
	Discard     []*Card               `json:"di"`
	Phase       AcesUpPhase           `json:"ps"`
	MoveCount   int                   `json:"mc"`
	IsStalemate bool                  `json:"sm"`
}

// MarshalJSON implements json.Marshaler for acesUpSnapshot.
func (s *acesUpSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(acesUpSnapshotJSON{
		Columns:     s.columns,
		Stock:       s.stock,
		Discard:     s.discard,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for acesUpSnapshot.
func (s *acesUpSnapshot) UnmarshalJSON(data []byte) error {
	var j acesUpSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > acesUpMaxSliceLen || len(j.Discard) > acesUpMaxSliceLen {
		return fmt.Errorf("acesup: snapshot array exceeds maximum allowed size")
	}
	for c := range AcesUpColCnt {
		if len(j.Columns[c]) > acesUpMaxSliceLen {
			return fmt.Errorf("acesup: snapshot column exceeds maximum allowed size")
		}
	}
	s.columns = j.Columns
	s.stock = nilToEmptyCards(j.Stock)
	s.discard = nilToEmptyCards(j.Discard)
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (a *AcesUp) MarshalJSON() ([]byte, error) {
	return json.Marshal(acesUpJSON{
		TrumpCards:  a.trumpCards,
		Columns:     a.columns,
		Stock:       a.stock,
		Discard:     a.discard,
		Phase:       a.phase,
		MoveCount:   a.moveCount,
		ActionLog:   a.actionLog,
		IsStalemate: a.isStalemate,
		History:     a.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *AcesUp) UnmarshalJSON(data []byte) error {
	var j acesUpJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > acesUpMaxSliceLen || len(j.Discard) > acesUpMaxSliceLen ||
		len(j.ActionLog) > acesUpMaxSliceLen || len(j.History) > acesUpMaxSliceLen {
		return fmt.Errorf("acesup: input array exceeds maximum allowed size")
	}
	for c := range AcesUpColCnt {
		if len(j.Columns[c]) > acesUpMaxSliceLen {
			return fmt.Errorf("acesup: input column exceeds maximum allowed size")
		}
	}

	a.trumpCards = j.TrumpCards
	if a.trumpCards == nil {
		a.trumpCards = NewTrumpCards(0)
	}
	a.columns = j.Columns
	for c := range AcesUpColCnt {
		a.columns[c] = nilToEmptyCards(a.columns[c])
	}
	a.stock = nilToEmptyCards(j.Stock)
	a.discard = nilToEmptyCards(j.Discard)
	a.phase = j.Phase
	a.moveCount = j.MoveCount
	a.actionLog = j.ActionLog
	if a.actionLog == nil {
		a.actionLog = make([]*ActionLogEntry, 0)
	}
	a.history = j.History
	if a.history == nil {
		a.history = make([]*acesUpSnapshot, 0)
	}
	a.isStalemate = j.IsStalemate
	return nil
}

// nilToEmptyCards は nil スライスを空スライスに正規化する。
func nilToEmptyCards(src []*Card) []*Card {
	if src == nil {
		return make([]*Card, 0)
	}
	return src
}
