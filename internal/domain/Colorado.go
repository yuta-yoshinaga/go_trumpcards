//go:build !js || !wasm || classic

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ColoradoPhase コロラドのゲームフェーズ
type ColoradoPhase int

// Coloradoのフェーズ定数
const (
	// ColoradoPhasePlaying プレイ中
	ColoradoPhasePlaying ColoradoPhase = iota
	// ColoradoPhaseGameClear ゲームクリア
	ColoradoPhaseGameClear
	// ColoradoPhaseGameOver ゲームオーバー
	ColoradoPhaseGameOver
)

// ColoradoTableauCnt タブローの山数。10 山 2 段に並ぶ。
const ColoradoTableauCnt = 20

// ColoradoFoundationCnt 基礎札の数。前半 4 つが A からの昇順、後半 4 つが K からの降順。
const ColoradoFoundationCnt = 8

// ColoradoFoundationTarget 基礎札 1 つあたりの完成枚数
const ColoradoFoundationTarget = CardValueMax

// ColoradoTotalCards 使用する総枚数（52 枚 2 組）
const ColoradoTotalCards = CardCnt * 2

// ColoradoAscendingCnt 昇順の基礎札の数（インデックス 0 以上この値未満が昇順）
const ColoradoAscendingCnt = ColoradoFoundationCnt / 2

// coloradoMaxSliceLen caps slice sizes during deserialisation.
const coloradoMaxSliceLen = 1000

// coloradoSuitOrder 基礎札インデックスとスートの対応。前半 4 つが昇順、後半 4 つが
// 降順で、どちらも同じスート順に並ぶ。固定しておくと配り直しても UI の位置が動かない。
var coloradoSuitOrder = [ColoradoFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// ColoradoHint コロラドのヒント
type ColoradoHint struct {
	// FromZone 移動元 "tableau" / "waste" / "stock"
	FromZone string
	// FromIdx 移動元のタブロー山（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// Colorado コロラド ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**20 山のタブロー**に 1 枚ずつ配り、
// 残り 84 枚が山札になる。基礎札は 8 つで、**4 つが A から K への昇順、残り 4 つが
// K から A への降順**。8×13 = 104 枚すべてを積み切ればクリア。
//
// 山札は 1 枚ずつ捨て札へめくり、**めくり直しは無い（1 巡のみ）**。捨て札の一番上は
// 基礎札へ送るか、**スートも数字も問わず好きなタブロー山へ置ける**。タブローの札は
// 基礎札へしか送れず、タブロー同士の移動は無い。空いた山は山札から直接埋められる。
//
// このゲームの緊張はすべて「捨て札をどの山へ置くか」に集約される。置いた札は下の
// 札を埋めてしまうので、埋めても痛くない山を選ぶのが上手さになる。
//
// issue #5277 の仕様案とは 3 点異なり、いずれも実際の規則に合わせた:
//   - **基礎札は 16 本ではなく 8 本**。16 本だと 16×13 = 208 枚となり、104 枚の
//     デッキでは半分も埋まらない。昇順 4 本 + 降順 4 本の計 8 本が 104 枚と一致する
//   - **捨て札は「空き山」だけでなく任意のタブロー山へ置ける**。空き山限定にすると
//     置き場所の選択という本作の核心が消え、ただ山札をめくるだけのゲームになる
//   - **捨て札に「積むだけ」の専用山は無い**。捨て札は 1 山で、一番上だけが使える
//     通常のウェイストパイルである
type Colorado struct {
	trumpCards *TrumpCards
	tableau    [ColoradoTableauCnt][]*Card
	foundation [ColoradoFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      ColoradoPhase
	moveCount  int
	actionLogBase
	history     []*coloradoSnapshot
	isStalemate bool
}

// coloradoSnapshot アンドゥ用スナップショット
type coloradoSnapshot struct {
	tableau     [ColoradoTableauCnt][]*Card
	foundation  [ColoradoFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	phase       ColoradoPhase
	moveCount   int
	isStalemate bool
}

// NewColorado コンストラクタ
func NewColorado(trumpCards *TrumpCards) *Colorado {
	return &Colorado{trumpCards: trumpCards}
}

// NewDefaultColorado returns Colorado with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultColorado() *Colorado {
	return NewColorado(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *Colorado) Reset() {
	c.trumpCards.Shuffle()
	c.phase = ColoradoPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil

	for i := range ColoradoFoundationCnt {
		c.foundation[i] = nil
	}
	// タブローは 20 山に 1 枚ずつ。残りはすべて山札。
	for i := range ColoradoTableauCnt {
		c.tableau[i] = nil
		if card := c.trumpCards.DrawCard(); card != nil {
			c.tableau[i] = append(c.tableau[i], card)
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
func (c *Colorado) Draw() error {
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

// MoveTableauToFoundation タブローの一番上を基礎札へ送る
func (c *Colorado) MoveTableauToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validColoradoPile(pile); err != nil {
		return err
	}
	card := c.tableauTop(pile)
	if card == nil {
		return fmt.Errorf("pile %d is empty", pile)
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	c.takeSnapshot()
	c.tableau[pile] = dropLast(c.tableau[pile])
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("タブロー山%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *Colorado) MoveWasteToFoundation() error {
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

// MoveWasteToTableau 捨て札の一番上をタブローへ置く。
//
// **スートも数字も問わず、どの山にも置ける**のがこのゲームの要。置いた札は下の札を
// 埋めてしまうので、どこへ置くかが唯一にして最大の判断になる。
func (c *Colorado) MoveWasteToTableau(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validColoradoPile(pile); err != nil {
		return err
	}
	card := c.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	c.takeSnapshot()
	c.popWaste()
	c.tableau[pile] = append(c.tableau[pile], card)
	c.afterMove("move", fmt.Sprintf("捨て札→タブロー山%d", pile), card)
	return nil
}

// MoveStockToTableau 山札の一番上で空き山を直接埋める。
//
// 規則上、タブローの空きは山札から埋める。捨て札を経由させると山札の 1 巡が
// 1 枚分早く減るので、山札から直接置く手を別に用意している。
func (c *Colorado) MoveStockToTableau(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validColoradoPile(pile); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return errors.New("stock is empty")
	}
	if len(c.tableau[pile]) != 0 {
		return errors.New("the stock may only fill an empty pile")
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.tableau[pile] = append(c.tableau[pile], card)
	c.afterMove("move", fmt.Sprintf("山札→タブロー山%d", pile), card)
	return nil
}

// GiveUp ギブアップ
func (c *Colorado) GiveUp() {
	if c.phase == ColoradoPhasePlaying {
		c.phase = ColoradoPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → 空き山埋め → 山めくり → 捨て札の置き場の順。
// 手詰まり判定も兼ねる。
//
// 捨て札はどの山にも置けるので「捨て札→タブロー」は常に成立してしまう。それを
// 先に返すとヒントが毎回同じ無意味な手になるため、最後に回している。
func (c *Colorado) GetHint() *ColoradoHint {
	if c.phase != ColoradoPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	// 空き山は山札から埋めるのが定石。1 巡しか無い山札を捨て札へ流さずに済む。
	if len(c.stock) > 0 {
		for pile := range ColoradoTableauCnt {
			if len(c.tableau[pile]) == 0 {
				return &ColoradoHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: pile}
			}
		}
		return &ColoradoHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	if c.wasteTop() != nil {
		return &ColoradoHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: c.bestBuryPile()}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (c *Colorado) foundationHint() *ColoradoHint {
	if c.phase != ColoradoPhasePlaying {
		return nil
	}
	for pile := range ColoradoTableauCnt {
		card := c.tableauTop(pile)
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &ColoradoHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &ColoradoHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// bestBuryPile 捨て札を置くのに一番損の小さい山を返す。
//
// 空き山があればそこ（何も埋めない）。無ければ、一番上の札が基礎札に必要になるまで
// 最も遠い山を選ぶ。同点なら添字の小さい方。
func (c *Colorado) bestBuryPile() int {
	best, bestCost := 0, -1
	for pile := range ColoradoTableauCnt {
		top := c.tableauTop(pile)
		if top == nil {
			return pile
		}
		if cost := c.buryCost(top); cost > bestCost {
			best, bestCost = pile, cost
		}
	}
	return best
}

// buryCost その札が基礎札に必要になるまでに、あと何枚積まれるのを待つかの最小値。
// 大きいほど今埋めても痛くない。どの基礎札にも入れない札は最大値を返す。
func (c *Colorado) buryCost(card *Card) int {
	if card == nil {
		return ColoradoFoundationTarget + 1
	}
	best := ColoradoFoundationTarget + 1
	for i := range ColoradoFoundationCnt {
		if coloradoSuitOrder[i] != card.GetDesign() {
			continue
		}
		need := c.foundationNeed(i, card.GetValue())
		// 負なら、その基礎札は既にこの数字を通り過ぎている（もう入らない）。
		if need >= 0 && need < best {
			best = need
		}
	}
	return best
}

// foundationNeed 基礎札 fIdx が値 v を受け入れるまでに、あと何枚必要かを返す。
// 0 なら今すぐ置ける。負なら通り過ぎていて二度と置けない。
func (c *Colorado) foundationNeed(fIdx, v int) int {
	filled := len(c.foundation[fIdx])
	if coloradoIsAscending(fIdx) {
		// 昇順は A(1) から。filled 枚積んだ次に必要なのは filled+1。
		return v - (filled + 1)
	}
	// 降順は K(CardValueMax) から。filled 枚積んだ次に必要なのは CardValueMax-filled。
	return (CardValueMax - filled) - v
}

// coloradoIsAscending その基礎札が A からの昇順か（false なら K からの降順）
func coloradoIsAscending(fIdx int) bool { return fIdx < ColoradoAscendingCnt }

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *Colorado) AutoComplete() error {
	if c.phase != ColoradoPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := c.foundationHint()
		if h == nil {
			break
		}
		var err error
		if h.FromZone == "waste" {
			err = c.MoveWasteToFoundation()
		} else {
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
func (c *Colorado) Undo() error {
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
	return nil
}

// CanUndo アンドゥ可能か
func (c *Colorado) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *Colorado) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *Colorado) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *coloradoSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *Colorado) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *Colorado) GetPhase() ColoradoPhase { return c.phase }

// GetMoveCount 手数取得
func (c *Colorado) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *Colorado) GetStockCount() int { return len(c.stock) }

// GetWaste 捨て札を取得
func (c *Colorado) GetWaste() []*Card { return c.waste }

// GetTableau タブローを取得
func (c *Colorado) GetTableau() [ColoradoTableauCnt][]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *Colorado) GetFoundation() [ColoradoFoundationCnt][]*Card { return c.foundation }

// IsAscendingFoundation その基礎札が A からの昇順かを返す（表示の向き分け用）
func (c *Colorado) IsAscendingFoundation(fIdx int) bool {
	if fIdx < 0 || fIdx >= ColoradoFoundationCnt {
		return false
	}
	return coloradoIsAscending(fIdx)
}

// GetGameEndFlag ゲーム終了フラグ
func (c *Colorado) GetGameEndFlag() bool { return c.phase != ColoradoPhasePlaying }

// IsStalemate 手詰まりか
func (c *Colorado) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *Colorado) requirePlaying() error {
	if c.phase != ColoradoPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validColoradoPile タブロー山のインデックスを検証する
func validColoradoPile(pile int) error {
	if pile < 0 || pile >= ColoradoTableauCnt {
		return fmt.Errorf("invalid pile: %d", pile)
	}
	return nil
}

// tableauTop 山の一番上（空なら nil）
func (c *Colorado) tableauTop(pile int) *Card {
	return discardTop(c.tableau[pile])
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *Colorado) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *Colorado) popWaste() {
	c.waste = dropLast(c.waste)
}

// canPlaceOnFoundation 基礎札に置けるか。昇順は空なら A、以降は 1 つ上。
// 降順は空なら K、以降は 1 つ下。どちらも同スートに限る。
func (c *Colorado) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= ColoradoFoundationCnt {
		return false
	}
	if coloradoSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	if len(c.foundation[fIdx]) >= ColoradoFoundationTarget {
		return false
	}
	return c.foundationNeed(fIdx, card.GetValue()) == 0
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
//
// 1 枚の札が昇順と降順の両方に置けることはある（例: 昇順が Q まで積まれていて
// 降順が空のときの K）。どちらへ置いても詰まない — 同スート同値の札は 2 組ぶんで
// ちょうど 2 枚あり、昇順と降順がそれぞれ 1 枚ずつ必要とするので、先に埋めた方に
// 1 枚目、もう一方に 2 枚目が回る。よって最初に見つかったものを使ってよい。
func (c *Colorado) findFoundation(card *Card) int {
	for i := range ColoradoFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *Colorado) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚積まれたか
func (c *Colorado) checkGameClear() {
	for i := range ColoradoFoundationCnt {
		if len(c.foundation[i]) != ColoradoFoundationTarget {
			return
		}
	}
	c.phase = ColoradoPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
//
// 捨て札はどの山へも置けるので、手詰まりは「山札も捨て札も尽き、タブローの
// どの一番上も基礎札へ送れない」ときにだけ起きる。捨て札は山札からしか増えず
// 置くたびに減るので、この状態には必ず到達する。
func (c *Colorado) checkStalemate() {
	if c.phase != ColoradoPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *Colorado) takeSnapshot() {
	snap := &coloradoSnapshot{
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range ColoradoTableauCnt {
		snap.tableau[i] = append([]*Card(nil), c.tableau[i]...)
	}
	for i := range ColoradoFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = appendSnapshot(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *Colorado) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// coloradoSnapshotJSON is the wire format for a single undo snapshot.
// coloradoSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type coloradoSnapshotJSON struct {
	Tableau     [ColoradoTableauCnt][]*Card    `json:"tb"`
	Foundation  [ColoradoFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       ColoradoPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for coloradoSnapshot.
func (s *coloradoSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(coloradoSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for coloradoSnapshot.
func (s *coloradoSnapshot) UnmarshalJSON(data []byte) error {
	var j coloradoSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > coloradoMaxSliceLen || len(j.Waste) > coloradoMaxSliceLen {
		return errors.New("colorado: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > coloradoMaxSliceLen {
			return errors.New("colorado: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > coloradoMaxSliceLen {
			return errors.New("colorado: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// coloradoJSON is the JSON wire format for Colorado.
type coloradoJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [ColoradoTableauCnt][]*Card    `json:"tb"`
	Foundation  [ColoradoFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       ColoradoPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*coloradoSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *Colorado) MarshalJSON() ([]byte, error) {
	return json.Marshal(&coloradoJSON{
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
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *Colorado) UnmarshalJSON(data []byte) error {
	var j coloradoJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < ColoradoPhasePlaying || j.Phase > ColoradoPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > ColoradoTotalCards || len(j.Waste) > ColoradoTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > coloradoMaxSliceLen || len(j.History) > coloradoMaxSliceLen {
		return errors.New("colorado: input array exceeds maximum allowed size")
	}
	for i := range ColoradoFoundationCnt {
		if len(j.Foundation[i]) > ColoradoFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range ColoradoTableauCnt {
		if len(j.Tableau[i]) > ColoradoTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
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
	return nil
}
