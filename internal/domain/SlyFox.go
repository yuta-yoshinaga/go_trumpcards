//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// SlyFoxPhase スライ・フォックスのゲームフェーズ
type SlyFoxPhase int

// SlyFoxのフェーズ定数
const (
	// SlyFoxPhasePlaying プレイ中
	SlyFoxPhasePlaying SlyFoxPhase = iota
	// SlyFoxPhaseGameClear ゲームクリア
	SlyFoxPhaseGameClear
	// SlyFoxPhaseGameOver ゲームオーバー
	SlyFoxPhaseGameOver
)

// SlyFoxTableauCnt リザーブの枠数。10 枠 2 段に並ぶ。
const SlyFoxTableauCnt = 20

// SlyFoxDealCycle 1 周で配る枚数。**この枚数を配り切るまでリザーブから組札へは
// 送れない。**20 回の置き場所を先に決めきってから収穫する、という構造がこの
// ゲームそのもので、クローン元のコロラドにはこの縛りが無い。
const SlyFoxDealCycle = 20

// SlyFoxFoundationCnt 基礎札の数。前半 4 つが A からの昇順、後半 4 つが K からの降順。
const SlyFoxFoundationCnt = 8

// SlyFoxFoundationTarget 基礎札 1 つあたりの完成枚数
const SlyFoxFoundationTarget = CardValueMax

// SlyFoxTotalCards 使用する総枚数（52 枚 2 組）
const SlyFoxTotalCards = CardCnt * 2

// SlyFoxAscendingCnt 昇順の基礎札の数（インデックス 0 以上この値未満が昇順）
const SlyFoxAscendingCnt = SlyFoxFoundationCnt / 2

// slyFoxMaxSliceLen caps slice sizes during deserialisation.
const slyFoxMaxSliceLen = 1000

// slyFoxSuitOrder 基礎札インデックスとスートの対応。前半 4 つが昇順、後半 4 つが
// 降順で、どちらも同じスート順に並ぶ。固定しておくと配り直しても UI の位置が動かない。
var slyFoxSuitOrder = [SlyFoxFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// SlyFoxHint スライ・フォックスのヒント
type SlyFoxHint struct {
	// FromZone 移動元 "tableau"、または配りを勧める "stock"
	FromZone string
	// FromIdx 移動元のタブロー山（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation"、または配り先の枠を指す "tableau"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// SlyFox スライ・フォックス ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**20 枠のリザーブ**に 1 枚ずつ配って
// 始め、残り 84 枚が山札になる。基礎札は 8 つで、**4 つが A から K への昇順、
// 残り 4 つが K から A への降順**。8×13 = 104 枚すべてを積み切ればクリア。
//
// **山札は 20 枚を 1 周として配る。**めくった札はその場でリザーブの好きな枠へ
// 置くか、置ける組札があればそちらへ直接送る（組札行きは 20 枚に数えない）。
// **20 枚を配り切るまで、リザーブから組札へは送れない。**置き場所を 20 回
// 先に決めきってから収穫する、この一方通行がゲームの緊張そのもの。
//
// リザーブは**スートも数字も問わず**どの枠にも置けるが、置いた札は下の札を
// 埋めてしまう。**空いた枠は補充されない**ので、埋めても痛くない枠を選ぶのが
// 上手さになる。
//
// クローン元のコロラドとは 3 点異なる（両者は資料上しばしば混同され、Wikipedia は
// 一方から他方へリダイレクトする。PySolFC は別ゲームとして実装しており、差はここ）:
//   - コロラドは**捨て札を挟む**。Sly Fox はめくった札をその場で置く
//   - コロラドは**いつでも**タブローから組札へ送れる。Sly Fox は 20 枚配り切るまで送れない
//   - コロラドは**空き山を山札から直接埋められる**。Sly Fox に補充は無い
type SlyFox struct {
	trumpCards *TrumpCards
	tableau    [SlyFoxTableauCnt][]*Card
	foundation [SlyFoxFoundationCnt][]*Card
	stock      []*Card
	// dealtThisCycle この周でリザーブに置いた枚数 (0..SlyFoxDealCycle)。
	// 組札へ直接送った札は数えない。
	dealtThisCycle int
	phase          SlyFoxPhase
	moveCount      int
	actionLogBase
	history     []*slyFoxSnapshot
	isStalemate bool
}

// slyFoxSnapshot アンドゥ用スナップショット
type slyFoxSnapshot struct {
	tableau        [SlyFoxTableauCnt][]*Card
	foundation     [SlyFoxFoundationCnt][]*Card
	stock          []*Card
	dealtThisCycle int
	phase          SlyFoxPhase
	moveCount      int
	isStalemate    bool
}

// NewSlyFox コンストラクタ
func NewSlyFox(trumpCards *TrumpCards) *SlyFox {
	return &SlyFox{trumpCards: trumpCards}
}

// NewDefaultSlyFox returns SlyFox with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSlyFox() *SlyFox {
	return NewSlyFox(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *SlyFox) Reset() {
	c.trumpCards.Shuffle()
	c.phase = SlyFoxPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	// 最初の 20 枚は配り終えた状態から始まるので、カウンタも配り切った形にする。
	// 0 から始めると、開幕から 20 枚配らないと 1 枚も収穫できない盤になる。
	c.dealtThisCycle = SlyFoxDealCycle

	for i := range SlyFoxFoundationCnt {
		c.foundation[i] = nil
	}
	// リザーブは 20 枠に 1 枚ずつ。残りはすべて山札。
	for i := range SlyFoxTableauCnt {
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

// DealToPile 山札から 1 枚めくって、選んだリザーブ枠に置く。
//
// **捨て札は無い。**めくった札はその場で行き先が決まる ── 置き直しは効かない。
// この 1 手が 20 回積み重なって 1 周になり、周を終えるまでリザーブから組札へは
// 送れない。
func (c *SlyFox) DealToPile(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validSlyFoxPile(pile); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "slyfox.errStockEmptyNoRedeal", nil)
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.tableau[pile] = append(c.tableau[pile], card)
	// **周をまたいだら数え直す。**20 で頭打ちにすると、1 周開いたあと
	// 開きっぱなしになってクローン元のコロラドと同じ挙動になる。
	if c.dealtThisCycle >= SlyFoxDealCycle {
		c.dealtThisCycle = 0
	}
	c.dealtThisCycle++
	c.afterMove("deal", fmt.Sprintf("山札→リザーブ枠%d", pile), card)
	return nil
}

// DealToFoundation 山札から 1 枚めくって、そのまま基礎札へ送る。
//
// **この札は 20 枚に数えない。**数えてしまうと、運良く積める札を引いたぶんだけ
// 配れる枚数が減り、良い引きが罰になる。
func (c *SlyFox) DealToFoundation(fIdx int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if fIdx < 0 || fIdx >= SlyFoxFoundationCnt {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errInvalidFoundation", nil)
	}
	if len(c.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "slyfox.errStockEmptyNoRedeal", nil)
	}
	if !c.canPlaceOnFoundation(c.stock[0], fIdx) {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errCannotPlaceFoundation", nil)
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("山札→基礎札%d", fIdx), card)
	return nil
}

// reserveIsLocked は、この周がまだ終わっていないためリザーブから組札へ送れない
// 状態かを返す。
//
// **山札が尽きたら開く。**終盤は残ったリザーブだけで詰めるので、閉じたままだと
// 必ず手詰まりになる。
func (c *SlyFox) reserveIsLocked() bool {
	if len(c.stock) == 0 {
		return false
	}
	return c.dealtThisCycle < SlyFoxDealCycle
}

// ReserveIsLocked は、この周を配り切るまでリザーブが閉じているかを返す。
func (c *SlyFox) ReserveIsLocked() bool { return c.reserveIsLocked() }

// DealtThisCycle はこの周でリザーブに置いた枚数を返す。
func (c *SlyFox) DealtThisCycle() int { return c.dealtThisCycle }

// MoveTableauToFoundation リザーブ枠の一番上を基礎札へ送る
func (c *SlyFox) MoveTableauToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validSlyFoxPile(pile); err != nil {
		return err
	}
	// **周を配り切るまで閉じている。**これが Sly Fox の核で、クローン元の
	// コロラドはいつでも送れる。
	if c.reserveIsLocked() {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errDealInProgress",
			map[string]string{"dealt": strconv.Itoa(c.dealtThisCycle), "cycle": strconv.Itoa(SlyFoxDealCycle)})
	}
	card := c.tableauTop(pile)
	if card == nil {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errPileEmpty", map[string]string{"pile": strconv.Itoa(pile)})
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errCannotPlaceFoundation", nil)
	}
	c.takeSnapshot()
	c.tableau[pile] = dropLast(c.tableau[pile])
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("リザーブ枠%d→基礎札%d", pile, fIdx), card)
	return nil
}

// GiveUp ギブアップ
func (c *SlyFox) GiveUp() {
	if c.phase == SlyFoxPhasePlaying {
		c.phase = SlyFoxPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → 山札から直接組札 → 配る の順。
// 手詰まり判定も兼ねる。
//
// 「配る」は常に成立してしまうので最後に回している。先に返すと、ヒントが毎回
// 同じ無意味な手になる。
func (c *SlyFox) GetHint() *SlyFoxHint {
	if c.phase != SlyFoxPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	if len(c.stock) > 0 {
		// めくった札がそのまま組札へ行けるなら、それが最良。20 枚に数えない。
		for fIdx := range SlyFoxFoundationCnt {
			if c.canPlaceOnFoundation(c.stock[0], fIdx) {
				return &SlyFoxHint{FromZone: "stock", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
			}
		}
		// そうでなければ、一番損の小さい枠へ配る。
		return &SlyFoxHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: c.bestBuryPile()}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (c *SlyFox) foundationHint() *SlyFoxHint {
	if c.phase != SlyFoxPhasePlaying {
		return nil
	}
	// **閉じている間は 1 件も返さない。**手で送れば拒まれる手を指し続けると、
	// ヒントが「サーバに拒まれる手」の一覧になる。オートコンプリートもここを
	// 読むので、閉じている間は動かないのが正しい。
	if c.reserveIsLocked() {
		return nil
	}
	for pile := range SlyFoxTableauCnt {
		card := c.tableauTop(pile)
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &SlyFoxHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// bestBuryPile めくった札を置くのに一番損の小さいリザーブ枠を返す。
//
// 空き枠があればそこ（何も埋めない）。無ければ、一番上の札が基礎札に必要になるまで
// 最も遠い枠を選ぶ。同点なら添字の小さい方。
func (c *SlyFox) bestBuryPile() int {
	best, bestCost := 0, -1
	for pile := range SlyFoxTableauCnt {
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
func (c *SlyFox) buryCost(card *Card) int {
	if card == nil {
		return SlyFoxFoundationTarget + 1
	}
	best := SlyFoxFoundationTarget + 1
	for i := range SlyFoxFoundationCnt {
		if slyFoxSuitOrder[i] != card.GetDesign() {
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
func (c *SlyFox) foundationNeed(fIdx, v int) int {
	filled := len(c.foundation[fIdx])
	if slyFoxIsAscending(fIdx) {
		// 昇順は A(1) から。filled 枚積んだ次に必要なのは filled+1。
		return v - (filled + 1)
	}
	// 降順は K(CardValueMax) から。filled 枚積んだ次に必要なのは CardValueMax-filled。
	return (CardValueMax - filled) - v
}

// slyFoxIsAscending その基礎札が A からの昇順か（false なら K からの降順）
func slyFoxIsAscending(fIdx int) bool { return fIdx < SlyFoxAscendingCnt }

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *SlyFox) AutoComplete() error {
	if c.phase != SlyFoxPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := c.foundationHint()
		if h == nil {
			break
		}
		if err := c.MoveTableauToFoundation(h.FromIdx); err != nil {
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
func (c *SlyFox) Undo() error {
	if len(c.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.tableau = snap.tableau
	c.foundation = snap.foundation
	c.stock = snap.stock
	c.dealtThisCycle = snap.dealtThisCycle
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (c *SlyFox) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *SlyFox) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *SlyFox) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *slyFoxSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *SlyFox) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *SlyFox) GetPhase() SlyFoxPhase { return c.phase }

// GetMoveCount 手数取得
func (c *SlyFox) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *SlyFox) GetStockCount() int { return len(c.stock) }

// GetTableau リザーブ枠を取得
func (c *SlyFox) GetTableau() [SlyFoxTableauCnt][]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *SlyFox) GetFoundation() [SlyFoxFoundationCnt][]*Card { return c.foundation }

// IsAscendingFoundation その基礎札が A からの昇順かを返す（表示の向き分け用）
func (c *SlyFox) IsAscendingFoundation(fIdx int) bool {
	if fIdx < 0 || fIdx >= SlyFoxFoundationCnt {
		return false
	}
	return slyFoxIsAscending(fIdx)
}

// GetGameEndFlag ゲーム終了フラグ
func (c *SlyFox) GetGameEndFlag() bool { return c.phase != SlyFoxPhasePlaying }

// IsStalemate 手詰まりか
func (c *SlyFox) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *SlyFox) requirePlaying() error {
	if c.phase != SlyFoxPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validSlyFoxPile リザーブ枠のインデックスを検証する
func validSlyFoxPile(pile int) error {
	if pile < 0 || pile >= SlyFoxTableauCnt {
		return NewDomainErrorCode(ErrInvalidPlay, "slyfox.errInvalidPile",
			map[string]string{"pile": strconv.Itoa(pile)})
	}
	return nil
}

// tableauTop リザーブ枠の一番上（空なら nil）
func (c *SlyFox) tableauTop(pile int) *Card {
	return discardTop(c.tableau[pile])
}

// canPlaceOnFoundation 基礎札に置けるか。昇順は空なら A、以降は 1 つ上。
// 降順は空なら K、以降は 1 つ下。どちらも同スートに限る。
func (c *SlyFox) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= SlyFoxFoundationCnt {
		return false
	}
	if slyFoxSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	if len(c.foundation[fIdx]) >= SlyFoxFoundationTarget {
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
func (c *SlyFox) findFoundation(card *Card) int {
	for i := range SlyFoxFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *SlyFox) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて 13 枚積まれたか
func (c *SlyFox) checkGameClear() {
	for i := range SlyFoxFoundationCnt {
		if len(c.foundation[i]) != SlyFoxFoundationTarget {
			return
		}
	}
	c.phase = SlyFoxPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
//
// 山札がある限り必ず配れるので、手詰まりは「山札が尽き、リザーブのどの一番上も
// 基礎札へ送れない」ときにだけ起きる。山札は減る一方なので、この状態には必ず
// 到達する。
func (c *SlyFox) checkStalemate() {
	if c.phase != SlyFoxPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *SlyFox) takeSnapshot() {
	snap := &slyFoxSnapshot{
		stock:          append([]*Card(nil), c.stock...),
		dealtThisCycle: c.dealtThisCycle,
		phase:          c.phase,
		moveCount:      c.moveCount,
		isStalemate:    c.isStalemate,
	}
	for i := range SlyFoxTableauCnt {
		snap.tableau[i] = append([]*Card(nil), c.tableau[i]...)
	}
	for i := range SlyFoxFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = append(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *SlyFox) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// slyFoxSnapshotJSON is the wire format for a single undo snapshot.
// slyFoxSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type slyFoxSnapshotJSON struct {
	Tableau        [SlyFoxTableauCnt][]*Card    `json:"tb"`
	Foundation     [SlyFoxFoundationCnt][]*Card `json:"fd"`
	Stock          []*Card                      `json:"st"`
	DealtThisCycle int                          `json:"dc"`
	Phase          SlyFoxPhase                  `json:"ps"`
	MoveCount      int                          `json:"mc"`
	IsStalemate    bool                         `json:"sl"`
}

// MarshalJSON implements json.Marshaler for slyFoxSnapshot.
func (s *slyFoxSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(slyFoxSnapshotJSON{
		Tableau:        s.tableau,
		Foundation:     s.foundation,
		Stock:          s.stock,
		DealtThisCycle: s.dealtThisCycle,
		Phase:          s.phase,
		MoveCount:      s.moveCount,
		IsStalemate:    s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for slyFoxSnapshot.
func (s *slyFoxSnapshot) UnmarshalJSON(data []byte) error {
	var j slyFoxSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > slyFoxMaxSliceLen {
		return errors.New("slyfox: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > slyFoxMaxSliceLen {
			return errors.New("slyfox: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > slyFoxMaxSliceLen {
			return errors.New("slyfox: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.dealtThisCycle = j.DealtThisCycle
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// slyFoxJSON is the JSON wire format for SlyFox.
type slyFoxJSON struct {
	TrumpCards     *TrumpCards                  `json:"tc"`
	Tableau        [SlyFoxTableauCnt][]*Card    `json:"tb"`
	Foundation     [SlyFoxFoundationCnt][]*Card `json:"fd"`
	Stock          []*Card                      `json:"st"`
	DealtThisCycle int                          `json:"dc"`
	Phase          SlyFoxPhase                  `json:"ps"`
	MoveCount      int                          `json:"mc"`
	ActionLog      []*ActionLogEntry            `json:"al"`
	IsStalemate    bool                         `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*slyFoxSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *SlyFox) MarshalJSON() ([]byte, error) {
	return json.Marshal(&slyFoxJSON{
		TrumpCards:     c.trumpCards,
		Tableau:        c.tableau,
		Foundation:     c.foundation,
		Stock:          c.stock,
		DealtThisCycle: c.dealtThisCycle,
		Phase:          c.phase,
		MoveCount:      c.moveCount,
		ActionLog:      c.actionLog,
		IsStalemate:    c.isStalemate,
		History:        c.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *SlyFox) UnmarshalJSON(data []byte) error {
	var j slyFoxJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < SlyFoxPhasePlaying || j.Phase > SlyFoxPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > SlyFoxTotalCards {
		return fmt.Errorf("stock too large: %d", len(j.Stock))
	}
	if j.DealtThisCycle < 0 || j.DealtThisCycle > SlyFoxDealCycle {
		return fmt.Errorf("invalid deal count: %d", j.DealtThisCycle)
	}
	if len(j.ActionLog) > slyFoxMaxSliceLen || len(j.History) > slyFoxMaxSliceLen {
		return errors.New("slyfox: input array exceeds maximum allowed size")
	}
	for i := range SlyFoxFoundationCnt {
		if len(j.Foundation[i]) > SlyFoxFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range SlyFoxTableauCnt {
		if len(j.Tableau[i]) > SlyFoxTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		c.trumpCards = j.TrumpCards
	}
	c.tableau = j.Tableau
	c.foundation = j.Foundation
	c.stock = j.Stock
	c.dealtThisCycle = j.DealtThisCycle
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	c.isStalemate = j.IsStalemate
	c.history = j.History
	return nil
}
