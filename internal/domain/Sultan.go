//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SultanPhase スルタンゲームフェーズ
type SultanPhase int

// Sultanのフェーズ定数
const (
	// SultanPhasePlaying プレイ中
	SultanPhasePlaying SultanPhase = iota
	// SultanPhaseGameClear ゲームクリア
	SultanPhaseGameClear
	// SultanPhaseGameOver ゲームオーバー
	SultanPhaseGameOver
)

// SultanFoundationCnt ファンデーション（王）の数（2デッキ × 4スート = 8）
const SultanFoundationCnt = 8

// SultanDivanCnt ディヴァン（リザーブ）のスロット数
const SultanDivanCnt = 8

// SultanMaxRedeal リディール可能回数（合計3パス）
const SultanMaxRedeal = 2

// SultanFoundationFull 完成したファンデーションの枚数（King + A..Q の 12 枚 = 13）
const SultanFoundationFull = 13

// SultanHint ヒント
type SultanHint struct {
	// FromZone は移動元ゾーン ("divan" または "waste")。
	FromZone string
	// FromIdx は divan の場合のスロット番号、waste の場合は -1。
	FromIdx int
	// ToFoundation は移動先ファンデーション番号。
	ToFoundation int
}

// SultanConfig スルタンゲーム設定
type SultanConfig struct{}

// Sultan スルタンゲームクラス
type Sultan struct {
	trumpCards *TrumpCards
	foundation [SultanFoundationCnt][]*Card
	divan      []*Card // 8 スロット。プレイ済みでストックが空のスロットは nil。
	stock      []*Card
	waste      []*Card
	phase      SultanPhase
	moveCount  int
	actionLogBase
	history     []*sultanSnapshot
	isStalemate bool
	redealCount int
}

// sultanSnapshot アンドゥ用スナップショット
type sultanSnapshot struct {
	foundation  [SultanFoundationCnt][]*Card
	divan       []*Card
	stock       []*Card
	waste       []*Card
	phase       SultanPhase
	moveCount   int
	isStalemate bool
	redealCount int
}

// NewSultan コンストラクタ
func NewSultan(trumpCards *TrumpCards) *Sultan {
	return &Sultan{
		trumpCards: trumpCards,
	}
}

// NewDefaultSultan returns Sultan with two combined 52-card decks (104 cards).
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSultan() *Sultan {
	return NewSultan(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (su *Sultan) Reset() {
	su.trumpCards.Shuffle()
	su.phase = SultanPhasePlaying
	su.moveCount = 0
	su.actionLog = nil
	su.history = nil
	su.isStalemate = false
	su.redealCount = 0

	// 全 104 枚を引いて、8 枚の King をファンデーションの土台に据える。
	// 残り 96 枚のうち 8 枚を Divan、88 枚を Stock に配る。
	for i := range SultanFoundationCnt {
		su.foundation[i] = nil
	}
	su.divan = make([]*Card, 0, SultanDivanCnt)
	su.stock = nil
	su.waste = nil

	kingIdx := 0
	var rest []*Card
	for su.trumpCards.GetRemainingCount() > 0 {
		card := su.trumpCards.DrawCard()
		if card.GetValue() == CardValueMax && kingIdx < SultanFoundationCnt {
			// King を順にファンデーションの土台へ。
			su.foundation[kingIdx] = []*Card{card}
			kingIdx++
		} else {
			rest = append(rest, card)
		}
	}

	// 残り 96 枚: 先頭 8 枚を Divan、残り 88 枚を Stock へ。
	for i, card := range rest {
		if i < SultanDivanCnt {
			su.divan = append(su.divan, card)
		} else {
			su.stock = append(su.stock, card)
		}
	}
}

// Draw ストックからウェイストにカードを1枚引く（リサイクルなし、リディールは別）
func (su *Sultan) Draw() error {
	if su.phase != SultanPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(su.stock) == 0 {
		return errors.New("no cards in stock")
	}
	su.takeSnapshot()
	card := su.stock[len(su.stock)-1]
	su.stock = su.stock[:len(su.stock)-1]
	su.waste = append(su.waste, card)
	su.moveCount++
	su.appendLog("draw", "ストックからカードを引きました", []*Card{card})
	su.checkSultanStalemate()
	return nil
}

// Redeal ストックを使い切った後、ウェイストを集めて新しいストックを作る（最大2回）。
// ウェイストを反転してストックへ戻すことで、最初に引いたカードが再び最初に引かれる。
func (su *Sultan) Redeal() error {
	if su.phase != SultanPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(su.stock) != 0 {
		return errors.New("cannot redeal: stock is not empty")
	}
	if su.redealCount >= SultanMaxRedeal {
		return errors.New("cannot redeal: no redeals left")
	}
	if len(su.waste) == 0 {
		return errors.New("cannot redeal: waste is empty")
	}
	su.takeSnapshot()
	// ウェイストを反転してストックへ（最初に引いたカードが再び最初に引かれる）
	su.stock = make([]*Card, len(su.waste))
	for i, c := range su.waste {
		su.stock[len(su.waste)-1-i] = c
	}
	su.waste = nil
	su.redealCount++
	su.moveCount++
	su.appendLog("redeal", "ウェイストを集めて新しいストックを作りました", nil)
	su.checkSultanStalemate()
	return nil
}

// MoveDivanToFoundation ディヴァンからファンデーションにカードを移動し、空いたスロットをストックから補充する。
func (su *Sultan) MoveDivanToFoundation(divanIdx int) error {
	if su.phase != SultanPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if divanIdx < 0 || divanIdx >= len(su.divan) {
		return errors.New("invalid divan index")
	}
	card := su.divan[divanIdx]
	if card == nil {
		return errors.New("divan slot is empty")
	}
	fIdx := su.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	su.takeSnapshot()
	su.foundation[fIdx] = append(su.foundation[fIdx], card)
	// スロットをストックから補充。ストックが空ならスロットは nil のまま。
	if len(su.stock) > 0 {
		su.divan[divanIdx] = su.stock[len(su.stock)-1]
		su.stock = su.stock[:len(su.stock)-1]
	} else {
		su.divan[divanIdx] = nil
	}
	su.moveCount++
	su.appendLog("move", fmt.Sprintf("ディヴァン%d→ファンデーション", divanIdx), []*Card{card})
	su.checkGameClear()
	su.checkSultanStalemate()
	return nil
}

// MoveWasteToFoundation ウェイストからファンデーションにカードを移動
func (su *Sultan) MoveWasteToFoundation() error {
	if su.phase != SultanPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(su.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := su.waste[len(su.waste)-1]
	fIdx := su.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	su.takeSnapshot()
	su.waste = su.waste[:len(su.waste)-1]
	su.foundation[fIdx] = append(su.foundation[fIdx], card)
	su.moveCount++
	su.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	su.checkGameClear()
	su.checkSultanStalemate()
	return nil
}

// GiveUp ギブアップ
func (su *Sultan) GiveUp() {
	if su.phase == SultanPhasePlaying {
		su.phase = SultanPhaseGameOver
		su.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (su *Sultan) GetHint() *SultanHint {
	if su.phase != SultanPhasePlaying {
		return nil
	}
	// 優先度1: ディヴァンからファンデーションへ
	for i, card := range su.divan {
		if card == nil {
			continue
		}
		fIdx := su.findFoundation(card)
		if fIdx >= 0 {
			return &SultanHint{FromZone: "divan", FromIdx: i, ToFoundation: fIdx}
		}
	}
	// 優先度2: ウェイストからファンデーションへ
	if len(su.waste) > 0 {
		card := su.waste[len(su.waste)-1]
		fIdx := su.findFoundation(card)
		if fIdx >= 0 {
			return &SultanHint{FromZone: "waste", FromIdx: -1, ToFoundation: fIdx}
		}
	}
	return nil
}

// AutoComplete オートコンプリート（ディヴァン/ウェイストから自動でファンデーションへ移動）
func (su *Sultan) AutoComplete() error {
	if su.phase != SultanPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !su.AllFaceUp() {
		return errors.New("not all cards are face up")
	}
	su.takeSnapshot()
	for {
		moved := false
		// ウェイストからファンデーションへ
		for len(su.waste) > 0 {
			card := su.waste[len(su.waste)-1]
			fIdx := su.findFoundation(card)
			if fIdx < 0 {
				break
			}
			su.waste = su.waste[:len(su.waste)-1]
			su.foundation[fIdx] = append(su.foundation[fIdx], card)
			su.moveCount++
			moved = true
		}
		// ディヴァンからファンデーションへ
		for i, card := range su.divan {
			if card == nil {
				continue
			}
			fIdx := su.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			su.foundation[fIdx] = append(su.foundation[fIdx], card)
			su.divan[i] = nil
			su.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	su.appendLog("autocomplete", "オートコンプリートを実行しました", nil)
	su.checkGameClear()
	return nil
}

// AllFaceUp 全カードが可視かどうか（ストックが空 = 全カード可視）
func (su *Sultan) AllFaceUp() bool {
	return len(su.stock) == 0
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (su *Sultan) GetPhase() SultanPhase { return su.phase }

// SetPhase フェーズ設定 (テスト用)
func (su *Sultan) SetPhase(phase SultanPhase) { su.phase = phase }

// GetMoveCount 移動回数取得
func (su *Sultan) GetMoveCount() int { return su.moveCount }

// GetStockCount ストック枚数取得
func (su *Sultan) GetStockCount() int { return len(su.stock) }

// GetWaste ウェイスト取得
func (su *Sultan) GetWaste() []*Card { return su.waste }

// GetDivan ディヴァン取得
func (su *Sultan) GetDivan() []*Card { return su.divan }

// GetFoundation ファンデーション取得
func (su *Sultan) GetFoundation() [SultanFoundationCnt][]*Card { return su.foundation }

// GetGameEndFlag returns true once the game has left the playing phase.
func (su *Sultan) GetGameEndFlag() bool { return su.phase != SultanPhasePlaying }

// IsStalemate 手詰まり状態取得
func (su *Sultan) IsStalemate() bool { return su.isStalemate }

// GetRedealCount リディール使用回数を返す
func (su *Sultan) GetRedealCount() int { return su.redealCount }

// CanRedeal リディール可能かどうかを返す（ストックが空 && 残回数あり && ウェイストに残あり && プレイ中）
func (su *Sultan) CanRedeal() bool {
	return su.phase == SultanPhasePlaying && len(su.stock) == 0 &&
		su.redealCount < SultanMaxRedeal && len(su.waste) > 0
}

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (su *Sultan) SetIsStalemate(v bool) { su.isStalemate = v }

// SetRedealCount リディール使用回数設定 (テスト用)
func (su *Sultan) SetRedealCount(v int) { su.redealCount = v }

// SetDivan ディヴァン設定 (テスト用)
func (su *Sultan) SetDivan(divan []*Card) { su.divan = divan }

// SetStock ストック設定 (テスト用)
func (su *Sultan) SetStock(stock []*Card) { su.stock = stock }

// SetWaste ウェイスト設定 (テスト用)
func (su *Sultan) SetWaste(waste []*Card) { su.waste = waste }

// SetFoundation ファンデーション設定 (テスト用)
func (su *Sultan) SetFoundation(foundation [SultanFoundationCnt][]*Card) {
	su.foundation = foundation
}

// Undo 直前の操作を取り消す
func (su *Sultan) Undo() error {
	if su.phase != SultanPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(su.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := su.history[len(su.history)-1]
	su.history = su.history[:len(su.history)-1]
	su.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能かどうか
func (su *Sultan) CanUndo() bool {
	return len(su.history) > 0 && su.phase == SultanPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (su *Sultan) UndoToEscape() int {
	return undoToEscape(su.isStalemate, su.history, func(s *sultanSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (su *Sultan) UndoN(n int) error {
	return undoN(su, n)
}

// --- Private helpers ---

// canPlaceOnFoundation ファンデーションにカードを置けるか判定。
// ファンデーションは王の上に同スートで King→A→2→…→10→J→Q の順に積み上げる。
func (su *Sultan) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := su.foundation[fIdx]
	if len(pile) == 0 {
		// 土台に王が据えられているため、空のファンデーションは存在しない。
		return false
	}
	topCard := pile[len(pile)-1]
	if card.GetDesign() != topCard.GetDesign() {
		return false
	}
	if topCard.GetValue() == CardValueMax {
		// 王の上には Ace (1) のみ置ける。
		return card.GetValue() == 1
	}
	if topCard.GetValue() >= SultanFoundationFull-1 {
		// Queen (12) が積まれていれば完成しており、これ以上置けない。
		return false
	}
	// それ以外は同スートで1つ上の値。
	return card.GetValue() == topCard.GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（見つからない場合-1）。
// 同スートのファンデーションが2つあるため、最初に置けるものを返す。
func (su *Sultan) findFoundation(card *Card) int {
	for i := range SultanFoundationCnt {
		if su.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定（全ファンデーションが完成）
func (su *Sultan) checkGameClear() {
	for i := range SultanFoundationCnt {
		if len(su.foundation[i]) != SultanFoundationFull {
			return
		}
	}
	su.phase = SultanPhaseGameClear
}

// checkSultanStalemate 手詰まり判定
func (su *Sultan) checkSultanStalemate() {
	if su.phase != SultanPhasePlaying {
		return
	}
	if su.GetHint() != nil {
		su.isStalemate = false
		return
	}
	// ヒントがない場合
	if len(su.stock) == 0 {
		// ストック空。リディール可能ならまだ手がある。
		if su.CanRedeal() {
			su.isStalemate = false
			return
		}
		su.isStalemate = true
		return
	}
	// ストックにカードが残っている場合はまだ引ける
	su.isStalemate = false
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (su *Sultan) takeSnapshot() {
	snap := &sultanSnapshot{
		phase:       su.phase,
		moveCount:   su.moveCount,
		isStalemate: su.isStalemate,
		redealCount: su.redealCount,
	}
	for i := range SultanFoundationCnt {
		snap.foundation[i] = make([]*Card, len(su.foundation[i]))
		copy(snap.foundation[i], su.foundation[i])
	}
	snap.divan = make([]*Card, len(su.divan))
	copy(snap.divan, su.divan)
	snap.stock = make([]*Card, len(su.stock))
	copy(snap.stock, su.stock)
	snap.waste = make([]*Card, len(su.waste))
	copy(snap.waste, su.waste)
	su.history = appendSnapshot(su.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (su *Sultan) restoreSnapshot(snap *sultanSnapshot) {
	su.foundation = snap.foundation
	su.divan = snap.divan
	su.stock = snap.stock
	su.waste = snap.waste
	su.phase = snap.phase
	su.moveCount = snap.moveCount
	su.isStalemate = snap.isStalemate
	su.redealCount = snap.redealCount
}

// appendLog 棋譜エントリを追加
func (su *Sultan) appendLog(actionType, detail string, cards []*Card) {
	su.appendLogAt(su.moveCount, 0, actionType, detail, cards)
}

// sultanJSON is the JSON wire format for Sultan.
type sultanJSON struct {
	TrumpCards  *TrumpCards                  `json:"tc"`
	Foundation  [SultanFoundationCnt][]*Card `json:"fd"`
	Divan       []*Card                      `json:"dv"`
	Stock       []*Card                      `json:"st"`
	Waste       []*Card                      `json:"wa"`
	Phase       SultanPhase                  `json:"ps"`
	MoveCount   int                          `json:"mc"`
	ActionLog   []*ActionLogEntry            `json:"al"`
	IsStalemate bool                         `json:"sl"`
	RedealCount int                          `json:"rc"`
	History     []*sultanSnapshot            `json:"hi,omitempty"`
}

// sultanSnapshotJSON is the wire format for a single undo snapshot.
// sultanSnapshot uses unexported fields, so we project to/from this shape
// with explicit Marshal/Unmarshal methods. Field names match sultanJSON's
// short keys to keep the KV payload compact (#1654).
type sultanSnapshotJSON struct {
	Foundation  [SultanFoundationCnt][]*Card `json:"fd"`
	Divan       []*Card                      `json:"dv"`
	Stock       []*Card                      `json:"st"`
	Waste       []*Card                      `json:"wa"`
	Phase       SultanPhase                  `json:"ps"`
	MoveCount   int                          `json:"mc"`
	IsStalemate bool                         `json:"sl"`
	RedealCount int                          `json:"rc"`
}

// sultanMaxSliceLen caps slice sizes during deserialisation.
const sultanMaxSliceLen = 1000

// MarshalJSON implements json.Marshaler for sultanSnapshot.
func (s *sultanSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(sultanSnapshotJSON{
		Foundation:  s.foundation,
		Divan:       s.divan,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
		RedealCount: s.redealCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for sultanSnapshot.
func (s *sultanSnapshot) UnmarshalJSON(data []byte) error {
	var j sultanSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > sultanMaxSliceLen || len(j.Waste) > sultanMaxSliceLen || len(j.Divan) > sultanMaxSliceLen {
		return fmt.Errorf("sultan: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundation {
		if len(pile) > sultanMaxSliceLen {
			return fmt.Errorf("sultan: snapshot foundation pile exceeds maximum allowed size")
		}
	}
	s.foundation = j.Foundation
	for _, pile := range s.foundation {
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("sultan: snapshot foundation contains a nil card")
			}
		}
	}
	// Divan: a played, unrefilled slot is intentionally nil, so nil is valid here.
	s.divan = j.Divan
	if s.divan == nil {
		s.divan = make([]*Card, 0)
	}
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	for _, c := range s.stock {
		if c == nil {
			return fmt.Errorf("sultan: snapshot stock contains a nil card")
		}
	}
	s.waste = j.Waste
	if s.waste == nil {
		s.waste = make([]*Card, 0)
	}
	for _, c := range s.waste {
		if c == nil {
			return fmt.Errorf("sultan: snapshot waste contains a nil card")
		}
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.redealCount = j.RedealCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (su *Sultan) MarshalJSON() ([]byte, error) {
	return json.Marshal(sultanJSON{
		TrumpCards:  su.trumpCards,
		Foundation:  su.foundation,
		Divan:       su.divan,
		Stock:       su.stock,
		Waste:       su.waste,
		Phase:       su.phase,
		MoveCount:   su.moveCount,
		ActionLog:   su.actionLog,
		IsStalemate: su.isStalemate,
		RedealCount: su.redealCount,
		History:     su.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (su *Sultan) UnmarshalJSON(data []byte) error {
	var j sultanJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > sultanMaxSliceLen || len(j.Waste) > sultanMaxSliceLen ||
		len(j.Divan) > sultanMaxSliceLen ||
		len(j.ActionLog) > sultanMaxSliceLen || len(j.History) > sultanMaxSliceLen {
		return fmt.Errorf("sultan: input array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundation {
		if len(pile) > sultanMaxSliceLen {
			return fmt.Errorf("sultan: foundation pile exceeds maximum allowed size")
		}
	}

	su.trumpCards = j.TrumpCards
	if su.trumpCards == nil {
		su.trumpCards = NewTrumpCardsWithDecks(2, 0)
	}
	su.foundation = j.Foundation
	for _, pile := range su.foundation {
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("sultan: foundation contains a nil card")
			}
		}
	}
	// Divan: a played, unrefilled slot is intentionally nil, so nil is valid here.
	su.divan = j.Divan
	if su.divan == nil {
		su.divan = make([]*Card, 0)
	}
	su.stock = j.Stock
	if su.stock == nil {
		su.stock = make([]*Card, 0)
	}
	for _, c := range su.stock {
		if c == nil {
			return fmt.Errorf("sultan: stock contains a nil card")
		}
	}
	su.waste = j.Waste
	if su.waste == nil {
		su.waste = make([]*Card, 0)
	}
	for _, c := range su.waste {
		if c == nil {
			return fmt.Errorf("sultan: waste contains a nil card")
		}
	}
	su.phase = j.Phase
	su.moveCount = j.MoveCount
	su.actionLog = j.ActionLog
	if su.actionLog == nil {
		su.actionLog = make([]*ActionLogEntry, 0)
	}
	su.history = j.History
	if su.history == nil {
		su.history = make([]*sultanSnapshot, 0)
	}
	su.isStalemate = j.IsStalemate
	su.redealCount = j.RedealCount
	return nil
}
