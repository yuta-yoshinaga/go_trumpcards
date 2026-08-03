//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// SirTommyPhase サー・トミーのゲームフェーズ
type SirTommyPhase int

// SirTommyのフェーズ定数
const (
	// SirTommyPhasePlaying プレイ中
	SirTommyPhasePlaying SirTommyPhase = iota
	// SirTommyPhaseGameClear ゲームクリア
	SirTommyPhaseGameClear
	// SirTommyPhaseGameOver ゲームオーバー
	SirTommyPhaseGameOver
)

// SirTommyFoundationCnt ファンデーション数（各列がエースから K まで積み上がる）
const SirTommyFoundationCnt = 4

// SirTommyWasteCnt ウェイストパイルの数
const SirTommyWasteCnt = 4

// SirTommyHint サー・トミーのヒント
type SirTommyHint struct {
	// FromZone 移動元ゾーン "stock" または "waste"
	FromZone string
	// WasteIdx 移動元がウェイストの場合のインデックス、stock の場合は -1
	WasteIdx int
	// FoundationIdx 移動先ファンデーションのインデックス
	FoundationIdx int
}

// SirTommy サー・トミーゲームクラス。
//
// 現存最古級のペイシェンス。山札を 1 枚ずつめくり、エースはファンデーションを開き、
// それ以外は 4 つのウェイストのいずれかへ置く。ファンデーションは**スートを問わず**
// A→K へ 1 ずつ昇順に積み上げ、ウェイストは**最上段のみ**動かせる（列内の入れ替えも
// ウェイスト同士の移動もできない）。この「置き場所を選ぶ判断が後で効いてくる」点が
// ゲーム性の中心で、Calculation と構造は似ているが、あちらがファンデーションを
// 最初から 4 枚据えるのに対し、こちらは引いたエースで開く。
type SirTommy struct {
	trumpCards  *TrumpCards
	foundations [SirTommyFoundationCnt][]*Card
	wastes      [SirTommyWasteCnt][]*Card
	stock       []*Card
	phase       SirTommyPhase
	moveCount   int
	actionLog   []*ActionLogEntry
	history     []*sirTommySnapshot
	isStalemate bool
}

// sirTommySnapshot アンドゥ用スナップショット
type sirTommySnapshot struct {
	foundations [SirTommyFoundationCnt][]*Card
	wastes      [SirTommyWasteCnt][]*Card
	stock       []*Card
	phase       SirTommyPhase
	moveCount   int
	isStalemate bool
}

// NewSirTommy コンストラクタ
func NewSirTommy(trumpCards *TrumpCards) *SirTommy {
	return &SirTommy{trumpCards: trumpCards}
}

// NewDefaultSirTommy returns SirTommy with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSirTommy() *SirTommy {
	return NewSirTommy(NewTrumpCards(0))
}

// Reset ゲームリセット
func (s *SirTommy) Reset() {
	s.trumpCards.Shuffle()
	s.phase = SirTommyPhasePlaying
	s.moveCount = 0
	s.actionLog = nil
	s.history = nil
	s.isStalemate = false

	for i := range SirTommyFoundationCnt {
		s.foundations[i] = nil
	}
	for i := range SirTommyWasteCnt {
		s.wastes[i] = nil
	}

	// 全 52 枚がストックに入る。エースを事前に抜かないのがカルキュレーションとの違い。
	s.stock = nil
	for s.trumpCards.GetRemainingCount() > 0 {
		s.stock = append(s.stock, s.trumpCards.DrawCard())
	}
}

// PlayStockToFoundation ストック最上段をファンデーションに置く
func (s *SirTommy) PlayStockToFoundation(fIdx int) error {
	if s.phase != SirTommyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fIdx < 0 || fIdx >= SirTommyFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(s.stock) == 0 {
		return errors.New("stock is empty")
	}
	card := s.stock[len(s.stock)-1]
	if !s.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	s.takeSnapshot()
	s.stock = s.stock[:len(s.stock)-1]
	s.foundations[fIdx] = append(s.foundations[fIdx], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("ストック→ファンデーション%d", fIdx+1), []*Card{card})
	s.checkGameClear()
	s.checkStalemate()
	return nil
}

// PlayStockToWaste ストック最上段を指定ウェイストパイルに置く
func (s *SirTommy) PlayStockToWaste(wasteIdx int) error {
	if s.phase != SirTommyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if wasteIdx < 0 || wasteIdx >= SirTommyWasteCnt {
		return errors.New("invalid waste index")
	}
	if len(s.stock) == 0 {
		return errors.New("stock is empty")
	}
	s.takeSnapshot()
	card := s.stock[len(s.stock)-1]
	s.stock = s.stock[:len(s.stock)-1]
	s.wastes[wasteIdx] = append(s.wastes[wasteIdx], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("ストック→ウェイスト%d", wasteIdx+1), []*Card{card})
	s.checkStalemate()
	return nil
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (s *SirTommy) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	if s.phase != SirTommyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if wasteIdx < 0 || wasteIdx >= SirTommyWasteCnt {
		return errors.New("invalid waste index")
	}
	if fIdx < 0 || fIdx >= SirTommyFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(s.wastes[wasteIdx]) == 0 {
		return errors.New("waste is empty")
	}
	card := s.wastes[wasteIdx][len(s.wastes[wasteIdx])-1]
	if !s.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	s.takeSnapshot()
	s.wastes[wasteIdx] = s.wastes[wasteIdx][:len(s.wastes[wasteIdx])-1]
	s.foundations[fIdx] = append(s.foundations[fIdx], card)
	s.moveCount++
	s.appendLog("move", fmt.Sprintf("ウェイスト%d→ファンデーション%d", wasteIdx+1, fIdx+1), []*Card{card})
	s.checkGameClear()
	s.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (s *SirTommy) GiveUp() {
	if s.phase == SirTommyPhasePlaying {
		s.phase = SirTommyPhaseGameOver
		s.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得（ファンデーションに置けるカードを提示する）
func (s *SirTommy) GetHint() *SirTommyHint {
	if s.phase != SirTommyPhasePlaying {
		return nil
	}
	if len(s.stock) > 0 {
		card := s.stock[len(s.stock)-1]
		if fIdx := s.findFoundation(card); fIdx >= 0 {
			return &SirTommyHint{FromZone: "stock", WasteIdx: -1, FoundationIdx: fIdx}
		}
	}
	for wIdx := range SirTommyWasteCnt {
		pile := s.wastes[wIdx]
		if len(pile) == 0 {
			continue
		}
		if fIdx := s.findFoundation(pile[len(pile)-1]); fIdx >= 0 {
			return &SirTommyHint{FromZone: "waste", WasteIdx: wIdx, FoundationIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 置けるカードがなくなるまで自動でファンデーションへ積む
func (s *SirTommy) AutoComplete() error {
	if s.phase != SirTommyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	// Same gate as Calculation: auto-complete finishes an endgame, it does not
	// play for you mid-deal. The frontend button already requires an empty stock
	// (autoCompleteReady), so without this the CLI and API would disagree with it.
	if !s.AllFaceUp() {
		return errors.New("stock is not empty")
	}
	moved := false
	// 1 手ごとに全体を見直す。1 枚積むと別のウェイスト最上段が解放されるため、
	// 単純な一巡では取りこぼす。
	for {
		h := s.GetHint()
		if h == nil {
			break
		}
		var err error
		if h.FromZone == "stock" {
			err = s.PlayStockToFoundation(h.FoundationIdx)
		} else {
			err = s.PlayWasteToFoundation(h.WasteIdx, h.FoundationIdx)
		}
		if err != nil {
			return err
		}
		moved = true
	}
	if !moved {
		return errors.New("no card can be auto-completed")
	}
	return nil
}

// Undo 直前の 1 手を取り消す
func (s *SirTommy) Undo() error {
	if len(s.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := s.history[len(s.history)-1]
	s.history = s.history[:len(s.history)-1]
	s.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能か
func (s *SirTommy) CanUndo() bool { return len(s.history) > 0 }

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ 0、履歴を遡っても抜けられなければ -1。
func (s *SirTommy) UndoToEscape() int {
	if !s.isStalemate {
		return 0
	}
	for i := len(s.history) - 1; i >= 0; i-- {
		if !s.history[i].isStalemate {
			return len(s.history) - i
		}
	}
	return -1
}

// AllFaceUp すべてのカードが見えているか。
// サー・トミーは伏せ札を持たず、ストックを引き切った時点で全札が可視になる。
func (s *SirTommy) AllFaceUp() bool { return len(s.stock) == 0 }

// UndoN n 手戻す
func (s *SirTommy) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(s.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := s.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// GetPhase フェーズ取得
func (s *SirTommy) GetPhase() SirTommyPhase { return s.phase }

// GetMoveCount 手数取得
func (s *SirTommy) GetMoveCount() int { return s.moveCount }

// GetStockCount ストック残枚数取得
func (s *SirTommy) GetStockCount() int { return len(s.stock) }

// GetWastes ウェイスト取得
func (s *SirTommy) GetWastes() [SirTommyWasteCnt][]*Card { return s.wastes }

// GetFoundations ファンデーション取得
func (s *SirTommy) GetFoundations() [SirTommyFoundationCnt][]*Card { return s.foundations }

// GetActionLog 棋譜取得
func (s *SirTommy) GetActionLog() []*ActionLogEntry { return s.actionLog }

// GetGameEndFlag ゲーム終了フラグ
func (s *SirTommy) GetGameEndFlag() bool { return s.phase != SirTommyPhasePlaying }

// IsStalemate 手詰まりか
func (s *SirTommy) IsStalemate() bool { return s.isStalemate }

// GetStockTop ストック最上段を返す（空なら nil）
func (s *SirTommy) GetStockTop() *Card {
	if len(s.stock) == 0 {
		return nil
	}
	return s.stock[len(s.stock)-1]
}

// --- Private helpers ---

// canPlaceOnFoundation ファンデーションにカードを置けるか判定。
// 空の列にはエースのみ、それ以外は最上段の 1 つ上のランクのみ。スートは問わない。
func (s *SirTommy) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := s.foundations[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	if len(pile) >= CardValueMax {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（無ければ -1）
func (s *SirTommy) findFoundation(card *Card) int {
	for i := range SirTommyFoundationCnt {
		if s.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (s *SirTommy) checkGameClear() {
	for i := range SirTommyFoundationCnt {
		if len(s.foundations[i]) != CardValueMax {
			return
		}
	}
	s.phase = SirTommyPhaseGameClear
}

// checkStalemate 手詰まり判定。
// ストックが残っている限りウェイストへ逃がせるので手詰まりにはならない。
func (s *SirTommy) checkStalemate() {
	if s.phase != SirTommyPhasePlaying {
		return
	}
	if len(s.stock) > 0 {
		s.isStalemate = false
		return
	}
	s.isStalemate = s.GetHint() == nil
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (s *SirTommy) takeSnapshot() {
	snap := &sirTommySnapshot{
		phase:       s.phase,
		moveCount:   s.moveCount,
		isStalemate: s.isStalemate,
	}
	for i := range SirTommyFoundationCnt {
		snap.foundations[i] = make([]*Card, len(s.foundations[i]))
		copy(snap.foundations[i], s.foundations[i])
	}
	for i := range SirTommyWasteCnt {
		snap.wastes[i] = make([]*Card, len(s.wastes[i]))
		copy(snap.wastes[i], s.wastes[i])
	}
	snap.stock = make([]*Card, len(s.stock))
	copy(snap.stock, s.stock)
	s.history = append(s.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (s *SirTommy) restoreSnapshot(snap *sirTommySnapshot) {
	s.foundations = snap.foundations
	s.wastes = snap.wastes
	s.stock = snap.stock
	s.phase = snap.phase
	s.moveCount = snap.moveCount
	s.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (s *SirTommy) appendLog(actionType, detail string, cards []*Card) {
	s.actionLog = append(s.actionLog, &ActionLogEntry{
		TurnNumber: s.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// sirTommyMaxSliceLen caps slice sizes during deserialisation.
const sirTommyMaxSliceLen = 1000

// sirTommySnapshotJSON is the wire format for a single undo snapshot.
// sirTommySnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type sirTommySnapshotJSON struct {
	Foundations [SirTommyFoundationCnt][]*Card `json:"fs"`
	Wastes      [SirTommyWasteCnt][]*Card      `json:"ws"`
	Stock       []*Card                        `json:"st"`
	Phase       SirTommyPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for sirTommySnapshot.
func (s *sirTommySnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(sirTommySnapshotJSON{
		Foundations: s.foundations,
		Wastes:      s.wastes,
		Stock:       s.stock,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for sirTommySnapshot.
func (s *sirTommySnapshot) UnmarshalJSON(data []byte) error {
	var j sirTommySnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > sirTommyMaxSliceLen {
		return errors.New("sirtommy: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundations {
		if len(pile) > sirTommyMaxSliceLen {
			return errors.New("sirtommy: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Wastes {
		if len(pile) > sirTommyMaxSliceLen {
			return errors.New("sirtommy: snapshot pile exceeds maximum allowed size")
		}
	}
	s.foundations = j.Foundations
	s.wastes = j.Wastes
	s.stock = j.Stock
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// sirTommyJSON is the JSON wire format for SirTommy.
type sirTommyJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Foundations [SirTommyFoundationCnt][]*Card `json:"fd"`
	Wastes      [SirTommyWasteCnt][]*Card      `json:"wa"`
	Stock       []*Card                        `json:"st"`
	Phase       SirTommyPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*sirTommySnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (s *SirTommy) MarshalJSON() ([]byte, error) {
	return json.Marshal(&sirTommyJSON{
		TrumpCards:  s.trumpCards,
		Foundations: s.foundations,
		Wastes:      s.wastes,
		Stock:       s.stock,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		ActionLog:   s.actionLog,
		IsStalemate: s.isStalemate,
		History:     s.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。
// 値域を検証するのは、KV に入っているのは以前のバージョンが書いた任意のバイト列で
// あり、壊れた状態でゲームを開始させないため。
func (s *SirTommy) UnmarshalJSON(data []byte) error {
	var j sirTommyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > sirTommyMaxSliceLen || len(j.History) > sirTommyMaxSliceLen {
		return errors.New("sirtommy: input array exceeds maximum allowed size")
	}
	if j.Phase < SirTommyPhasePlaying || j.Phase > SirTommyPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	for i := range SirTommyFoundationCnt {
		if len(j.Foundations[i]) > CardValueMax {
			return fmt.Errorf("foundation %d has %d cards", i, len(j.Foundations[i]))
		}
	}
	if j.TrumpCards != nil {
		s.trumpCards = j.TrumpCards
	}
	s.foundations = j.Foundations
	s.wastes = j.Wastes
	s.stock = j.Stock
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.actionLog = j.ActionLog
	s.isStalemate = j.IsStalemate
	s.history = j.History
	return nil
}
