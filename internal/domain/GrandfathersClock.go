//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// GrandfathersClockPhase グランドファーザーズ・クロックのゲームフェーズ
type GrandfathersClockPhase int

// GrandfathersClockのフェーズ定数
const (
	// GrandfathersClockPhasePlaying プレイ中
	GrandfathersClockPhasePlaying GrandfathersClockPhase = iota
	// GrandfathersClockPhaseGameClear ゲームクリア
	GrandfathersClockPhaseGameClear
	// GrandfathersClockPhaseGameOver ゲームオーバー
	GrandfathersClockPhaseGameOver
)

// GrandfathersClockFoundationCnt 基礎札の数。時計の文字盤 12 個ぶん。
const GrandfathersClockFoundationCnt = 12

// GrandfathersClockTableauCnt タブローの列数
const GrandfathersClockTableauCnt = 8

// GrandfathersClockColumnLen 1 列あたりの配り枚数
const GrandfathersClockColumnLen = 5

// grandfathersClockStarter は文字盤 1 つぶんの初期カードと目標ランク。
type grandfathersClockStarter struct {
	design int
	value  int
}

// grandfathersClockStarters 文字盤に最初から置かれる 12 枚。
//
// インデックス i は「i+1 時」の位置で、その位置の目標ランクは i+1（1 時→A、
// 12 時→Q）。並びはランダムではなく固定で、これがゲームを成立させている：
// 各札は目標ランクまで 3 枚か 4 枚を必要とし、合計がちょうど 40 枚
// （= 52 − 12）になってタブローの 8 列×5 枚と一致する。
var grandfathersClockStarters = [GrandfathersClockFoundationCnt]grandfathersClockStarter{
	{CardDesignHeart, 10},            // 1 時: 10♥ → J,Q,K,A（4 枚）
	{CardDesignSpade, 11},            // 2 時: J♠ → Q,K,A,2（4 枚）
	{CardDesignDiamond, 12},          // 3 時: Q♦ → K,A,2,3（4 枚）
	{CardDesignClover, CardValueMax}, // 4 時: K♣ → A,2,3,4（4 枚）
	{CardDesignHeart, 2},             // 5 時: 2♥ → 3,4,5（3 枚）
	{CardDesignSpade, 3},             // 6 時: 3♠ → 4,5,6（3 枚）
	{CardDesignDiamond, 4},           // 7 時: 4♦ → 5,6,7（3 枚）
	{CardDesignClover, 5},            // 8 時: 5♣ → 6,7,8（3 枚）
	{CardDesignHeart, 6},             // 9 時: 6♥ → 7,8,9（3 枚）
	{CardDesignSpade, 7},             // 10 時: 7♠ → 8,9,10（3 枚）
	{CardDesignDiamond, 8},           // 11 時: 8♦ → 9,10,J（3 枚）
	{CardDesignClover, 9},            // 12 時: 9♣ → 10,J,Q（3 枚）
}

// GrandfathersClockTargetRank 文字盤 idx（0 始まり）の目標ランクを返す。
// 1 時が A、12 時が Q。
func GrandfathersClockTargetRank(idx int) int { return idx + 1 }

// GrandfathersClockTableauCard タブロー上のカード。全札が表向きだが、他の
// ソリティアと同じ形にしておくとプレゼンターを使い回せる。
type GrandfathersClockTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// GrandfathersClockHint グランドファーザーズ・クロックのヒント
type GrandfathersClockHint struct {
	// FromCol 移動元のタブロー列
	FromCol int
	// ToZone 移動先 "foundation" / "tableau"
	ToZone string
	// ToIdx 移動先のインデックス（文字盤またはタブロー列）
	ToIdx int
}

// GrandfathersClock グランドファーザーズ・クロックゲームクラス。
//
// 52 枚 1 組。12 枚が時計の文字盤として最初から置かれ、残り 40 枚が 8 列×5 枚の
// タブローに表向きで配られる。**山札は無い** — issue #4399 の仕様案は「山札が
// 尽きるか」と書いているが、12 + 8×5 = 52 で全札が配り切られるため山札は存在
// しえない。
//
// 文字盤は同スートで昇順（K の次は A に折り返す）に積み、その位置の時刻に対応
// するランクに達したら完成で、それ以上は受け付けない。タブローはスートを無視
// して降順に 1 枚ずつ動かせ、空き列には任意の札を置ける。
type GrandfathersClock struct {
	trumpCards *TrumpCards
	foundation [GrandfathersClockFoundationCnt][]*Card
	tableau    [GrandfathersClockTableauCnt][]*GrandfathersClockTableauCard
	phase      GrandfathersClockPhase
	moveCount  int
	actionLogBase
	history     []*grandfathersClockSnapshot
	isStalemate bool
}

// grandfathersClockSnapshot アンドゥ用スナップショット
type grandfathersClockSnapshot struct {
	foundation  [GrandfathersClockFoundationCnt][]*Card
	tableau     [GrandfathersClockTableauCnt][]*GrandfathersClockTableauCard
	phase       GrandfathersClockPhase
	moveCount   int
	isStalemate bool
}

// NewGrandfathersClock コンストラクタ
func NewGrandfathersClock(trumpCards *TrumpCards) *GrandfathersClock {
	return &GrandfathersClock{trumpCards: trumpCards}
}

// NewDefaultGrandfathersClock returns GrandfathersClock with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultGrandfathersClock() *GrandfathersClock {
	return NewGrandfathersClock(NewTrumpCards(0))
}

// Reset ゲームリセット
func (gc *GrandfathersClock) Reset() {
	gc.trumpCards.Shuffle()
	gc.phase = GrandfathersClockPhasePlaying
	gc.moveCount = 0
	gc.actionLog = nil
	gc.history = nil
	gc.isStalemate = false

	for i := range GrandfathersClockFoundationCnt {
		gc.foundation[i] = nil
	}
	for i := range GrandfathersClockTableauCnt {
		gc.tableau[i] = nil
	}

	// 文字盤の 12 枚は固定なので、山から抜き出して所定の位置に置く。残りが
	// ちょうど 40 枚になり、8 列×5 枚に収まる。
	remaining := make([]*Card, 0, CardCnt-GrandfathersClockFoundationCnt)
	for {
		card := gc.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if idx := grandfathersClockStarterIndex(card); idx >= 0 && len(gc.foundation[idx]) == 0 {
			gc.foundation[idx] = []*Card{card}
			continue
		}
		remaining = append(remaining, card)
	}

	for col := range GrandfathersClockTableauCnt {
		for range GrandfathersClockColumnLen {
			if len(remaining) == 0 {
				break
			}
			card := remaining[0]
			remaining = remaining[1:]
			gc.tableau[col] = append(gc.tableau[col],
				&GrandfathersClockTableauCard{Card: card, FaceUp: true})
		}
	}

	gc.checkStalemate()
}

// grandfathersClockStarterIndex カードが文字盤の初期札ならその位置を返す（違えば -1）。
func grandfathersClockStarterIndex(card *Card) int {
	if card == nil {
		return -1
	}
	for i, s := range grandfathersClockStarters {
		if s.design == card.GetDesign() && s.value == card.GetValue() {
			return i
		}
	}
	return -1
}

// MoveTableauToFoundation タブロー最上段を文字盤へ移す
func (gc *GrandfathersClock) MoveTableauToFoundation(col, fIdx int) error {
	if err := gc.requirePlaying(); err != nil {
		return err
	}
	card, err := gc.topOf(col)
	if err != nil {
		return err
	}
	if fIdx < 0 || fIdx >= GrandfathersClockFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if !gc.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on that clock face")
	}
	gc.takeSnapshot()
	gc.popTop(col)
	gc.foundation[fIdx] = append(gc.foundation[fIdx], card)
	gc.afterMove("move", fmt.Sprintf("タブロー列%d→文字盤%d", col, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー最上段を別の列へ移す（スート無視の降順、空き列は任意）
func (gc *GrandfathersClock) MoveTableauToTableau(fromCol, toCol int) error {
	if err := gc.requirePlaying(); err != nil {
		return err
	}
	card, err := gc.topOf(fromCol)
	if err != nil {
		return err
	}
	if toCol < 0 || toCol >= GrandfathersClockTableauCnt {
		return errors.New("invalid destination column")
	}
	if fromCol == toCol {
		return errors.New("source and destination are the same column")
	}
	if !gc.canPlaceOnTableau(card, toCol) {
		return errors.New("tableau builds down by one, any suit")
	}
	gc.takeSnapshot()
	gc.popTop(fromCol)
	gc.tableau[toCol] = append(gc.tableau[toCol],
		&GrandfathersClockTableauCard{Card: card, FaceUp: true})
	gc.afterMove("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), card)
	return nil
}

// GiveUp ギブアップ
func (gc *GrandfathersClock) GiveUp() {
	if gc.phase == GrandfathersClockPhasePlaying {
		gc.phase = GrandfathersClockPhaseGameOver
		gc.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。文字盤へ送れる手を優先し、無ければタブローの手を返す。
// 手詰まり判定もこの関数を使う。
func (gc *GrandfathersClock) GetHint() *GrandfathersClockHint {
	if h := gc.foundationHint(); h != nil {
		return h
	}
	return gc.tableauHint()
}

// foundationHint 文字盤へ送れる手を 1 つ返す（オートコンプリート用）。
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 文字盤に上げるだけの操作である。
func (gc *GrandfathersClock) foundationHint() *GrandfathersClockHint {
	if gc.phase != GrandfathersClockPhasePlaying {
		return nil
	}
	for col := range GrandfathersClockTableauCnt {
		card := gc.topCard(col)
		if card == nil {
			continue
		}
		for fIdx := range GrandfathersClockFoundationCnt {
			if gc.canPlaceOnFoundation(card, fIdx) {
				return &GrandfathersClockHint{FromCol: col, ToZone: "foundation", ToIdx: fIdx}
			}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
//
// 空き列への移動は、その列に 1 枚しかない場合は「出して戻す」だけで盤面が進ま
// ないため提案しない。放置するとヒントが無限に同じ手を勧める。
func (gc *GrandfathersClock) tableauHint() *GrandfathersClockHint {
	if gc.phase != GrandfathersClockPhasePlaying {
		return nil
	}
	for from := range GrandfathersClockTableauCnt {
		card := gc.topCard(from)
		if card == nil {
			continue
		}
		for to := range GrandfathersClockTableauCnt {
			if to == from {
				continue
			}
			if len(gc.tableau[to]) == 0 && len(gc.tableau[from]) == 1 {
				continue
			}
			if gc.canPlaceOnTableau(card, to) {
				return &GrandfathersClockHint{FromCol: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 文字盤へ送れる札がなくなるまで自動で送る
func (gc *GrandfathersClock) AutoComplete() error {
	if gc.phase != GrandfathersClockPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := gc.foundationHint()
		if h == nil {
			break
		}
		if err := gc.MoveTableauToFoundation(h.FromCol, h.ToIdx); err != nil {
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
func (gc *GrandfathersClock) Undo() error {
	if len(gc.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := gc.history[len(gc.history)-1]
	gc.history = gc.history[:len(gc.history)-1]
	gc.foundation = snap.foundation
	gc.tableau = snap.tableau
	gc.phase = snap.phase
	gc.moveCount = snap.moveCount
	gc.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (gc *GrandfathersClock) CanUndo() bool { return len(gc.history) > 0 }

// UndoN n 手戻す
func (gc *GrandfathersClock) UndoN(n int) error {
	return undoNChecked(gc, n, len(gc.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (gc *GrandfathersClock) UndoToEscape() int {
	return undoToEscape(gc.isStalemate, gc.history, func(s *grandfathersClockSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (gc *GrandfathersClock) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (gc *GrandfathersClock) GetPhase() GrandfathersClockPhase { return gc.phase }

// GetMoveCount 手数取得
func (gc *GrandfathersClock) GetMoveCount() int { return gc.moveCount }

// GetFoundation 文字盤を取得
func (gc *GrandfathersClock) GetFoundation() [GrandfathersClockFoundationCnt][]*Card {
	return gc.foundation
}

// GetTableau タブローを取得
func (gc *GrandfathersClock) GetTableau() [GrandfathersClockTableauCnt][]*GrandfathersClockTableauCard {
	return gc.tableau
}

// GetGameEndFlag ゲーム終了フラグ
func (gc *GrandfathersClock) GetGameEndFlag() bool {
	return gc.phase != GrandfathersClockPhasePlaying
}

// IsStalemate 手詰まりか
func (gc *GrandfathersClock) IsStalemate() bool { return gc.isStalemate }

// IsFoundationComplete 文字盤 fIdx が目標ランクに達しているか
func (gc *GrandfathersClock) IsFoundationComplete(fIdx int) bool {
	if fIdx < 0 || fIdx >= GrandfathersClockFoundationCnt {
		return false
	}
	pile := gc.foundation[fIdx]
	if len(pile) == 0 {
		return false
	}
	return pile[len(pile)-1].GetValue() == GrandfathersClockTargetRank(fIdx)
}

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (gc *GrandfathersClock) requirePlaying() error {
	if gc.phase != GrandfathersClockPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// topOf 指定列の最上段カードを返す（列が空・範囲外ならエラー）
func (gc *GrandfathersClock) topOf(col int) (*Card, error) {
	if col < 0 || col >= GrandfathersClockTableauCnt {
		return nil, errors.New("invalid column index")
	}
	pile := gc.tableau[col]
	if len(pile) == 0 {
		return nil, errors.New("column is empty")
	}
	return pile[len(pile)-1].Card, nil
}

// topCard 指定列の最上段カードを返す（空なら nil）。手の探索用。
func (gc *GrandfathersClock) topCard(col int) *Card {
	pile := gc.tableau[col]
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1].Card
}

// popTop 指定列の最上段を取り除く
func (gc *GrandfathersClock) popTop(col int) {
	gc.tableau[col] = gc.tableau[col][:len(gc.tableau[col])-1]
}

// canPlaceOnTableau タブローに置けるか（空き列は任意、以降はスート無視で 1 つ下）。
// 文字盤と違い、こちらは A の下へ K を置くような折り返しをしない。
func (gc *GrandfathersClock) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := gc.tableau[col]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1].Card
	return card.GetValue() == top.GetValue()-1
}

// canPlaceOnFoundation 文字盤に置けるか。
// 同スートで 1 つ上、K の次は A に折り返す。目標ランクに達した文字盤は完成
// なので何も受け付けない — 折り返しを許すと 1 周して積み続けられてしまう。
func (gc *GrandfathersClock) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := gc.foundation[fIdx]
	if len(pile) == 0 {
		return false
	}
	if gc.IsFoundationComplete(fIdx) {
		return false
	}
	top := pile[len(pile)-1]
	if card.GetDesign() != top.GetDesign() {
		return false
	}
	return card.GetValue() == grandfathersClockNextRank(top.GetValue())
}

// grandfathersClockNextRank K の次を A に折り返す昇順の次ランク
func grandfathersClockNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (gc *GrandfathersClock) afterMove(actionType, detail string, card *Card) {
	gc.moveCount++
	gc.appendLog(actionType, detail, []*Card{card})
	gc.checkGameClear()
	gc.checkStalemate()
}

// checkGameClear 12 の文字盤すべてが目標ランクに達したか
func (gc *GrandfathersClock) checkGameClear() {
	for i := range GrandfathersClockFoundationCnt {
		if !gc.IsFoundationComplete(i) {
			return
		}
	}
	gc.phase = GrandfathersClockPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は文字盤とタブローの両方を見るので、「ヒントが無い」と「手詰まり」は
// 同じ条件になる。二重に持つと片方だけ直したときに静かに食い違う。
func (gc *GrandfathersClock) checkStalemate() {
	if gc.phase != GrandfathersClockPhasePlaying {
		return
	}
	gc.isStalemate = gc.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (gc *GrandfathersClock) takeSnapshot() {
	snap := &grandfathersClockSnapshot{
		phase:       gc.phase,
		moveCount:   gc.moveCount,
		isStalemate: gc.isStalemate,
	}
	for i := range GrandfathersClockFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), gc.foundation[i]...)
	}
	for i := range GrandfathersClockTableauCnt {
		snap.tableau[i] = append([]*GrandfathersClockTableauCard(nil), gc.tableau[i]...)
	}
	gc.history = append(gc.history, snap)
}

// appendLog 棋譜エントリを追加
func (gc *GrandfathersClock) appendLog(actionType, detail string, cards []*Card) {
	gc.appendLogAt(gc.moveCount, 0, actionType, detail, cards)
}

// grandfathersClockMaxSliceLen caps slice sizes during deserialisation.
const grandfathersClockMaxSliceLen = 1000

// grandfathersClockSnapshotJSON is the wire format for a single undo snapshot.
// grandfathersClockSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type grandfathersClockSnapshotJSON struct {
	Foundation  [GrandfathersClockFoundationCnt][]*Card                      `json:"fd"`
	Tableau     [GrandfathersClockTableauCnt][]*GrandfathersClockTableauCard `json:"tb"`
	Phase       GrandfathersClockPhase                                       `json:"ps"`
	MoveCount   int                                                          `json:"mc"`
	IsStalemate bool                                                         `json:"sl"`
}

// MarshalJSON implements json.Marshaler for grandfathersClockSnapshot.
func (s *grandfathersClockSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(grandfathersClockSnapshotJSON{
		Foundation:  s.foundation,
		Tableau:     s.tableau,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for grandfathersClockSnapshot.
func (s *grandfathersClockSnapshot) UnmarshalJSON(data []byte) error {
	var j grandfathersClockSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, pile := range j.Foundation {
		if len(pile) > grandfathersClockMaxSliceLen {
			return errors.New("grandfathersclock: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Tableau {
		if len(pile) > grandfathersClockMaxSliceLen {
			return errors.New("grandfathersclock: snapshot pile exceeds maximum allowed size")
		}
	}
	s.foundation = j.Foundation
	s.tableau = j.Tableau
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// grandfathersClockJSON is the JSON wire format for GrandfathersClock.
type grandfathersClockJSON struct {
	TrumpCards  *TrumpCards                                                  `json:"tc"`
	Foundation  [GrandfathersClockFoundationCnt][]*Card                      `json:"fd"`
	Tableau     [GrandfathersClockTableauCnt][]*GrandfathersClockTableauCard `json:"tb"`
	Phase       GrandfathersClockPhase                                       `json:"ps"`
	MoveCount   int                                                          `json:"mc"`
	ActionLog   []*ActionLogEntry                                            `json:"al"`
	IsStalemate bool                                                         `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*grandfathersClockSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (gc *GrandfathersClock) MarshalJSON() ([]byte, error) {
	return json.Marshal(&grandfathersClockJSON{
		TrumpCards:  gc.trumpCards,
		Foundation:  gc.foundation,
		Tableau:     gc.tableau,
		Phase:       gc.phase,
		MoveCount:   gc.moveCount,
		ActionLog:   gc.actionLog,
		IsStalemate: gc.isStalemate,
		History:     gc.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (gc *GrandfathersClock) UnmarshalJSON(data []byte) error {
	var j grandfathersClockJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > grandfathersClockMaxSliceLen || len(j.History) > grandfathersClockMaxSliceLen {
		return errors.New("grandfathersclock: input array exceeds maximum allowed size")
	}
	if j.Phase < GrandfathersClockPhasePlaying || j.Phase > GrandfathersClockPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	for i := range GrandfathersClockFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("clock face %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range GrandfathersClockTableauCnt {
		if len(j.Tableau[i]) > CardCnt {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		gc.trumpCards = j.TrumpCards
	}
	gc.foundation = j.Foundation
	gc.tableau = j.Tableau
	gc.phase = j.Phase
	gc.moveCount = j.MoveCount
	gc.actionLog = j.ActionLog
	gc.isStalemate = j.IsStalemate
	gc.history = j.History
	return nil
}
