//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BraidPhase ブレイドのゲームフェーズ
type BraidPhase int

// Braidのフェーズ定数
const (
	// BraidPhasePlaying プレイ中
	BraidPhasePlaying BraidPhase = iota
	// BraidPhaseGameClear ゲームクリア
	BraidPhaseGameClear
	// BraidPhaseGameOver ゲームオーバー
	BraidPhaseGameOver
)

// BraidSize 三つ編み（ブレイド）の初期枚数
const BraidSize = 20

// BraidFieldCnt ブレイド札の枠数。ここが空くとブレイドから自動補充される。
const BraidFieldCnt = 4

// BraidHelperCnt ヘルパー枠の数。捨て札からのみ埋められる。
const BraidHelperCnt = 8

// BraidFoundationCnt 基礎札の数
const BraidFoundationCnt = 8

// BraidFoundationTarget 基礎札 1 つあたりの完成枚数
const BraidFoundationTarget = CardValueMax

// BraidMaxPasses 山札を通せる回数（めくり直し 2 回を含む）
const BraidMaxPasses = 3

// BraidTotalCards 使用する総枚数（52 枚 2 組）
const BraidTotalCards = CardCnt * 2

// braidMaxSliceLen caps slice sizes during deserialisation.
const braidMaxSliceLen = 1000

// BraidDirection 基礎札を積む向き
type BraidDirection int

// BraidDirectionの定数
const (
	// BraidDirectionUnset 未選択
	BraidDirectionUnset BraidDirection = iota
	// BraidDirectionAscending 昇順
	BraidDirectionAscending
	// BraidDirectionDescending 降順
	BraidDirectionDescending
)

// BraidHint ブレイドのヒント
type BraidHint struct {
	// FromZone 移動元 "braid" / "field" / "helper" / "waste" / "stock"
	FromZone string
	// FromIdx 移動元のブレイド札／ヘルパーの枠（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "helper" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の枠を指さない）
	ToIdx int
}

// Braid ブレイド（三つ編み）ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**20 枚を三つ編み状に重ねたブレイド**、
// その周りに **4 つのブレイド札**と **8 つのヘルパー枠**、そして基礎札 8 つを置く。
// 開始札 1 枚を基礎札に置き、残り 71 枚が山札になる。
//
// このゲームの要は「札の流れが一方通行」であること:
//
//   - **カードは基礎札にしか動かせない。** 枠同士やブレイド→枠の移動は無い
//   - **ブレイド札の枠が空くと、ブレイドの末尾から自動補充される。** ブレイドを
//     掘る唯一の手段がこれで、issue はこの 4 枠に触れていない
//   - **ヘルパー枠は捨て札からしか埋められない。** ブレイドからは埋められない
//
// 基礎札は 8 つ。開始札のランクが 8 つすべての起点になり、**プレイヤーが選んだ
// 一方向**（昇順か降順）に同スートで 13 枚積む。K の次は A に折り返す。
//
// 山札は 1 枚ずつ捨て札へめくり、**めくり直しは 2 回**（通算 3 巡）。
//
// issue #4382 の仕様案とは 5 点異なり、いずれも実際の規則に合わせた:
//   - 向きは「昇順・降順どちらにも進められる」のではなく、**開始時に一方を選び、
//     以後は全基礎札がその向き**で揃う
//   - 枠は「供給札 8 枚」だけでなく、**ヘルパー 8 + ブレイド札 4 の計 12**。
//     後者がブレイドを消化する唯一の経路で、issue はこれを落としている
//   - **めくり直しが 2 回ある**（issue は触れていない）
//   - **カードは基礎札にしか動かせない**（issue は触れていない）
//   - 基礎札は**同スート**で積む（issue は触れていない）
type Braid struct {
	trumpCards *TrumpCards
	braid      []*Card
	fields     [BraidFieldCnt]*Card
	helpers    [BraidHelperCnt]*Card
	foundation [BraidFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	baseRank   int
	direction  BraidDirection
	passesUsed int
	phase      BraidPhase
	moveCount  int
	actionLogBase
	history     []*braidSnapshot
	isStalemate bool
}

// braidSnapshot アンドゥ用スナップショット
type braidSnapshot struct {
	braid       []*Card
	fields      [BraidFieldCnt]*Card
	helpers     [BraidHelperCnt]*Card
	foundation  [BraidFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	baseRank    int
	direction   BraidDirection
	passesUsed  int
	phase       BraidPhase
	moveCount   int
	isStalemate bool
}

// NewBraid コンストラクタ
func NewBraid(trumpCards *TrumpCards) *Braid {
	return &Braid{trumpCards: trumpCards}
}

// NewDefaultBraid returns Braid with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBraid() *Braid {
	return NewBraid(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (b *Braid) Reset() {
	b.trumpCards.Shuffle()
	b.phase = BraidPhasePlaying
	b.moveCount = 0
	b.actionLog = nil
	b.history = nil
	b.isStalemate = false
	b.direction = BraidDirectionUnset
	b.baseRank = 0
	b.passesUsed = 0
	b.braid = nil
	b.stock = nil
	b.waste = nil

	for i := range BraidFoundationCnt {
		b.foundation[i] = nil
	}
	for range BraidSize {
		if card := b.trumpCards.DrawCard(); card != nil {
			b.braid = append(b.braid, card)
		}
	}
	for i := range BraidFieldCnt {
		b.fields[i] = b.trumpCards.DrawCard()
	}
	for i := range BraidHelperCnt {
		b.helpers[i] = b.trumpCards.DrawCard()
	}
	// 開始札は自動で配られ、そのランクが 8 つすべての起点になる。向きだけは
	// プレイヤーが選ぶ。
	if starter := b.trumpCards.DrawCard(); starter != nil {
		b.baseRank = starter.GetValue()
		b.foundation[0] = []*Card{starter}
	}
	for {
		card := b.trumpCards.DrawCard()
		if card == nil {
			break
		}
		b.stock = append(b.stock, card)
	}

	b.checkStalemate()
}

// IsAwaitingDirection 積む向きがまだ選ばれていないか。
//
// フェーズを増やさず専用フラグにしているのは、他のソリティアが Playing/GameClear/
// GameOver の 3 値で揃っており、ここだけ 4 値にすると Web/フロントの phase 番号が
// 全ゲームでずれるため（Duchess・Terrace と同じ判断）。
func (b *Braid) IsAwaitingDirection() bool {
	return b.phase == BraidPhasePlaying && b.direction == BraidDirectionUnset
}

// GetDirection 積む向きを取得する（未選択なら BraidDirectionUnset）
func (b *Braid) GetDirection() BraidDirection { return b.direction }

// GetBaseRank 基礎札の開始ランク
func (b *Braid) GetBaseRank() int { return b.baseRank }

// ChooseDirection 基礎札を積む向きを決める。ゲーム中に一度だけ。
func (b *Braid) ChooseDirection(ascending bool) error {
	if err := b.requirePlaying(); err != nil {
		return err
	}
	if b.direction != BraidDirectionUnset {
		return errors.New("the direction is already fixed")
	}
	b.takeSnapshot()
	if ascending {
		b.direction = BraidDirectionAscending
	} else {
		b.direction = BraidDirectionDescending
	}
	b.afterMove("direction", b.directionDetail(), nil)
	return nil
}

// directionDetail 棋譜に残す向きの説明
func (b *Braid) directionDetail() string {
	if b.direction == BraidDirectionAscending {
		return "基礎札を昇順に決めた"
	}
	return "基礎札を降順に決めた"
}

// Draw 山札から捨て札へ 1 枚めくる。山札が空ならめくり直す（2 回まで）。
func (b *Braid) Draw() error {
	if err := b.requirePlaying(); err != nil {
		return err
	}
	if len(b.stock) == 0 {
		if !b.CanRedeal() {
			return errors.New("no redeal left")
		}
		b.takeSnapshot()
		b.stock = b.waste
		b.waste = nil
		b.passesUsed++
		b.afterMove("redeal", "捨て札を山札に戻した", nil)
		return nil
	}
	b.takeSnapshot()
	card := b.stock[0]
	b.stock = b.stock[1:]
	b.waste = append(b.waste, card)
	b.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// CanRedeal もう一度めくり直せるか
func (b *Braid) CanRedeal() bool {
	return b.phase == BraidPhasePlaying &&
		len(b.stock) == 0 && len(b.waste) > 0 && b.passesUsed < BraidMaxPasses-1
}

// MoveBraidToFoundation ブレイドの末尾を基礎札へ送る
func (b *Braid) MoveBraidToFoundation() error {
	if err := b.requireDirection(); err != nil {
		return err
	}
	card := b.braidTail()
	if card == nil {
		return errors.New("the braid is empty")
	}
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	b.takeSnapshot()
	b.braid = b.braid[:len(b.braid)-1]
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.afterMove("move", fmt.Sprintf("ブレイド→基礎札%d", fIdx), card)
	return nil
}

// MoveFieldToFoundation ブレイド札の枠から基礎札へ送る。
//
// 空いた枠はブレイドの末尾から自動補充される。ブレイドを掘る唯一の経路。
func (b *Braid) MoveFieldToFoundation(idx int) error {
	if err := b.requireDirection(); err != nil {
		return err
	}
	if err := validBraidField(idx); err != nil {
		return err
	}
	card := b.fields[idx]
	if card == nil {
		return fmt.Errorf("braid field %d is empty", idx)
	}
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	b.takeSnapshot()
	b.fields[idx] = nil
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.refillFields()
	b.afterMove("move", fmt.Sprintf("ブレイド札%d→基礎札%d", idx, fIdx), card)
	return nil
}

// MoveHelperToFoundation ヘルパー枠から基礎札へ送る
func (b *Braid) MoveHelperToFoundation(idx int) error {
	if err := b.requireDirection(); err != nil {
		return err
	}
	if err := validBraidHelper(idx); err != nil {
		return err
	}
	card := b.helpers[idx]
	if card == nil {
		return fmt.Errorf("helper %d is empty", idx)
	}
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	b.takeSnapshot()
	b.helpers[idx] = nil
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.afterMove("move", fmt.Sprintf("ヘルパー%d→基礎札%d", idx, fIdx), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (b *Braid) MoveWasteToFoundation() error {
	if err := b.requireDirection(); err != nil {
		return err
	}
	card := b.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	fIdx := b.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	b.takeSnapshot()
	b.popWaste()
	b.foundation[fIdx] = append(b.foundation[fIdx], card)
	b.afterMove("move", fmt.Sprintf("捨て札→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToHelper 捨て札の一番上で空のヘルパー枠を埋める。
//
// ヘルパー枠を埋められるのは捨て札だけ。ブレイドからは埋められない。
func (b *Braid) MoveWasteToHelper(idx int) error {
	if err := b.requirePlaying(); err != nil {
		return err
	}
	if err := validBraidHelper(idx); err != nil {
		return err
	}
	if b.helpers[idx] != nil {
		return fmt.Errorf("helper %d is not empty", idx)
	}
	card := b.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	b.takeSnapshot()
	b.popWaste()
	b.helpers[idx] = card
	b.afterMove("move", fmt.Sprintf("捨て札→ヘルパー%d", idx), card)
	return nil
}

// GiveUp ギブアップ
func (b *Braid) GiveUp() {
	if b.phase == BraidPhasePlaying {
		b.phase = BraidPhaseGameOver
		b.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。手詰まり判定も兼ねる。
func (b *Braid) GetHint() *BraidHint {
	if b.phase != BraidPhasePlaying {
		return nil
	}
	// 向きが未選択なら、まずそれを決める以外にできることはない。
	if b.direction == BraidDirectionUnset {
		return &BraidHint{FromZone: "direction", FromIdx: -1, ToZone: "foundation", ToIdx: -1}
	}
	if h := b.foundationHint(); h != nil {
		return h
	}
	// 捨て札を空きヘルパーへ逃がす手。基礎札に行けない札の唯一の置き場。
	if b.wasteTop() != nil {
		for i := range BraidHelperCnt {
			if b.helpers[i] == nil {
				return &BraidHint{FromZone: "waste", FromIdx: -1, ToZone: "helper", ToIdx: i}
			}
		}
	}
	if len(b.stock) > 0 || b.CanRedeal() {
		return &BraidHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// ブレイド札を先に見るのは、そこがブレイドを消化する唯一の経路だから。
func (b *Braid) foundationHint() *BraidHint {
	if b.phase != BraidPhasePlaying || b.direction == BraidDirectionUnset {
		return nil
	}
	for i := range BraidFieldCnt {
		if card := b.fields[i]; card != nil && b.findFoundation(card) >= 0 {
			return &BraidHint{FromZone: "field", FromIdx: i, ToZone: "foundation", ToIdx: b.findFoundation(card)}
		}
	}
	if card := b.braidTail(); card != nil {
		if fIdx := b.findFoundation(card); fIdx >= 0 {
			return &BraidHint{FromZone: "braid", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for i := range BraidHelperCnt {
		if card := b.helpers[i]; card != nil && b.findFoundation(card) >= 0 {
			return &BraidHint{FromZone: "helper", FromIdx: i, ToZone: "foundation", ToIdx: b.findFoundation(card)}
		}
	}
	if card := b.wasteTop(); card != nil {
		if fIdx := b.findFoundation(card); fIdx >= 0 {
			return &BraidHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (b *Braid) AutoComplete() error {
	if b.phase != BraidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if b.direction == BraidDirectionUnset {
		return errors.New("choose the direction first")
	}
	moved := false
	for {
		h := b.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch h.FromZone {
		case "field":
			err = b.MoveFieldToFoundation(h.FromIdx)
		case "braid":
			err = b.MoveBraidToFoundation()
		case "helper":
			err = b.MoveHelperToFoundation(h.FromIdx)
		default:
			err = b.MoveWasteToFoundation()
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
func (b *Braid) Undo() error {
	if len(b.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := b.history[len(b.history)-1]
	b.history = b.history[:len(b.history)-1]
	b.braid = snap.braid
	b.fields = snap.fields
	b.helpers = snap.helpers
	b.foundation = snap.foundation
	b.stock = snap.stock
	b.waste = snap.waste
	b.baseRank = snap.baseRank
	b.direction = snap.direction
	b.passesUsed = snap.passesUsed
	b.phase = snap.phase
	b.moveCount = snap.moveCount
	b.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (b *Braid) CanUndo() bool { return len(b.history) > 0 }

// UndoN n 手戻す
func (b *Braid) UndoN(n int) error {
	return undoNChecked(b, n, len(b.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (b *Braid) UndoToEscape() int {
	return undoToEscape(b.isStalemate, b.history, func(s *braidSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (b *Braid) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (b *Braid) GetPhase() BraidPhase { return b.phase }

// GetMoveCount 手数取得
func (b *Braid) GetMoveCount() int { return b.moveCount }

// GetStockCount 山札の残り枚数
func (b *Braid) GetStockCount() int { return len(b.stock) }

// GetWaste 捨て札を取得
func (b *Braid) GetWaste() []*Card { return b.waste }

// GetBraid 三つ編みを取得（末尾が使える 1 枚）
func (b *Braid) GetBraid() []*Card { return b.braid }

// GetFields ブレイド札の 4 枠を取得。空き枠は nil。
func (b *Braid) GetFields() [BraidFieldCnt]*Card { return b.fields }

// GetHelpers ヘルパーの 8 枠を取得。空き枠は nil。
func (b *Braid) GetHelpers() [BraidHelperCnt]*Card { return b.helpers }

// GetFoundation 基礎札を取得
func (b *Braid) GetFoundation() [BraidFoundationCnt][]*Card { return b.foundation }

// GetPassesUsed 山札を通した回数
func (b *Braid) GetPassesUsed() int { return b.passesUsed }

// GetGameEndFlag ゲーム終了フラグ
func (b *Braid) GetGameEndFlag() bool { return b.phase != BraidPhasePlaying }

// IsStalemate 手詰まりか
func (b *Braid) IsStalemate() bool { return b.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (b *Braid) requirePlaying() error {
	if b.phase != BraidPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// requireDirection 向きが決まっていなければエラーを返す。
// 基礎札に積むあらゆる手はこれを通る。
func (b *Braid) requireDirection() error {
	if err := b.requirePlaying(); err != nil {
		return err
	}
	if b.direction == BraidDirectionUnset {
		return errors.New("choose the direction first")
	}
	return nil
}

// validBraidField ブレイド札の枠インデックスを検証する
func validBraidField(idx int) error {
	if idx < 0 || idx >= BraidFieldCnt {
		return fmt.Errorf("invalid braid field: %d", idx)
	}
	return nil
}

// validBraidHelper ヘルパー枠のインデックスを検証する
func validBraidHelper(idx int) error {
	if idx < 0 || idx >= BraidHelperCnt {
		return fmt.Errorf("invalid helper: %d", idx)
	}
	return nil
}

// braidNextRank 向きに応じた次のランク。K の次は A、A の次は K に折り返す。
func braidNextRank(v int, dir BraidDirection) int {
	if dir == BraidDirectionDescending {
		if v <= 1 {
			return CardValueMax
		}
		return v - 1
	}
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// braidTail 三つ編みの末尾（空なら nil）
func (b *Braid) braidTail() *Card {
	if len(b.braid) == 0 {
		return nil
	}
	return b.braid[len(b.braid)-1]
}

// wasteTop 捨て札の一番上（空なら nil）
func (b *Braid) wasteTop() *Card {
	if len(b.waste) == 0 {
		return nil
	}
	return b.waste[len(b.waste)-1]
}

// popWaste 捨て札の一番上を取り除く
func (b *Braid) popWaste() {
	if len(b.waste) > 0 {
		b.waste = b.waste[:len(b.waste)-1]
	}
}

// canPlaceOnFoundation 基礎札に置けるか（同スート、選んだ向きに 1 つずつ）
func (b *Braid) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil || b.direction == BraidDirectionUnset {
		return false
	}
	pile := b.foundation[fIdx]
	if len(pile) == 0 {
		// 空の基礎札を開けるのは開始ランクの札だけ。
		return card.GetValue() == b.baseRank
	}
	if len(pile) >= BraidFoundationTarget {
		return false
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() &&
		card.GetValue() == braidNextRank(top.GetValue(), b.direction)
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
// 空の山より先に積み増せる山を見る。
func (b *Braid) findFoundation(card *Card) int {
	for i := range BraidFoundationCnt {
		if len(b.foundation[i]) > 0 && b.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	for i := range BraidFoundationCnt {
		if len(b.foundation[i]) == 0 && b.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// refillFields 空いたブレイド札の枠をブレイドの末尾から自動補充する。
// ヘルパー枠はここでは埋めない — あちらは捨て札専用。
func (b *Braid) refillFields() {
	for i := range BraidFieldCnt {
		if b.fields[i] != nil {
			continue
		}
		if len(b.braid) == 0 {
			return
		}
		b.fields[i] = b.braid[len(b.braid)-1]
		b.braid = b.braid[:len(b.braid)-1]
	}
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (b *Braid) afterMove(actionType, detail string, card *Card) {
	afterMove(&b.moveCount, b, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚になったか
func (b *Braid) checkGameClear() {
	for i := range BraidFoundationCnt {
		if len(b.foundation[i]) != BraidFoundationTarget {
			return
		}
	}
	b.phase = BraidPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は向きの選択・基礎札・ヘルパー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (b *Braid) checkStalemate() {
	if b.phase != BraidPhasePlaying {
		return
	}
	b.isStalemate = b.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (b *Braid) takeSnapshot() {
	snap := &braidSnapshot{
		braid:       append([]*Card(nil), b.braid...),
		fields:      b.fields,
		helpers:     b.helpers,
		stock:       append([]*Card(nil), b.stock...),
		waste:       append([]*Card(nil), b.waste...),
		baseRank:    b.baseRank,
		direction:   b.direction,
		passesUsed:  b.passesUsed,
		phase:       b.phase,
		moveCount:   b.moveCount,
		isStalemate: b.isStalemate,
	}
	for i := range BraidFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), b.foundation[i]...)
	}
	b.history = append(b.history, snap)
}

// appendLog 棋譜エントリを追加
func (b *Braid) appendLog(actionType, detail string, cards []*Card) {
	b.appendLogAt(b.moveCount, 0, actionType, detail, cards)
}

// braidSnapshotJSON is the wire format for a single undo snapshot.
// braidSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type braidSnapshotJSON struct {
	Braid       []*Card                     `json:"br"`
	Fields      [BraidFieldCnt]*Card        `json:"fl"`
	Helpers     [BraidHelperCnt]*Card       `json:"hp"`
	Foundation  [BraidFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                     `json:"st"`
	Waste       []*Card                     `json:"ws"`
	BaseRank    int                         `json:"bk"`
	Direction   BraidDirection              `json:"dr"`
	PassesUsed  int                         `json:"pu"`
	Phase       BraidPhase                  `json:"ps"`
	MoveCount   int                         `json:"mc"`
	IsStalemate bool                        `json:"sl"`
}

// MarshalJSON implements json.Marshaler for braidSnapshot.
func (s *braidSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(braidSnapshotJSON{
		Braid:       s.braid,
		Fields:      s.fields,
		Helpers:     s.helpers,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		BaseRank:    s.baseRank,
		Direction:   s.direction,
		PassesUsed:  s.passesUsed,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for braidSnapshot.
func (s *braidSnapshot) UnmarshalJSON(data []byte) error {
	var j braidSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Braid) > braidMaxSliceLen || len(j.Stock) > braidMaxSliceLen ||
		len(j.Waste) > braidMaxSliceLen {
		return errors.New("braid: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundation {
		if len(pile) > braidMaxSliceLen {
			return errors.New("braid: snapshot pile exceeds maximum allowed size")
		}
	}
	s.braid = j.Braid
	s.fields = j.Fields
	s.helpers = j.Helpers
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.baseRank = j.BaseRank
	s.direction = j.Direction
	s.passesUsed = j.PassesUsed
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// braidJSON is the JSON wire format for Braid.
type braidJSON struct {
	TrumpCards  *TrumpCards                 `json:"tc"`
	Braid       []*Card                     `json:"br"`
	Fields      [BraidFieldCnt]*Card        `json:"fl"`
	Helpers     [BraidHelperCnt]*Card       `json:"hp"`
	Foundation  [BraidFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                     `json:"st"`
	Waste       []*Card                     `json:"ws"`
	BaseRank    int                         `json:"bk"`
	Direction   BraidDirection              `json:"dr"`
	PassesUsed  int                         `json:"pu"`
	Phase       BraidPhase                  `json:"ps"`
	MoveCount   int                         `json:"mc"`
	ActionLog   []*ActionLogEntry           `json:"al"`
	IsStalemate bool                        `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*braidSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (b *Braid) MarshalJSON() ([]byte, error) {
	return json.Marshal(&braidJSON{
		TrumpCards:  b.trumpCards,
		Braid:       b.braid,
		Fields:      b.fields,
		Helpers:     b.helpers,
		Foundation:  b.foundation,
		Stock:       b.stock,
		Waste:       b.waste,
		BaseRank:    b.baseRank,
		Direction:   b.direction,
		PassesUsed:  b.passesUsed,
		Phase:       b.phase,
		MoveCount:   b.moveCount,
		ActionLog:   b.actionLog,
		IsStalemate: b.isStalemate,
		History:     b.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (b *Braid) UnmarshalJSON(data []byte) error {
	var j braidJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < BraidPhasePlaying || j.Phase > BraidPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if j.BaseRank < 0 || j.BaseRank > CardValueMax {
		return fmt.Errorf("invalid base rank: %d", j.BaseRank)
	}
	if j.Direction < BraidDirectionUnset || j.Direction > BraidDirectionDescending {
		return fmt.Errorf("invalid direction: %d", j.Direction)
	}
	if j.PassesUsed < 0 || j.PassesUsed >= BraidMaxPasses {
		return fmt.Errorf("invalid pass count: %d", j.PassesUsed)
	}
	if len(j.Braid) > BraidTotalCards || len(j.Stock) > BraidTotalCards ||
		len(j.Waste) > BraidTotalCards {
		return errors.New("braid: braid/stock/waste too large")
	}
	if len(j.ActionLog) > braidMaxSliceLen || len(j.History) > braidMaxSliceLen {
		return errors.New("braid: input array exceeds maximum allowed size")
	}
	for i := range BraidFoundationCnt {
		if len(j.Foundation[i]) > BraidFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	if j.TrumpCards != nil {
		b.trumpCards = j.TrumpCards
	}
	b.braid = j.Braid
	b.fields = j.Fields
	b.helpers = j.Helpers
	b.foundation = j.Foundation
	b.stock = j.Stock
	b.waste = j.Waste
	b.baseRank = j.BaseRank
	b.direction = j.Direction
	b.passesUsed = j.PassesUsed
	b.phase = j.Phase
	b.moveCount = j.MoveCount
	b.actionLog = j.ActionLog
	b.isStalemate = j.IsStalemate
	b.history = j.History
	return nil
}
