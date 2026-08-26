//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// DiplomatPhase ディプロマットのゲームフェーズ
type DiplomatPhase int

// Diplomatのフェーズ定数
const (
	// DiplomatPhasePlaying プレイ中
	DiplomatPhasePlaying DiplomatPhase = iota
	// DiplomatPhaseGameClear ゲームクリア
	DiplomatPhaseGameClear
	// DiplomatPhaseGameOver ゲームオーバー
	DiplomatPhaseGameOver
)

// DiplomatTableauCnt タブローの列数
const DiplomatTableauCnt = 8

// DiplomatDealPerColumn 配りきりの 1 列あたり枚数（8 列 × 4 枚 = 32 枚）
const DiplomatDealPerColumn = 4

// DiplomatFoundationCnt 基礎札の数（スートごとに 2 つ、A→K の昇順）
const DiplomatFoundationCnt = 8

// DiplomatFoundationTarget 基礎札 1 つあたりの完成枚数（A→K）
const DiplomatFoundationTarget = CardValueMax

// DiplomatTotalCards 使用する総枚数（52 枚 2 組）
const DiplomatTotalCards = CardCnt * 2

// diplomatMaxSliceLen caps slice sizes during deserialisation.
const diplomatMaxSliceLen = 1000

// diplomatSuitOrder 基礎札インデックスとスートの対応。スートごとに 2 つあり、
// 固定しておくと配り直しても UI の位置が動かない。
var diplomatSuitOrder = [DiplomatFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// DiplomatHint ディプロマットのヒント
type DiplomatHint struct {
	// FromZone 移動元 "tableau" / "waste" / "stock"
	FromZone string
	// FromIdx 移動元のタブロー山（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// Diplomat ディプロマット ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。**8 列に 4 枚ずつ計 32 枚**を表向きで
// 配り、残り 72 枚が山札になる。
//
// 基礎札は 8 つ。**最初は空**で、A が出てきたときにそこへ置き、同スートで K まで
// 13 枚積む。8×13 = 104 枚すべてを積み切ればクリア。
//
// タブローは**スートを無視した降順**で、**1 枚ずつしか動かせない**（連番のまとめ
// 移動は無い）。空いた列には**タブローか捨て札から 1 枚**を置ける。
//
// 山札は 1 枚ずつ捨て札へめくり、**めくり直しは無い（1 巡のみ）**。
//
// Forty Thieves 系の易しい版と呼ばれるのは、あちらがタブローを**同スート**降順に
// 限るのに対し、こちらはスートを問わないため。Congress は同じ 104 枚・8 基礎札・
// スート無視の降順だが、配りが 8 山 1 枚ずつで、空き山をタブローから埋められない。
//
// issue #5276 の仕様案とは 3 点異なり、いずれも実際の規則に合わせた:
//   - **4 枚のリザーブは無い。** 8×4 = 32 枚を配って残り 72 枚が山札、というのが
//     どの規則書も一致している枚数配分で、リザーブがあると山札は 68 枚になる
//   - **山札は 3 枚ずつではなく 1 枚ずつめくる。** 3 枚めくりは Klondike 系の作法で、
//     この系統には無い
//   - **空いた列はタブローからも埋められる。**「別の列か捨て札から 1 枚」が規則で、
//     山札から直接置くことはできない
type Diplomat struct {
	trumpCards *TrumpCards
	tableau    [DiplomatTableauCnt][]*Card
	foundation [DiplomatFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      DiplomatPhase
	moveCount  int
	actionLogBase
	history     []*diplomatSnapshot
	isStalemate bool
}

// diplomatSnapshot アンドゥ用スナップショット
type diplomatSnapshot struct {
	tableau     [DiplomatTableauCnt][]*Card
	foundation  [DiplomatFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	phase       DiplomatPhase
	moveCount   int
	isStalemate bool
}

// NewDiplomat コンストラクタ
func NewDiplomat(trumpCards *TrumpCards) *Diplomat {
	return &Diplomat{trumpCards: trumpCards}
}

// NewDefaultDiplomat returns Diplomat with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultDiplomat() *Diplomat {
	return NewDiplomat(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *Diplomat) Reset() {
	c.trumpCards.Shuffle()
	c.phase = DiplomatPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil

	for i := range DiplomatFoundationCnt {
		c.foundation[i] = nil
	}
	// タブローは 8 列に 4 枚ずつ、計 32 枚。残り 72 枚がすべて山札。
	for i := range DiplomatTableauCnt {
		c.tableau[i] = nil
		for range DiplomatDealPerColumn {
			card := c.trumpCards.DrawCard()
			if card == nil {
				break
			}
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
func (c *Diplomat) Draw() error {
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
func (c *Diplomat) MoveTableauToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validDiplomatPile(pile); err != nil {
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
	c.tableau[pile] = c.tableau[pile][:len(c.tableau[pile])-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("タブロー山%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間で 1 枚だけ動かす。
//
// このゲームに連番のまとめ移動は無い。**空いた列にはどのカードでも置ける**ので、
// 空き列は Congress と違ってタブローの逃がし先として使える。
func (c *Diplomat) MoveTableauToTableau(fromPile, toPile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validDiplomatPile(fromPile); err != nil {
		return err
	}
	if err := validDiplomatPile(toPile); err != nil {
		return err
	}
	if fromPile == toPile {
		return errors.New("source and destination are the same pile")
	}
	card := c.tableauTop(fromPile)
	if card == nil {
		return fmt.Errorf("pile %d is empty", fromPile)
	}
	if !c.canPlaceOnTableau(card, toPile) {
		return errors.New("card cannot be placed on that pile")
	}
	c.takeSnapshot()
	c.tableau[fromPile] = c.tableau[fromPile][:len(c.tableau[fromPile])-1]
	c.tableau[toPile] = append(c.tableau[toPile], card)
	c.afterMove("move", fmt.Sprintf("タブロー山%d→タブロー山%d", fromPile, toPile), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *Diplomat) MoveWasteToFoundation() error {
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

// MoveWasteToTableau 捨て札の一番上をタブローへ送る。空いた列も埋められる。
func (c *Diplomat) MoveWasteToTableau(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validDiplomatPile(pile); err != nil {
		return err
	}
	card := c.wasteTop()
	if card == nil {
		return errors.New("waste is empty")
	}
	if !c.canPlaceOnTableau(card, pile) {
		return errors.New("card cannot be placed on that pile")
	}
	c.takeSnapshot()
	c.popWaste()
	c.tableau[pile] = append(c.tableau[pile], card)
	c.afterMove("move", fmt.Sprintf("捨て札→タブロー山%d", pile), card)
	return nil
}

// GiveUp ギブアップ
func (c *Diplomat) GiveUp() {
	if c.phase == DiplomatPhasePlaying {
		c.phase = DiplomatPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (c *Diplomat) GetHint() *DiplomatHint {
	if c.phase != DiplomatPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	if h := c.tableauHint(); h != nil {
		return h
	}
	if len(c.stock) > 0 {
		return &DiplomatHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (c *Diplomat) foundationHint() *DiplomatHint {
	if c.phase != DiplomatPhasePlaying {
		return nil
	}
	for pile := range DiplomatTableauCnt {
		card := c.tableauTop(pile)
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &DiplomatHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &DiplomatHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
func (c *Diplomat) tableauHint() *DiplomatHint {
	if c.phase != DiplomatPhasePlaying {
		return nil
	}
	if card := c.wasteTop(); card != nil {
		for pile := range DiplomatTableauCnt {
			if c.canPlaceOnTableau(card, pile) {
				return &DiplomatHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: pile}
			}
		}
	}
	// 列同士。空いた列も置き先になるが、**1 枚しか埋まっていない列を空き列へ
	// 動かす手は勧めない** — 空き列がもう一方へ移るだけで盤面が進まず、
	// 同じ手を延々と提示し続けることになる。
	for from := range DiplomatTableauCnt {
		card := c.tableauTop(from)
		if card == nil {
			continue
		}
		for to := range DiplomatTableauCnt {
			if from == to {
				continue
			}
			if len(c.tableau[to]) == 0 && len(c.tableau[from]) == 1 {
				continue
			}
			if c.canPlaceOnTableau(card, to) {
				return &DiplomatHint{FromZone: "tableau", FromIdx: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *Diplomat) AutoComplete() error {
	if c.phase != DiplomatPhasePlaying {
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
func (c *Diplomat) Undo() error {
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
func (c *Diplomat) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *Diplomat) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *Diplomat) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *diplomatSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *Diplomat) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *Diplomat) GetPhase() DiplomatPhase { return c.phase }

// GetMoveCount 手数取得
func (c *Diplomat) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *Diplomat) GetStockCount() int { return len(c.stock) }

// GetWaste 捨て札を取得
func (c *Diplomat) GetWaste() []*Card { return c.waste }

// GetTableau タブローを取得
func (c *Diplomat) GetTableau() [DiplomatTableauCnt][]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *Diplomat) GetFoundation() [DiplomatFoundationCnt][]*Card { return c.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (c *Diplomat) GetGameEndFlag() bool { return c.phase != DiplomatPhasePlaying }

// IsStalemate 手詰まりか
func (c *Diplomat) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *Diplomat) requirePlaying() error {
	if c.phase != DiplomatPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// validDiplomatPile タブロー山のインデックスを検証する
func validDiplomatPile(pile int) error {
	if pile < 0 || pile >= DiplomatTableauCnt {
		return fmt.Errorf("invalid pile: %d", pile)
	}
	return nil
}

// tableauTop 山の一番上（空なら nil）
func (c *Diplomat) tableauTop(pile int) *Card {
	if len(c.tableau[pile]) == 0 {
		return nil
	}
	return c.tableau[pile][len(c.tableau[pile])-1]
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *Diplomat) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *Diplomat) popWaste() {
	c.waste = dropLast(c.waste)
}

// canPlaceOnTableau タブローに置けるか（スート無視の降順、A の下には何も置けない）。
// 空いた列にはどのカードでも置ける。
func (c *Diplomat) canPlaceOnTableau(card *Card, pile int) bool {
	if card == nil {
		return false
	}
	if len(c.tableau[pile]) == 0 {
		return true
	}
	top := c.tableau[pile][len(c.tableau[pile])-1]
	if top == nil {
		return false
	}
	// 折り返しは無い。A の上には何も置けず、その山はそこで止まる。
	return card.GetValue() == top.GetValue()-1
}

// DiplomatIsDeadEndTop は、その札が一番上にある列がタブロー間移動の
// 受け皿として死んでいるかを返す。
//
// **A の上には何も置けない** (canPlaceOnTableau は top-1 しか通さないので、
// A のとき通る値は 0 = 存在しないランク)。空き列が主要な逃げ道なので、
// 詰んだ列を早く見分けられるかどうかは実際の判断に効く (#5741)。
func DiplomatIsDeadEndTop(card *Card) bool {
	return card != nil && card.GetValue() == 1
}

// canPlaceOnFoundation 基礎札に置けるか（空なら A、以降は同スートで 1 つ上）
func (c *Diplomat) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if diplomatSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	pile := c.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	if len(pile) >= DiplomatFoundationTarget {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()+1
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）
func (c *Diplomat) findFoundation(card *Card) int {
	for i := range DiplomatFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *Diplomat) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて K まで積まれたか
func (c *Diplomat) checkGameClear() {
	for i := range DiplomatFoundationCnt {
		if len(c.foundation[i]) != DiplomatFoundationTarget {
			return
		}
	}
	c.phase = DiplomatPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (c *Diplomat) checkStalemate() {
	if c.phase != DiplomatPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *Diplomat) takeSnapshot() {
	snap := &diplomatSnapshot{
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range DiplomatTableauCnt {
		snap.tableau[i] = append([]*Card(nil), c.tableau[i]...)
	}
	for i := range DiplomatFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = appendSnapshot(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *Diplomat) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// diplomatSnapshotJSON is the wire format for a single undo snapshot.
// diplomatSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type diplomatSnapshotJSON struct {
	Tableau     [DiplomatTableauCnt][]*Card    `json:"tb"`
	Foundation  [DiplomatFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       DiplomatPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for diplomatSnapshot.
func (s *diplomatSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(diplomatSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for diplomatSnapshot.
func (s *diplomatSnapshot) UnmarshalJSON(data []byte) error {
	var j diplomatSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > diplomatMaxSliceLen || len(j.Waste) > diplomatMaxSliceLen {
		return errors.New("diplomat: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > diplomatMaxSliceLen {
			return errors.New("diplomat: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > diplomatMaxSliceLen {
			return errors.New("diplomat: snapshot pile exceeds maximum allowed size")
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

// diplomatJSON is the JSON wire format for Diplomat.
type diplomatJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [DiplomatTableauCnt][]*Card    `json:"tb"`
	Foundation  [DiplomatFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       DiplomatPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*diplomatSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *Diplomat) MarshalJSON() ([]byte, error) {
	return json.Marshal(&diplomatJSON{
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
func (c *Diplomat) UnmarshalJSON(data []byte) error {
	var j diplomatJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < DiplomatPhasePlaying || j.Phase > DiplomatPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > DiplomatTotalCards || len(j.Waste) > DiplomatTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > diplomatMaxSliceLen || len(j.History) > diplomatMaxSliceLen {
		return errors.New("diplomat: input array exceeds maximum allowed size")
	}
	for i := range DiplomatFoundationCnt {
		if len(j.Foundation[i]) > DiplomatFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range DiplomatTableauCnt {
		if len(j.Tableau[i]) > DiplomatTotalCards {
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
