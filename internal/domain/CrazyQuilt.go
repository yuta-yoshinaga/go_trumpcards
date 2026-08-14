//go:build !js || !wasm || solo

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// CrazyQuiltPhase クレイジーキルトのゲームフェーズ
type CrazyQuiltPhase int

// CrazyQuiltのフェーズ定数
const (
	// CrazyQuiltPhasePlaying プレイ中
	CrazyQuiltPhasePlaying CrazyQuiltPhase = iota
	// CrazyQuiltPhaseGameClear ゲームクリア
	CrazyQuiltPhaseGameClear
	// CrazyQuiltPhaseGameOver ゲームオーバー
	CrazyQuiltPhaseGameOver
)

// CrazyQuiltGridSize キルトの一辺（8×8 = 64 枚）
const CrazyQuiltGridSize = 8

// CrazyQuiltCells キルトのマス数
const CrazyQuiltCells = CrazyQuiltGridSize * CrazyQuiltGridSize

// CrazyQuiltFoundationCnt 基礎札の数。スートごとに A 昇順と K 降順の 2 本。
const CrazyQuiltFoundationCnt = 8

// CrazyQuiltAscendingCnt A 始まりの基礎札の数（インデックス 0 以上この値未満）
const CrazyQuiltAscendingCnt = CrazyQuiltFoundationCnt / 2

// CrazyQuiltFoundationTarget 基礎札 1 つあたりの完成枚数
const CrazyQuiltFoundationTarget = CardValueMax

// CrazyQuiltTotalCards 使用する総枚数（52 枚 2 組）
const CrazyQuiltTotalCards = CardCnt * 2

// CrazyQuiltRedealCnt 山札を組み直せる回数。捨て札を伏せて 1 度だけ引き直せる。
const CrazyQuiltRedealCnt = 1

// crazyQuiltMaxSliceLen caps slice sizes during deserialisation.
const crazyQuiltMaxSliceLen = 1000

// crazyQuiltSuitOrder 基礎札インデックスとスートの対応。前半 4 つが A からの昇順、
// 後半 4 つが K からの降順で、どちらも同じスート順に並ぶ。
var crazyQuiltSuitOrder = [CrazyQuiltFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// CrazyQuiltHint クレイジーキルトのヒント
type CrazyQuiltHint struct {
	// FromZone 移動元 "quilt" / "waste" / "stock"
	FromZone string
	// FromIdx キルトのマス番号（row*8+col）。それ以外は -1
	FromIdx int
	// ToZone 移動先 "foundation" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の場所を指さない）
	ToIdx int
}

// CrazyQuilt クレイジーキルト（インディアンカーペット）ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**各スートの A と K を 1 枚ずつ、
// 計 8 枚を先に抜いて基礎札に据える**。残り 96 枚のうち 64 枚を 8×8 のキルトに
// 並べ、残り 32 枚が山札になる。
//
// キルトは**縦置きと横置きを market 目状に交互**に敷く。取れるのは**短辺が
// 露出している札だけ** — 縦置きなら上か下、横置きなら左か右が空いている札。
// 1 枚取るたびに隣の短辺が開くので、どこから崩すかがパズルになる。
//
// 基礎札は 8 本。前半 4 本は A から K への昇順、後半 4 本は K から A への降順。
// 8×13 = 104 枚すべてを積み切ればクリア。
//
// 山札は 1 枚ずつ捨て札へめくり、**尽きたら捨て札を伏せて 1 度だけ組み直せる**
// （シャッフルはしない）。キルトの札は基礎札のほか、**捨て札の一番上と数字が
// 1 つ違いなら捨て札へも置ける**（スートは問わない）。これがキルトを崩す主要な
// 手段になる。
//
// issue #5274 の仕様案とは 4 点異なり、いずれも実際の規則に合わせた:
//   - **山札は 40 枚ではなく 32 枚。** 基礎札に据える 8 枚を先に抜くので
//     104 − 8 − 64 = 32 になる。issue はその 8 枚を数え落としている
//   - **可動判定は「隣が空いている」ではなく「短辺が露出している」。**
//     縦置きの札は上下、横置きの札は左右しか見ない。左右が空いた縦置きの札は
//     まだ動かせず、この向きの区別こそがキルトのパズル性そのもの
//   - **基礎札は空から始まらない。** 各スートの A と K を最初から据える
//   - **山札の組み直しと、キルト→捨て札の連番置きがある。** issue はどちらも
//     触れていないが、後者が無いとキルトはほとんど崩せない
type CrazyQuilt struct {
	trumpCards *TrumpCards
	// quilt はマス番号（row*8+col）順。取り除いたマスは nil。
	quilt       [CrazyQuiltCells]*Card
	foundation  [CrazyQuiltFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	redealsLeft int
	phase       CrazyQuiltPhase
	moveCount   int
	actionLogBase
	history     []*crazyQuiltSnapshot
	isStalemate bool
}

// crazyQuiltSnapshot アンドゥ用スナップショット
type crazyQuiltSnapshot struct {
	quilt       [CrazyQuiltCells]*Card
	foundation  [CrazyQuiltFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	redealsLeft int
	phase       CrazyQuiltPhase
	moveCount   int
	isStalemate bool
}

// NewCrazyQuilt コンストラクタ
func NewCrazyQuilt(trumpCards *TrumpCards) *CrazyQuilt {
	return &CrazyQuilt{trumpCards: trumpCards}
}

// NewDefaultCrazyQuilt returns CrazyQuilt with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCrazyQuilt() *CrazyQuilt {
	return NewCrazyQuilt(NewTrumpCardsWithDecks(2, 0))
}

// CrazyQuiltIsVertical そのマスが縦置きかを返す（false なら横置き）。
//
// 縦横を market 目状に交互に置くので、行と列の和の偶奇で決まる。
func CrazyQuiltIsVertical(idx int) bool {
	if idx < 0 || idx >= CrazyQuiltCells {
		return false
	}
	return (idx/CrazyQuiltGridSize+idx%CrazyQuiltGridSize)%2 == 0
}

// Reset ゲームリセット
func (c *CrazyQuilt) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CrazyQuiltPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil
	c.redealsLeft = CrazyQuiltRedealCnt

	// 各スートの A と K を 1 枚ずつ抜いて基礎札に据える。抜いた 8 枚は
	// キルトにも山札にも入らない。
	for i := range CrazyQuiltFoundationCnt {
		v := 1
		if !crazyQuiltIsAscending(i) {
			v = CardValueMax
		}
		c.foundation[i] = []*Card{NewCard(crazyQuiltSuitOrder[i], v, true)}
	}

	remaining := make([]*Card, 0, CrazyQuiltTotalCards)
	need := map[[2]int]int{}
	for i := range CrazyQuiltFoundationCnt {
		v := 1
		if !crazyQuiltIsAscending(i) {
			v = CardValueMax
		}
		need[[2]int{crazyQuiltSuitOrder[i], v}]++
	}
	for {
		card := c.trumpCards.DrawCard()
		if card == nil {
			break
		}
		key := [2]int{card.GetDesign(), card.GetValue()}
		if need[key] > 0 {
			// この 1 枚は基礎札に据えた札の実体なので、盤面には出さない。
			need[key]--
			continue
		}
		remaining = append(remaining, card)
	}

	for i := range CrazyQuiltCells {
		c.quilt[i] = nil
		if i < len(remaining) {
			c.quilt[i] = remaining[i]
		}
	}
	if len(remaining) > CrazyQuiltCells {
		c.stock = append(c.stock, remaining[CrazyQuiltCells:]...)
	}

	c.checkStalemate()
}

// IsAvailable そのマスの札が取れるか（短辺のどちらかが露出しているか）。
//
// 縦置きは上下、横置きは左右しか見ない。**長辺が空いていても取れない**のが
// このゲームの肝で、「隣が空いていれば取れる」ではない。
func (c *CrazyQuilt) IsAvailable(idx int) bool {
	if idx < 0 || idx >= CrazyQuiltCells || c.quilt[idx] == nil {
		return false
	}
	row, col := idx/CrazyQuiltGridSize, idx%CrazyQuiltGridSize
	if CrazyQuiltIsVertical(idx) {
		return c.cellEmpty(row-1, col) || c.cellEmpty(row+1, col)
	}
	return c.cellEmpty(row, col-1) || c.cellEmpty(row, col+1)
}

// cellEmpty そのマスが盤外か、札が取り除かれているか
func (c *CrazyQuilt) cellEmpty(row, col int) bool {
	if row < 0 || row >= CrazyQuiltGridSize || col < 0 || col >= CrazyQuiltGridSize {
		return true
	}
	return c.quilt[row*CrazyQuiltGridSize+col] == nil
}

// Draw 山札から捨て札へ 1 枚めくる。尽きたら 1 度だけ組み直せる。
func (c *CrazyQuilt) Draw() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		if c.redealsLeft <= 0 {
			return errors.New("stock is empty and no redeal is left")
		}
		if len(c.waste) == 0 {
			return errors.New("nothing to redeal")
		}
		c.takeSnapshot()
		// **シャッフルしない。**伏せて裏返すだけなので順序は保たれる。
		c.stock = c.waste
		c.waste = nil
		c.redealsLeft--
		c.afterMove("redeal", "捨て札を伏せて山札に戻した", nil)
		return nil
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.waste = append(c.waste, card)
	c.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveQuiltToFoundation キルトの札を基礎札へ送る
func (c *CrazyQuilt) MoveQuiltToFoundation(idx int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCrazyQuiltCell(idx); err != nil {
		return err
	}
	if c.quilt[idx] == nil {
		return fmt.Errorf("cell %d is already empty", idx)
	}
	if !c.IsAvailable(idx) {
		return errors.New("the card has no exposed short side")
	}
	card := c.quilt[idx]
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return errors.New("card cannot be placed on a foundation")
	}
	c.takeSnapshot()
	c.quilt[idx] = nil
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("キルト%d→基礎札%d", idx, fIdx), card)
	return nil
}

// MoveQuiltToWaste キルトの札を捨て札の上へ置く。
//
// 捨て札の一番上と**数字が 1 つ違い**なら置ける（スートは問わない）。
// これが無いとキルトはほとんど崩せない。
func (c *CrazyQuilt) MoveQuiltToWaste(idx int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCrazyQuiltCell(idx); err != nil {
		return err
	}
	if c.quilt[idx] == nil {
		return fmt.Errorf("cell %d is already empty", idx)
	}
	if !c.IsAvailable(idx) {
		return errors.New("the card has no exposed short side")
	}
	top := c.wasteTop()
	if top == nil {
		return errors.New("waste is empty")
	}
	card := c.quilt[idx]
	if !crazyQuiltAdjacentRank(card.GetValue(), top.GetValue()) {
		return errors.New("card is not in sequence with the waste top")
	}
	c.takeSnapshot()
	c.quilt[idx] = nil
	c.waste = append(c.waste, card)
	c.afterMove("move", fmt.Sprintf("キルト%d→捨て札", idx), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *CrazyQuilt) MoveWasteToFoundation() error {
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

// GiveUp ギブアップ
func (c *CrazyQuilt) GiveUp() {
	if c.phase == CrazyQuiltPhasePlaying {
		c.phase = CrazyQuiltPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → キルト崩し → 山めくりの順。手詰まり判定も兼ねる。
func (c *CrazyQuilt) GetHint() *CrazyQuiltHint {
	if c.phase != CrazyQuiltPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	// キルト→捨て札は基礎札に上げられないときの崩し手。1 枚どけると隣の短辺が
	// 開くので、盤面が動く唯一の手になることが多い。
	if top := c.wasteTop(); top != nil {
		for idx := range CrazyQuiltCells {
			card := c.quilt[idx]
			if card == nil || !c.IsAvailable(idx) {
				continue
			}
			if crazyQuiltAdjacentRank(card.GetValue(), top.GetValue()) {
				return &CrazyQuiltHint{FromZone: "quilt", FromIdx: idx, ToZone: "waste", ToIdx: -1}
			}
		}
	}
	if len(c.stock) > 0 || (c.redealsLeft > 0 && len(c.waste) > 0) {
		return &CrazyQuiltHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// キルトを優先する。1 枚どけると隣の短辺が開くので、同じ 1 点でも捨て札より
// 盤面が動く。
func (c *CrazyQuilt) foundationHint() *CrazyQuiltHint {
	if c.phase != CrazyQuiltPhasePlaying {
		return nil
	}
	for idx := range CrazyQuiltCells {
		card := c.quilt[idx]
		if card == nil || !c.IsAvailable(idx) {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &CrazyQuiltHint{FromZone: "quilt", FromIdx: idx, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &CrazyQuiltHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *CrazyQuilt) AutoComplete() error {
	if c.phase != CrazyQuiltPhasePlaying {
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
			err = c.MoveQuiltToFoundation(h.FromIdx)
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
func (c *CrazyQuilt) Undo() error {
	if len(c.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.quilt = snap.quilt
	c.foundation = snap.foundation
	c.stock = snap.stock
	c.waste = snap.waste
	c.redealsLeft = snap.redealsLeft
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (c *CrazyQuilt) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *CrazyQuilt) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *CrazyQuilt) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *crazyQuiltSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *CrazyQuilt) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *CrazyQuilt) GetPhase() CrazyQuiltPhase { return c.phase }

// GetMoveCount 手数取得
func (c *CrazyQuilt) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *CrazyQuilt) GetStockCount() int { return len(c.stock) }

// GetRedealsLeft 残りの組み直し回数
func (c *CrazyQuilt) GetRedealsLeft() int { return c.redealsLeft }

// GetWaste 捨て札を取得
func (c *CrazyQuilt) GetWaste() []*Card { return c.waste }

// GetQuilt キルトを取得（マス番号順、取り除いたマスは nil）
func (c *CrazyQuilt) GetQuilt() [CrazyQuiltCells]*Card { return c.quilt }

// GetFoundation 基礎札を取得
func (c *CrazyQuilt) GetFoundation() [CrazyQuiltFoundationCnt][]*Card { return c.foundation }

// IsAscendingFoundation その基礎札が A からの昇順かを返す
func (c *CrazyQuilt) IsAscendingFoundation(fIdx int) bool {
	if fIdx < 0 || fIdx >= CrazyQuiltFoundationCnt {
		return false
	}
	return crazyQuiltIsAscending(fIdx)
}

// GetGameEndFlag ゲーム終了フラグ
func (c *CrazyQuilt) GetGameEndFlag() bool { return c.phase != CrazyQuiltPhasePlaying }

// IsStalemate 手詰まりか
func (c *CrazyQuilt) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *CrazyQuilt) requirePlaying() error {
	if c.phase != CrazyQuiltPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validCrazyQuiltCell マス番号を検証する
func validCrazyQuiltCell(idx int) error {
	if idx < 0 || idx >= CrazyQuiltCells {
		return fmt.Errorf("invalid cell: %d", idx)
	}
	return nil
}

// crazyQuiltIsAscending その基礎札が A からの昇順か
func crazyQuiltIsAscending(fIdx int) bool { return fIdx < CrazyQuiltAscendingCnt }

// crazyQuiltAdjacentRank 2 つの値が 1 つ違いか。折り返しは無いので K と A は繋がらない。
func crazyQuiltAdjacentRank(a, b int) bool {
	d := a - b
	return d == 1 || d == -1
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *CrazyQuilt) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *CrazyQuilt) popWaste() {
	c.waste = dropLast(c.waste)
}

// canPlaceOnFoundation 基礎札に置けるか。昇順は 1 つ上、降順は 1 つ下。同スートに限る。
func (c *CrazyQuilt) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= CrazyQuiltFoundationCnt {
		return false
	}
	if crazyQuiltSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	pile := c.foundation[fIdx]
	if len(pile) == 0 || len(pile) >= CrazyQuiltFoundationTarget {
		return false
	}
	top := pile[len(pile)-1].GetValue()
	if crazyQuiltIsAscending(fIdx) {
		return card.GetValue() == top+1
	}
	return card.GetValue() == top-1
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）。
//
// 同スート同値の札は 2 組ぶんで 2 枚あり、そのスートの昇順・降順の 2 本が
// それぞれ 1 枚ずつ必要とするので、最初に見つかった本を使ってよい。
func (c *CrazyQuilt) findFoundation(card *Card) int {
	for i := range CrazyQuiltFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *CrazyQuilt) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚積まれたか
func (c *CrazyQuilt) checkGameClear() {
	for i := range CrazyQuiltFoundationCnt {
		if len(c.foundation[i]) != CrazyQuiltFoundationTarget {
			return
		}
	}
	c.phase = CrazyQuiltPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か
func (c *CrazyQuilt) checkStalemate() {
	if c.phase != CrazyQuiltPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *CrazyQuilt) takeSnapshot() {
	snap := &crazyQuiltSnapshot{
		quilt:       c.quilt,
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		redealsLeft: c.redealsLeft,
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range CrazyQuiltFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = append(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *CrazyQuilt) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// crazyQuiltSnapshotJSON is the wire format for a single undo snapshot.
// crazyQuiltSnapshot uses unexported fields, so marshalling it directly would
// emit `[{},{}]` -- the undo depth would survive but every snapshot would be
// blank, and Undo would wipe the board instead of rewinding it (#4478).
type crazyQuiltSnapshotJSON struct {
	Quilt       [CrazyQuiltCells]*Card           `json:"ql"`
	Foundation  [CrazyQuiltFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                          `json:"st"`
	Waste       []*Card                          `json:"ws"`
	RedealsLeft int                              `json:"rd"`
	Phase       CrazyQuiltPhase                  `json:"ps"`
	MoveCount   int                              `json:"mc"`
	IsStalemate bool                             `json:"sl"`
}

// MarshalJSON implements json.Marshaler for crazyQuiltSnapshot.
func (s *crazyQuiltSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(crazyQuiltSnapshotJSON{
		Quilt:       s.quilt,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		RedealsLeft: s.redealsLeft,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for crazyQuiltSnapshot.
func (s *crazyQuiltSnapshot) UnmarshalJSON(data []byte) error {
	var j crazyQuiltSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > crazyQuiltMaxSliceLen || len(j.Waste) > crazyQuiltMaxSliceLen {
		return errors.New("crazyquilt: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Foundation {
		if len(pile) > crazyQuiltMaxSliceLen {
			return errors.New("crazyquilt: snapshot foundation exceeds maximum allowed size")
		}
	}
	if j.RedealsLeft < 0 || j.RedealsLeft > CrazyQuiltRedealCnt {
		return fmt.Errorf("crazyquilt: snapshot redealsLeft out of range: %d", j.RedealsLeft)
	}
	s.quilt = j.Quilt
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.waste = j.Waste
	s.redealsLeft = j.RedealsLeft
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// crazyQuiltJSON is the JSON wire format for CrazyQuilt.
type crazyQuiltJSON struct {
	TrumpCards  *TrumpCards                      `json:"tc"`
	Quilt       [CrazyQuiltCells]*Card           `json:"ql"`
	Foundation  [CrazyQuiltFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                          `json:"st"`
	Waste       []*Card                          `json:"ws"`
	RedealsLeft int                              `json:"rd"`
	Phase       CrazyQuiltPhase                  `json:"ps"`
	MoveCount   int                              `json:"mc"`
	ActionLog   []*ActionLogEntry                `json:"al"`
	IsStalemate bool                             `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*crazyQuiltSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *CrazyQuilt) MarshalJSON() ([]byte, error) {
	return json.Marshal(&crazyQuiltJSON{
		TrumpCards:  c.trumpCards,
		Quilt:       c.quilt,
		Foundation:  c.foundation,
		Stock:       c.stock,
		Waste:       c.waste,
		RedealsLeft: c.redealsLeft,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *CrazyQuilt) UnmarshalJSON(data []byte) error {
	var j crazyQuiltJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < CrazyQuiltPhasePlaying || j.Phase > CrazyQuiltPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if j.RedealsLeft < 0 || j.RedealsLeft > CrazyQuiltRedealCnt {
		return fmt.Errorf("invalid redeals left: %d", j.RedealsLeft)
	}
	if len(j.Stock) > CrazyQuiltTotalCards || len(j.Waste) > CrazyQuiltTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > crazyQuiltMaxSliceLen || len(j.History) > crazyQuiltMaxSliceLen {
		return errors.New("crazyquilt: input array exceeds maximum allowed size")
	}
	for i := range CrazyQuiltFoundationCnt {
		if len(j.Foundation[i]) > CrazyQuiltFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	if j.TrumpCards != nil {
		c.trumpCards = j.TrumpCards
	}
	c.quilt = j.Quilt
	c.foundation = j.Foundation
	c.stock = j.Stock
	c.waste = j.Waste
	c.redealsLeft = j.RedealsLeft
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	c.isStalemate = j.IsStalemate
	c.history = j.History
	return nil
}
