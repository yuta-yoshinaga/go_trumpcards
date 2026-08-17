//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// CongressPhase コングレスのゲームフェーズ
type CongressPhase int

// Congressのフェーズ定数
const (
	// CongressPhasePlaying プレイ中
	CongressPhasePlaying CongressPhase = iota
	// CongressPhaseGameClear ゲームクリア
	CongressPhaseGameClear
	// CongressPhaseGameOver ゲームオーバー
	CongressPhaseGameOver
)

// CongressTableauCnt タブローの山数。左右に 4 つずつ、基礎札を挟んで並ぶ。
const CongressTableauCnt = 8

// CongressFoundationCnt 基礎札の数（スートごとに 2 つ）
const CongressFoundationCnt = 8

// CongressFoundationTarget 基礎札 1 つあたりの完成枚数（A→K）
const CongressFoundationTarget = CardValueMax

// CongressTotalCards 使用する総枚数（52 枚 2 組）
const CongressTotalCards = CardCnt * 2

// congressMaxSliceLen caps slice sizes during deserialisation.
const congressMaxSliceLen = 1000

// congressSuitOrder 基礎札インデックスとスートの対応。スートごとに 2 つあり、
// 固定しておくと配り直しても UI の位置が動かない。
var congressSuitOrder = [CongressFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// CongressHint コングレスのヒント
type CongressHint struct {
	// FromZone 移動元 "tableau" / "waste" / "stock"
	FromZone string
	// FromIdx 移動元のタブロー山（それ以外は -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の山を指さない）
	ToIdx int
}

// Congress コングレス ゲームクラス。
//
// 52 枚 2 組（104 枚）の 1 人用ソリティア。基礎札 8 つを中央に置き、その左右に
// **4 つずつ計 8 つのタブロー山**を並べる。タブローには**最初 1 枚ずつ**しか配らず、
// 残り 96 枚が山札になる。
//
// 基礎札は**最初は空**で、A が出てきたときにそこへ置き、同スートで K まで 13 枚
// 積む。8×13 = 104 枚すべてを積み切ればクリア。
//
// タブローは**スート・色を無視した降順**で、**1 枚ずつしか動かせない**（連番の
// まとめ移動は無い）。空いた山は**山札か捨て札の一番上から**埋める。
//
// 山札は 1 枚ずつ捨て札へめくり、**めくり直しは無い（1 巡のみ）**。
//
// issue #4419 の仕様案とは 4 点異なり、いずれも実際の規則に合わせた:
//   - **8 枚の A を最初から基礎札に並べるのではない**。基礎札は空で始まり、A が
//     出た順に置く（並べてしまうと山札から A を探し出す必要があり、別のゲームになる）
//   - **残り 96 枚はタブローではなく山札**。タブローは 8 山に 1 枚ずつの計 8 枚
//     （issue の配り方だと山札が 0 枚になり、規則 5 の「山札をめくる」と矛盾する）
//   - 基礎札は「各スート 8 枚まで」ではなく **8 つの山に 13 枚ずつ**（A→K）。
//     前者は 32 枚にしかならず 104 枚を吸収できない
//   - **空き山は山札か捨て札から埋める**（issue は触れていない）
type Congress struct {
	trumpCards *TrumpCards
	tableau    [CongressTableauCnt][]*Card
	foundation [CongressFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	phase      CongressPhase
	moveCount  int
	actionLogBase
	history     []*congressSnapshot
	isStalemate bool
}

// congressSnapshot アンドゥ用スナップショット
type congressSnapshot struct {
	tableau     [CongressTableauCnt][]*Card
	foundation  [CongressFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	phase       CongressPhase
	moveCount   int
	isStalemate bool
}

// NewCongress コンストラクタ
func NewCongress(trumpCards *TrumpCards) *Congress {
	return &Congress{trumpCards: trumpCards}
}

// NewDefaultCongress returns Congress with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultCongress() *Congress {
	return NewCongress(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *Congress) Reset() {
	c.trumpCards.Shuffle()
	c.phase = CongressPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.waste = nil

	for i := range CongressFoundationCnt {
		c.foundation[i] = nil
	}
	// タブローは 8 山に 1 枚ずつ。残りはすべて山札。
	for i := range CongressTableauCnt {
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
func (c *Congress) Draw() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "congress.errStockEmptyNoRedeal", nil)
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.waste = append(c.waste, card)
	c.afterMove("draw", "山札から1枚めくった", card)
	return nil
}

// MoveTableauToFoundation タブローの一番上を基礎札へ送る
func (c *Congress) MoveTableauToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCongressPile(pile); err != nil {
		return err
	}
	card := c.tableauTop(pile)
	if card == nil {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errPileEmpty", map[string]string{"pile": strconv.Itoa(pile)})
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errNoFoundationForCard", nil)
	}
	c.takeSnapshot()
	c.tableau[pile] = c.tableau[pile][:len(c.tableau[pile])-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("タブロー山%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間で 1 枚だけ動かす。
//
// このゲームに連番のまとめ移動は無い。空き山はタブローからは埋められず、
// 山札か捨て札の出口として使う。
func (c *Congress) MoveTableauToTableau(fromPile, toPile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCongressPile(fromPile); err != nil {
		return err
	}
	if err := validCongressPile(toPile); err != nil {
		return err
	}
	if fromPile == toPile {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errSamePile", nil)
	}
	card := c.tableauTop(fromPile)
	if card == nil {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errPileEmpty", map[string]string{"pile": strconv.Itoa(fromPile)})
	}
	if len(c.tableau[toPile]) == 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errEmptyPileNeedsStockOrWaste", nil)
	}
	if !c.canPlaceOnTableau(card, toPile) {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errCannotPlaceOnPile", nil)
	}
	c.takeSnapshot()
	c.tableau[fromPile] = c.tableau[fromPile][:len(c.tableau[fromPile])-1]
	c.tableau[toPile] = append(c.tableau[toPile], card)
	c.afterMove("move", fmt.Sprintf("タブロー山%d→タブロー山%d", fromPile, toPile), card)
	return nil
}

// MoveWasteToFoundation 捨て札の一番上を基礎札へ送る
func (c *Congress) MoveWasteToFoundation() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	card := c.wasteTop()
	if card == nil {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errWasteEmpty", nil)
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errNoFoundationForCard", nil)
	}
	c.takeSnapshot()
	c.popWaste()
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("捨て札→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToTableau 捨て札の一番上をタブローへ送る。空き山も埋められる。
func (c *Congress) MoveWasteToTableau(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCongressPile(pile); err != nil {
		return err
	}
	card := c.wasteTop()
	if card == nil {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errWasteEmpty", nil)
	}
	if !c.canPlaceOnTableau(card, pile) {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errCannotPlaceOnPile", nil)
	}
	c.takeSnapshot()
	c.popWaste()
	c.tableau[pile] = append(c.tableau[pile], card)
	c.afterMove("move", fmt.Sprintf("捨て札→タブロー山%d", pile), card)
	return nil
}

// MoveStockToTableau 山札の一番上で空き山を直接埋める。
//
// 規則上、空き山は「山札か捨て札の一番上」から埋められる。捨て札を経由させると
// 山札の 1 巡が 1 枚分早く減るので、山札から直接置く手を別に用意している。
func (c *Congress) MoveStockToTableau(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validCongressPile(pile); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "congress.errStockEmpty", nil)
	}
	if len(c.tableau[pile]) != 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errStockFillsGapsOnly", nil)
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	c.tableau[pile] = append(c.tableau[pile], card)
	c.afterMove("move", fmt.Sprintf("山札→タブロー山%d", pile), card)
	return nil
}

// GiveUp ギブアップ
func (c *Congress) GiveUp() {
	if c.phase == CongressPhasePlaying {
		c.phase = CongressPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (c *Congress) GetHint() *CongressHint {
	if c.phase != CongressPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	if h := c.tableauHint(); h != nil {
		return h
	}
	if len(c.stock) > 0 {
		return &CongressHint{FromZone: "stock", FromIdx: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (c *Congress) foundationHint() *CongressHint {
	if c.phase != CongressPhasePlaying {
		return nil
	}
	for pile := range CongressTableauCnt {
		card := c.tableauTop(pile)
		if card == nil {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &CongressHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	if card := c.wasteTop(); card != nil {
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &CongressHint{FromZone: "waste", FromIdx: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
func (c *Congress) tableauHint() *CongressHint {
	if c.phase != CongressPhasePlaying {
		return nil
	}
	if card := c.wasteTop(); card != nil {
		for pile := range CongressTableauCnt {
			if c.canPlaceOnTableau(card, pile) {
				return &CongressHint{FromZone: "waste", FromIdx: -1, ToZone: "tableau", ToIdx: pile}
			}
		}
	}
	// 空き山があれば山札から直接埋められる。
	if len(c.stock) > 0 {
		for pile := range CongressTableauCnt {
			if len(c.tableau[pile]) == 0 {
				return &CongressHint{FromZone: "stock", FromIdx: -1, ToZone: "tableau", ToIdx: pile}
			}
		}
	}
	for from := range CongressTableauCnt {
		card := c.tableauTop(from)
		if card == nil {
			continue
		}
		for to := range CongressTableauCnt {
			// 空き山はタブローからは埋められないので候補にしない。
			if from == to || len(c.tableau[to]) == 0 {
				continue
			}
			if c.canPlaceOnTableau(card, to) {
				return &CongressHint{FromZone: "tableau", FromIdx: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *Congress) AutoComplete() error {
	if c.phase != CongressPhasePlaying {
		return NewDomainErrorCode(ErrWrongPhase, "congress.errNotPlaying", nil)
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
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errNothingToAutoComplete", nil)
	}
	return nil
}

// Undo 直前の 1 手を取り消す
func (c *Congress) Undo() error {
	if len(c.history) == 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errNothingToUndo", nil)
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
func (c *Congress) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *Congress) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *Congress) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *congressSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *Congress) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *Congress) GetPhase() CongressPhase { return c.phase }

// GetMoveCount 手数取得
func (c *Congress) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *Congress) GetStockCount() int { return len(c.stock) }

// GetWaste 捨て札を取得
func (c *Congress) GetWaste() []*Card { return c.waste }

// GetTableau タブローを取得
func (c *Congress) GetTableau() [CongressTableauCnt][]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *Congress) GetFoundation() [CongressFoundationCnt][]*Card { return c.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (c *Congress) GetGameEndFlag() bool { return c.phase != CongressPhasePlaying }

// IsStalemate 手詰まりか
func (c *Congress) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *Congress) requirePlaying() error {
	if c.phase != CongressPhasePlaying {
		return NewDomainErrorCode(ErrWrongPhase, "congress.errNotPlaying", nil)
	}
	return nil
}

// validCongressPile タブロー山のインデックスを検証する
func validCongressPile(pile int) error {
	if pile < 0 || pile >= CongressTableauCnt {
		return NewDomainErrorCode(ErrInvalidPlay, "congress.errInvalidPile", map[string]string{"pile": strconv.Itoa(pile)})
	}
	return nil
}

// tableauTop 山の一番上（空なら nil）
func (c *Congress) tableauTop(pile int) *Card {
	if len(c.tableau[pile]) == 0 {
		return nil
	}
	return c.tableau[pile][len(c.tableau[pile])-1]
}

// wasteTop 捨て札の一番上（空なら nil）
func (c *Congress) wasteTop() *Card {
	return discardTop(c.waste)
}

// popWaste 捨て札の一番上を取り除く
func (c *Congress) popWaste() {
	c.waste = dropLast(c.waste)
}

// canPlaceOnTableau タブローに置けるか（スート無視の降順、A の下には何も置けない）。
// 空き山は山札と捨て札からのみ埋められるので、呼び出し側で出所を見る。
func (c *Congress) canPlaceOnTableau(card *Card, pile int) bool {
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

// canPlaceOnFoundation 基礎札に置けるか（空なら A、以降は同スートで 1 つ上）
func (c *Congress) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if congressSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	pile := c.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	if len(pile) >= CongressFoundationTarget {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()+1
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）
func (c *Congress) findFoundation(card *Card) int {
	for i := range CongressFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *Congress) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて K まで積まれたか
func (c *Congress) checkGameClear() {
	for i := range CongressFoundationCnt {
		if len(c.foundation[i]) != CongressFoundationTarget {
			return
		}
	}
	c.phase = CongressPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (c *Congress) checkStalemate() {
	if c.phase != CongressPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *Congress) takeSnapshot() {
	snap := &congressSnapshot{
		stock:       append([]*Card(nil), c.stock...),
		waste:       append([]*Card(nil), c.waste...),
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range CongressTableauCnt {
		snap.tableau[i] = append([]*Card(nil), c.tableau[i]...)
	}
	for i := range CongressFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = append(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *Congress) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// congressSnapshotJSON is the wire format for a single undo snapshot.
// congressSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type congressSnapshotJSON struct {
	Tableau     [CongressTableauCnt][]*Card    `json:"tb"`
	Foundation  [CongressFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       CongressPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for congressSnapshot.
func (s *congressSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(congressSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		Waste:       s.waste,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for congressSnapshot.
func (s *congressSnapshot) UnmarshalJSON(data []byte) error {
	var j congressSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > congressMaxSliceLen || len(j.Waste) > congressMaxSliceLen {
		return errors.New("congress: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > congressMaxSliceLen {
			return errors.New("congress: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > congressMaxSliceLen {
			return errors.New("congress: snapshot pile exceeds maximum allowed size")
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

// congressJSON is the JSON wire format for Congress.
type congressJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [CongressTableauCnt][]*Card    `json:"tb"`
	Foundation  [CongressFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	Waste       []*Card                        `json:"ws"`
	Phase       CongressPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*congressSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *Congress) MarshalJSON() ([]byte, error) {
	return json.Marshal(&congressJSON{
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
func (c *Congress) UnmarshalJSON(data []byte) error {
	var j congressJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < CongressPhasePlaying || j.Phase > CongressPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > CongressTotalCards || len(j.Waste) > CongressTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	if len(j.ActionLog) > congressMaxSliceLen || len(j.History) > congressMaxSliceLen {
		return errors.New("congress: input array exceeds maximum allowed size")
	}
	for i := range CongressFoundationCnt {
		if len(j.Foundation[i]) > CongressFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range CongressTableauCnt {
		if len(j.Tableau[i]) > CongressTotalCards {
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
