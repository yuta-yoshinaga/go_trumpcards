//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// OsmosisPhase オズモシスゲームフェーズ
type OsmosisPhase int

// Osmosisのフェーズ定数
const (
	// OsmosisPhasePlaying プレイ中
	OsmosisPhasePlaying OsmosisPhase = iota
	// OsmosisPhaseGameClear ゲームクリア
	OsmosisPhaseGameClear
	// OsmosisPhaseGameOver ゲームオーバー
	OsmosisPhaseGameOver
)

// OsmosisFoundationCnt ファンデーションの段数（縦4段）
const OsmosisFoundationCnt = 4

// OsmosisReserveCnt リザーブの列数（4列）
const OsmosisReserveCnt = 4

// OsmosisReservePileSize リザーブ1列あたりの枚数
const OsmosisReservePileSize = 4

// OsmosisDrawCount ウェイストへの1回のドロー枚数
const OsmosisDrawCount = 1

// OsmosisHint ヒント
type OsmosisHint struct {
	// FromZone は移動元ゾーン ("waste" / "reserve")。
	FromZone string
	// FromCol はリザーブ列インデックス（ウェイスト元の場合は -1）。
	FromCol int
	// ToCol は移動先ファンデーション段インデックス。
	ToCol int
}

// Osmosis オズモシス（浸透）ソリティアゲームクラス
type Osmosis struct {
	trumpCards *TrumpCards
	reserve    [OsmosisReserveCnt][]*Card
	stock      []*Card
	waste      []*Card
	foundation [OsmosisFoundationCnt][]*Card
	baseRank   int
	phase      OsmosisPhase
	moveCount  int
	actionLogBase
	history []*osmosisSnapshot
}

// osmosisSnapshot アンドゥ用スナップショット
type osmosisSnapshot struct {
	reserve    [OsmosisReserveCnt][]*Card
	stock      []*Card
	waste      []*Card
	foundation [OsmosisFoundationCnt][]*Card
	phase      OsmosisPhase
	moveCount  int
}

// NewOsmosis コンストラクタ
func NewOsmosis(trumpCards *TrumpCards) *Osmosis {
	return &Osmosis{trumpCards: trumpCards}
}

// NewDefaultOsmosis returns Osmosis with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultOsmosis() *Osmosis {
	return NewOsmosis(NewTrumpCards(0))
}

// Reset ゲームリセット
func (o *Osmosis) Reset() {
	o.trumpCards.Shuffle()
	o.phase = OsmosisPhasePlaying
	o.moveCount = 0
	o.actionLog = nil
	o.history = nil

	// リザーブに4列、各4枚
	for i := 0; i < OsmosisReserveCnt; i++ {
		o.reserve[i] = make([]*Card, 0, OsmosisReservePileSize)
		for j := 0; j < OsmosisReservePileSize; j++ {
			o.reserve[i] = append(o.reserve[i], o.trumpCards.DrawCard())
		}
	}

	// ファンデーション初期化
	for i := 0; i < OsmosisFoundationCnt; i++ {
		o.foundation[i] = nil
	}

	// ベースランク決定: 次の1枚を1段目のファンデーションに置く
	base := o.trumpCards.DrawCard()
	o.baseRank = base.GetValue()
	o.foundation[0] = []*Card{base}

	// 残りをストックへ
	o.stock = nil
	o.waste = nil
	for o.trumpCards.GetRemainingCount() > 0 {
		o.stock = append(o.stock, o.trumpCards.DrawCard())
	}
}

// Draw ストックからウェイストにカードを引く
func (o *Osmosis) Draw() error {
	if o.phase != OsmosisPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(o.stock) == 0 {
		if len(o.waste) == 0 {
			return errors.New("no cards in stock or waste")
		}
		o.takeSnapshot()
		for i := len(o.waste) - 1; i >= 0; i-- {
			o.stock = append(o.stock, o.waste[i])
		}
		o.waste = nil
		o.appendLog("recycle", "ウェイストをストックに戻しました", nil)
		return nil
	}
	o.takeSnapshot()
	count := OsmosisDrawCount
	if count > len(o.stock) {
		count = len(o.stock)
	}
	drawn := make([]*Card, 0, count)
	for i := 0; i < count; i++ {
		card := o.stock[len(o.stock)-1]
		o.stock = o.stock[:len(o.stock)-1]
		o.waste = append(o.waste, card)
		drawn = append(drawn, card)
	}
	o.moveCount++
	o.appendLog("draw", "ストックからカードを引きました", drawn)
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションに移動
func (o *Osmosis) MoveWasteToFoundation(fIdx int) error {
	if o.phase != OsmosisPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fIdx < 0 || fIdx >= OsmosisFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(o.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := o.waste[len(o.waste)-1]
	if !o.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	o.takeSnapshot()
	o.waste = o.waste[:len(o.waste)-1]
	o.foundation[fIdx] = append(o.foundation[fIdx], card)
	o.moveCount++
	o.appendLog("move", fmt.Sprintf("ウェイスト→ファンデーション%d段目", fIdx), []*Card{card})
	o.checkGameClear()
	return nil
}

// MoveReserveToFoundation リザーブからファンデーションに移動
func (o *Osmosis) MoveReserveToFoundation(rIdx, fIdx int) error {
	if o.phase != OsmosisPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if rIdx < 0 || rIdx >= OsmosisReserveCnt {
		return errors.New("invalid reserve index")
	}
	if fIdx < 0 || fIdx >= OsmosisFoundationCnt {
		return errors.New("invalid foundation index")
	}
	pile := o.reserve[rIdx]
	if len(pile) == 0 {
		return errors.New("reserve pile is empty")
	}
	card := pile[len(pile)-1]
	if !o.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	o.takeSnapshot()
	o.reserve[rIdx] = pile[:len(pile)-1]
	o.foundation[fIdx] = append(o.foundation[fIdx], card)
	o.moveCount++
	o.appendLog("move", fmt.Sprintf("リザーブ列%d→ファンデーション%d段目", rIdx, fIdx), []*Card{card})
	o.checkGameClear()
	return nil
}

// GiveUp ギブアップ
func (o *Osmosis) GiveUp() {
	if o.phase == OsmosisPhasePlaying {
		o.phase = OsmosisPhaseGameOver
		o.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (o *Osmosis) GetHint() *OsmosisHint {
	if o.phase != OsmosisPhasePlaying {
		return nil
	}
	// 優先度1: リザーブからファンデーションへ
	for rIdx := 0; rIdx < OsmosisReserveCnt; rIdx++ {
		pile := o.reserve[rIdx]
		if len(pile) == 0 {
			continue
		}
		if fIdx := o.findFoundationFor(pile[len(pile)-1]); fIdx >= 0 {
			return &OsmosisHint{FromZone: "reserve", FromCol: rIdx, ToCol: fIdx}
		}
	}
	// 優先度2: ウェイストからファンデーションへ
	if len(o.waste) > 0 {
		if fIdx := o.findFoundationFor(o.waste[len(o.waste)-1]); fIdx >= 0 {
			return &OsmosisHint{FromZone: "waste", FromCol: -1, ToCol: fIdx}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（移動可能なカードをファンデーションへ繰り返し移動）
func (o *Osmosis) AutoComplete() error {
	if o.phase != OsmosisPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	o.takeSnapshot()
	moved := false
	for {
		stepMoved := false
		// リザーブの各列トップ
		for rIdx := 0; rIdx < OsmosisReserveCnt; rIdx++ {
			pile := o.reserve[rIdx]
			if len(pile) == 0 {
				continue
			}
			card := pile[len(pile)-1]
			if fIdx := o.findFoundationFor(card); fIdx >= 0 {
				o.reserve[rIdx] = pile[:len(pile)-1]
				o.foundation[fIdx] = append(o.foundation[fIdx], card)
				o.moveCount++
				stepMoved = true
			}
		}
		// ウェイストトップ
		if len(o.waste) > 0 {
			card := o.waste[len(o.waste)-1]
			if fIdx := o.findFoundationFor(card); fIdx >= 0 {
				o.waste = o.waste[:len(o.waste)-1]
				o.foundation[fIdx] = append(o.foundation[fIdx], card)
				o.moveCount++
				stepMoved = true
			}
		}
		if !stepMoved {
			break
		}
		moved = true
	}
	if !moved {
		// 何も動かなければスナップショットを取り消す
		o.history = o.history[:len(o.history)-1]
		return errors.New("no card can be auto-completed")
	}
	o.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	o.checkGameClear()
	return nil
}

// --- Getters / Setters ---

// IsStalemate は、どのカードもファンデーションへ送れなくなった手詰まりを報告する。
//
// **山札の残枚数は条件にならない。**Draw はウェイストをストックへ戻して循環させる
// ので、山札の中身は何周でも見に行ける。逆に、リザーブのトップ・ストック・ウェイスト
// のどれ一つ置けないなら盤面はもう動かず、めくり続けても同じ状態に戻るだけになる。
//
// 判定は findFoundationFor に委ねる。移動側と同じ関数を読まないと、詰みと言った
// 直後に打てる手が残る (あるいはその逆になる)。
func (o *Osmosis) IsStalemate() bool {
	if o.phase != OsmosisPhasePlaying {
		return false
	}
	for rIdx := range OsmosisReserveCnt {
		pile := o.reserve[rIdx]
		if len(pile) == 0 {
			continue
		}
		if o.findFoundationFor(pile[len(pile)-1]) >= 0 {
			return false
		}
	}
	for _, card := range o.stock {
		if o.findFoundationFor(card) >= 0 {
			return false
		}
	}
	for _, card := range o.waste {
		if o.findFoundationFor(card) >= 0 {
			return false
		}
	}
	return true
}

// GetPhase フェーズ取得
func (o *Osmosis) GetPhase() OsmosisPhase { return o.phase }

// SetPhase フェーズ設定
func (o *Osmosis) SetPhase(p OsmosisPhase) { o.phase = p }

// GetMoveCount 移動回数取得
func (o *Osmosis) GetMoveCount() int { return o.moveCount }

// GetStockCount ストック残枚数
func (o *Osmosis) GetStockCount() int { return len(o.stock) }

// GetWaste ウェイスト取得
func (o *Osmosis) GetWaste() []*Card { return o.waste }

// GetReserve リザーブ取得（4列）
func (o *Osmosis) GetReserve() [OsmosisReserveCnt][]*Card { return o.reserve }

// GetFoundation ファンデーション取得
func (o *Osmosis) GetFoundation() [OsmosisFoundationCnt][]*Card { return o.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (o *Osmosis) GetGameEndFlag() bool { return o.phase != OsmosisPhasePlaying }

// GetBaseRank ベースランク取得
func (o *Osmosis) GetBaseRank() int { return o.baseRank }

// SetBaseRank ベースランク設定 (テスト用)
func (o *Osmosis) SetBaseRank(r int) { o.baseRank = r }

// SetStock ストック設定 (テスト用)
func (o *Osmosis) SetStock(s []*Card) { o.stock = s }

// SetWaste ウェイスト設定 (テスト用)
func (o *Osmosis) SetWaste(w []*Card) { o.waste = w }

// SetReserve リザーブ設定 (テスト用)
func (o *Osmosis) SetReserve(r [OsmosisReserveCnt][]*Card) { o.reserve = r }

// SetFoundation ファンデーション設定 (テスト用)
func (o *Osmosis) SetFoundation(f [OsmosisFoundationCnt][]*Card) { o.foundation = f }

// Undo 直前の操作を取り消す
func (o *Osmosis) Undo() error {
	if o.phase != OsmosisPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(o.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := o.history[len(o.history)-1]
	o.history = o.history[:len(o.history)-1]
	o.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能か
func (o *Osmosis) CanUndo() bool {
	return len(o.history) > 0 && o.phase == OsmosisPhasePlaying
}

// UndoN n回連続アンドゥ
func (o *Osmosis) UndoN(n int) error {
	return undoN(o, n)
}

// --- Private helpers ---

// canPlaceOnFoundation はカードを fIdx 段目のファンデーションに置けるか判定する。
//
// オズモシスのルール:
//   - 各段は1スート専用。空段に最初に置くカードはベースランクでなければならず、
//     そのスートが他の段で未使用である必要がある。
//   - 1段目（最上段）は同一スートであれば順不同で積める。
//   - 2段目以降は、置こうとするランクのカードが「すぐ上の段」に既に存在する場合のみ置ける
//     （上段の進行が下段に「浸透」する）。空段に最初のカード（ベースランク）を置くには、
//     すぐ上の段が開始済み（1枚以上）である必要がある。
func (o *Osmosis) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if fIdx < 0 || fIdx >= OsmosisFoundationCnt {
		return false
	}
	pile := o.foundation[fIdx]
	r := card.GetValue()
	s := card.GetDesign()
	if len(pile) == 0 {
		if r != o.baseRank {
			return false
		}
		if o.suitAssigned(s, fIdx) {
			return false
		}
		if fIdx == 0 {
			return true
		}
		return len(o.foundation[fIdx-1]) > 0
	}
	if pile[0].GetDesign() != s {
		return false
	}
	if fIdx == 0 {
		return true
	}
	return o.rankInPile(o.foundation[fIdx-1], r)
}

// findFoundationFor はカードを置けるファンデーション段を返す（無ければ -1）。
func (o *Osmosis) findFoundationFor(card *Card) int {
	for fIdx := 0; fIdx < OsmosisFoundationCnt; fIdx++ {
		if o.canPlaceOnFoundation(card, fIdx) {
			return fIdx
		}
	}
	return -1
}

// suitAssigned は except 以外の段にスート s が既に割り当てられているか判定する。
func (o *Osmosis) suitAssigned(s, except int) bool {
	for i := 0; i < OsmosisFoundationCnt; i++ {
		if i == except {
			continue
		}
		if len(o.foundation[i]) > 0 && o.foundation[i][0].GetDesign() == s {
			return true
		}
	}
	return false
}

// rankInPile は pile 内にランク r のカードが存在するか判定する。
func (o *Osmosis) rankInPile(pile []*Card, r int) bool {
	for _, c := range pile {
		if c.GetValue() == r {
			return true
		}
	}
	return false
}

func (o *Osmosis) checkGameClear() {
	for i := 0; i < OsmosisFoundationCnt; i++ {
		if len(o.foundation[i]) != CardValueMax {
			return
		}
	}
	o.phase = OsmosisPhaseGameClear
}

func (o *Osmosis) takeSnapshot() {
	snap := &osmosisSnapshot{
		phase:     o.phase,
		moveCount: o.moveCount,
	}
	for i := 0; i < OsmosisReserveCnt; i++ {
		snap.reserve[i] = make([]*Card, len(o.reserve[i]))
		copy(snap.reserve[i], o.reserve[i])
	}
	snap.stock = make([]*Card, len(o.stock))
	copy(snap.stock, o.stock)
	snap.waste = make([]*Card, len(o.waste))
	copy(snap.waste, o.waste)
	for i := 0; i < OsmosisFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(o.foundation[i]))
		copy(snap.foundation[i], o.foundation[i])
	}
	o.history = appendSnapshot(o.history, snap)
}

func (o *Osmosis) restoreSnapshot(snap *osmosisSnapshot) {
	o.reserve = snap.reserve
	o.stock = snap.stock
	o.waste = snap.waste
	o.foundation = snap.foundation
	o.phase = snap.phase
	o.moveCount = snap.moveCount
}

func (o *Osmosis) appendLog(actionType, detail string, cards []*Card) {
	o.appendLogAt(o.moveCount, 0, actionType, detail, cards)
}

// osmosisJSON is the JSON wire format for Osmosis.
type osmosisJSON struct {
	TrumpCards *TrumpCards                   `json:"tc"`
	Reserve    [OsmosisReserveCnt][]*Card    `json:"rv"`
	Stock      []*Card                       `json:"st"`
	Waste      []*Card                       `json:"wa"`
	Foundation [OsmosisFoundationCnt][]*Card `json:"fd"`
	BaseRank   int                           `json:"br"`
	Phase      OsmosisPhase                  `json:"ps"`
	MoveCount  int                           `json:"mc"`
	ActionLog  []*ActionLogEntry             `json:"al"`
	History    []*osmosisSnapshot            `json:"hi,omitempty"`
}

// osmosisSnapshotJSON is the wire format for a single undo snapshot.
type osmosisSnapshotJSON struct {
	Reserve    [OsmosisReserveCnt][]*Card    `json:"rv"`
	Stock      []*Card                       `json:"st"`
	Waste      []*Card                       `json:"wa"`
	Foundation [OsmosisFoundationCnt][]*Card `json:"fd"`
	Phase      OsmosisPhase                  `json:"ps"`
	MoveCount  int                           `json:"mc"`
}

// osmosisMaxSliceLen caps slice sizes during deserialisation.
const osmosisMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler for osmosisSnapshot.
func (s *osmosisSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(osmosisSnapshotJSON{
		Reserve:    s.reserve,
		Stock:      s.stock,
		Waste:      s.waste,
		Foundation: s.foundation,
		Phase:      s.phase,
		MoveCount:  s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for osmosisSnapshot.
func (s *osmosisSnapshot) UnmarshalJSON(data []byte) error {
	var j osmosisSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > osmosisMaxSliceLen || len(j.Waste) > osmosisMaxSliceLen {
		return fmt.Errorf("osmosis: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Reserve {
		if len(pile) > osmosisMaxSliceLen {
			return fmt.Errorf("osmosis: snapshot reserve pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > osmosisMaxSliceLen {
			return fmt.Errorf("osmosis: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.reserve = osmosisNormalizeReserve(j.Reserve)
	s.stock = osmosisNonNil(j.Stock)
	s.waste = osmosisNonNil(j.Waste)
	for i := 0; i < OsmosisFoundationCnt; i++ {
		s.foundation[i] = osmosisNonNil(j.Foundation[i])
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (o *Osmosis) MarshalJSON() ([]byte, error) {
	return json.Marshal(osmosisJSON{
		TrumpCards: o.trumpCards,
		Reserve:    o.reserve,
		Stock:      o.stock,
		Waste:      o.waste,
		Foundation: o.foundation,
		BaseRank:   o.baseRank,
		Phase:      o.phase,
		MoveCount:  o.moveCount,
		ActionLog:  o.actionLog,
		History:    o.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Osmosis) UnmarshalJSON(data []byte) error {
	var j osmosisJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > osmosisMaxSliceLen || len(j.Waste) > osmosisMaxSliceLen ||
		len(j.ActionLog) > osmosisMaxSliceLen || len(j.History) > osmosisMaxSliceLen {
		return fmt.Errorf("osmosis: input array exceeds maximum allowed size")
	}
	for _, pile := range j.Reserve {
		if len(pile) > osmosisMaxSliceLen {
			return fmt.Errorf("osmosis: reserve pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > osmosisMaxSliceLen {
			return fmt.Errorf("osmosis: foundation pile exceeds maximum allowed size")
		}
	}
	o.trumpCards = j.TrumpCards
	if o.trumpCards == nil {
		o.trumpCards = NewTrumpCards(0)
	}
	o.reserve = osmosisNormalizeReserve(j.Reserve)
	o.stock = osmosisNonNil(j.Stock)
	o.waste = osmosisNonNil(j.Waste)
	for i := 0; i < OsmosisFoundationCnt; i++ {
		o.foundation[i] = osmosisNonNil(j.Foundation[i])
	}
	o.baseRank = j.BaseRank
	o.phase = j.Phase
	o.moveCount = j.MoveCount
	o.actionLog = j.ActionLog
	if o.actionLog == nil {
		o.actionLog = make([]*ActionLogEntry, 0)
	}
	o.history = j.History
	if o.history == nil {
		o.history = make([]*osmosisSnapshot, 0)
	}
	return nil
}

// osmosisNonNil returns a non-nil slice with nil elements removed.
func osmosisNonNil(s []*Card) []*Card {
	res := make([]*Card, 0, len(s))
	for _, c := range s {
		if c != nil {
			res = append(res, c)
		}
	}
	return res
}

// osmosisNormalizeReserve ensures every reserve pile is non-nil and free of nil elements.
func osmosisNormalizeReserve(r [OsmosisReserveCnt][]*Card) [OsmosisReserveCnt][]*Card {
	for i := 0; i < OsmosisReserveCnt; i++ {
		r[i] = osmosisNonNil(r[i])
	}
	return r
}
