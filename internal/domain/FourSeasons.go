//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// FourSeasonsPhase フォーシーズンズのゲームフェーズ
type FourSeasonsPhase int

// FourSeasonsのフェーズ定数
const (
	// FourSeasonsPhasePlaying プレイ中
	FourSeasonsPhasePlaying FourSeasonsPhase = iota
	// FourSeasonsPhaseGameClear ゲームクリア
	FourSeasonsPhaseGameClear
	// FourSeasonsPhaseGameOver ゲームオーバー
	FourSeasonsPhaseGameOver
)

// FourSeasonsTableauCnt タブロー（十字）の山数。中央1＋四方4。
const FourSeasonsTableauCnt = 5

// FourSeasonsFoundationCnt ファンデーション数（四隅）
const FourSeasonsFoundationCnt = 4

// FourSeasonsHint フォーシーズンズのヒント
type FourSeasonsHint struct {
	// FromZone 移動元 "waste" または "tableau"
	FromZone string
	// FromIdx 移動元がタブローの場合の山インデックス、ウェイストなら -1
	FromIdx int
	// ToZone 移動先 "foundation" または "tableau"
	ToZone string
	// ToIdx 移動先のインデックス
	ToIdx int
}

// FourSeasons フォーシーズンズ（コーナーズ）ゲームクラス。
//
// 十字に並べた5つのタブローと四隅の4つのファンデーションで遊ぶイギリスの古典。
// ファンデーションは**最初にめくれた1枚のランクから始まり**、同じスートで昇順に
// 積む。タブローは**スートを問わず**降順。どちらも**ラップアラウンド**する
// （K の次は A、A の下は K）。
//
// ベースランクが配りごとに変わる点は Canfield と同じだが、Canfield のタブローが
// 色違いの降順でラップしないのに対し、こちらはスート無視かつラップする。空いた
// 十字のマスは**どのカードでも**置ける（Canfield はリザーブが尽きるまで置けない）。
type FourSeasons struct {
	trumpCards *TrumpCards
	tableau    [FourSeasonsTableauCnt][]*Card
	foundation [FourSeasonsFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	baseRank   int
	phase      FourSeasonsPhase
	moveCount  int
	actionLogBase
	history []*fourSeasonsSnapshot
}

// fourSeasonsSnapshot アンドゥ用スナップショット
type fourSeasonsSnapshot struct {
	tableau    [FourSeasonsTableauCnt][]*Card
	foundation [FourSeasonsFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      FourSeasonsPhase
	moveCount  int
}

// NewFourSeasons コンストラクタ
func NewFourSeasons(trumpCards *TrumpCards) *FourSeasons {
	return &FourSeasons{trumpCards: trumpCards}
}

// NewDefaultFourSeasons returns FourSeasons with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultFourSeasons() *FourSeasons {
	return NewFourSeasons(NewTrumpCards(0))
}

// Reset ゲームリセット
func (f *FourSeasons) Reset() {
	f.trumpCards.Shuffle()
	f.phase = FourSeasonsPhasePlaying
	f.moveCount = 0
	f.actionLog = nil
	f.history = nil

	// 十字に5枚
	for i := range FourSeasonsTableauCnt {
		f.tableau[i] = []*Card{f.trumpCards.DrawCard()}
	}

	// ファンデーション初期化後、次の1枚でベースランクを決める。
	// **どのランクで始まるかは配り次第**で、A 固定ではない。以降のすべての
	// 置ける／置けない判定がこの値に乗るので、KV へ必ず永続化する。
	for i := range FourSeasonsFoundationCnt {
		f.foundation[i] = nil
	}
	base := f.trumpCards.DrawCard()
	f.baseRank = base.GetValue()
	f.foundation[0] = []*Card{base}

	f.stock = nil
	f.waste = nil
	for f.trumpCards.GetRemainingCount() > 0 {
		f.stock = append(f.stock, f.trumpCards.DrawCard())
	}
}

// Draw ストックからウェイストへ1枚引く
func (f *FourSeasons) Draw() error {
	if f.phase != FourSeasonsPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(f.stock) == 0 {
		return errors.New("stock is empty")
	}
	f.takeSnapshot()
	card := f.stock[len(f.stock)-1]
	f.stock = f.stock[:len(f.stock)-1]
	f.waste = append(f.waste, card)
	f.moveCount++
	f.appendLog("draw", "ストックから1枚引きました", []*Card{card})
	return nil
}

// MoveWasteToFoundation ウェイスト最上段をファンデーションに置く
func (f *FourSeasons) MoveWasteToFoundation(fIdx int) error {
	if err := f.checkPlaying(); err != nil {
		return err
	}
	if err := f.checkFoundationIdx(fIdx); err != nil {
		return err
	}
	card, err := f.wasteTop()
	if err != nil {
		return err
	}
	if !f.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	f.takeSnapshot()
	f.waste = f.waste[:len(f.waste)-1]
	f.foundation[fIdx] = append(f.foundation[fIdx], card)
	f.afterMove(fmt.Sprintf("ウェイスト→ファンデーション%d", fIdx+1), card)
	return nil
}

// MoveWasteToTableau ウェイスト最上段をタブローに置く
func (f *FourSeasons) MoveWasteToTableau(col int) error {
	if err := f.checkPlaying(); err != nil {
		return err
	}
	if err := f.checkTableauIdx(col); err != nil {
		return err
	}
	card, err := f.wasteTop()
	if err != nil {
		return err
	}
	if !f.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	f.takeSnapshot()
	f.waste = f.waste[:len(f.waste)-1]
	f.tableau[col] = append(f.tableau[col], card)
	f.afterMove(fmt.Sprintf("ウェイスト→タブロー%d", col+1), card)
	return nil
}

// MoveTableauToFoundation タブロー最上段をファンデーションに置く
func (f *FourSeasons) MoveTableauToFoundation(col, fIdx int) error {
	if err := f.checkPlaying(); err != nil {
		return err
	}
	if err := f.checkTableauIdx(col); err != nil {
		return err
	}
	if err := f.checkFoundationIdx(fIdx); err != nil {
		return err
	}
	if len(f.tableau[col]) == 0 {
		return errors.New("tableau pile is empty")
	}
	card := f.tableau[col][len(f.tableau[col])-1]
	if !f.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	f.takeSnapshot()
	f.tableau[col] = f.tableau[col][:len(f.tableau[col])-1]
	f.foundation[fIdx] = append(f.foundation[fIdx], card)
	f.afterMove(fmt.Sprintf("タブロー%d→ファンデーション%d", col+1, fIdx+1), card)
	return nil
}

// MoveTableauToTableau タブロー最上段を別のタブローへ移す。
// 動くのは最上段の1枚だけで、列をまとめて動かすことはできない。
func (f *FourSeasons) MoveTableauToTableau(fromCol, toCol int) error {
	if err := f.checkPlaying(); err != nil {
		return err
	}
	if err := f.checkTableauIdx(fromCol); err != nil {
		return err
	}
	if err := f.checkTableauIdx(toCol); err != nil {
		return err
	}
	if fromCol == toCol {
		return errors.New("cannot move a pile onto itself")
	}
	if len(f.tableau[fromCol]) == 0 {
		return errors.New("tableau pile is empty")
	}
	card := f.tableau[fromCol][len(f.tableau[fromCol])-1]
	if !f.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	f.takeSnapshot()
	f.tableau[fromCol] = f.tableau[fromCol][:len(f.tableau[fromCol])-1]
	f.tableau[toCol] = append(f.tableau[toCol], card)
	f.afterMove(fmt.Sprintf("タブロー%d→タブロー%d", fromCol+1, toCol+1), card)
	return nil
}

// GiveUp ギブアップ
func (f *FourSeasons) GiveUp() {
	if f.phase == FourSeasonsPhasePlaying {
		f.phase = FourSeasonsPhaseGameOver
		f.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ファンデーションへ送れる手を1つ提示する。
// タブローからの手を先に見るのは、ウェイストは引き直せばまた出てくるのに対し、
// タブロー最上段は下の札を塞いだままだから。
func (f *FourSeasons) GetHint() *FourSeasonsHint {
	if f.phase != FourSeasonsPhasePlaying {
		return nil
	}
	for col := range FourSeasonsTableauCnt {
		pile := f.tableau[col]
		if len(pile) == 0 {
			continue
		}
		if fIdx := f.findFoundation(pile[len(pile)-1]); fIdx >= 0 {
			return &FourSeasonsHint{FromZone: "tableau", FromIdx: col, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if len(f.waste) > 0 {
		if fIdx := f.findFoundation(f.waste[len(f.waste)-1]); fIdx >= 0 {
			return &FourSeasonsHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 送れるカードが無くなるまでファンデーションへ積む
func (f *FourSeasons) AutoComplete() error {
	if f.phase != FourSeasonsPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	// 1手ごとに全体を見直す。1枚送ると下の札が最上段に出るため、一巡では取りこぼす。
	for {
		h := f.GetHint()
		if h == nil {
			break
		}
		var err error
		if h.FromZone == "tableau" {
			err = f.MoveTableauToFoundation(h.FromIdx, h.ToIdx)
		} else {
			err = f.MoveWasteToFoundation(h.ToIdx)
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

// Undo 直前の1手を取り消す
func (f *FourSeasons) Undo() error {
	if len(f.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := f.history[len(f.history)-1]
	f.history = f.history[:len(f.history)-1]
	f.tableau = snap.tableau
	f.foundation = snap.foundation
	f.stock = snap.stock
	f.waste = snap.waste
	f.phase = snap.phase
	f.moveCount = snap.moveCount
	return nil
}

// CanUndo アンドゥ可能か
func (f *FourSeasons) CanUndo() bool { return len(f.history) > 0 }

// UndoN n手戻す
func (f *FourSeasons) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(f.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := f.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// GetPhase フェーズ取得
func (f *FourSeasons) GetPhase() FourSeasonsPhase { return f.phase }

// GetMoveCount 手数取得
func (f *FourSeasons) GetMoveCount() int { return f.moveCount }

// GetBaseRank ベースランク取得
func (f *FourSeasons) GetBaseRank() int { return f.baseRank }

// GetStockCount ストック残枚数取得
func (f *FourSeasons) GetStockCount() int { return len(f.stock) }

// GetWaste ウェイスト取得
func (f *FourSeasons) GetWaste() []*Card { return f.waste }

// GetTableau タブロー取得
func (f *FourSeasons) GetTableau() [FourSeasonsTableauCnt][]*Card { return f.tableau }

// GetFoundations ファンデーション取得
func (f *FourSeasons) GetFoundations() [FourSeasonsFoundationCnt][]*Card { return f.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (f *FourSeasons) GetGameEndFlag() bool { return f.phase != FourSeasonsPhasePlaying }

// --- Private helpers ---

func (f *FourSeasons) checkPlaying() error {
	if f.phase != FourSeasonsPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

func (f *FourSeasons) checkTableauIdx(col int) error {
	if col < 0 || col >= FourSeasonsTableauCnt {
		return errors.New("invalid tableau index")
	}
	return nil
}

func (f *FourSeasons) checkFoundationIdx(fIdx int) error {
	if fIdx < 0 || fIdx >= FourSeasonsFoundationCnt {
		return errors.New("invalid foundation index")
	}
	return nil
}

func (f *FourSeasons) wasteTop() (*Card, error) {
	if len(f.waste) == 0 {
		return nil, errors.New("waste is empty")
	}
	return f.waste[len(f.waste)-1], nil
}

// afterMove は移動後の共通処理（手数・棋譜・クリア判定）。
func (f *FourSeasons) afterMove(detail string, card *Card) {
	f.moveCount++
	f.appendLog("move", detail, []*Card{card})
	f.checkGameClear()
}

// canPlaceOnFoundation 四隅は**同じスート**でベースランクから昇順、ラップあり。
// 空の隅はベースランクのみ受け付ける（どのスートがどの隅に載るかは先着順）。
func (f *FourSeasons) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := f.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == f.baseRank
	}
	if len(pile) >= CardValueMax {
		return false
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == fourSeasonsNextRank(top.GetValue())
}

// canPlaceOnTableau 十字は**スートを問わず**降順、ラップあり。
// 空きマスはどのカードでも受け付ける。
func (f *FourSeasons) canPlaceOnTableau(card *Card, col int) bool {
	pile := f.tableau[col]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1]
	return card.GetValue() == fourSeasonsPrevRank(top.GetValue())
}

// fourSeasonsNextRank は K の次を A に戻す昇順の次ランク。
func fourSeasonsNextRank(r int) int { return (r % CardValueMax) + 1 }

// fourSeasonsPrevRank は A の下を K に戻す降順の次ランク。
func fourSeasonsPrevRank(r int) int { return ((r + CardValueMax - 2) % CardValueMax) + 1 }

func (f *FourSeasons) findFoundation(card *Card) int {
	for i := range FourSeasonsFoundationCnt {
		if f.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

func (f *FourSeasons) checkGameClear() {
	for i := range FourSeasonsFoundationCnt {
		if len(f.foundation[i]) != CardValueMax {
			return
		}
	}
	f.phase = FourSeasonsPhaseGameClear
}

func (f *FourSeasons) takeSnapshot() {
	snap := &fourSeasonsSnapshot{phase: f.phase, moveCount: f.moveCount}
	for i := range FourSeasonsTableauCnt {
		snap.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(snap.tableau[i], f.tableau[i])
	}
	for i := range FourSeasonsFoundationCnt {
		snap.foundation[i] = make([]*Card, len(f.foundation[i]))
		copy(snap.foundation[i], f.foundation[i])
	}
	snap.stock = make([]*Card, len(f.stock))
	copy(snap.stock, f.stock)
	snap.waste = make([]*Card, len(f.waste))
	copy(snap.waste, f.waste)
	f.history = append(f.history, snap)
}

func (f *FourSeasons) appendLog(actionType, detail string, cards []*Card) {
	f.appendLogAt(f.moveCount, 0, actionType, detail, cards)
}

// fourSeasonsMaxSliceLen caps slice sizes during deserialisation.
const fourSeasonsMaxSliceLen = 1000

// fourSeasonsSnapshotJSON is the wire format for a single undo snapshot.
// fourSeasonsSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type fourSeasonsSnapshotJSON struct {
	Tableau    [FourSeasonsTableauCnt][]*Card    `json:"tb"`
	Foundation [FourSeasonsFoundationCnt][]*Card `json:"fd"`
	Stock      []*Card                           `json:"st"`
	Waste      []*Card                           `json:"wa"`
	Phase      FourSeasonsPhase                  `json:"ps"`
	MoveCount  int                               `json:"mc"`
}

// MarshalJSON implements json.Marshaler for fourSeasonsSnapshot.
func (s *fourSeasonsSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(fourSeasonsSnapshotJSON{
		Tableau: s.tableau, Foundation: s.foundation, Stock: s.stock,
		Waste: s.waste, Phase: s.phase, MoveCount: s.moveCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for fourSeasonsSnapshot.
func (s *fourSeasonsSnapshot) UnmarshalJSON(data []byte) error {
	var j fourSeasonsSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > fourSeasonsMaxSliceLen || len(j.Waste) > fourSeasonsMaxSliceLen {
		return errors.New("fourseasons: snapshot array exceeds maximum allowed size")
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}

// fourSeasonsJSON is the JSON wire format for FourSeasons.
type fourSeasonsJSON struct {
	TrumpCards *TrumpCards                       `json:"tc"`
	Tableau    [FourSeasonsTableauCnt][]*Card    `json:"tb"`
	Foundation [FourSeasonsFoundationCnt][]*Card `json:"fd"`
	Stock      []*Card                           `json:"st"`
	Waste      []*Card                           `json:"wa"`
	// BaseRank must round-trip: every placement rule reads it, so a session
	// restored without it would reject legal moves and accept illegal ones.
	BaseRank  int               `json:"br"`
	Phase     FourSeasonsPhase  `json:"ps"`
	MoveCount int               `json:"mc"`
	ActionLog []*ActionLogEntry `json:"al"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call (#4478).
	History []*fourSeasonsSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (f *FourSeasons) MarshalJSON() ([]byte, error) {
	return json.Marshal(&fourSeasonsJSON{
		TrumpCards: f.trumpCards, Tableau: f.tableau, Foundation: f.foundation,
		Stock: f.stock, Waste: f.waste, BaseRank: f.baseRank, Phase: f.phase,
		MoveCount: f.moveCount, ActionLog: f.actionLog, History: f.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。
// 値域を検証するのは、KV に入っているのは以前のバージョンが書いた任意のバイト列で
// あり、壊れた状態でゲームを開始させないため。
func (f *FourSeasons) UnmarshalJSON(data []byte) error {
	var j fourSeasonsJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	// **生きている山も含めて全部に上限をかける。** ここは Worker が KV から毎
	// リクエスト復元する経路なので、壊れた／細工されたスナップショットの
	// 巨大配列がそのまま通る。ActionLog と History だけ見ていたのでは、
	// st / wa / tb が素通りする（Canfield.UnmarshalJSON は5つとも見ている）。
	if len(j.ActionLog) > fourSeasonsMaxSliceLen || len(j.History) > fourSeasonsMaxSliceLen ||
		len(j.Stock) > fourSeasonsMaxSliceLen || len(j.Waste) > fourSeasonsMaxSliceLen {
		return errors.New("fourseasons: input array exceeds maximum allowed size")
	}
	for i := range FourSeasonsTableauCnt {
		if len(j.Tableau[i]) > fourSeasonsMaxSliceLen {
			return fmt.Errorf("tableau %d exceeds maximum allowed size", i)
		}
	}
	if j.Phase < FourSeasonsPhasePlaying || j.Phase > FourSeasonsPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	// 0 は「まだ配っていない」を表すゼロ値なので許す。それ以外は 1..13。
	if j.BaseRank < 0 || j.BaseRank > CardValueMax {
		return fmt.Errorf("invalid base rank: %d", j.BaseRank)
	}
	for i := range FourSeasonsFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("foundation %d has %d cards", i, len(j.Foundation[i]))
		}
	}
	if j.TrumpCards != nil {
		f.trumpCards = j.TrumpCards
	}
	f.tableau = j.Tableau
	f.foundation = j.Foundation
	f.stock = j.Stock
	f.waste = j.Waste
	f.baseRank = j.BaseRank
	f.phase = j.Phase
	f.moveCount = j.MoveCount
	f.actionLog = j.ActionLog
	f.history = j.History
	return nil
}
