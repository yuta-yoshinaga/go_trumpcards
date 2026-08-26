//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AuldLangSynePhase オールド・ラング・サインのゲームフェーズ
type AuldLangSynePhase int

// AuldLangSyneのフェーズ定数
const (
	// AuldLangSynePhasePlaying プレイ中
	AuldLangSynePhasePlaying AuldLangSynePhase = iota
	// AuldLangSynePhaseGameClear ゲームクリア
	AuldLangSynePhaseGameClear
	// AuldLangSynePhaseGameOver ゲームオーバー
	AuldLangSynePhaseGameOver
)

// AuldLangSyneFoundationCnt ファンデーション数（各列がエースから K まで積み上がる）
const AuldLangSyneFoundationCnt = 4

// AuldLangSyneWasteCnt ウェイストパイルの数
const AuldLangSyneWasteCnt = 4

// AuldLangSyneHint オールド・ラング・サインのヒント
type AuldLangSyneHint struct {
	// WasteIdx 移動元ウェイストのインデックス
	WasteIdx int
	// FoundationIdx 移動先ファンデーションのインデックス
	FoundationIdx int
}

// AuldLangSyne オールド・ラング・サインゲームクラス。
//
// イギリスの古典ペイシェンス。4 枚のエースを最初からファンデーションに据え、
// 残る 48 枚を 4 つのウェイストへ**1 段 4 枚ずつ**配る。ファンデーションは
// スートを問わず A→K へ 1 ずつ昇順に積み、動かせるのはウェイストの最上段のみ
// （ウェイスト同士の移動も列内の入れ替えもできない）。
//
// Sir Tommy と構造は近いが決定的に違うのは**配り先を選べない**点。あちらは
// 引いた 1 枚をどのウェイストに置くか選ぶのがゲーム性の中心で、こちらは 4 列へ
// 強制的に配られるため、プレイヤーに残る判断は「いつ配るか」と「積める札が
// 複数あるときどれを先に積むか」だけになる。運の要素が非常に強い古典として
// 知られるのはこのため。
type AuldLangSyne struct {
	trumpCards  *TrumpCards
	foundations [AuldLangSyneFoundationCnt][]*Card
	wastes      [AuldLangSyneWasteCnt][]*Card
	stock       []*Card
	phase       AuldLangSynePhase
	moveCount   int
	actionLogBase
	history     []*auldLangSyneSnapshot
	isStalemate bool
}

// auldLangSyneSnapshot アンドゥ用スナップショット
type auldLangSyneSnapshot struct {
	foundations [AuldLangSyneFoundationCnt][]*Card
	wastes      [AuldLangSyneWasteCnt][]*Card
	stock       []*Card
	phase       AuldLangSynePhase
	moveCount   int
	isStalemate bool
}

// NewAuldLangSyne コンストラクタ
func NewAuldLangSyne(trumpCards *TrumpCards) *AuldLangSyne {
	return &AuldLangSyne{trumpCards: trumpCards}
}

// NewDefaultAuldLangSyne returns AuldLangSyne with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultAuldLangSyne() *AuldLangSyne {
	return NewAuldLangSyne(NewTrumpCards(0))
}

// Reset ゲームリセット
func (a *AuldLangSyne) Reset() {
	a.trumpCards.Shuffle()
	a.phase = AuldLangSynePhasePlaying
	a.moveCount = 0
	a.actionLog = nil
	a.history = nil
	a.isStalemate = false

	for i := range AuldLangSyneFoundationCnt {
		a.foundations[i] = nil
	}
	for i := range AuldLangSyneWasteCnt {
		a.wastes[i] = nil
	}

	// エースは山札から抜いてファンデーションに据える。どのエースがどの列に載るかは
	// 引いた順で決まるが、スートを問わず積むので列は互換であり結果に影響しない。
	a.stock = nil
	aceIdx := 0
	for a.trumpCards.GetRemainingCount() > 0 {
		card := a.trumpCards.DrawCard()
		if card.GetValue() == 1 && aceIdx < AuldLangSyneFoundationCnt {
			a.foundations[aceIdx] = []*Card{card}
			aceIdx++
			continue
		}
		a.stock = append(a.stock, card)
	}

	// 初期配り。以降の Deal と同じ「4 列へ 1 枚ずつ」だが、こちらは手数にも
	// 棋譜にも数えない（ゲーム開始前の配置なので、アンドゥ対象でもない）。
	a.dealRow()
}

// Deal ストックから 4 つのウェイストへ 1 枚ずつ配る
func (a *AuldLangSyne) Deal() error {
	if a.phase != AuldLangSynePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(a.stock) == 0 {
		return errors.New("stock is empty")
	}
	a.takeSnapshot()
	dealt := a.dealRow()
	a.moveCount++
	a.appendLog("deal", fmt.Sprintf("%d枚を配りました", len(dealt)), dealt)
	a.checkStalemate()
	return nil
}

// PlayWasteToFoundation ウェイスト最上段をファンデーションに置く
func (a *AuldLangSyne) PlayWasteToFoundation(wasteIdx, fIdx int) error {
	if a.phase != AuldLangSynePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if wasteIdx < 0 || wasteIdx >= AuldLangSyneWasteCnt {
		return errors.New("invalid waste index")
	}
	if fIdx < 0 || fIdx >= AuldLangSyneFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if len(a.wastes[wasteIdx]) == 0 {
		return errors.New("waste is empty")
	}
	card := a.wastes[wasteIdx][len(a.wastes[wasteIdx])-1]
	if !a.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	a.takeSnapshot()
	a.wastes[wasteIdx] = a.wastes[wasteIdx][:len(a.wastes[wasteIdx])-1]
	a.foundations[fIdx] = append(a.foundations[fIdx], card)
	a.moveCount++
	a.appendLog("move", fmt.Sprintf("ウェイスト%d→ファンデーション%d", wasteIdx+1, fIdx+1), []*Card{card})
	a.checkGameClear()
	a.checkStalemate()
	return nil
}

// GiveUp ギブアップ
func (a *AuldLangSyne) GiveUp() {
	if a.phase == AuldLangSynePhasePlaying {
		a.phase = AuldLangSynePhaseGameOver
		a.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得（ファンデーションに置けるウェイスト最上段を提示する）
func (a *AuldLangSyne) GetHint() *AuldLangSyneHint {
	if a.phase != AuldLangSynePhasePlaying {
		return nil
	}
	for wIdx := range AuldLangSyneWasteCnt {
		pile := a.wastes[wIdx]
		if len(pile) == 0 {
			continue
		}
		if fIdx := a.findFoundation(pile[len(pile)-1]); fIdx >= 0 {
			return &AuldLangSyneHint{WasteIdx: wIdx, FoundationIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 置けるカードがなくなるまで自動でファンデーションへ積む
func (a *AuldLangSyne) AutoComplete() error {
	if a.phase != AuldLangSynePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	// Same gate as Sir Tommy: auto-complete finishes an endgame, it does not play
	// the deal for you. The frontend button already requires an empty stock
	// (autoCompleteReady), so without this the CLI and API would disagree with it.
	if !a.AllFaceUp() {
		return errors.New("stock is not empty")
	}
	moved := false
	// 1 手ごとに全体を見直す。1 枚積むと下の札が最上段に出るため、単純な一巡では
	// 取りこぼす。
	for {
		h := a.GetHint()
		if h == nil {
			break
		}
		if err := a.PlayWasteToFoundation(h.WasteIdx, h.FoundationIdx); err != nil {
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
func (a *AuldLangSyne) Undo() error {
	if len(a.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能か
func (a *AuldLangSyne) CanUndo() bool { return len(a.history) > 0 }

// UndoN n 手戻す
func (a *AuldLangSyne) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(a.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := a.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。
// 膠着状態でなければ 0、履歴を遡っても抜けられなければ -1。
func (a *AuldLangSyne) UndoToEscape() int {
	return undoToEscape(a.isStalemate, a.history, func(s *auldLangSyneSnapshot) bool { return s.isStalemate })
}

// AllFaceUp すべてのカードが見えているか。
// オールド・ラング・サインは伏せ札を持たず、ストックを配り切った時点で全札が可視になる。
func (a *AuldLangSyne) AllFaceUp() bool { return len(a.stock) == 0 }

// GetPhase フェーズ取得
func (a *AuldLangSyne) GetPhase() AuldLangSynePhase { return a.phase }

// GetMoveCount 手数取得
func (a *AuldLangSyne) GetMoveCount() int { return a.moveCount }

// GetStockCount ストック残枚数取得
func (a *AuldLangSyne) GetStockCount() int { return len(a.stock) }

// GetWastes ウェイスト取得
func (a *AuldLangSyne) GetWastes() [AuldLangSyneWasteCnt][]*Card { return a.wastes }

// GetFoundations ファンデーション取得
func (a *AuldLangSyne) GetFoundations() [AuldLangSyneFoundationCnt][]*Card { return a.foundations }

// GetGameEndFlag ゲーム終了フラグ
func (a *AuldLangSyne) GetGameEndFlag() bool { return a.phase != AuldLangSynePhasePlaying }

// IsStalemate 手詰まりか
func (a *AuldLangSyne) IsStalemate() bool { return a.isStalemate }

// --- Private helpers ---

// dealRow ストックから各ウェイストへ 1 枚ずつ配り、配った札を返す。
// 48 は 4 の倍数なので通常は 4 枚配り切るが、壊れた KV スナップショットから
// 復元した場合に備えてストックが尽きたら途中で止める。
func (a *AuldLangSyne) dealRow() []*Card {
	dealt := make([]*Card, 0, AuldLangSyneWasteCnt)
	for i := range AuldLangSyneWasteCnt {
		if len(a.stock) == 0 {
			break
		}
		card := a.stock[len(a.stock)-1]
		a.stock = a.stock[:len(a.stock)-1]
		a.wastes[i] = append(a.wastes[i], card)
		dealt = append(dealt, card)
	}
	return dealt
}

// canPlaceOnFoundation ファンデーションにカードを置けるか判定。
// エースは最初から据わっているので空の列は通常存在しないが、壊れた KV
// スナップショットから復元した場合に備えて空列はエースのみ受け付ける。
// それ以外は最上段の 1 つ上のランクのみ。スートは問わない。
func (a *AuldLangSyne) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := a.foundations[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	if len(pile) >= CardValueMax {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()+1
}

// findFoundation カードを置けるファンデーションのインデックスを探す（無ければ -1）
func (a *AuldLangSyne) findFoundation(card *Card) int {
	for i := range AuldLangSyneFoundationCnt {
		if a.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// checkGameClear ゲームクリア判定
func (a *AuldLangSyne) checkGameClear() {
	for i := range AuldLangSyneFoundationCnt {
		if len(a.foundations[i]) != CardValueMax {
			return
		}
	}
	a.phase = AuldLangSynePhaseGameClear
}

// checkStalemate 手詰まり判定。
// ストックが残っている限り配り直せるので手詰まりにはならない。
func (a *AuldLangSyne) checkStalemate() {
	if a.phase != AuldLangSynePhasePlaying {
		return
	}
	if len(a.stock) > 0 {
		a.isStalemate = false
		return
	}
	a.isStalemate = a.GetHint() == nil
}

// takeSnapshot 現在の状態をスナップショットとして保存
func (a *AuldLangSyne) takeSnapshot() {
	snap := &auldLangSyneSnapshot{
		phase:       a.phase,
		moveCount:   a.moveCount,
		isStalemate: a.isStalemate,
	}
	for i := range AuldLangSyneFoundationCnt {
		snap.foundations[i] = make([]*Card, len(a.foundations[i]))
		copy(snap.foundations[i], a.foundations[i])
	}
	for i := range AuldLangSyneWasteCnt {
		snap.wastes[i] = make([]*Card, len(a.wastes[i]))
		copy(snap.wastes[i], a.wastes[i])
	}
	snap.stock = make([]*Card, len(a.stock))
	copy(snap.stock, a.stock)
	a.history = appendSnapshot(a.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (a *AuldLangSyne) restoreSnapshot(snap *auldLangSyneSnapshot) {
	a.foundations = snap.foundations
	a.wastes = snap.wastes
	a.stock = snap.stock
	a.phase = snap.phase
	a.moveCount = snap.moveCount
	a.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (a *AuldLangSyne) appendLog(actionType, detail string, cards []*Card) {
	a.appendLogAt(a.moveCount, 0, actionType, detail, cards)
}

// auldLangSyneMaxSliceLen caps slice sizes during deserialisation.
const auldLangSyneMaxSliceLen = 1000

// auldLangSyneSnapshotJSON is the wire format for a single undo snapshot.
// auldLangSyneSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type auldLangSyneSnapshotJSON struct {
	Foundations [AuldLangSyneFoundationCnt][]*Card `json:"fs"`
	Wastes      [AuldLangSyneWasteCnt][]*Card      `json:"ws"`
	Stock       []*Card                            `json:"st"`
	Phase       AuldLangSynePhase                  `json:"ps"`
	MoveCount   int                                `json:"mc"`
	IsStalemate bool                               `json:"sl"`
}

// MarshalJSON implements json.Marshaler for auldLangSyneSnapshot.
func (s *auldLangSyneSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(auldLangSyneSnapshotJSON{
		Foundations: s.foundations,
		Wastes:      s.wastes,
		Stock:       s.stock,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for auldLangSyneSnapshot.
func (s *auldLangSyneSnapshot) UnmarshalJSON(data []byte) error {
	var j auldLangSyneSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > auldLangSyneMaxSliceLen {
		return errors.New("auldlangsyne: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundations {
		if len(pile) > auldLangSyneMaxSliceLen {
			return errors.New("auldlangsyne: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Wastes {
		if len(pile) > auldLangSyneMaxSliceLen {
			return errors.New("auldlangsyne: snapshot pile exceeds maximum allowed size")
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

// auldLangSyneJSON is the JSON wire format for AuldLangSyne.
type auldLangSyneJSON struct {
	TrumpCards  *TrumpCards                        `json:"tc"`
	Foundations [AuldLangSyneFoundationCnt][]*Card `json:"fd"`
	Wastes      [AuldLangSyneWasteCnt][]*Card      `json:"wa"`
	Stock       []*Card                            `json:"st"`
	Phase       AuldLangSynePhase                  `json:"ps"`
	MoveCount   int                                `json:"mc"`
	ActionLog   []*ActionLogEntry                  `json:"al"`
	IsStalemate bool                               `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*auldLangSyneSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (a *AuldLangSyne) MarshalJSON() ([]byte, error) {
	return json.Marshal(&auldLangSyneJSON{
		TrumpCards:  a.trumpCards,
		Foundations: a.foundations,
		Wastes:      a.wastes,
		Stock:       a.stock,
		Phase:       a.phase,
		MoveCount:   a.moveCount,
		ActionLog:   a.actionLog,
		IsStalemate: a.isStalemate,
		History:     a.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。
// 値域を検証するのは、KV に入っているのは以前のバージョンが書いた任意のバイト列で
// あり、壊れた状態でゲームを開始させないため。
func (a *AuldLangSyne) UnmarshalJSON(data []byte) error {
	var j auldLangSyneJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > auldLangSyneMaxSliceLen || len(j.History) > auldLangSyneMaxSliceLen {
		return errors.New("auldlangsyne: input array exceeds maximum allowed size")
	}
	if j.Phase < AuldLangSynePhasePlaying || j.Phase > AuldLangSynePhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	for i := range AuldLangSyneFoundationCnt {
		if len(j.Foundations[i]) > CardValueMax {
			return fmt.Errorf("foundation %d has %d cards", i, len(j.Foundations[i]))
		}
	}
	if j.TrumpCards != nil {
		a.trumpCards = j.TrumpCards
	}
	a.foundations = j.Foundations
	a.wastes = j.Wastes
	a.stock = j.Stock
	a.phase = j.Phase
	a.moveCount = j.MoveCount
	a.actionLog = j.ActionLog
	a.isStalemate = j.IsStalemate
	a.history = j.History
	return nil
}
