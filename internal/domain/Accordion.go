//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AccordionPhase アコーディオンゲームフェーズ
type AccordionPhase int

// Accordionのフェーズ定数
const (
	// AccordionPhasePlaying プレイ中
	AccordionPhasePlaying AccordionPhase = iota
	// AccordionPhaseGameClear ゲームクリア
	AccordionPhaseGameClear
	// AccordionPhaseGameOver ゲームオーバー
	AccordionPhaseGameOver
)

// AccordionPileCnt 初期の山数 (52 = 1 deck)
const AccordionPileCnt = CardCnt

// AccordionWinPileCnt クリア時の山数
const AccordionWinPileCnt = 1

// AccordionOffsetOne 左隣 (offset 1) への移動
const AccordionOffsetOne = 1

// AccordionOffsetThree 3つ左 (offset 3) への移動
const AccordionOffsetThree = 3

// AccordionHint ヒント。FromIdx のパイルを ToIdx のパイルに重ねる手を表す。
type AccordionHint struct {
	FromIdx int
	ToIdx   int
}

// Accordion アコーディオンゲームクラス
type Accordion struct {
	trumpCards *TrumpCards
	piles      [][]*Card
	phase      AccordionPhase
	moveCount  int
	actionLogBase
	history     []*accordionSnapshot
	isStalemate bool
}

// accordionSnapshot アンドゥ用スナップショット
type accordionSnapshot struct {
	piles       [][]*Card
	phase       AccordionPhase
	moveCount   int
	isStalemate bool
}

// NewAccordion コンストラクタ
func NewAccordion(trumpCards *TrumpCards) *Accordion {
	return &Accordion{trumpCards: trumpCards}
}

// NewDefaultAccordion 標準1デッキのAccordionを返す
func NewDefaultAccordion() *Accordion {
	return NewAccordion(NewTrumpCards(0))
}

// Reset ゲーム初期化
func (a *Accordion) Reset() {
	a.trumpCards.Shuffle()
	a.phase = AccordionPhasePlaying
	a.moveCount = 0
	a.actionLog = nil
	a.history = nil
	a.isStalemate = false

	a.piles = make([][]*Card, 0, AccordionPileCnt)
	for a.trumpCards.GetRemainingCount() > 0 {
		card := a.trumpCards.DrawCard()
		a.piles = append(a.piles, []*Card{card})
	}

	a.checkAccordionStalemate()
}

// Move fromIdx のパイルを toIdx のパイルに重ねる
func (a *Accordion) Move(fromIdx, toIdx int) error {
	if a.phase != AccordionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromIdx < 0 || fromIdx >= len(a.piles) {
		return errors.New("invalid from index")
	}
	if toIdx < 0 || toIdx >= len(a.piles) {
		return errors.New("invalid to index")
	}
	offset := fromIdx - toIdx
	if offset != AccordionOffsetOne && offset != AccordionOffsetThree {
		return errors.New("can only move 1 or 3 positions to the left")
	}
	if !a.canMerge(fromIdx, toIdx) {
		return errors.New("cards do not match in rank or suit")
	}

	a.takeSnapshot()
	a.piles[toIdx] = append(a.piles[toIdx], a.piles[fromIdx]...)
	// Nil out the vacated tail slot before truncating so the removed pile's
	// card slice can be GC'd. Without this, the backing array keeps the old
	// inner slice alive for the lifetime of a.piles.
	copy(a.piles[fromIdx:], a.piles[fromIdx+1:])
	a.piles[len(a.piles)-1] = nil
	a.piles = a.piles[:len(a.piles)-1]
	a.moveCount++
	top := a.piles[toIdx][len(a.piles[toIdx])-1]
	a.appendLog("move", fmt.Sprintf("パイル%d→パイル%d", fromIdx, toIdx), []*Card{top})
	a.checkGameClear()
	a.checkAccordionStalemate()
	return nil
}

// AutoComplete はヒントが示す手を、手が尽きるまで繰り返す (#4793)。
//
// **Web は同じことを1クリックでやっている**のに、ネイティブ CUI には
// オートコンプリートが存在せず、同じ手を延々と打たされていた。姉妹の Wasp は
// ac/autocomplete を公開している。
//
// **手を選ぶ規則は GetHint と共有する。**別実装だと、ヒントが勧める手と自動で
// 打たれる手が食い違う。
//
// 1手も動かせなければエラー。押しても何も起きないより、押せない理由が返る
// ほうがよい。
func (a *Accordion) AutoComplete() error {
	if a.phase != AccordionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := 0
	// **上限を置く。**Move はパイルを1つ減らすので理屈の上では停まるが、
	// 規則を変えたときに無限ループへ落ちない保険。
	for range len(a.piles) {
		hint := a.GetHint()
		if hint == nil {
			break
		}
		if err := a.Move(hint.FromIdx, hint.ToIdx); err != nil {
			break
		}
		moved++
	}
	if moved == 0 {
		return errors.New("no automatic move is available")
	}
	return nil
}

// GiveUp ギブアップ
func (a *Accordion) GiveUp() {
	if a.phase == AccordionPhasePlaying {
		a.phase = AccordionPhaseGameOver
		a.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (a *Accordion) GetHint() *AccordionHint {
	if a.phase != AccordionPhasePlaying {
		return nil
	}
	// 優先度1: offset=3 の移動 (選択肢を温存できる重要な手)
	// 優先度2: offset=1 の移動
	for _, offset := range []int{AccordionOffsetThree, AccordionOffsetOne} {
		for from := offset; from < len(a.piles); from++ {
			to := from - offset
			if a.canMerge(from, to) {
				return &AccordionHint{FromIdx: from, ToIdx: to}
			}
		}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (a *Accordion) Undo() error {
	if a.phase != AccordionPhasePlaying {
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

// CanUndo アンドゥ可能か
func (a *Accordion) CanUndo() bool {
	return len(a.history) > 0 && a.phase == AccordionPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (a *Accordion) UndoToEscape() int {
	return undoToEscape(a.isStalemate, a.history, func(s *accordionSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する
func (a *Accordion) UndoN(n int) error {
	return undoN(a, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (a *Accordion) GetPhase() AccordionPhase { return a.phase }

// SetPhase フェーズ設定 (テスト用)
func (a *Accordion) SetPhase(phase AccordionPhase) { a.phase = phase }

// GetMoveCount 移動回数取得
func (a *Accordion) GetMoveCount() int { return a.moveCount }

// GetPiles パイル一覧取得
func (a *Accordion) GetPiles() [][]*Card { return a.piles }

// GetPileCount 残りパイル数取得
func (a *Accordion) GetPileCount() int { return len(a.piles) }

// GetGameEndFlag returns true once the game has left the playing phase.
func (a *Accordion) GetGameEndFlag() bool { return a.phase != AccordionPhasePlaying }

// IsStalemate 手詰まり状態取得
func (a *Accordion) IsStalemate() bool { return a.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (a *Accordion) SetIsStalemate(v bool) { a.isStalemate = v }

// SetPiles パイル設定 (テスト用)
func (a *Accordion) SetPiles(piles [][]*Card) { a.piles = piles }

// --- Private helpers ---

// canMerge fromIdx と toIdx の山を合流できるか (スートまたはランクが一致するか)
func (a *Accordion) canMerge(fromIdx, toIdx int) bool {
	fromPile := a.piles[fromIdx]
	toPile := a.piles[toIdx]
	if len(fromPile) == 0 || len(toPile) == 0 {
		return false
	}
	fromTop := fromPile[len(fromPile)-1]
	toTop := toPile[len(toPile)-1]
	return fromTop.GetDesign() == toTop.GetDesign() || fromTop.GetValue() == toTop.GetValue()
}

// checkGameClear ゲームクリア判定
func (a *Accordion) checkGameClear() {
	if len(a.piles) == AccordionWinPileCnt {
		a.phase = AccordionPhaseGameClear
	}
}

// checkAccordionStalemate 手詰まり判定
func (a *Accordion) checkAccordionStalemate() {
	if a.phase != AccordionPhasePlaying {
		return
	}
	if a.GetHint() != nil {
		a.isStalemate = false
		return
	}
	// プレイ中に有効手がない。クリアしていないなら膠着状態。
	a.isStalemate = len(a.piles) > AccordionWinPileCnt
}

// takeSnapshot 現在の状態をスナップショットとして保存
//
// actionLog は意図的に含めない。他のソリティア(Scorpion/Klondike 等)と同様に
// 棋譜は追記専用として扱い、Undo で巻き戻されない。プレゼンターは Playing
// フェーズの間 棋譜を非表示にするため、ユーザーには「undo で消えた手」が
// 露出しない。
func (a *Accordion) takeSnapshot() {
	snap := &accordionSnapshot{
		phase:       a.phase,
		moveCount:   a.moveCount,
		isStalemate: a.isStalemate,
	}
	snap.piles = make([][]*Card, len(a.piles))
	for i, pile := range a.piles {
		snap.piles[i] = make([]*Card, len(pile))
		copy(snap.piles[i], pile)
	}
	a.history = append(a.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (a *Accordion) restoreSnapshot(snap *accordionSnapshot) {
	a.piles = snap.piles
	a.phase = snap.phase
	a.moveCount = snap.moveCount
	a.isStalemate = snap.isStalemate
}

// appendLog 棋譜エントリを追加
func (a *Accordion) appendLog(actionType, detail string, cards []*Card) {
	a.appendLogAt(a.moveCount, 0, actionType, detail, cards)
}

// accordionJSON is the JSON wire format for Accordion.
type accordionJSON struct {
	TrumpCards  *TrumpCards          `json:"tc"`
	Piles       [][]*Card            `json:"pl"`
	Phase       AccordionPhase       `json:"ps"`
	MoveCount   int                  `json:"mc"`
	ActionLog   []*ActionLogEntry    `json:"al"`
	IsStalemate bool                 `json:"sl"`
	History     []*accordionSnapshot `json:"hi,omitempty"`
}

// accordionSnapshotJSON is the wire format for a single undo snapshot.
// accordionSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// accordionJSON's short keys to keep the KV payload compact (#1654).
type accordionSnapshotJSON struct {
	Piles       [][]*Card      `json:"pl"`
	Phase       AccordionPhase `json:"ps"`
	MoveCount   int            `json:"mc"`
	IsStalemate bool           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for accordionSnapshot.
func (s *accordionSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(accordionSnapshotJSON{
		Piles:       s.piles,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for accordionSnapshot.
func (s *accordionSnapshot) UnmarshalJSON(data []byte) error {
	var j accordionSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Piles) > accordionMaxSliceLen {
		return fmt.Errorf("accordion: snapshot piles exceeds maximum allowed size")
	}
	for _, pile := range j.Piles {
		if len(pile) > accordionMaxSliceLen {
			return fmt.Errorf("accordion: snapshot pile exceeds maximum allowed size")
		}
	}
	s.piles = j.Piles
	if s.piles == nil {
		s.piles = make([][]*Card, 0)
	}
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON implements json.Marshaler.
func (a *Accordion) MarshalJSON() ([]byte, error) {
	return json.Marshal(accordionJSON{
		TrumpCards:  a.trumpCards,
		Piles:       a.piles,
		Phase:       a.phase,
		MoveCount:   a.moveCount,
		ActionLog:   a.actionLog,
		IsStalemate: a.isStalemate,
		History:     a.history,
	})
}

// accordionMaxSliceLen caps slice sizes during deserialisation. A real game
// has at most CardCnt (52) piles, each holding at most CardCnt cards, and the
// action log grows by one entry per move. 200 gives a comfortable margin while
// still bounding unbounded allocation from a hostile payload.
const accordionMaxSliceLen = 200

// UnmarshalJSON implements json.Unmarshaler.
func (a *Accordion) UnmarshalJSON(data []byte) error {
	var j accordionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Piles) > accordionMaxSliceLen || len(j.ActionLog) > accordionMaxSliceLen ||
		len(j.History) > accordionMaxSliceLen {
		return fmt.Errorf("accordion: input array exceeds maximum allowed size")
	}
	for _, pile := range j.Piles {
		if len(pile) > accordionMaxSliceLen {
			return fmt.Errorf("accordion: input array exceeds maximum allowed size")
		}
	}
	a.trumpCards = j.TrumpCards
	if a.trumpCards == nil {
		a.trumpCards = NewTrumpCards(0)
	}
	a.piles = j.Piles
	if a.piles == nil {
		a.piles = make([][]*Card, 0)
	}
	a.phase = j.Phase
	a.moveCount = j.MoveCount
	a.actionLog = j.ActionLog
	if a.actionLog == nil {
		a.actionLog = make([]*ActionLogEntry, 0)
	}
	a.history = j.History
	if a.history == nil {
		a.history = make([]*accordionSnapshot, 0)
	}
	a.isStalemate = j.IsStalemate
	return nil
}
