//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// NarcoticPhase ナルコティック（Perpetual Motion）ゲームフェーズ
type NarcoticPhase int

// Narcoticのフェーズ定数
const (
	// NarcoticPhasePlaying プレイ中
	NarcoticPhasePlaying NarcoticPhase = iota
	// NarcoticPhaseGameClear ゲームクリア
	NarcoticPhaseGameClear
	// NarcoticPhaseGameOver ゲームオーバー
	NarcoticPhaseGameOver
)

// NarcoticColCnt 場札の列数
const NarcoticColCnt = 4

// narcoticMaxSliceLen デシリアライズ時のスライス長上限
const narcoticMaxSliceLen = 1000

// NarcoticHint ヒント
type NarcoticHint struct {
	// Type は "remove" / "move" / "draw" のいずれか
	Type string
	// Col は対象の列番号（draw のときは -1）
	Col int
}

// Narcotic ナルコティック（Perpetual Motion）ゲームクラス
type Narcotic struct {
	trumpCards *TrumpCards
	columns    [NarcoticColCnt][]*Card
	stock      []*Card
	discard    []*Card
	phase      NarcoticPhase
	moveCount  int
	actionLogBase
	history     []*narcoticSnapshot
	isStalemate bool
	redealCount int
	// seen は既に現れた盤面の指紋。**再配りに上限が無いので、これが唯一の
	// 終了保証。**同じ盤面が二度現れ、かつ合法手が無ければ、あとは同じ手順を
	// 永遠に繰り返すだけなのでそこで詰みとする。
	seen map[string]bool
}

// narcoticSnapshot アンドゥ用スナップショット
type narcoticSnapshot struct {
	columns     [NarcoticColCnt][]*Card
	stock       []*Card
	discard     []*Card
	phase       NarcoticPhase
	moveCount   int
	isStalemate bool
	redealCount int
}

// NewNarcotic コンストラクタ
func NewNarcotic(trumpCards *TrumpCards) *Narcotic {
	return &Narcotic{
		trumpCards: trumpCards,
	}
}

// NewDefaultNarcotic returns Narcotic with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultNarcotic() *Narcotic {
	return NewNarcotic(NewTrumpCards(0))
}

// Reset ゲームリセット
func (a *Narcotic) Reset() {
	a.trumpCards.Shuffle()
	a.phase = NarcoticPhasePlaying
	a.moveCount = 0
	a.actionLog = nil
	a.history = nil
	a.isStalemate = false
	a.discard = nil

	// 4列を空に初期化し、各列に1枚ずつ配る
	for c := range NarcoticColCnt {
		a.columns[c] = nil
	}
	for c := range NarcoticColCnt {
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
func (a *Narcotic) Draw() error {
	if a.phase != NarcoticPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(a.stock) == 0 {
		return errors.New("no cards in stock")
	}
	a.takeSnapshot()
	dealt := make([]*Card, 0, NarcoticColCnt)
	for c := range NarcoticColCnt {
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

// Remove は露出している4枚のランクが揃っているとき、その4枚を捨てる。
//
// **列を指定しない。**クローン元の Aces Up は「同スートで下位の1枚」を捨てるので
// どの列かを言う必要があるが、Narcotic は4枚まとめてしか捨てられないので、
// 指定できる自由度が無い。
func (a *Narcotic) Remove() error {
	if a.phase != NarcoticPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if !a.canRemoveSet() {
		return errors.New("the four exposed cards do not share a rank")
	}
	a.takeSnapshot()
	removed := make([]*Card, 0, NarcoticColCnt)
	for c := range NarcoticColCnt {
		removed = append(removed, a.popTop(c))
	}
	a.discard = append(a.discard, removed...)
	a.moveCount++
	a.appendLog("remove", "4枚のランクが揃ったので取り除きました", removed)
	a.checkGameClear()
	a.checkStalemate()
	return nil
}

// Move は列 col の露出札を、**同じランクの札を露出している最も左の列**へ重ねる。
//
// **空き列へは動かさない。**クローン元の Aces Up は空き列を作業スペースとして
// 使うが、Narcotic の移動は「重複を1つにまとめる」ためだけにあり、行き先は
// 盤面から一意に決まる。
func (a *Narcotic) Move(col int) error {
	if a.phase != NarcoticPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	dest := a.moveTarget(col)
	if dest < 0 {
		return errors.New("no column to the left exposes the same rank")
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

// Redeal は山札が尽きたとき、場の札を集めて配り直す。
//
// **右の列から左へ順に重ね、シャッフルしない。**したがって次の山札は現在の盤面
// から一意に決まる ── 運で変わる要素は無く、だからこそ同じ盤面が再び現れたら
// それは本物のループになる。回数に制限は無い。
func (a *Narcotic) Redeal() error {
	if a.phase != NarcoticPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(a.stock) > 0 {
		return errors.New("cannot redeal while the stock still has cards")
	}
	if a.remainingOnTable() == 0 {
		return errors.New("no cards left to gather")
	}
	a.takeSnapshot()
	gathered := make([]*Card, 0, CardCnt)
	for c := NarcoticColCnt - 1; c >= 0; c-- {
		gathered = append(gathered, a.columns[c]...)
		a.columns[c] = nil
	}
	a.stock = gathered
	a.redealCount++
	a.moveCount++
	a.appendLog("redeal", fmt.Sprintf("集めて配り直しました (%d回目)", a.redealCount), nil)
	a.checkStalemate()
	return nil
}

// GetRedealCount は配り直した回数を返す。上限は無い。
func (a *Narcotic) GetRedealCount() int { return a.redealCount }

// remainingOnTable は場に残っている札の枚数。
func (a *Narcotic) remainingOnTable() int {
	n := 0
	for c := range NarcoticColCnt {
		n += len(a.columns[c])
	}
	return n
}

// GiveUp ギブアップ
func (a *Narcotic) GiveUp() {
	if a.phase == NarcoticPhasePlaying {
		a.phase = NarcoticPhaseGameOver
		a.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (a *Narcotic) GetHint() *NarcoticHint {
	if a.phase != NarcoticPhasePlaying {
		return nil
	}
	// **4枚揃いが最優先。**捨てられるものを残して先に進む理由は無い。
	if a.canRemoveSet() {
		return &NarcoticHint{Type: "remove", Col: -1}
	}
	// 重ねられる札があれば重ねる。露出が減って揃う目が出る。
	if col := a.firstMovableCol(); col >= 0 {
		return &NarcoticHint{Type: "move", Col: col}
	}
	// 山札があれば配る。
	if len(a.stock) > 0 {
		return &NarcoticHint{Type: "draw", Col: -1}
	}
	// 山札が尽きても、場に札が残っていれば集めて配り直せる。
	if a.remainingOnTable() > 0 && !a.isStalemate {
		return &NarcoticHint{Type: "redeal", Col: -1}
	}
	return nil
}

// Undo 直前の操作を取り消す
func (a *Narcotic) Undo() error {
	if a.phase != NarcoticPhasePlaying {
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
func (a *Narcotic) CanUndo() bool {
	return len(a.history) > 0 && a.phase == NarcoticPhasePlaying
}

// UndoToEscape 膠着状態から抜けるために必要なアンドゥ回数を返す。膠着状態でなければ0、脱出不可なら-1。
func (a *Narcotic) UndoToEscape() int {
	return undoToEscape(a.isStalemate, a.history, func(s *narcoticSnapshot) bool { return s.isStalemate })
}

// UndoN n回連続でアンドゥを実行する。
func (a *Narcotic) UndoN(n int) error {
	return undoN(a, n)
}

// --- State getters/setters ---

// GetPhase フェーズ取得
func (a *Narcotic) GetPhase() NarcoticPhase { return a.phase }

// SetPhase フェーズ設定 (テスト用)
func (a *Narcotic) SetPhase(phase NarcoticPhase) { a.phase = phase }

// GetMoveCount 移動回数取得
func (a *Narcotic) GetMoveCount() int { return a.moveCount }

// GetStockCount ストック枚数取得
func (a *Narcotic) GetStockCount() int { return len(a.stock) }

// GetDiscardCount 除去済み枚数取得
func (a *Narcotic) GetDiscardCount() int { return len(a.discard) }

// GetDiscardTop 捨て札の一番上（直近に除去した札）を返す。捨て札が空なら nil。
func (a *Narcotic) GetDiscardTop() *Card {
	return discardTop(a.discard)
}

// GetColumns 場札の列を取得
func (a *Narcotic) GetColumns() [NarcoticColCnt][]*Card { return a.columns }

// GetGameEndFlag returns true once the game has left the playing phase.
func (a *Narcotic) GetGameEndFlag() bool { return a.phase != NarcoticPhasePlaying }

// IsStalemate 手詰まり状態取得
func (a *Narcotic) IsStalemate() bool { return a.isStalemate }

// SetIsStalemate 手詰まり状態設定 (テスト用)
func (a *Narcotic) SetIsStalemate(v bool) { a.isStalemate = v }

// SetColumns 場札の列設定 (テスト用)
func (a *Narcotic) SetColumns(columns [NarcoticColCnt][]*Card) { a.columns = columns }

// SetStock ストック設定 (テスト用)
func (a *Narcotic) SetStock(stock []*Card) { a.stock = stock }

// --- Private helpers ---

// topCard 指定列の一番上のカードを返す (空なら nil)
func (a *Narcotic) topCard(col int) *Card {
	if col < 0 || col >= NarcoticColCnt || len(a.columns[col]) == 0 {
		return nil
	}
	return a.columns[col][len(a.columns[col])-1]
}

// popTop 指定列の一番上のカードを取り出して返す。
func (a *Narcotic) popTop(col int) *Card {
	top := a.columns[col][len(a.columns[col])-1]
	a.columns[col] = a.columns[col][:len(a.columns[col])-1]
	return top
}

// canRemoveSet は露出している4枚のランクが全て揃っているかを返す。
//
// **これが Narcotic の唯一の除去条件。**クローン元の Aces Up は「同スートで
// 下位の1枚」を捨てるので、そちらの述語を残すと全く違うゲームになる。
// 4列のどれかが空なら、揃えようが無いので false。
func (a *Narcotic) canRemoveSet() bool {
	first := a.topCard(0)
	if first == nil {
		return false
	}
	for c := 1; c < NarcoticColCnt; c++ {
		other := a.topCard(c)
		if other == nil || other.GetValue() != first.GetValue() {
			return false
		}
	}
	return true
}

// CanRemoveSet は露出4枚が揃っているかを外部に公開する。
func (a *Narcotic) CanRemoveSet() bool { return a.canRemoveSet() }

// moveTarget は列 col の露出札を重ねられる列を返す (-1 = 無し)。
//
// 行き先は「同じランクを露出している**最も左**の列」で、自分より左に限る。
// 右へ動かせてしまうと同じ2枚を往復させられ、手数だけが増える。
func (a *Narcotic) moveTarget(col int) int {
	top := a.topCard(col)
	if top == nil {
		return -1
	}
	for c := range col {
		other := a.topCard(c)
		if other != nil && other.GetValue() == top.GetValue() {
			return c
		}
	}
	return -1
}

// CanMove は列 col の露出札を重ねられるかを返す。
func (a *Narcotic) CanMove(col int) bool {
	if col < 0 || col >= NarcoticColCnt {
		return false
	}
	return a.moveTarget(col) >= 0
}

// MoveTarget は列 col の行き先を外部に公開する (-1 = 無し)。
func (a *Narcotic) MoveTarget(col int) int { return a.moveTarget(col) }

// firstMovableCol は重ねられる最初の列を返す (-1 = なし)。
func (a *Narcotic) firstMovableCol() int {
	for c := range NarcoticColCnt {
		if a.moveTarget(c) >= 0 {
			return c
		}
	}
	return -1
}

// checkGameClear は全 52 枚を捨て切ったかを判定する。
func (a *Narcotic) checkGameClear() {
	if len(a.discard) >= CardCnt {
		a.phase = NarcoticPhaseGameClear
	}
}

// checkStalemate は手詰まりを判定する。
//
// **再配りに上限が無いので「山札が尽きた」では終わらない。**終わるのは、
//  1. 4枚揃いも重ねる手も無く、山札も空で、集める札も無い（＝物理的に不可能）か、
//  2. **同じ盤面が二度現れ、かつ今すぐ指せる手が無い**とき。
//
// 2 は再配りが決定的（無シャッフル・右から左）だからこそ成り立つ ── 同じ盤面から
// は必ず同じ展開になるので、以後は永久に同じ循環をたどる。
func (a *Narcotic) checkStalemate() {
	if a.phase != NarcoticPhasePlaying {
		return
	}
	if a.canRemoveSet() || a.firstMovableCol() >= 0 {
		a.isStalemate = false
		a.rememberBoard()
		return
	}
	// 指せる手は無い。配れる/集められるなら、まだ進展の余地がある。
	if len(a.stock) > 0 || a.remainingOnTable() > 0 {
		a.isStalemate = a.boardSeenBefore()
		a.rememberBoard()
		return
	}
	a.isStalemate = true
}

// CheckNarcoticStalemate は外部から手詰まり判定を再評価する公開ラッパー。
// SetColumns/SetStock で盤を組み立てたあとに呼ぶ。
func (a *Narcotic) CheckNarcoticStalemate() { a.checkStalemate() }

// SetDiscardCount は捨て札の枚数を設定する (テスト用)。
func (a *Narcotic) SetDiscardCount(n int) {
	a.discard = make([]*Card, n)
	for i := range a.discard {
		a.discard[i] = NewCard(CardDesignSpade, 1, true)
	}
}

// rememberBoard は現在の盤面を既出として記録する。
func (a *Narcotic) rememberBoard() {
	if a.seen == nil {
		a.seen = map[string]bool{}
	}
	a.seen[a.boardFingerprint()] = true
}

// boardSeenBefore は現在の盤面が既に現れたことがあるかを返す。
func (a *Narcotic) boardSeenBefore() bool {
	return a.seen != nil && a.seen[a.boardFingerprint()]
}

// boardFingerprint は盤面 + 山札の指紋。ループ検出に使う。
//
// 再配りはシャッフルしないので、**盤面と山札が一致すれば以後の展開も完全に
// 一致する。**だから指紋の再出現は本物のループを意味する。
func (a *Narcotic) boardFingerprint() string {
	var b strings.Builder
	for c := range NarcoticColCnt {
		for _, card := range a.columns[c] {
			fmt.Fprintf(&b, "%d.%d,", card.GetDesign(), card.GetValue())
		}
		b.WriteByte('|')
	}
	b.WriteByte('#')
	for _, card := range a.stock {
		fmt.Fprintf(&b, "%d.%d,", card.GetDesign(), card.GetValue())
	}
	return b.String()
}

func (a *Narcotic) takeSnapshot() {
	snap := &narcoticSnapshot{
		stock:       narcoticCloneCards(a.stock),
		discard:     narcoticCloneCards(a.discard),
		phase:       a.phase,
		moveCount:   a.moveCount,
		isStalemate: a.isStalemate,
		redealCount: a.redealCount,
	}
	for c := range NarcoticColCnt {
		snap.columns[c] = narcoticCloneCards(a.columns[c])
	}
	a.history = appendSnapshot(a.history, snap)
}

// restoreSnapshot スナップショットから状態を復元
func (a *Narcotic) restoreSnapshot(snap *narcoticSnapshot) {
	a.columns = snap.columns
	a.stock = snap.stock
	a.discard = snap.discard
	a.phase = snap.phase
	a.moveCount = snap.moveCount
	a.isStalemate = snap.isStalemate
	a.redealCount = snap.redealCount
	// **seen は巻き戻さない。**Undo で「まだ見ていないこと」にしてしまうと、
	// 同じ循環を何度でも再訪できてしまい、ループ検出が意味を失う。
}

// cloneCards はカードスライスの浅いコピーを返す。
func narcoticCloneCards(src []*Card) []*Card {
	dst := make([]*Card, len(src))
	copy(dst, src)
	return dst
}

// appendLog 棋譜エントリを追加
func (a *Narcotic) appendLog(actionType, detail string, cards []*Card) {
	a.appendLogAt(a.moveCount, 0, actionType, detail, cards)
}

// narcoticJSON is the JSON wire format for Narcotic.
type narcoticJSON struct {
	TrumpCards  *TrumpCards             `json:"tc"`
	Columns     [NarcoticColCnt][]*Card `json:"co"`
	Stock       []*Card                 `json:"st"`
	Discard     []*Card                 `json:"di"`
	Phase       NarcoticPhase           `json:"ps"`
	MoveCount   int                     `json:"mc"`
	ActionLog   []*ActionLogEntry       `json:"al"`
	IsStalemate bool                    `json:"sm"`
	RedealCount int                     `json:"rd"`
	// Seen は既出の盤面指紋。**Worker はリクエストごとに作り直すので、これを
	// 載せないとループ検出の記憶が毎回消え、永久に詰まなくなる。**
	Seen    []string            `json:"sn,omitempty"`
	History []*narcoticSnapshot `json:"hi,omitempty"`
}

// narcoticSnapshotJSON is the wire format for a single undo snapshot.
// narcoticSnapshot uses unexported fields, so we project to/from this shape
// with explicit Marshal/Unmarshal methods. Field names match narcoticJSON's
// short keys to keep the KV payload compact (#1654).
type narcoticSnapshotJSON struct {
	Columns     [NarcoticColCnt][]*Card `json:"co"`
	Stock       []*Card                 `json:"st"`
	Discard     []*Card                 `json:"di"`
	Phase       NarcoticPhase           `json:"ps"`
	MoveCount   int                     `json:"mc"`
	IsStalemate bool                    `json:"sm"`
	RedealCount int                     `json:"rd"`
}

// MarshalJSON implements json.Marshaler for narcoticSnapshot.
func (s *narcoticSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(narcoticSnapshotJSON{
		Columns:     s.columns,
		Stock:       s.stock,
		Discard:     s.discard,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
		RedealCount: s.redealCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for narcoticSnapshot.
func (s *narcoticSnapshot) UnmarshalJSON(data []byte) error {
	var j narcoticSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > narcoticMaxSliceLen || len(j.Discard) > narcoticMaxSliceLen {
		return fmt.Errorf("narcotic: snapshot array exceeds maximum allowed size")
	}
	for c := range NarcoticColCnt {
		if len(j.Columns[c]) > narcoticMaxSliceLen {
			return fmt.Errorf("narcotic: snapshot column exceeds maximum allowed size")
		}
	}
	s.columns = j.Columns
	s.stock = narcoticNilToEmptyCards(j.Stock)
	s.discard = narcoticNilToEmptyCards(j.Discard)
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.redealCount = j.RedealCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (a *Narcotic) MarshalJSON() ([]byte, error) {
	return json.Marshal(narcoticJSON{
		TrumpCards:  a.trumpCards,
		Columns:     a.columns,
		Stock:       a.stock,
		Discard:     a.discard,
		Phase:       a.phase,
		MoveCount:   a.moveCount,
		ActionLog:   a.actionLog,
		IsStalemate: a.isStalemate,
		RedealCount: a.redealCount,
		Seen:        a.seenKeys(),
		History:     a.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Narcotic) UnmarshalJSON(data []byte) error {
	var j narcoticJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > narcoticMaxSliceLen || len(j.Discard) > narcoticMaxSliceLen ||
		len(j.ActionLog) > narcoticMaxSliceLen || len(j.History) > narcoticMaxSliceLen {
		return fmt.Errorf("narcotic: input array exceeds maximum allowed size")
	}
	for c := range NarcoticColCnt {
		if len(j.Columns[c]) > narcoticMaxSliceLen {
			return fmt.Errorf("narcotic: input column exceeds maximum allowed size")
		}
	}

	a.trumpCards = j.TrumpCards
	if a.trumpCards == nil {
		a.trumpCards = NewTrumpCards(0)
	}
	a.columns = j.Columns
	for c := range NarcoticColCnt {
		a.columns[c] = narcoticNilToEmptyCards(a.columns[c])
	}
	a.stock = narcoticNilToEmptyCards(j.Stock)
	a.discard = narcoticNilToEmptyCards(j.Discard)
	a.phase = j.Phase
	a.moveCount = j.MoveCount
	a.actionLog = j.ActionLog
	if a.actionLog == nil {
		a.actionLog = make([]*ActionLogEntry, 0)
	}
	a.history = j.History
	if a.history == nil {
		a.history = make([]*narcoticSnapshot, 0)
	}
	a.isStalemate = j.IsStalemate
	if j.RedealCount < 0 {
		return fmt.Errorf("narcotic: redeal count out of range: %d", j.RedealCount)
	}
	a.redealCount = j.RedealCount
	if len(j.Seen) > narcoticMaxSliceLen {
		return fmt.Errorf("narcotic: seen-board set exceeds maximum allowed size")
	}
	a.seen = make(map[string]bool, len(j.Seen))
	for _, k := range j.Seen {
		a.seen[k] = true
	}
	return nil
}

// seenKeys は既出盤面の指紋を安定順で返す。
//
// **map をそのまま並べると順序が毎回変わり、同じ状態が別の JSON になる。**
// KV の値が毎リクエスト変わると差分が読めないので、並べ替えてから出す。
func (a *Narcotic) seenKeys() []string {
	if len(a.seen) == 0 {
		return nil
	}
	keys := make([]string, 0, len(a.seen))
	for k := range a.seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// nilToEmptyCards は nil スライスを空スライスに正規化する。
func narcoticNilToEmptyCards(src []*Card) []*Card {
	if src == nil {
		return make([]*Card, 0)
	}
	return src
}
