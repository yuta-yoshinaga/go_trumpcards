//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MatrimonyPhase マトリモニーのゲームフェーズ
type MatrimonyPhase int

// Matrimonyのフェーズ定数
const (
	// MatrimonyPhasePlaying プレイ中
	MatrimonyPhasePlaying MatrimonyPhase = iota
	// MatrimonyPhaseGameClear ゲームクリア
	MatrimonyPhaseGameClear
	// MatrimonyPhaseGameOver ゲームオーバー
	MatrimonyPhaseGameOver
)

// MatrimonyTableauCnt タブローの枠数。各枠にカードは 1 枚しか置けない。
const MatrimonyTableauCnt = 16

// MatrimonyFoundationCnt 基礎札の数。各起点のカードが2組ずつある。
const MatrimonyFoundationCnt = 4

// MatrimonyFoundationTarget 基礎札 1 つあたりの完成枚数。
//
// 2 つ飛ばしでも**折り返す**ので、1 本で 13 枚すべてを通る。
// A→3→…→K→2→4→…→Q の順で 13 枚。
const MatrimonyFoundationTarget = CardValueMax

// MatrimonyTotalCards 使用する総枚数（52 枚 2 組）
const MatrimonyTotalCards = CardCnt * 2

// MatrimonyMaxRedeals is the maximum number of times the waste may be collected.
const MatrimonyMaxRedeals = 3

const (
	matrimonyAce   = 1
	matrimonyJack  = 11
	matrimonyQueen = 12
	matrimonyKing  = 13
)

// matrimonyMaxSliceLen caps slice sizes during deserialisation.
const matrimonyMaxSliceLen = 1000

// matrimonySuitOrder 基礎札インデックスとスートの対応。前半 4 つが A 始まり、
// 後半 4 つが 2 始まりで、どちらも同じスート順に並ぶ。
var matrimonyFoundationStart = [MatrimonyFoundationCnt]struct {
	design int
	value  int
	desc   bool
}{
	{CardDesignSpade, matrimonyQueen, true}, {CardDesignSpade, matrimonyQueen, true},
	{CardDesignDiamond, matrimonyJack, false}, {CardDesignDiamond, matrimonyJack, false},
}

// MatrimonyHint マトリモニーのヒント
type MatrimonyHint struct {
	// FromZone is "tableau", "waste", or "stock".
	FromZone string
	// FromIdx 移動元の枠・山（捨て札と山札は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の場所を指さない）
	ToIdx int
}

// Matrimony マトリモニー ゲームクラス。
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
type Matrimony struct {
	trumpCards *TrumpCards
	// tableau は 1 枠 1 枚。空き枠は nil。
	tableau    [MatrimonyTableauCnt]*Card
	foundation [MatrimonyFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      MatrimonyPhase
	moveCount  int
	actionLogBase
	history     []*matrimonySnapshot
	isStalemate bool
	redealCount int
}

// matrimonySnapshot アンドゥ用スナップショット
type matrimonySnapshot struct {
	tableau     [MatrimonyTableauCnt]*Card
	foundation  [MatrimonyFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	phase       MatrimonyPhase
	moveCount   int
	isStalemate bool
	redealCount int
}

// NewMatrimony コンストラクタ
func NewMatrimony(trumpCards *TrumpCards) *Matrimony {
	return &Matrimony{trumpCards: trumpCards}
}

// NewDefaultMatrimony returns Matrimony with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultMatrimony() *Matrimony {
	return NewMatrimony(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *Matrimony) Reset() {
	c.trumpCards.Shuffle()
	c.phase = MatrimonyPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil

	for i := range MatrimonyFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range MatrimonyTableauCnt {
		c.tableau[i] = c.trumpCards.DrawCard()
	}
	for {
		card := c.trumpCards.DrawCard()
		if card == nil {
			break
		}
		c.stock = append(c.stock, card)
	}
	c.redealCount = 0

	c.checkStalemate()
}

// Draw draws one card, collecting the waste for a new pass when necessary.
func (c *Matrimony) Draw() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		if len(c.waste) == 0 || c.redealCount >= MatrimonyMaxRedeals {
			return errors.New("stock is empty and there is no redeal")
		}
		c.takeSnapshot()
		for i := len(c.waste) - 1; i >= 0; i-- {
			c.stock = append(c.stock, c.waste[i])
		}
		c.waste = nil
		c.redealCount++
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.waste = append(c.waste, card)
	c.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveTableauToFoundation タブロー枠の札を基礎札へ送る
func (c *Matrimony) MoveTableauToFoundation(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validMatrimonySlot(slot); err != nil {
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

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *Matrimony) MoveWasteToFoundation() error {
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
func (c *Matrimony) MoveWasteToTableau(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validMatrimonySlot(slot); err != nil {
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
func (c *Matrimony) MoveStockToTableau(slot int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validMatrimonySlot(slot); err != nil {
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
func (c *Matrimony) GiveUp() {
	if c.phase == MatrimonyPhasePlaying {
		c.phase = MatrimonyPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → 空き枠埋め → 山めくり の順。手詰まり判定も兼ねる。
func (c *Matrimony) GetHint() *MatrimonyHint {
	if c.phase != MatrimonyPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	// 空き枠は埋めておくほど選択肢が増える。山札から直接埋めれば捨て札を節約できる。
	for slot := range MatrimonyTableauCnt {
		if c.tableau[slot] != nil {
			continue
		}
		if len(c.stock) > 0 {
			return &MatrimonyHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: slot}
		}
		if c.wasteTop() != nil {
			return &MatrimonyHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: slot}
		}
	}
	if len(c.stock) > 0 {
		return &MatrimonyHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブロー枠を空ける手を優先する。枠が空けば山札・捨て札の出口が増えるので、
// 同じ 1 点でもリザーブより盤面が動く。
func (c *Matrimony) foundationHint() *MatrimonyHint {
	if c.phase != MatrimonyPhasePlaying {
		return nil
	}
	for slot := range MatrimonyTableauCnt {
		card := c.tableau[slot]
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &MatrimonyHint{FromZone: "tableau", FromIdx: slot, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &MatrimonyHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *Matrimony) AutoComplete() error {
	if c.phase != MatrimonyPhasePlaying {
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
func (c *Matrimony) Undo() error {
	if len(c.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.tableau = snap.tableau
	c.foundation = snap.foundation
	c.stock = snap.stock
	c.waste = snap.waste
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
	c.redealCount = snap.redealCount
	return nil
}

// CanUndo アンドゥ可能か
func (c *Matrimony) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *Matrimony) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *Matrimony) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *matrimonySnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *Matrimony) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *Matrimony) GetPhase() MatrimonyPhase { return c.phase }

// GetMoveCount 手数取得
func (c *Matrimony) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *Matrimony) GetStockCount() int { return len(c.stock) }

// GetRedealCount returns how many redeals have been used.
func (c *Matrimony) GetRedealCount() int { return c.redealCount }

// GetWaste 捨て札を取得
func (c *Matrimony) GetWaste() []*Card { return c.waste }

// GetTableau タブロー枠を取得
func (c *Matrimony) GetTableau() [MatrimonyTableauCnt]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *Matrimony) GetFoundation() [MatrimonyFoundationCnt][]*Card { return c.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (c *Matrimony) GetGameEndFlag() bool { return c.phase != MatrimonyPhasePlaying }

// IsStalemate 手詰まりか
func (c *Matrimony) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *Matrimony) requirePlaying() error {
	if c.phase != MatrimonyPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validMatrimonySlot タブロー枠のインデックスを検証する
func validMatrimonySlot(slot int) error {
	if slot < 0 || slot >= MatrimonyTableauCnt {
		return fmt.Errorf("invalid slot: %d", slot)
	}
	return nil
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *Matrimony) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *Matrimony) popWaste() {
	c.waste = dropLast(c.waste)
}

// matrimonyPreviousRank returns the rank before value, wrapping A to K.
func matrimonyPreviousRank(value int) int {
	if value == matrimonyAce {
		return matrimonyKing
	}
	return value - 1
}

// matrimonyNextRank returns the rank after value, wrapping K to A.
func matrimonyNextRank(value int) int {
	if value == matrimonyKing {
		return matrimonyAce
	}
	return value + 1
}

// canPlaceOnFoundation 基礎札に置けるか（同スートで、次に必要な値ちょうど）
func (c *Matrimony) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= MatrimonyFoundationCnt {
		return false
	}
	start := matrimonyFoundationStart[fIdx]
	if start.design != card.GetDesign() {
		return false
	}
	filled := len(c.foundation[fIdx])
	if filled >= MatrimonyFoundationTarget {
		return false
	}
	value := start.value
	for range filled {
		if start.desc {
			value = matrimonyPreviousRank(value)
		} else {
			value = matrimonyNextRank(value)
		}
	}
	return card.GetValue() == value
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
//
// 同スート同値の札は 2 組ぶんでちょうど 2 枚あり、そのスートの 2 本
// （A 始まりと 2 始まり）はどちらも全 13 値を 1 度ずつ通る。よって 2 枚は
// 必ず別々の本に収まり、最初に見つかった本を使ってよい。
func (c *Matrimony) findFoundation(card *Card) int {
	for i := range MatrimonyFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *Matrimony) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚積まれたか
func (c *Matrimony) checkGameClear() {
	for i := range MatrimonyFoundationCnt {
		if len(c.foundation[i]) != MatrimonyFoundationTarget {
			return
		}
	}
	c.phase = MatrimonyPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint はすべての手を見るので、「ヒントが無い」と「手詰まり」は同じ条件になる。
func (c *Matrimony) checkStalemate() {
	if c.phase != MatrimonyPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *Matrimony) takeSnapshot() {
	snap := &matrimonySnapshot{
		tableau:     c.tableau,
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	snap.redealCount = c.redealCount
	for i := range MatrimonyFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = appendSnapshot(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *Matrimony) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// matrimonySnapshotJSON is the wire format for a single undo snapshot.
// matrimonySnapshot uses unexported fields, so marshalling it directly
// would emit `[{},{}]` -- the undo depth would survive but every snapshot would
// be blank, and Undo would wipe the board instead of rewinding it (#4478).
type matrimonySnapshotJSON struct {
	Tableau     [MatrimonyTableauCnt]*Card      `json:"tb"`
	Foundation  [MatrimonyFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                         `json:"st"`
	Waste       []*Card                         `json:"ws"`
	Phase       MatrimonyPhase                  `json:"ps"`
	MoveCount   int                             `json:"mc"`
	IsStalemate bool                            `json:"sl"`
	RedealCount int                             `json:"rd"`
}

// MarshalJSON implements json.Marshaler for matrimonySnapshot.
func (s *matrimonySnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(matrimonySnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
		RedealCount: s.redealCount,
	})
}

// UnmarshalJSON implements json.Unmarshaler for matrimonySnapshot.
func (s *matrimonySnapshot) UnmarshalJSON(data []byte) error {
	var j matrimonySnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > matrimonyMaxSliceLen || len(j.Waste) > matrimonyMaxSliceLen {
		return errors.New("matrimony: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundation {
		if len(pile) > matrimonyMaxSliceLen {
			return errors.New("matrimony: snapshot foundation exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	s.redealCount = j.RedealCount
	return nil
}

// matrimonyJSON is the JSON wire format for Matrimony.
type matrimonyJSON struct {
	TrumpCards  *TrumpCards                     `json:"tc"`
	Tableau     [MatrimonyTableauCnt]*Card      `json:"tb"`
	Foundation  [MatrimonyFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                         `json:"st"`
	Waste       []*Card                         `json:"ws"`
	Phase       MatrimonyPhase                  `json:"ps"`
	MoveCount   int                             `json:"mc"`
	ActionLog   []*ActionLogEntry               `json:"al"`
	IsStalemate bool                            `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History     []*matrimonySnapshot `json:"hi,omitempty"`
	RedealCount int                  `json:"rd"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *Matrimony) MarshalJSON() ([]byte, error) {
	return json.Marshal(&matrimonyJSON{
		TrumpCards:  c.trumpCards,
		Tableau:     c.tableau,
		Foundation:  c.foundation,
		Stock:       c.stock,
		Waste:       c.waste,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
		RedealCount: c.redealCount,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *Matrimony) UnmarshalJSON(data []byte) error {
	var j matrimonyJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < MatrimonyPhasePlaying || j.Phase > MatrimonyPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > MatrimonyTotalCards || len(j.Waste) > MatrimonyTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > matrimonyMaxSliceLen || len(j.History) > matrimonyMaxSliceLen {
		return errors.New("matrimony: input array exceeds maximum allowed size")
	}
	for i := range MatrimonyFoundationCnt {
		if len(j.Foundation[i]) > MatrimonyFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	if j.TrumpCards != nil {
		c.trumpCards = j.TrumpCards
	}
	c.tableau = j.Tableau
	c.foundation = j.Foundation
	c.stock = j.Stock
	c.waste = j.Waste
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	c.isStalemate = j.IsStalemate
	c.history = j.History
	c.redealCount = j.RedealCount
	return nil
}
