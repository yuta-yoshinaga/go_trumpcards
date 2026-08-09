//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// TerracePhase テラスのゲームフェーズ
type TerracePhase int

// Terraceのフェーズ定数
const (
	// TerracePhasePlaying プレイ中
	TerracePhasePlaying TerracePhase = iota
	// TerracePhaseGameClear ゲームクリア
	TerracePhaseGameClear
	// TerracePhaseGameOver ゲームオーバー
	TerracePhaseGameOver
)

// TerraceReserveSize テラス（リザーブ）の初期枚数
const TerraceReserveSize = 11

// TerraceTableauCnt タブローの山数
const TerraceTableauCnt = 9

// TerraceFoundationCnt 基礎札の数
const TerraceFoundationCnt = 8

// TerraceFoundationTarget 基礎札 1 つあたりの完成枚数
const TerraceFoundationTarget = CardValueMax

// TerraceTotalCards 使用する総枚数（52 枚 2 組）
const TerraceTotalCards = CardCnt * 2

// terraceMaxSliceLen caps slice sizes during deserialisation.
const terraceMaxSliceLen = 1000

// TerraceHint テラスのヒント
type TerraceHint struct {
	// FromZone 移動元 "reserve" / "waste" / "tableau" / "stock"
	FromZone string
	// FromIdx 移動元のタブロー山（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// Terrace テラス（別名クイーン・オブ・イタリー）ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**11 枚のテラス**（リザーブ）を表向きに
// 並べ、**9 つのタブロー山**に 1 枚ずつ配り、残り 84 枚が山札になる。
//
// このゲームの署名的な規則は 2 つ:
//
//   - **基礎札は「色違い」の昇順で積む。同スートではない。** 開始ランクはプレイヤー
//     が選び、K の次は A に折り返して 13 枚で 1 つ完成。8×13 = 104 枚でクリア。
//     色違いなので基礎札はスートに紐づかない。
//   - **テラスの一番上は基礎札にしか出せない。** タブローへは動かせない。テラスは
//     「置き場」ではなく「基礎札への供給口」であり、これがゲームの緊張を作っている。
//
// タブローも色違いの降順だが、**1 枚ずつしか動かせない**（連番のまとめ移動は無い）。
// タブローの空きは**捨て札か山札から自動的に**埋まる。山札は 1 枚ずつ捨て札へめくり、
// **めくり直しは無い**。
//
// issue #4381 の仕様案とは 4 点異なり、いずれも実際の規則に合わせた:
//   - 基礎札は「同スートで昇順」ではなく **色違いで昇順**。これはこのゲームの要で、
//     同スートにすると別のゲームになる
//   - **テラスの札はタブローへ動かせない**（issue は「テラスとタブローの最上段カード
//     のみ移動可能」とだけ書いており、テラス→タブローを許してしまう）
//   - 補充されるのはテラスではなく**タブローの空き**で、しかも捨て札か山札から自動。
//     テラスは減る一方で補充されない
//   - タブローは 9 山（issue は数に触れていない）
type Terrace struct {
	trumpCards *TrumpCards
	reserve    []*Card
	tableau    [TerraceTableauCnt][]*Card
	foundation [TerraceFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	baseRank   int
	phase      TerracePhase
	moveCount  int
	actionLogBase
	history     []*terraceSnapshot
	isStalemate bool
}

// terraceSnapshot アンドゥ用スナップショット
type terraceSnapshot struct {
	reserve     []*Card
	tableau     [TerraceTableauCnt][]*Card
	foundation  [TerraceFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	baseRank    int
	phase       TerracePhase
	moveCount   int
	isStalemate bool
}

// NewTerrace コンストラクタ
func NewTerrace(trumpCards *TrumpCards) *Terrace {
	return &Terrace{trumpCards: trumpCards}
}

// NewDefaultTerrace returns Terrace with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultTerrace() *Terrace {
	return NewTerrace(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (t *Terrace) Reset() {
	t.trumpCards.Shuffle()
	t.phase = TerracePhasePlaying
	t.moveCount = 0
	t.actionLog = nil
	t.history = nil
	t.isStalemate = false
	t.baseRank = 0
	t.reserve = nil
	t.stock = nil
	t.waste = nil

	for i := range TerraceFoundationCnt {
		t.foundation[i] = nil
	}
	for range TerraceReserveSize {
		if card := t.trumpCards.DrawCard(); card != nil {
			t.reserve = append(t.reserve, card)
		}
	}
	for i := range TerraceTableauCnt {
		t.tableau[i] = nil
		if card := t.trumpCards.DrawCard(); card != nil {
			t.tableau[i] = append(t.tableau[i], card)
		}
	}
	for {
		card := t.trumpCards.DrawCard()
		if card == nil {
			break
		}
		t.stock = append(t.stock, card)
	}

	t.checkStalemate()
}

// IsAwaitingBaseRank 開始ランクがまだ選ばれていないか。
//
// フェーズを増やさず専用フラグにしているのは、他のソリティアが Playing/GameClear/
// GameOver の 3 値で揃っており、ここだけ 4 値にすると Web/フロントの phase 番号が
// 全ゲームでずれるため（Duchess と同じ判断）。
func (t *Terrace) IsAwaitingBaseRank() bool {
	return t.phase == TerracePhasePlaying && t.baseRank == 0
}

// GetBaseRank 基礎札の開始ランク（未選択なら 0）
func (t *Terrace) GetBaseRank() int { return t.baseRank }

// Draw 山札から捨て札へ 1 枚めくる。めくり直しは無い。
func (t *Terrace) Draw() error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	if len(t.stock) == 0 {
		return errors.New("stock is empty and there is no redeal")
	}
	t.takeSnapshot()
	card := t.stock[0]
	t.stock = t.stock[1:]
	t.waste = append(t.waste, card)
	t.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveReserveToFoundation テラスの一番上を基礎札へ送る。
//
// テラスの札の行き先はここだけ。開始ランクが未選択ならこの手が決定にもなる。
func (t *Terrace) MoveReserveToFoundation() error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	card := t.reserveTop()
	if card == nil {
		return errors.New("reserve is empty")
	}
	fIdx := t.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	t.takeSnapshot()
	t.popReserve()
	t.placeOnFoundation(card, fIdx)
	t.fillEmptyColumns()
	t.afterMove("move", fmt.Sprintf("テラス→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (t *Terrace) MoveWasteToFoundation() error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	card := t.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	fIdx := t.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	t.takeSnapshot()
	t.popWaste()
	t.placeOnFoundation(card, fIdx)
	t.fillEmptyColumns()
	t.afterMove("move", fmt.Sprintf("捨て札→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToTableau 捨て札の一番上をタブローへ送る
func (t *Terrace) MoveWasteToTableau(pile int) error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	if err := validTerracePile(pile); err != nil {
		return err
	}
	card := t.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	if !t.canPlaceOnTableau(card, pile) {
		return errors.New("card cannot be placed on that pile")
	}
	t.takeSnapshot()
	t.popWaste()
	t.tableau[pile] = append(t.tableau[pile], card)
	t.fillEmptyColumns()
	t.afterMove("move", fmt.Sprintf("捨て札→タブロー山%d", pile), card)
	return nil
}

// MoveTableauToFoundation タブローの一番上を基礎札へ送る
func (t *Terrace) MoveTableauToFoundation(pile int) error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	if err := validTerracePile(pile); err != nil {
		return err
	}
	card := t.tableauTop(pile)
	if card == nil {
		return fmt.Errorf("pile %d is empty", pile)
	}
	fIdx := t.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	t.takeSnapshot()
	t.tableau[pile] = t.tableau[pile][:len(t.tableau[pile])-1]
	t.placeOnFoundation(card, fIdx)
	t.fillEmptyColumns()
	t.afterMove("move", fmt.Sprintf("タブロー山%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間で 1 枚だけ動かす。連番のまとめ移動は無い。
func (t *Terrace) MoveTableauToTableau(fromPile, toPile int) error {
	if err := t.requirePlaying(); err != nil {
		return err
	}
	if err := validTerracePile(fromPile); err != nil {
		return err
	}
	if err := validTerracePile(toPile); err != nil {
		return err
	}
	if fromPile == toPile {
		return errors.New("source and destination are the same pile")
	}
	card := t.tableauTop(fromPile)
	if card == nil {
		return fmt.Errorf("pile %d is empty", fromPile)
	}
	// 空き山は捨て札か山札から自動で埋まるので、手で埋める対象ではない。
	if len(t.tableau[toPile]) == 0 {
		return errors.New("an empty pile fills itself from the waste or the stock")
	}
	if !t.canPlaceOnTableau(card, toPile) {
		return errors.New("card cannot be placed on that pile")
	}
	t.takeSnapshot()
	t.tableau[fromPile] = t.tableau[fromPile][:len(t.tableau[fromPile])-1]
	t.tableau[toPile] = append(t.tableau[toPile], card)
	t.fillEmptyColumns()
	t.afterMove("move", fmt.Sprintf("タブロー山%d→タブロー山%d", fromPile, toPile), card)
	return nil
}

// GiveUp ギブアップ
func (t *Terrace) GiveUp() {
	if t.phase == TerracePhasePlaying {
		t.phase = TerracePhaseGameOver
		t.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (t *Terrace) GetHint() *TerraceHint {
	if t.phase != TerracePhasePlaying {
		return nil
	}
	if h := t.foundationHint(); h != nil {
		return h
	}
	if h := t.tableauHint(); h != nil {
		return h
	}
	if len(t.stock) > 0 {
		return &TerraceHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (t *Terrace) foundationHint() *TerraceHint {
	if t.phase != TerracePhasePlaying {
		return nil
	}
	// テラスは基礎札にしか出せないので、詰まりやすいこちらを先に消化させる。
	if card := t.reserveTop(); card != nil {
		if fIdx := t.findFoundation(card); fIdx >= 0 {
			return &TerraceHint{FromZone: "reserve", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := t.wasteTop(); card != nil {
		if fIdx := t.findFoundation(card); fIdx >= 0 {
			return &TerraceHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for pile := range TerraceTableauCnt {
		card := t.tableauTop(pile)
		if card == nil {
			continue
		}
		if fIdx := t.findFoundation(card); fIdx >= 0 {
			return &TerraceHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
func (t *Terrace) tableauHint() *TerraceHint {
	if t.phase != TerracePhasePlaying {
		return nil
	}
	if card := t.wasteTop(); card != nil {
		for pile := range TerraceTableauCnt {
			if t.canPlaceOnTableau(card, pile) {
				return &TerraceHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: pile}
			}
		}
	}
	for from := range TerraceTableauCnt {
		card := t.tableauTop(from)
		if card == nil {
			continue
		}
		for to := range TerraceTableauCnt {
			// 空き山は自動補充の対象なので手では埋めない。
			if from == to || len(t.tableau[to]) == 0 {
				continue
			}
			if t.canPlaceOnTableau(card, to) {
				return &TerraceHint{FromZone: "tableau", FromIdx: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (t *Terrace) AutoComplete() error {
	if t.phase != TerracePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := t.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch h.FromZone {
		case "reserve":
			err = t.MoveReserveToFoundation()
		case "waste":
			err = t.MoveWasteToFoundation()
		default:
			err = t.MoveTableauToFoundation(h.FromIdx)
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
func (t *Terrace) Undo() error {
	if len(t.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := t.history[len(t.history)-1]
	t.history = t.history[:len(t.history)-1]
	t.reserve = snap.reserve
	t.tableau = snap.tableau
	t.foundation = snap.foundation
	t.stock = snap.stock
	t.waste = snap.waste
	t.baseRank = snap.baseRank
	t.phase = snap.phase
	t.moveCount = snap.moveCount
	t.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (t *Terrace) CanUndo() bool { return len(t.history) > 0 }

// UndoN n 手戻す
func (t *Terrace) UndoN(n int) error {
	return undoNChecked(t, n, len(t.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (t *Terrace) UndoToEscape() int {
	return undoToEscape(t.isStalemate, t.history, func(s *terraceSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (t *Terrace) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (t *Terrace) GetPhase() TerracePhase { return t.phase }

// GetMoveCount 手数取得
func (t *Terrace) GetMoveCount() int { return t.moveCount }

// GetStockCount 山札の残り枚数
func (t *Terrace) GetStockCount() int { return len(t.stock) }

// GetWaste 捨て札を取得
func (t *Terrace) GetWaste() []*Card { return t.waste }

// GetReserve テラス（リザーブ）を取得
func (t *Terrace) GetReserve() []*Card { return t.reserve }

// GetTableau タブローを取得
func (t *Terrace) GetTableau() [TerraceTableauCnt][]*Card { return t.tableau }

// GetFoundation 基礎札を取得
func (t *Terrace) GetFoundation() [TerraceFoundationCnt][]*Card { return t.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (t *Terrace) GetGameEndFlag() bool { return t.phase != TerracePhasePlaying }

// IsStalemate 手詰まりか
func (t *Terrace) IsStalemate() bool { return t.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (t *Terrace) requirePlaying() error {
	if t.phase != TerracePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validTerracePile タブロー山のインデックスを検証する
func validTerracePile(pile int) error {
	if pile < 0 || pile >= TerraceTableauCnt {
		return fmt.Errorf("invalid pile: %d", pile)
	}
	return nil
}

// terraceIsRed 赤いスートか
func terraceIsRed(design int) bool {
	return design == CardDesignHeart || design == CardDesignDiamond
}

// terraceNextRank K の次は A に折り返す昇順
func terraceNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// terracePrevRank A の次は K に折り返す降順
func terracePrevRank(v int) int {
	if v <= 1 {
		return CardValueMax
	}
	return v - 1
}

// reserveTop テラスの一番上（空なら nil）
func (t *Terrace) reserveTop() *Card {
	if len(t.reserve) == 0 {
		return nil
	}
	return t.reserve[len(t.reserve)-1]
}

// popReserve テラスの一番上を取り除く
func (t *Terrace) popReserve() {
	if len(t.reserve) > 0 {
		t.reserve = t.reserve[:len(t.reserve)-1]
	}
}

// wasteTop 捨て札の一番上（空なら nil）
func (t *Terrace) wasteTop() *Card {
	if len(t.waste) == 0 {
		return nil
	}
	return t.waste[len(t.waste)-1]
}

// popWaste 捨て札の一番上を取り除く
func (t *Terrace) popWaste() {
	if len(t.waste) > 0 {
		t.waste = t.waste[:len(t.waste)-1]
	}
}

// tableauTop 山の一番上（空なら nil）
func (t *Terrace) tableauTop(pile int) *Card {
	if len(t.tableau[pile]) == 0 {
		return nil
	}
	return t.tableau[pile][len(t.tableau[pile])-1]
}

// canPlaceOnTableau タブローに置けるか（色違いの降順、A の下には K）。
// 空き山は自動補充の対象なので、置けるかは呼び出し側で見る。
func (t *Terrace) canPlaceOnTableau(card *Card, pile int) bool {
	if card == nil {
		return false
	}
	if len(t.tableau[pile]) == 0 {
		return true
	}
	top := t.tableau[pile][len(t.tableau[pile])-1]
	if top == nil {
		return false
	}
	return card.GetValue() == terracePrevRank(top.GetValue()) &&
		terraceIsRed(card.GetDesign()) != terraceIsRed(top.GetDesign())
}

// canPlaceOnFoundation 基礎札に置けるか。
//
// **色違いの昇順**で、同スートではない。これがこのゲームの要。開始ランクが未選択の
// うちは、どの札でも空の基礎札に置けて、その札のランクが開始ランクになる。
func (t *Terrace) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := t.foundation[fIdx]
	if len(pile) == 0 {
		// 開始ランク未選択なら、この 1 枚がそれを決める。
		return t.baseRank == 0 || card.GetValue() == t.baseRank
	}
	if len(pile) >= TerraceFoundationTarget {
		return false
	}
	top := pile[len(pile)-1]
	return card.GetValue() == terraceNextRank(top.GetValue()) &&
		terraceIsRed(card.GetDesign()) != terraceIsRed(top.GetDesign())
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
//
// 空の基礎札より先に積み増せる山を探す。開始ランク未選択のときは空の山しか
// 候補にならないので、この順序でも取りこぼさない。
func (t *Terrace) findFoundation(card *Card) int {
	for i := range TerraceFoundationCnt {
		if len(t.foundation[i]) > 0 && t.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	for i := range TerraceFoundationCnt {
		if len(t.foundation[i]) == 0 && t.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// placeOnFoundation 基礎札に置く。最初の 1 枚は開始ランクも決める。
func (t *Terrace) placeOnFoundation(card *Card, fIdx int) {
	if t.baseRank == 0 {
		t.baseRank = card.GetValue()
	}
	t.foundation[fIdx] = append(t.foundation[fIdx], card)
}

// fillEmptyColumns 空いたタブロー山を捨て札（無ければ山札）から自動的に埋める。
// テラスは補充されない — 減る一方である。
func (t *Terrace) fillEmptyColumns() {
	for pile := range TerraceTableauCnt {
		if len(t.tableau[pile]) > 0 {
			continue
		}
		switch {
		case len(t.waste) > 0:
			t.tableau[pile] = append(t.tableau[pile], t.waste[len(t.waste)-1])
			t.waste = t.waste[:len(t.waste)-1]
		case len(t.stock) > 0:
			t.tableau[pile] = append(t.tableau[pile], t.stock[0])
			t.stock = t.stock[1:]
		default:
			return
		}
	}
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (t *Terrace) afterMove(actionType, detail string, card *Card) {
	t.moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	t.appendLog(actionType, detail, cards)
	t.checkGameClear()
	t.checkStalemate()
}

// checkGameClear 8 つの基礎札がすべて 13 枚になったか
func (t *Terrace) checkGameClear() {
	for i := range TerraceFoundationCnt {
		if len(t.foundation[i]) != TerraceFoundationTarget {
			return
		}
	}
	t.phase = TerracePhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (t *Terrace) checkStalemate() {
	if t.phase != TerracePhasePlaying {
		return
	}
	t.isStalemate = t.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (t *Terrace) takeSnapshot() {
	snap := &terraceSnapshot{
		reserve:     append([]*Card(nil), t.reserve...),
		stock:       append([]*Card(nil), t.stock...),
		waste:       append([]*Card(nil), t.waste...),
		baseRank:    t.baseRank,
		phase:       t.phase,
		moveCount:   t.moveCount,
		isStalemate: t.isStalemate,
	}
	for i := range TerraceTableauCnt {
		snap.tableau[i] = append([]*Card(nil), t.tableau[i]...)
	}
	for i := range TerraceFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), t.foundation[i]...)
	}
	t.history = append(t.history, snap)
}

// appendLog 棋譜エントリを追加
func (t *Terrace) appendLog(actionType, detail string, cards []*Card) {
	t.appendLogAt(t.moveCount, 0, actionType, detail, cards)
}

// terraceSnapshotJSON is the wire format for a single undo snapshot.
// terraceSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type terraceSnapshotJSON struct {
	Reserve     []*Card                       `json:"rs"`
	Tableau     [TerraceTableauCnt][]*Card    `json:"tb"`
	Foundation  [TerraceFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                       `json:"st"`
	Waste       []*Card                       `json:"ws"`
	BaseRank    int                           `json:"br"`
	Phase       TerracePhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	IsStalemate bool                          `json:"sl"`
}

// MarshalJSON implements json.Marshaler for terraceSnapshot.
func (s *terraceSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(terraceSnapshotJSON{
		Reserve:     s.reserve,
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		BaseRank:    s.baseRank,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for terraceSnapshot.
func (s *terraceSnapshot) UnmarshalJSON(data []byte) error {
	var j terraceSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Reserve) > terraceMaxSliceLen || len(j.Stock) > terraceMaxSliceLen ||
		len(j.Waste) > terraceMaxSliceLen {
		return errors.New("terrace: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > terraceMaxSliceLen {
			return errors.New("terrace: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > terraceMaxSliceLen {
			return errors.New("terrace: snapshot pile exceeds maximum allowed size")
		}
	}
	s.reserve = j.Reserve
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.baseRank = j.BaseRank
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// terraceJSON is the JSON wire format for Terrace.
type terraceJSON struct {
	TrumpCards  *TrumpCards                   `json:"tc"`
	Reserve     []*Card                       `json:"rs"`
	Tableau     [TerraceTableauCnt][]*Card    `json:"tb"`
	Foundation  [TerraceFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                       `json:"st"`
	Waste       []*Card                       `json:"ws"`
	BaseRank    int                           `json:"br"`
	Phase       TerracePhase                  `json:"ps"`
	MoveCount   int                           `json:"mc"`
	ActionLog   []*ActionLogEntry             `json:"al"`
	IsStalemate bool                          `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*terraceSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (t *Terrace) MarshalJSON() ([]byte, error) {
	return json.Marshal(&terraceJSON{
		TrumpCards:  t.trumpCards,
		Reserve:     t.reserve,
		Tableau:     t.tableau,
		Foundation:  t.foundation,
		Stock:       t.stock,
		Waste:       t.waste,
		BaseRank:    t.baseRank,
		Phase:       t.phase,
		MoveCount:   t.moveCount,
		ActionLog:   t.actionLog,
		IsStalemate: t.isStalemate,
		History:     t.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (t *Terrace) UnmarshalJSON(data []byte) error {
	var j terraceJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < TerracePhasePlaying || j.Phase > TerracePhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	// 0 は「未選択」なので許すが、それ以外は実在するランクでなければならない。
	if j.BaseRank < 0 || j.BaseRank > CardValueMax {
		return fmt.Errorf("invalid base rank: %d", j.BaseRank)
	}
	if len(j.Reserve) > TerraceTotalCards || len(j.Stock) > TerraceTotalCards ||
		len(j.Waste) > TerraceTotalCards {
		return errors.New("terrace: reserve/stock/waste too large")
	}
	if len(j.ActionLog) > terraceMaxSliceLen || len(j.History) > terraceMaxSliceLen {
		return errors.New("terrace: input array exceeds maximum allowed size")
	}
	for i := range TerraceFoundationCnt {
		if len(j.Foundation[i]) > TerraceFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range TerraceTableauCnt {
		if len(j.Tableau[i]) > TerraceTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		t.trumpCards = j.TrumpCards
	}
	t.reserve = j.Reserve
	t.tableau = j.Tableau
	t.foundation = j.Foundation
	t.stock = j.Stock
	t.waste = j.Waste
	t.baseRank = j.BaseRank
	t.phase = j.Phase
	t.moveCount = j.MoveCount
	t.actionLog = j.ActionLog
	t.isStalemate = j.IsStalemate
	t.history = j.History
	return nil
}
