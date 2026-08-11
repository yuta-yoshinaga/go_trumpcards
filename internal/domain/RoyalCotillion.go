//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// RoyalCotillionPhase ロイヤルコティヨンのゲームフェーズ
type RoyalCotillionPhase int

// RoyalCotillionのフェーズ定数
const (
	// RoyalCotillionPhasePlaying プレイ中
	RoyalCotillionPhasePlaying RoyalCotillionPhase = iota
	// RoyalCotillionPhaseGameClear ゲームクリア
	RoyalCotillionPhaseGameClear
	// RoyalCotillionPhaseGameOver ゲームオーバー
	RoyalCotillionPhaseGameOver
)

// RoyalCotillionTableauCnt タブローの枠数。各枠にカードは 1 枚しか置けない。
const RoyalCotillionTableauCnt = 16

// RoyalCotillionReserveCnt リザーブの山数
const RoyalCotillionReserveCnt = 4

// RoyalCotillionReserveDepth リザーブ 1 山あたりの枚数
const RoyalCotillionReserveDepth = 3

// RoyalCotillionFoundationCnt 基礎札の数。スートごとに A 始まりと 2 始まりの 2 本。
const RoyalCotillionFoundationCnt = 8

// RoyalCotillionOddCnt A 始まりの基礎札の数（インデックス 0 以上この値未満）
const RoyalCotillionOddCnt = RoyalCotillionFoundationCnt / 2

// RoyalCotillionFoundationTarget 基礎札 1 つあたりの完成枚数。
//
// 2 つ飛ばしでも**折り返す**ので、1 本で 13 枚すべてを通る。
// A→3→…→K→2→4→…→Q の順で 13 枚。
const RoyalCotillionFoundationTarget = CardValueMax

// RoyalCotillionTotalCards 使用する総枚数（52 枚 2 組）
const RoyalCotillionTotalCards = CardCnt * 2

// royalCotillionMaxSliceLen caps slice sizes during deserialisation.
const royalCotillionMaxSliceLen = 1000

// royalCotillionSuitOrder 基礎札インデックスとスートの対応。前半 4 つが A 始まり、
// 後半 4 つが 2 始まりで、どちらも同じスート順に並ぶ。
var royalCotillionSuitOrder = [RoyalCotillionFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// RoyalCotillionHint ロイヤルコティヨンのヒント
type RoyalCotillionHint struct {
	// FromZone 移動元 "tableau" / "reserve" / "waste" / "stock"
	FromZone string
	// FromIdx 移動元の枠・山（捨て札と山札は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の場所を指さない）
	ToIdx int
}

// RoyalCotillion ロイヤルコティヨン ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**16 枠のタブロー**に 1 枚ずつ、
// **4 山のリザーブ**に 3 枚ずつ配り、残り 76 枚が山札になる。
//
// 基礎札は 8 本。**2 つ飛ばしで積む**のがこのゲームの特徴で、前半 4 本は A から、
// 後半 4 本は 2 から始まる。ただし**折り返す**ので、A 始まりの本は
// A→3→5→7→9→J→K→2→4→6→8→10→Q の 13 枚、2 始まりの本は
// 2→4→6→8→10→Q→A→3→5→7→9→J→K の 13 枚を通る。8×13 = 104 枚でクリア。
//
// タブローは**1 枠 1 枚**で重ねられず、空いた枠は山札か捨て札から補充される。
// リザーブは**一番上だけ**が使え、**空いた山は二度と埋まらない**。この非対称が
// このゲームの緊張で、リザーブを掘るほど選択肢は増えるが枠は戻らない。
//
// issue #5275 の仕様案とは 4 点異なり、いずれも実際の規則に合わせた:
//   - **基礎札は 16 本ではなく 8 本。** 8 本が折り返して 13 枚ずつ通るので
//     8×13 = 104 枚とちょうど一致する。issue の「奇数側 7 枚 + 偶数側 6 枚 ×
//     16 本」も合計 104 になってしまうため、**枚数だけでは見分けられない**。
//     決め手は折り返しの有無で、どの規則書も K の次は 2、Q の次は A と書いている
//   - **盤面は 4×3 のグリッド 2 つではない。** 1 枚ずつの枠が 16、3 枚重ねの
//     リザーブが 4 山（計 28 枚）で、山札は 76 枚になる
//   - **補充されるのはタブロー枠**で、リザーブは補充されない。issue は
//     「右グリッドが補充、左は補充されない」としており非対称の向きは合っているが、
//     補充される側は「1 枚ずつの枠」であって「3 枚重ねの山」ではない
//   - **スートは 4 つ。** issue の「8 スート×2 系統」は数え違いで、
//     8 という数は 4 スート × 2 系統の**基礎札の本数**である
type RoyalCotillion struct {
	trumpCards *TrumpCards
	// tableau は 1 枠 1 枚。空き枠は nil。
	tableau    [RoyalCotillionTableauCnt]*Card
	reserve    [RoyalCotillionReserveCnt][]*Card
	foundation [RoyalCotillionFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      RoyalCotillionPhase
	moveCount  int
	actionLogBase
	history     []*royalCotillionSnapshot
	isStalemate bool
}

// royalCotillionSnapshot アンドゥ用スナップショット
type royalCotillionSnapshot struct {
	tableau     [RoyalCotillionTableauCnt]*Card
	reserve     [RoyalCotillionReserveCnt][]*Card
	foundation  [RoyalCotillionFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	phase       RoyalCotillionPhase
	moveCount   int
	isStalemate bool
}

// NewRoyalCotillion コンストラクタ
func NewRoyalCotillion(trumpCards *TrumpCards) *RoyalCotillion {
	return &RoyalCotillion{trumpCards: trumpCards}
}

// NewDefaultRoyalCotillion returns RoyalCotillion with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultRoyalCotillion() *RoyalCotillion {
	return NewRoyalCotillion(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *RoyalCotillion) Reset() {
	c.trumpCards.Shuffle()
	c.phase = RoyalCotillionPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil

	for i := range RoyalCotillionFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range RoyalCotillionTableauCnt {
		c.tableau[i] = c.trumpCards.DrawCard()
	}
	for i := range RoyalCotillionReserveCnt {
		c.reserve[i] = nil
		for range RoyalCotillionReserveDepth {
			card := c.trumpCards.DrawCard()
			if card == nil {
				break
			}
			c.reserve[i] = append(c.reserve[i], card)
		}
	}
	for {
		card := c.trumpCards.DrawCard()
		if card == nil {
			break
		}
		c.stock = append(c.stock, card)
	}

	c.checkStalemate()
}

// Draw 山札から捨て札へ 1 枚めくる。めくり直しは無い。
func (c *RoyalCotillion) Draw() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return errors.New("stock is empty and there is no redeal")
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.waste = append(c.waste, card)
	c.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveTableauToFoundation タブロー枠の札を基礎札へ送る
func (c *RoyalCotillion) MoveTableauToFoundation(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validRoyalCotillionSlot(slot); err != nil {
		return err
	}
	card := c.tableau[slot]
	if card == nil {
		return fmt.Errorf("slot %d is empty", slot)
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	c.takeSnapshot()
	c.tableau[slot] = nil
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("タブロー枠%d→基礎札%d", slot, fIdx), card)
	return nil
}

// MoveReserveToFoundation リザーブの一番上を基礎札へ送る
func (c *RoyalCotillion) MoveReserveToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validRoyalCotillionReserve(pile); err != nil {
		return err
	}
	card := c.reserveTop(pile)
	if card == nil {
		return fmt.Errorf("reserve %d is empty", pile)
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	c.takeSnapshot()
	c.reserve[pile] = dropLast(c.reserve[pile])
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("リザーブ%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *RoyalCotillion) MoveWasteToFoundation() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	card := c.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	c.takeSnapshot()
	c.popWaste()
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("捨て札→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToTableau 捨て札の一番上で空いたタブロー枠を埋める。
//
// タブローは 1 枠 1 枚なので、埋められるのは空き枠だけ。
func (c *RoyalCotillion) MoveWasteToTableau(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validRoyalCotillionSlot(slot); err != nil {
		return err
	}
	if c.tableau[slot] != nil {
		return errors.New("the slot already holds a card")
	}
	card := c.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	c.takeSnapshot()
	c.popWaste()
	c.tableau[slot] = card
	c.afterMove("move", fmt.Sprintf("捨て札→タブロー枠%d", slot), card)
	return nil
}

// MoveStockToTableau 山札の一番上で空いたタブロー枠を直接埋める。
//
// 規則上、空き枠は「山札か捨て札」から埋められる。捨て札を経由させると
// 山札の 1 巡が 1 枚分早く減るので、山札から直接置く手を別に用意している。
func (c *RoyalCotillion) MoveStockToTableau(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validRoyalCotillionSlot(slot); err != nil {
		return err
	}
	if c.tableau[slot] != nil {
		return errors.New("the slot already holds a card")
	}
	if len(c.stock) == 0 {
		return errors.New("stock is empty")
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.tableau[slot] = card
	c.afterMove("move", fmt.Sprintf("山札→タブロー枠%d", slot), card)
	return nil
}

// GiveUp ギブアップ
func (c *RoyalCotillion) GiveUp() {
	if c.phase == RoyalCotillionPhasePlaying {
		c.phase = RoyalCotillionPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → 空き枠埋め → 山めくり の順。手詰まり判定も兼ねる。
func (c *RoyalCotillion) GetHint() *RoyalCotillionHint {
	if c.phase != RoyalCotillionPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	// 空き枠は埋めておくほど選択肢が増える。山札から直接埋めれば捨て札を節約できる。
	for slot := range RoyalCotillionTableauCnt {
		if c.tableau[slot] != nil {
			continue
		}
		if len(c.stock) > 0 {
			return &RoyalCotillionHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: slot}
		}
		if c.wasteTop() != nil {
			return &RoyalCotillionHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: slot}
		}
	}
	if len(c.stock) > 0 {
		return &RoyalCotillionHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブロー枠を空ける手を優先する。枠が空けば山札・捨て札の出口が増えるので、
// 同じ 1 点でもリザーブより盤面が動く。
func (c *RoyalCotillion) foundationHint() *RoyalCotillionHint {
	if c.phase != RoyalCotillionPhasePlaying {
		return nil
	}
	for slot := range RoyalCotillionTableauCnt {
		card := c.tableau[slot]
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &RoyalCotillionHint{FromZone: "tableau", FromIdx: slot, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for pile := range RoyalCotillionReserveCnt {
		card := c.reserveTop(pile)
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &RoyalCotillionHint{FromZone: "reserve", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &RoyalCotillionHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *RoyalCotillion) AutoComplete() error {
	if c.phase != RoyalCotillionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := c.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch h.FromZone {
		case "waste":
			err = c.MoveWasteToFoundation()
		case "reserve":
			err = c.MoveReserveToFoundation(h.FromIdx)
		default:
			err = c.MoveTableauToFoundation(h.FromIdx)
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
func (c *RoyalCotillion) Undo() error {
	if len(c.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.tableau = snap.tableau
	c.reserve = snap.reserve
	c.foundation = snap.foundation
	c.stock = snap.stock
	c.waste = snap.waste
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (c *RoyalCotillion) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *RoyalCotillion) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *RoyalCotillion) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *royalCotillionSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *RoyalCotillion) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *RoyalCotillion) GetPhase() RoyalCotillionPhase { return c.phase }

// GetMoveCount 手数取得
func (c *RoyalCotillion) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *RoyalCotillion) GetStockCount() int { return len(c.stock) }

// GetWaste 捨て札を取得
func (c *RoyalCotillion) GetWaste() []*Card { return c.waste }

// GetTableau タブロー枠を取得
func (c *RoyalCotillion) GetTableau() [RoyalCotillionTableauCnt]*Card { return c.tableau }

// GetReserve リザーブを取得
func (c *RoyalCotillion) GetReserve() [RoyalCotillionReserveCnt][]*Card { return c.reserve }

// GetFoundation 基礎札を取得
func (c *RoyalCotillion) GetFoundation() [RoyalCotillionFoundationCnt][]*Card { return c.foundation }

// IsOddFoundation その基礎札が A 始まりかを返す（false なら 2 始まり）
func (c *RoyalCotillion) IsOddFoundation(fIdx int) bool {
	if fIdx < 0 || fIdx >= RoyalCotillionFoundationCnt {
		return false
	}
	return royalCotillionIsOdd(fIdx)
}

// GetGameEndFlag ゲーム終了フラグ
func (c *RoyalCotillion) GetGameEndFlag() bool { return c.phase != RoyalCotillionPhasePlaying }

// IsStalemate 手詰まりか
func (c *RoyalCotillion) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *RoyalCotillion) requirePlaying() error {
	if c.phase != RoyalCotillionPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validRoyalCotillionSlot タブロー枠のインデックスを検証する
func validRoyalCotillionSlot(slot int) error {
	if slot < 0 || slot >= RoyalCotillionTableauCnt {
		return fmt.Errorf("invalid slot: %d", slot)
	}
	return nil
}

// validRoyalCotillionReserve リザーブ山のインデックスを検証する
func validRoyalCotillionReserve(pile int) error {
	if pile < 0 || pile >= RoyalCotillionReserveCnt {
		return fmt.Errorf("invalid reserve: %d", pile)
	}
	return nil
}

// reserveTop リザーブの一番上（空なら nil）
func (c *RoyalCotillion) reserveTop(pile int) *Card {
	return discardTop(c.reserve[pile])
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *RoyalCotillion) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *RoyalCotillion) popWaste() {
	c.waste = dropLast(c.waste)
}

// royalCotillionIsOdd その基礎札が A 始まりか（false なら 2 始まり）
func royalCotillionIsOdd(fIdx int) bool { return fIdx < RoyalCotillionOddCnt }

// royalCotillionNthValue 基礎札の n 枚目（0 起点）に必要な値を返す。
//
// 2 つ飛ばしで**折り返す**のがこのゲームの核心。A 始まりなら
// A,3,5,7,9,J,K,2,4,6,8,10,Q、2 始まりなら 2,4,6,8,10,Q,A,3,5,7,9,J,K。
// 13 で法をとるので、K(13) の次が 2、Q(12) の次が A になる。
func royalCotillionNthValue(base, n int) int {
	return ((base - 1 + 2*n) % CardValueMax) + 1
}

// canPlaceOnFoundation 基礎札に置けるか（同スートで、次に必要な値ちょうど）
func (c *RoyalCotillion) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= RoyalCotillionFoundationCnt {
		return false
	}
	if royalCotillionSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	filled := len(c.foundation[fIdx])
	if filled >= RoyalCotillionFoundationTarget {
		return false
	}
	base := 1
	if !royalCotillionIsOdd(fIdx) {
		base = 2
	}
	return card.GetValue() == royalCotillionNthValue(base, filled)
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
//
// 同スート同値の札は 2 組ぶんでちょうど 2 枚あり、そのスートの 2 本
// （A 始まりと 2 始まり）はどちらも全 13 値を 1 度ずつ通る。よって 2 枚は
// 必ず別々の本に収まり、最初に見つかった本を使ってよい。
func (c *RoyalCotillion) findFoundation(card *Card) int {
	for i := range RoyalCotillionFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *RoyalCotillion) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚積まれたか
func (c *RoyalCotillion) checkGameClear() {
	for i := range RoyalCotillionFoundationCnt {
		if len(c.foundation[i]) != RoyalCotillionFoundationTarget {
			return
		}
	}
	c.phase = RoyalCotillionPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint はすべての手を見るので、「ヒントが無い」と「手詰まり」は同じ条件になる。
func (c *RoyalCotillion) checkStalemate() {
	if c.phase != RoyalCotillionPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *RoyalCotillion) takeSnapshot() {
	snap := &royalCotillionSnapshot{
		tableau:     c.tableau,
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range RoyalCotillionReserveCnt {
		snap.reserve[i] = append([]*Card(nil), c.reserve[i]...)
	}
	for i := range RoyalCotillionFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = append(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *RoyalCotillion) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// royalCotillionSnapshotJSON is the wire format for a single undo snapshot.
// royalCotillionSnapshot uses unexported fields, so marshalling it directly
// would emit `[{},{}]` -- the undo depth would survive but every snapshot would
// be blank, and Undo would wipe the board instead of rewinding it (#4478).
type royalCotillionSnapshotJSON struct {
	Tableau     [RoyalCotillionTableauCnt]*Card      `json:"tb"`
	Reserve     [RoyalCotillionReserveCnt][]*Card    `json:"rs"`
	Foundation  [RoyalCotillionFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                              `json:"st"`
	Waste       []*Card                              `json:"ws"`
	Phase       RoyalCotillionPhase                  `json:"ps"`
	MoveCount   int                                  `json:"mc"`
	IsStalemate bool                                 `json:"sl"`
}

// MarshalJSON implements json.Marshaler for royalCotillionSnapshot.
func (s *royalCotillionSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(royalCotillionSnapshotJSON{
		Tableau:     s.tableau,
		Reserve:     s.reserve,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for royalCotillionSnapshot.
func (s *royalCotillionSnapshot) UnmarshalJSON(data []byte) error {
	var j royalCotillionSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > royalCotillionMaxSliceLen || len(j.Waste) > royalCotillionMaxSliceLen {
		return errors.New("royalcotillion: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Reserve {
		if len(pile) > royalCotillionMaxSliceLen {
			return errors.New("royalcotillion: snapshot reserve exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > royalCotillionMaxSliceLen {
			return errors.New("royalcotillion: snapshot foundation exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.reserve = j.Reserve
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// royalCotillionJSON is the JSON wire format for RoyalCotillion.
type royalCotillionJSON struct {
	TrumpCards  *TrumpCards                          `json:"tc"`
	Tableau     [RoyalCotillionTableauCnt]*Card      `json:"tb"`
	Reserve     [RoyalCotillionReserveCnt][]*Card    `json:"rs"`
	Foundation  [RoyalCotillionFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                              `json:"st"`
	Waste       []*Card                              `json:"ws"`
	Phase       RoyalCotillionPhase                  `json:"ps"`
	MoveCount   int                                  `json:"mc"`
	ActionLog   []*ActionLogEntry                    `json:"al"`
	IsStalemate bool                                 `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*royalCotillionSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *RoyalCotillion) MarshalJSON() ([]byte, error) {
	return json.Marshal(&royalCotillionJSON{
		TrumpCards:  c.trumpCards,
		Tableau:     c.tableau,
		Reserve:     c.reserve,
		Foundation:  c.foundation,
		Stock:       c.stock,
		Waste:       c.waste,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *RoyalCotillion) UnmarshalJSON(data []byte) error {
	var j royalCotillionJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < RoyalCotillionPhasePlaying || j.Phase > RoyalCotillionPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > RoyalCotillionTotalCards || len(j.Waste) > RoyalCotillionTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > royalCotillionMaxSliceLen || len(j.History) > royalCotillionMaxSliceLen {
		return errors.New("royalcotillion: input array exceeds maximum allowed size")
	}
	for i := range RoyalCotillionFoundationCnt {
		if len(j.Foundation[i]) > RoyalCotillionFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range RoyalCotillionReserveCnt {
		if len(j.Reserve[i]) > RoyalCotillionTotalCards {
			return fmt.Errorf("reserve %d holds %d cards", i, len(j.Reserve[i]))
		}
	}
	if j.TrumpCards != nil {
		c.trumpCards = j.TrumpCards
	}
	c.tableau = j.Tableau
	c.reserve = j.Reserve
	c.foundation = j.Foundation
	c.stock = j.Stock
	c.waste = j.Waste
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	c.isStalemate = j.IsStalemate
	c.history = j.History
	return nil
}
