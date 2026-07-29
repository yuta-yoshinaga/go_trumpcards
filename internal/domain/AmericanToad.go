//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AmericanToadPhase アメリカン・トードのゲームフェーズ
type AmericanToadPhase int

// AmericanToadのフェーズ定数
const (
	// AmericanToadPhasePlaying プレイ中
	AmericanToadPhasePlaying AmericanToadPhase = iota
	// AmericanToadPhaseGameClear ゲームクリア
	AmericanToadPhaseGameClear
	// AmericanToadPhaseGameOver ゲームオーバー
	AmericanToadPhaseGameOver
)

// AmericanToadTableauCnt タブローの列数
const AmericanToadTableauCnt = 8

// AmericanToadFoundationCnt 基礎札の数（スートごとに 2 つ）
const AmericanToadFoundationCnt = 8

// AmericanToadReserveSize リザーブの初期枚数
const AmericanToadReserveSize = 20

// AmericanToadFoundationTarget 基礎札 1 つあたりの完成枚数
const AmericanToadFoundationTarget = CardValueMax

// AmericanToadMaxPasses 山札を通せる回数（1 回のめくり直しを含む）
const AmericanToadMaxPasses = 2

// AmericanToadTotalCards 使用する総枚数（52 枚 2 組）
const AmericanToadTotalCards = CardCnt * 2

// americanToadSuitOrder 基礎札インデックスとスートの対応。スートごとに 2 つあり、
// 固定しておくと配り直しても UI の位置が動かない。
var americanToadSuitOrder = [AmericanToadFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// AmericanToadTableauCard タブロー上のカード。全札が表向きだが、他のソリティアと
// 同じ形にしておくとプレゼンターを使い回せる。
type AmericanToadTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// AmericanToadHint アメリカン・トードのヒント
type AmericanToadHint struct {
	// FromZone 移動元 "reserve" / "waste" / "tableau" / "stock"
	FromZone string
	// FromIdx 移動元のタブロー列（それ以外は -1）
	FromIdx int
	// CardIndex 移動元の列内インデックス。連番グループの先頭を指す（-1 は非該当）。
	CardIndex int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// AmericanToad アメリカン・トード ゲームクラス。
//
// カンフィールド系の 2 デッキ版。**リザーブ 20 枚**、タブロー 8 列に 1 枚ずつ、
// さらに 1 枚を最初の基礎札に置き、残り 75 枚が山札。
//
// 基礎札は 8 つ（スートごとに 2 つ）。最初に置かれた札のランクが 8 つすべての開始
// ランクになり、同スートで昇順、**K の次は A に折り返して 13 枚**で完成する。
// 8×13 = 104 枚すべてを積み切ればクリア。issue #4418 の「各スート 8 枚まで」は
// 32 枚にしかならず、104 枚を吸収できない。
//
// タブローは**同スートの降順**で、**A の下には K を置ける**（折り返す）。連番は
// まとめて動かせる。
//
// 空き列の扱いがこのゲームの要:
//   - リザーブが残っているうちは、空いた列は**自動的に**リザーブの一番上で埋まる
//   - リザーブを使い切ったあとは、**捨て札からのみ**空き列を埋められる。タブロー
//     から別の空き列へ移すことはできない（それを許すと列を自由に組み替えられる）
//
// 山札は 1 枚ずつ捨て札へめくり、**めくり直しは 1 回だけ**（通算 2 巡）。
//
// issue #4417 との相違点は 5 つで、いずれも実際の規則に合わせた:
//   - 基礎札は「各スート 8 枚」ではなく **8 つの山に 13 枚ずつ**
//   - 空き列はリザーブ専用ではなく、**リザーブが尽きたあとは捨て札から**埋められる
//     （さらに補充は自動で、プレイヤーの操作ではない）
//   - **めくり直しが 1 回ある**（issue は触れていない）
//   - タブローの連番は**まとめて動かせる**（issue は触れていない）
//   - タブローの降順も **A→K に折り返す**（issue は触れていない）
//
// なお Wikipedia 版は「連番の途中からは動かせない（全部か 1 枚か）」とするが、
// 主要な実装と bvssolitaire の記述に合わせ、任意の連番グループを動かせるものとした。
// このリポジトリの他のソリティアとも操作感が揃う。
type AmericanToad struct {
	trumpCards  *TrumpCards
	reserve     []*Card
	tableau     [AmericanToadTableauCnt][]*AmericanToadTableauCard
	foundation  [AmericanToadFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	baseRank    int
	passesUsed  int
	phase       AmericanToadPhase
	moveCount   int
	actionLog   []*ActionLogEntry
	history     []*americanToadSnapshot
	isStalemate bool
}

// americanToadSnapshot アンドゥ用スナップショット
type americanToadSnapshot struct {
	reserve     []*Card
	tableau     [AmericanToadTableauCnt][]*AmericanToadTableauCard
	foundation  [AmericanToadFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	baseRank    int
	passesUsed  int
	phase       AmericanToadPhase
	moveCount   int
	isStalemate bool
}

// NewAmericanToad コンストラクタ
func NewAmericanToad(trumpCards *TrumpCards) *AmericanToad {
	return &AmericanToad{trumpCards: trumpCards}
}

// NewDefaultAmericanToad returns AmericanToad with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultAmericanToad() *AmericanToad {
	return NewAmericanToad(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (at *AmericanToad) Reset() {
	at.trumpCards.Shuffle()
	at.phase = AmericanToadPhasePlaying
	at.moveCount = 0
	at.actionLog = nil
	at.history = nil
	at.isStalemate = false
	at.passesUsed = 0
	at.reserve = nil
	at.stock = nil
	at.waste = nil

	for i := range AmericanToadFoundationCnt {
		at.foundation[i] = nil
	}
	for i := range AmericanToadTableauCnt {
		at.tableau[i] = nil
	}

	for range AmericanToadReserveSize {
		if card := at.trumpCards.DrawCard(); card != nil {
			at.reserve = append(at.reserve, card)
		}
	}
	for i := range AmericanToadTableauCnt {
		if card := at.trumpCards.DrawCard(); card != nil {
			at.tableau[i] = append(at.tableau[i], &AmericanToadTableauCard{Card: card, FaceUp: true})
		}
	}
	// 次の 1 枚が開始ランクを決め、そのスートの基礎札に載る。
	if starter := at.trumpCards.DrawCard(); starter != nil {
		at.baseRank = starter.GetValue()
		at.foundation[americanToadFoundationForSuit(starter.GetDesign())] = []*Card{starter}
	}
	for {
		card := at.trumpCards.DrawCard()
		if card == nil {
			break
		}
		at.stock = append(at.stock, card)
	}

	at.checkStalemate()
}

// americanToadFoundationForSuit スートに対応する最初の基礎札インデックスを返す。
func americanToadFoundationForSuit(design int) int {
	for i, d := range americanToadSuitOrder {
		if d == design {
			return i
		}
	}
	return 0
}

// Draw 山札から捨て札へ 1 枚めくる。山札が空ならめくり直す（1 回だけ）。
func (at *AmericanToad) Draw() error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	if len(at.stock) == 0 {
		if at.passesUsed >= AmericanToadMaxPasses-1 || len(at.waste) == 0 {
			return errors.New("no redeal left")
		}
		at.takeSnapshot()
		// 捨て札を裏返して山札に戻す。順序は保たれる。
		at.stock = at.waste
		at.waste = nil
		at.passesUsed++
		at.afterMove("redeal", "捨て札を山札に戻した", nil)
		return nil
	}
	at.takeSnapshot()
	card := at.stock[0]
	at.stock = at.stock[1:]
	at.waste = append(at.waste, card)
	at.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveReserveToFoundation リザーブの一番上を基礎札へ送る
func (at *AmericanToad) MoveReserveToFoundation() error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	card := at.reserveTop()
	if card == nil {
		return errors.New("reserve is empty")
	}
	fIdx := at.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	at.takeSnapshot()
	at.popReserve()
	at.foundation[fIdx] = append(at.foundation[fIdx], card)
	at.fillEmptyColumnsFromReserve()
	at.afterMove("move", fmt.Sprintf("リザーブ→基礎札%d", fIdx), card)
	return nil
}

// MoveReserveToTableau リザーブの一番上をタブローへ送る
func (at *AmericanToad) MoveReserveToTableau(col int) error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	if err := validAmericanToadColumn(col); err != nil {
		return err
	}
	card := at.reserveTop()
	if card == nil {
		return errors.New("reserve is empty")
	}
	if !at.canPlaceOnTableau(card, col) {
		return errors.New("card cannot be placed on that column")
	}
	at.takeSnapshot()
	at.popReserve()
	at.tableau[col] = append(at.tableau[col], &AmericanToadTableauCard{Card: card, FaceUp: true})
	at.fillEmptyColumnsFromReserve()
	at.afterMove("move", fmt.Sprintf("リザーブ→タブロー列%d", col), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (at *AmericanToad) MoveWasteToFoundation() error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	card := at.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	fIdx := at.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	at.takeSnapshot()
	at.popWaste()
	at.foundation[fIdx] = append(at.foundation[fIdx], card)
	at.afterMove("move", fmt.Sprintf("捨て札→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToTableau 捨て札の一番上をタブローへ送る
func (at *AmericanToad) MoveWasteToTableau(col int) error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	if err := validAmericanToadColumn(col); err != nil {
		return err
	}
	card := at.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	if !at.canPlaceOnTableau(card, col) {
		return errors.New("card cannot be placed on that column")
	}
	at.takeSnapshot()
	at.popWaste()
	at.tableau[col] = append(at.tableau[col], &AmericanToadTableauCard{Card: card, FaceUp: true})
	at.afterMove("move", fmt.Sprintf("捨て札→タブロー列%d", col), card)
	return nil
}

// MoveTableauToFoundation タブローの一番上を基礎札へ送る
func (at *AmericanToad) MoveTableauToFoundation(col int) error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	if err := validAmericanToadColumn(col); err != nil {
		return err
	}
	card := at.tableauTop(col)
	if card == nil {
		return fmt.Errorf("column %d is empty", col)
	}
	fIdx := at.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	at.takeSnapshot()
	at.tableau[col] = at.tableau[col][:len(at.tableau[col])-1]
	at.foundation[fIdx] = append(at.foundation[fIdx], card)
	at.fillEmptyColumnsFromReserve()
	at.afterMove("move", fmt.Sprintf("タブロー列%d→基礎札%d", col, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間で移動する。cardIndex は動かす連番の先頭
// （-1 なら最上段 1 枚）。
func (at *AmericanToad) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if err := at.requirePlaying(); err != nil {
		return err
	}
	if err := validAmericanToadColumn(fromCol); err != nil {
		return err
	}
	if err := validAmericanToadColumn(toCol); err != nil {
		return err
	}
	if fromCol == toCol {
		return errors.New("source and destination are the same column")
	}
	from := at.tableau[fromCol]
	if len(from) == 0 {
		return fmt.Errorf("column %d is empty", fromCol)
	}
	idx := cardIndex
	if idx < 0 {
		idx = len(from) - 1
	}
	if idx >= len(from) {
		return fmt.Errorf("invalid card index: %d", cardIndex)
	}
	if !at.isRun(from, idx) {
		return errors.New("the selected cards are not a same-suit descending run")
	}
	// 空き列はリザーブと捨て札の出口であって、タブローの組み替えには使えない。
	if len(at.tableau[toCol]) == 0 {
		return errors.New("an empty column cannot be filled from another column")
	}
	if !at.canPlaceOnTableau(from[idx].Card, toCol) {
		return errors.New("card cannot be placed on that column")
	}
	at.takeSnapshot()
	moved := append([]*AmericanToadTableauCard(nil), from[idx:]...)
	at.tableau[fromCol] = from[:idx]
	at.tableau[toCol] = append(at.tableau[toCol], moved...)
	at.fillEmptyColumnsFromReserve()
	at.afterMove("move", fmt.Sprintf("タブロー列%d[%d]→タブロー列%d", fromCol, idx, toCol), moved[0].Card)
	return nil
}

// GiveUp ギブアップ
func (at *AmericanToad) GiveUp() {
	if at.phase == AmericanToadPhasePlaying {
		at.phase = AmericanToadPhaseGameOver
		at.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (at *AmericanToad) GetHint() *AmericanToadHint {
	if at.phase != AmericanToadPhasePlaying {
		return nil
	}
	if h := at.foundationHint(); h != nil {
		return h
	}
	if h := at.tableauHint(); h != nil {
		return h
	}
	if at.canDraw() {
		return &AmericanToadHint{FromZone: "stock", FromIdx: -1, CardIndex: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブロー間の手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (at *AmericanToad) foundationHint() *AmericanToadHint {
	if at.phase != AmericanToadPhasePlaying {
		return nil
	}
	if card := at.reserveTop(); card != nil {
		if fIdx := at.findFoundation(card); fIdx >= 0 {
			return &AmericanToadHint{FromZone: "reserve", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := at.wasteTop(); card != nil {
		if fIdx := at.findFoundation(card); fIdx >= 0 {
			return &AmericanToadHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for col := range AmericanToadTableauCnt {
		card := at.tableauTop(col)
		if card == nil {
			continue
		}
		if fIdx := at.findFoundation(card); fIdx >= 0 {
			return &AmericanToadHint{FromZone: "tableau", FromIdx: col, CardIndex: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
func (at *AmericanToad) tableauHint() *AmericanToadHint {
	if at.phase != AmericanToadPhasePlaying {
		return nil
	}
	if card := at.reserveTop(); card != nil {
		for col := range AmericanToadTableauCnt {
			if at.canPlaceOnTableau(card, col) {
				return &AmericanToadHint{FromZone: "reserve", FromIdx: -1, CardIndex: -1, ToZone: "tableau", ToIdx: col}
			}
		}
	}
	if card := at.wasteTop(); card != nil {
		for col := range AmericanToadTableauCnt {
			if at.canPlaceOnTableau(card, col) {
				return &AmericanToadHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "tableau", ToIdx: col}
			}
		}
	}
	for from := range AmericanToadTableauCnt {
		pile := at.tableau[from]
		for idx := range pile {
			if !at.isRun(pile, idx) {
				continue
			}
			for to := range AmericanToadTableauCnt {
				// 空き列はタブローからは埋められないので候補にしない。
				if from == to || len(at.tableau[to]) == 0 {
					continue
				}
				if at.canPlaceOnTableau(pile[idx].Card, to) {
					return &AmericanToadHint{FromZone: "tableau", FromIdx: from, CardIndex: idx, ToZone: "tableau", ToIdx: to}
				}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (at *AmericanToad) AutoComplete() error {
	if at.phase != AmericanToadPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := at.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch h.FromZone {
		case "reserve":
			err = at.MoveReserveToFoundation()
		case "waste":
			err = at.MoveWasteToFoundation()
		default:
			err = at.MoveTableauToFoundation(h.FromIdx)
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
func (at *AmericanToad) Undo() error {
	if len(at.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := at.history[len(at.history)-1]
	at.history = at.history[:len(at.history)-1]
	at.reserve = snap.reserve
	at.tableau = snap.tableau
	at.foundation = snap.foundation
	at.stock = snap.stock
	at.waste = snap.waste
	at.baseRank = snap.baseRank
	at.passesUsed = snap.passesUsed
	at.phase = snap.phase
	at.moveCount = snap.moveCount
	at.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (at *AmericanToad) CanUndo() bool { return len(at.history) > 0 }

// UndoN n 手戻す
func (at *AmericanToad) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(at.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := at.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (at *AmericanToad) UndoToEscape() int {
	if !at.isStalemate {
		return 0
	}
	for i := len(at.history) - 1; i >= 0; i-- {
		if !at.history[i].isStalemate {
			return len(at.history) - i
		}
	}
	return -1
}

// AllFaceUp 常に全札が表向き
func (at *AmericanToad) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (at *AmericanToad) GetPhase() AmericanToadPhase { return at.phase }

// GetMoveCount 手数取得
func (at *AmericanToad) GetMoveCount() int { return at.moveCount }

// GetStockCount 山札の残り枚数
func (at *AmericanToad) GetStockCount() int { return len(at.stock) }

// GetWaste 捨て札を取得
func (at *AmericanToad) GetWaste() []*Card { return at.waste }

// GetReserve リザーブを取得
func (at *AmericanToad) GetReserve() []*Card { return at.reserve }

// GetTableau タブローを取得
func (at *AmericanToad) GetTableau() [AmericanToadTableauCnt][]*AmericanToadTableauCard {
	return at.tableau
}

// GetFoundation 基礎札を取得
func (at *AmericanToad) GetFoundation() [AmericanToadFoundationCnt][]*Card { return at.foundation }

// GetBaseRank 基礎札の開始ランク
func (at *AmericanToad) GetBaseRank() int { return at.baseRank }

// GetPassesUsed 山札を通した回数
func (at *AmericanToad) GetPassesUsed() int { return at.passesUsed }

// CanRedeal もう一度めくり直せるか。捨て札が空のときは戻すものが無いので偽。
func (at *AmericanToad) CanRedeal() bool {
	return at.phase == AmericanToadPhasePlaying &&
		len(at.stock) == 0 && len(at.waste) > 0 && at.passesUsed < AmericanToadMaxPasses-1
}

// GetActionLog 棋譜取得
func (at *AmericanToad) GetActionLog() []*ActionLogEntry { return at.actionLog }

// GetGameEndFlag ゲーム終了フラグ
func (at *AmericanToad) GetGameEndFlag() bool { return at.phase != AmericanToadPhasePlaying }

// IsStalemate 手詰まりか
func (at *AmericanToad) IsStalemate() bool { return at.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (at *AmericanToad) requirePlaying() error {
	if at.phase != AmericanToadPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validAmericanToadColumn 列インデックスを検証する
func validAmericanToadColumn(col int) error {
	if col < 0 || col >= AmericanToadTableauCnt {
		return fmt.Errorf("invalid column: %d", col)
	}
	return nil
}

// americanToadNextRank K の次は A に折り返す昇順
func americanToadNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// americanToadPrevRank A の次は K に折り返す降順
func americanToadPrevRank(v int) int {
	if v <= 1 {
		return CardValueMax
	}
	return v - 1
}

// canDraw 山札をめくれるか（めくり直しも含む）
func (at *AmericanToad) canDraw() bool {
	return len(at.stock) > 0 || at.CanRedeal()
}

// reserveTop リザーブの一番上（空なら nil）
func (at *AmericanToad) reserveTop() *Card {
	if len(at.reserve) == 0 {
		return nil
	}
	return at.reserve[len(at.reserve)-1]
}

// popReserve リザーブの一番上を取り除く
func (at *AmericanToad) popReserve() {
	if len(at.reserve) > 0 {
		at.reserve = at.reserve[:len(at.reserve)-1]
	}
}

// wasteTop 捨て札の一番上（空なら nil）
func (at *AmericanToad) wasteTop() *Card {
	if len(at.waste) == 0 {
		return nil
	}
	return at.waste[len(at.waste)-1]
}

// popWaste 捨て札の一番上を取り除く
func (at *AmericanToad) popWaste() {
	if len(at.waste) > 0 {
		at.waste = at.waste[:len(at.waste)-1]
	}
}

// tableauTop 列の一番上（空なら nil）
func (at *AmericanToad) tableauTop(col int) *Card {
	pile := at.tableau[col]
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1].Card
}

// isRun idx 以降が同スートの降順（A→K 折り返しつき）に並んでいるか
func (at *AmericanToad) isRun(pile []*AmericanToadTableauCard, idx int) bool {
	if idx < 0 || idx >= len(pile) || pile[idx] == nil || pile[idx].Card == nil {
		return false
	}
	for i := idx; i+1 < len(pile); i++ {
		cur, next := pile[i].Card, pile[i+1].Card
		if cur == nil || next == nil {
			return false
		}
		if cur.GetDesign() != next.GetDesign() {
			return false
		}
		if next.GetValue() != americanToadPrevRank(cur.GetValue()) {
			return false
		}
	}
	return true
}

// canPlaceOnTableau タブローに置けるか（同スートの降順、A の下には K）。
// 空き列に置けるかは呼び出し側が判断する — 出所によって可否が変わるため。
func (at *AmericanToad) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := at.tableau[col]
	if len(pile) == 0 {
		// リザーブが残っているうちは自動補充の対象なので、手で置くことはできない。
		return len(at.reserve) == 0
	}
	top := pile[len(pile)-1].Card
	if top == nil {
		return false
	}
	return card.GetDesign() == top.GetDesign() &&
		card.GetValue() == americanToadPrevRank(top.GetValue())
}

// canPlaceOnFoundation 基礎札に置けるか（空なら開始ランク、以降は同スート昇順で K→A 折り返し）
func (at *AmericanToad) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil || at.baseRank == 0 {
		return false
	}
	if americanToadSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	pile := at.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == at.baseRank
	}
	if len(pile) >= AmericanToadFoundationTarget {
		return false
	}
	return card.GetValue() == americanToadNextRank(pile[len(pile)-1].GetValue())
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）
func (at *AmericanToad) findFoundation(card *Card) int {
	for i := range AmericanToadFoundationCnt {
		if at.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// fillEmptyColumnsFromReserve 空き列をリザーブの一番上で自動的に埋める。
// リザーブが尽きたら止まり、以降は捨て札から手で埋めることになる。
func (at *AmericanToad) fillEmptyColumnsFromReserve() {
	for col := range AmericanToadTableauCnt {
		if len(at.tableau[col]) > 0 {
			continue
		}
		card := at.reserveTop()
		if card == nil {
			return
		}
		at.popReserve()
		at.tableau[col] = append(at.tableau[col], &AmericanToadTableauCard{Card: card, FaceUp: true})
	}
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (at *AmericanToad) afterMove(actionType, detail string, card *Card) {
	at.moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	at.appendLog(actionType, detail, cards)
	at.checkGameClear()
	at.checkStalemate()
}

// checkGameClear 8 つの基礎札がすべて 13 枚になったか
func (at *AmericanToad) checkGameClear() {
	for i := range AmericanToadFoundationCnt {
		if len(at.foundation[i]) != AmericanToadFoundationTarget {
			return
		}
	}
	at.phase = AmericanToadPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (at *AmericanToad) checkStalemate() {
	if at.phase != AmericanToadPhasePlaying {
		return
	}
	at.isStalemate = at.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (at *AmericanToad) takeSnapshot() {
	snap := &americanToadSnapshot{
		reserve:     append([]*Card(nil), at.reserve...),
		stock:       append([]*Card(nil), at.stock...),
		waste:       append([]*Card(nil), at.waste...),
		baseRank:    at.baseRank,
		passesUsed:  at.passesUsed,
		phase:       at.phase,
		moveCount:   at.moveCount,
		isStalemate: at.isStalemate,
	}
	for i := range AmericanToadFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), at.foundation[i]...)
	}
	for i := range AmericanToadTableauCnt {
		snap.tableau[i] = append([]*AmericanToadTableauCard(nil), at.tableau[i]...)
	}
	at.history = append(at.history, snap)
}

// appendLog 棋譜エントリを追加
func (at *AmericanToad) appendLog(actionType, detail string, cards []*Card) {
	at.actionLog = append(at.actionLog, &ActionLogEntry{
		TurnNumber: at.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// americanToadJSON is the JSON wire format for AmericanToad.
type americanToadJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Reserve     []*Card                                            `json:"rs"`
	Tableau     [AmericanToadTableauCnt][]*AmericanToadTableauCard `json:"tb"`
	Foundation  [AmericanToadFoundationCnt][]*Card                 `json:"fd"`
	Stock       []*Card                                            `json:"st"`
	Waste       []*Card                                            `json:"ws"`
	BaseRank    int                                                `json:"br"`
	PassesUsed  int                                                `json:"pu"`
	Phase       AmericanToadPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production.
	History []*americanToadSnapshot `json:"hi,omitempty"`
}

// americanToadMaxSliceLen caps slice sizes during deserialisation.
const americanToadMaxSliceLen = 1000

// americanToadSnapshotJSON is the wire format for a single undo snapshot.
// americanToadSnapshot uses unexported fields, so we project to/from this
// shape with explicit Marshal/Unmarshal methods. Field names match
// americanToadJSON's short keys to keep the KV payload compact.
type americanToadSnapshotJSON struct {
	Reserve     []*Card                                            `json:"rs"`
	Tableau     [AmericanToadTableauCnt][]*AmericanToadTableauCard `json:"tb"`
	Foundation  [AmericanToadFoundationCnt][]*Card                 `json:"fd"`
	Stock       []*Card                                            `json:"st"`
	Waste       []*Card                                            `json:"ws"`
	BaseRank    int                                                `json:"br"`
	PassesUsed  int                                                `json:"pu"`
	Phase       AmericanToadPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
}

// MarshalJSON implements json.Marshaler for americanToadSnapshot.
func (s *americanToadSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(americanToadSnapshotJSON{
		Reserve:     s.reserve,
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		BaseRank:    s.baseRank,
		PassesUsed:  s.passesUsed,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for americanToadSnapshot.
func (s *americanToadSnapshot) UnmarshalJSON(data []byte) error {
	var j americanToadSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > americanToadMaxSliceLen || len(j.Stock) > americanToadMaxSliceLen ||
		len(j.Waste) > americanToadMaxSliceLen {
		return errors.New("americantoad: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > americanToadMaxSliceLen {
			return errors.New("americantoad: snapshot tableau column exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > americanToadMaxSliceLen {
			return errors.New("americantoad: snapshot foundation exceeds maximum allowed size")
		}
	}
	s.reserve = j.Reserve
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.baseRank = j.BaseRank
	s.passesUsed = j.PassesUsed
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// MarshalJSON KV スナップショット用のシリアライズ
func (at *AmericanToad) MarshalJSON() ([]byte, error) {
	return json.Marshal(&americanToadJSON{
		TrumpCards:  at.trumpCards,
		Reserve:     at.reserve,
		Tableau:     at.tableau,
		Foundation:  at.foundation,
		Stock:       at.stock,
		Waste:       at.waste,
		BaseRank:    at.baseRank,
		PassesUsed:  at.passesUsed,
		Phase:       at.phase,
		MoveCount:   at.moveCount,
		ActionLog:   at.actionLog,
		IsStalemate: at.isStalemate,
		History:     at.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (at *AmericanToad) UnmarshalJSON(data []byte) error {
	var j americanToadJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < AmericanToadPhasePlaying || j.Phase > AmericanToadPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if j.BaseRank < 0 || j.BaseRank > CardValueMax {
		return fmt.Errorf("invalid base rank: %d", j.BaseRank)
	}
	if j.PassesUsed < 0 || j.PassesUsed >= AmericanToadMaxPasses {
		return fmt.Errorf("invalid pass count: %d", j.PassesUsed)
	}
	if len(j.Reserve) > AmericanToadTotalCards {
		return fmt.Errorf("reserve holds %d cards", len(j.Reserve))
	}
	if len(j.Stock) > AmericanToadTotalCards || len(j.Waste) > AmericanToadTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > americanToadMaxSliceLen || len(j.History) > americanToadMaxSliceLen {
		return errors.New("americantoad: input array exceeds maximum allowed size")
	}
	for i := range AmericanToadFoundationCnt {
		if len(j.Foundation[i]) > AmericanToadFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range AmericanToadTableauCnt {
		if len(j.Tableau[i]) > AmericanToadTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		at.trumpCards = j.TrumpCards
	}
	at.reserve = j.Reserve
	at.tableau = j.Tableau
	at.foundation = j.Foundation
	at.stock = j.Stock
	at.waste = j.Waste
	at.baseRank = j.BaseRank
	at.passesUsed = j.PassesUsed
	at.phase = j.Phase
	at.moveCount = j.MoveCount
	at.actionLog = j.ActionLog
	at.isStalemate = j.IsStalemate
	at.history = j.History
	return nil
}
