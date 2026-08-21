//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// SalicLawPhase サリカ法典のゲームフェーズ
type SalicLawPhase int

// SalicLawのフェーズ定数
const (
	// SalicLawPhasePlaying プレイ中
	SalicLawPhasePlaying SalicLawPhase = iota
	// SalicLawPhaseGameClear ゲームクリア
	SalicLawPhaseGameClear
	// SalicLawPhaseGameOver ゲームオーバー
	SalicLawPhaseGameOver
)

// SalicLawTableauCnt タブローの列数。K 1 枚ずつが土台になるので、2 組ぶんの K と
// 同じ 8 列で固定される。
const SalicLawTableauCnt = 8

// SalicLawFoundationCnt 基礎札の数。K 列と 1 対 1 で対応する。
const SalicLawFoundationCnt = 8

// SalicLawFoundationTarget 基礎札 1 つあたりの完成枚数（A→J の 11 枚）。
// Q は場に無く K は土台なので、12 枚目・13 枚目は存在しない。
const SalicLawFoundationTarget = CardValueMax - 2

// SalicLawQueenValue クイーンのランク
const SalicLawQueenValue = CardValueMax - 1

// SalicLawQueenCnt 場から抜くクイーンの枚数（2 組 × 4 スート）
const SalicLawQueenCnt = 8

// SalicLawTotalCards 実際に場に出る枚数（104 枚から Q 8 枚を抜いた 96 枚）
const SalicLawTotalCards = CardCnt*2 - SalicLawQueenCnt

// salicLawMaxSliceLen caps slice sizes during deserialisation.
const salicLawMaxSliceLen = 1000

// SalicLawHint サリカ法典のヒント。
//
// 捨て札は無いので waste ゾーンは存在しない。**FromZone が "stock" のときは
// 移動ではなく「配れ」**で、ToZone も "stock"、両方のインデックスが -1 になる。
// 移動の体裁（A → B）に落とすと、存在しない列 -1 が画面に漏れる。
type SalicLawHint struct {
	// FromZone 移動元 "tableau"、または配りを勧める "stock"
	FromZone string
	// FromIdx 移動元のタブロー列（配りのときは -1）
	FromIdx int
	// ToZone 移動先 "foundation" / "tableau"、または配りを勧める "stock"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の列を指さない）
	ToIdx int
}

// SalicLaw サリカ法典 ゲームクラス。
//
// 52 枚 2 組から **Q 8 枚を抜いた 96 枚**で遊ぶ 1 人用ソリティア。名前は
// 女子の王位継承を認めないサリカ法典に由来し、抜いた Q は飾りとして脇に置く。
//
// **配りが独特。**まず K を 1 枚据えてそれを列の土台とし、以降 1 枚ずつめくって
// 今の列に積む。**K が出たらそれが次の列の土台**になり、96 枚を配り切ると 8 列に
// なる。配っている途中でも各列の一番上は動かせる。捨て札は無い。
//
// **組札は 8 つで、スートを見ない。**A を置いて J まで 1 つずつ上げる（Q は場に
// 無く K は土台なので 11 枚で完成）。組札 i は**列 i が開いてから**使える。
//
// **タブロー同士は積めない。**唯一の例外が「K だけの列」で、そこには任意の 1 枚を
// 置ける。土台の K そのものは動かせない。
//
// 8 × 11 = 88 枚すべてを組札へ積み切ればクリア。K 8 枚は盤に残る。
//
// issue #5447 の仕様案とは異なり、実際の規則に合わせた。issue は「Q を K 上部の
// 専用スロットへ移す」「組札は同スート昇順」「残り 96 枚をタブロー数列に配る」と
// するが、Q は**移すのではなく最初から場に出ない**し、組札は**スート不問**で、
// 列は**配りながら K で区切られて**できる（PySolFC / Wikipedia の規則）。
type SalicLaw struct {
	trumpCards *TrumpCards
	tableau    [SalicLawTableauCnt][]*Card
	foundation [SalicLawFoundationCnt][]*Card
	stock      []*Card
	// queens 場から抜いた Q 8 枚。飾りとして表示するだけで、動かせない。
	queens []*Card
	// openPiles 土台の K が据わって使えるようになった列の数。配りが進むと増える。
	openPiles int
	phase     SalicLawPhase
	moveCount int
	actionLogBase
	history     []*salicLawSnapshot
	isStalemate bool
}

// salicLawSnapshot アンドゥ用スナップショット
type salicLawSnapshot struct {
	tableau     [SalicLawTableauCnt][]*Card
	foundation  [SalicLawFoundationCnt][]*Card
	stock       []*Card
	openPiles   int
	phase       SalicLawPhase
	moveCount   int
	isStalemate bool
}

// NewSalicLaw コンストラクタ
func NewSalicLaw(trumpCards *TrumpCards) *SalicLaw {
	return &SalicLaw{trumpCards: trumpCards}
}

// NewDefaultSalicLaw returns SalicLaw with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultSalicLaw() *SalicLaw {
	return NewSalicLaw(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (c *SalicLaw) Reset() {
	c.trumpCards.Shuffle()
	c.phase = SalicLawPhasePlaying
	c.moveCount = 0
	c.actionLog = nil
	c.history = nil
	c.isStalemate = false
	c.stock = nil
	c.queens = nil
	c.openPiles = 0

	for i := range SalicLawFoundationCnt {
		c.foundation[i] = nil
	}
	for i := range SalicLawTableauCnt {
		c.tableau[i] = nil
	}

	// **Q は場に出さない。**ゲーム名の由来そのもので、抜いた 8 枚は飾りとして
	// 脇に置く。ここで抜かないと 104 枚のままになり、組札 8 × 11 = 88 枚と
	// K 8 枚という勘定が合わなくなる。
	var rest []*Card
	for {
		card := c.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if card.GetValue() == SalicLawQueenValue {
			c.queens = append(c.queens, card)
			continue
		}
		rest = append(rest, card)
	}

	// **最初の 1 列は K を据えて始める。**シャッフル順の先頭が K とは限らないので、
	// 最初に見つかった K を土台に回し、残りはその順のまま山札にする。
	for i, card := range rest {
		if card.GetValue() == CardValueMax {
			c.tableau[0] = []*Card{card}
			c.openPiles = 1
			c.stock = append(append([]*Card(nil), rest[:i]...), rest[i+1:]...)
			break
		}
	}

	c.checkStalemate()
}

// Draw 山札から 1 枚めくって場に置く。捨て札は無く、めくった札はそのまま今の列の
// 一番上に乗る（＝すぐ動かせる）。K なら次の列の土台になる。
func (c *SalicLaw) Draw() error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if len(c.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "saliclaw.errStockEmptyNoRedeal", nil)
	}
	c.takeSnapshot()
	card := c.stock[0]
	c.stock = c.stock[1:]
	if card.GetValue() == CardValueMax && c.openPiles < SalicLawTableauCnt {
		c.tableau[c.openPiles] = []*Card{card}
		c.openPiles++
		c.afterMove("draw", fmt.Sprintf("Kで列%dを開いた", c.openPiles-1), card)
		return nil
	}
	c.tableau[c.openPiles-1] = append(c.tableau[c.openPiles-1], card)
	c.afterMove("draw", fmt.Sprintf("列%dに1枚配った", c.openPiles-1), card)
	return nil
}

// MoveTableauToFoundation タブロー列の一番上を基礎札へ送る
func (c *SalicLaw) MoveTableauToFoundation(pile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validSalicLawPile(pile); err != nil {
		return err
	}
	card := c.tableauTop(pile)
	if card == nil {
		return salicLawPileEmpty(pile)
	}
	// **土台の K は動かせない。**K は列の底に据わったまま最後まで残る。
	if len(c.tableau[pile]) == 1 {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errKingIsTheBase", nil)
	}
	fIdx := c.findFoundation(card)
	if fIdx < 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errNoFoundationForCard", nil)
	}
	c.takeSnapshot()
	c.tableau[pile] = c.tableau[pile][:len(c.tableau[pile])-1]
	c.foundation[fIdx] = append(c.foundation[fIdx], card)
	c.afterMove("move", fmt.Sprintf("列%d→基礎札%d", pile, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー列の一番上を「K だけの列」へ移す。
// それ以外の積み方は無い。
func (c *SalicLaw) MoveTableauToTableau(fromPile, toPile int) error {
	if err := c.requirePlaying(); err != nil {
		return err
	}
	if err := validSalicLawPile(fromPile); err != nil {
		return err
	}
	if err := validSalicLawPile(toPile); err != nil {
		return err
	}
	if fromPile == toPile {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errSamePile", nil)
	}
	card := c.tableauTop(fromPile)
	if card == nil {
		return salicLawPileEmpty(fromPile)
	}
	if len(c.tableau[fromPile]) == 1 {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errKingIsTheBase", nil)
	}
	if !c.canPlaceOnTableau(card, toPile) {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errCannotPlaceOnPile", nil)
	}
	c.takeSnapshot()
	c.tableau[fromPile] = c.tableau[fromPile][:len(c.tableau[fromPile])-1]
	c.tableau[toPile] = append(c.tableau[toPile], card)
	c.afterMove("move", fmt.Sprintf("列%d→列%d", fromPile, toPile), card)
	return nil
}

// GiveUp ギブアップ
func (c *SalicLaw) GiveUp() {
	if c.phase == SalicLawPhasePlaying {
		c.phase = SalicLawPhaseGameOver
		c.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (c *SalicLaw) GetHint() *SalicLawHint {
	if c.phase != SalicLawPhasePlaying {
		return nil
	}
	if h := c.foundationHint(); h != nil {
		return h
	}
	if h := c.tableauHint(); h != nil {
		return h
	}
	// 捨て札は無いので、山札のヒントは「配れ」の意味。移動の体裁に落とすと
	// 行き先の無い -1 が漏れるので、ToZone も stock のままにする。
	if len(c.stock) > 0 {
		return &SalicLawHint{FromZone: "stock", FromIdx: -1, ToZone: "stock", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
//
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (c *SalicLaw) foundationHint() *SalicLawHint {
	if c.phase != SalicLawPhasePlaying {
		return nil
	}
	for pile := range SalicLawTableauCnt {
		card := c.tableauTop(pile)
		// 土台の K は上げられないので、K 1 枚だけの列は候補にしない。
		if card == nil || len(c.tableau[pile]) == 1 {
			continue
		}
		if fIdx := c.findFoundation(card); fIdx >= 0 {
			return &SalicLawHint{FromZone: "tableau", FromIdx: pile, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。置き先は「K だけの列」しかない。
func (c *SalicLaw) tableauHint() *SalicLawHint {
	if c.phase != SalicLawPhasePlaying {
		return nil
	}
	for from := range SalicLawTableauCnt {
		card := c.tableauTop(from)
		// 土台の K しか無い列は、動かす側にはならない。
		if card == nil || len(c.tableau[from]) == 1 {
			continue
		}
		for to := range SalicLawTableauCnt {
			if from == to {
				continue
			}
			if c.canPlaceOnTableau(card, to) {
				return &SalicLawHint{FromZone: "tableau", FromIdx: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (c *SalicLaw) AutoComplete() error {
	if c.phase != SalicLawPhasePlaying {
		return NewDomainErrorCode(ErrWrongPhase, "saliclaw.errNotPlaying", nil)
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
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errNothingToAutoComplete", nil)
	}
	return nil
}

// Undo 直前の 1 手を取り消す
func (c *SalicLaw) Undo() error {
	if len(c.history) == 0 {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errNothingToUndo", nil)
	}
	snap := c.history[len(c.history)-1]
	c.history = c.history[:len(c.history)-1]
	c.tableau = snap.tableau
	c.foundation = snap.foundation
	c.stock = snap.stock
	c.openPiles = snap.openPiles
	c.phase = snap.phase
	c.moveCount = snap.moveCount
	c.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (c *SalicLaw) CanUndo() bool { return len(c.history) > 0 }

// UndoN n 手戻す
func (c *SalicLaw) UndoN(n int) error {
	return undoNChecked(c, n, len(c.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (c *SalicLaw) UndoToEscape() int {
	return undoToEscape(c.isStalemate, c.history, func(s *salicLawSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (c *SalicLaw) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (c *SalicLaw) GetPhase() SalicLawPhase { return c.phase }

// GetMoveCount 手数取得
func (c *SalicLaw) GetMoveCount() int { return c.moveCount }

// GetStockCount 山札の残り枚数
func (c *SalicLaw) GetStockCount() int { return len(c.stock) }

// GetQueens 場から抜いた Q 8 枚。飾りとして表示するだけで、動かせない。
func (c *SalicLaw) GetQueens() []*Card { return c.queens }

// GetOpenPiles 土台の K が据わって使えるようになった列の数
func (c *SalicLaw) GetOpenPiles() int { return c.openPiles }

// GetTableau タブローを取得
func (c *SalicLaw) GetTableau() [SalicLawTableauCnt][]*Card { return c.tableau }

// GetFoundation 基礎札を取得
func (c *SalicLaw) GetFoundation() [SalicLawFoundationCnt][]*Card { return c.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (c *SalicLaw) GetGameEndFlag() bool { return c.phase != SalicLawPhasePlaying }

// IsStalemate 手詰まりか
func (c *SalicLaw) IsStalemate() bool { return c.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (c *SalicLaw) requirePlaying() error {
	if c.phase != SalicLawPhasePlaying {
		return NewDomainErrorCode(ErrWrongPhase, "saliclaw.errNotPlaying", nil)
	}
	return nil
}

// salicLawPileEmpty 空の列を指したエラー。**列番号を必ず載せる。**文言が
// {{pile}} を持つので、nil を渡すと画面にプレースホルダがそのまま出る。
func salicLawPileEmpty(pile int) error {
	return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errPileEmpty",
		map[string]string{"pile": strconv.Itoa(pile)})
}

// validSalicLawPile タブロー列のインデックスを検証する
func validSalicLawPile(pile int) error {
	if pile < 0 || pile >= SalicLawTableauCnt {
		return NewDomainErrorCode(ErrInvalidPlay, "saliclaw.errInvalidPile", map[string]string{"pile": strconv.Itoa(pile)})
	}
	return nil
}

// tableauTop 山の一番上（空なら nil）
func (c *SalicLaw) tableauTop(pile int) *Card {
	if len(c.tableau[pile]) == 0 {
		return nil
	}
	return c.tableau[pile][len(c.tableau[pile])-1]
}

// canPlaceOnTableau タブローに置けるか。
//
// **置けるのは「K だけの列」だけ。**タブロー同士の積み上げはこのゲームには無く、
// 土台の K が剥き出しの列が唯一の空きとして働く。まだ開いていない列（K が据わって
// いない列）は空スライスなので、そこには置けない。
func (c *SalicLaw) canPlaceOnTableau(card *Card, pile int) bool {
	if card == nil {
		return false
	}
	return len(c.tableau[pile]) == 1
}

// canPlaceOnFoundation 基礎札に置けるか。
//
// **スートを見ない。**空なら A、以降はランクが 1 つ上ならどのスートでもよい。
// J で止まる（Q は場に無く、K は土台）。組札 i は**列 i が開いてから**使える
// ── 実物では組札が K 列の上に載るので、列が無ければ置き場も無い。
func (c *SalicLaw) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	if fIdx < 0 || fIdx >= SalicLawFoundationCnt {
		return false
	}
	if fIdx >= c.openPiles {
		return false
	}
	pile := c.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	if len(pile) >= SalicLawFoundationTarget {
		return false
	}
	return card.GetValue() == pile[len(pile)-1].GetValue()+1
}

// findFoundation 置ける基礎札を探す（見つからなければ -1）
func (c *SalicLaw) findFoundation(card *Card) int {
	for i := range SalicLawFoundationCnt {
		if c.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (c *SalicLaw) afterMove(actionType, detail string, card *Card) {
	afterMove(&c.moveCount, c, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて J まで積まれたか。K 8 枚は土台に残るので、
// 盤が空になることはない。
func (c *SalicLaw) checkGameClear() {
	for i := range SalicLawFoundationCnt {
		if len(c.foundation[i]) != SalicLawFoundationTarget {
			return
		}
	}
	c.phase = SalicLawPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。
func (c *SalicLaw) checkStalemate() {
	if c.phase != SalicLawPhasePlaying {
		return
	}
	c.isStalemate = c.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (c *SalicLaw) takeSnapshot() {
	snap := &salicLawSnapshot{
		stock:       append([]*Card(nil), c.stock...),
		openPiles:   c.openPiles,
		phase:       c.phase,
		moveCount:   c.moveCount,
		isStalemate: c.isStalemate,
	}
	for i := range SalicLawTableauCnt {
		snap.tableau[i] = append([]*Card(nil), c.tableau[i]...)
	}
	for i := range SalicLawFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), c.foundation[i]...)
	}
	c.history = append(c.history, snap)
}

// appendLog 棋譜エントリを追加
func (c *SalicLaw) appendLog(actionType, detail string, cards []*Card) {
	c.appendLogAt(c.moveCount, 0, actionType, detail, cards)
}

// salicLawSnapshotJSON is the wire format for a single undo snapshot.
// salicLawSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type salicLawSnapshotJSON struct {
	Tableau     [SalicLawTableauCnt][]*Card    `json:"tb"`
	Foundation  [SalicLawFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	OpenPiles   int                            `json:"op"`
	Phase       SalicLawPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	IsStalemate bool                           `json:"sl"`
}

// MarshalJSON implements json.Marshaler for salicLawSnapshot.
func (s *salicLawSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(salicLawSnapshotJSON{
		Tableau:     s.tableau,
		Foundation:  s.foundation,
		Stock:       s.stock,
		OpenPiles:   s.openPiles,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for salicLawSnapshot.
func (s *salicLawSnapshot) UnmarshalJSON(data []byte) error {
	var j salicLawSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > salicLawMaxSliceLen {
		return errors.New("saliclaw: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > salicLawMaxSliceLen {
			return errors.New("saliclaw: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > salicLawMaxSliceLen {
			return errors.New("saliclaw: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.foundation = j.Foundation
	s.stock = j.Stock
	s.openPiles = j.OpenPiles
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// salicLawJSON is the JSON wire format for SalicLaw.
type salicLawJSON struct {
	TrumpCards  *TrumpCards                    `json:"tc"`
	Tableau     [SalicLawTableauCnt][]*Card    `json:"tb"`
	Foundation  [SalicLawFoundationCnt][]*Card `json:"fd"`
	Stock       []*Card                        `json:"st"`
	OpenPiles   int                            `json:"op"`
	Phase       SalicLawPhase                  `json:"ps"`
	MoveCount   int                            `json:"mc"`
	ActionLog   []*ActionLogEntry              `json:"al"`
	IsStalemate bool                           `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*salicLawSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (c *SalicLaw) MarshalJSON() ([]byte, error) {
	return json.Marshal(&salicLawJSON{
		TrumpCards:  c.trumpCards,
		Tableau:     c.tableau,
		Foundation:  c.foundation,
		Stock:       c.stock,
		OpenPiles:   c.openPiles,
		Phase:       c.phase,
		MoveCount:   c.moveCount,
		ActionLog:   c.actionLog,
		IsStalemate: c.isStalemate,
		History:     c.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (c *SalicLaw) UnmarshalJSON(data []byte) error {
	var j salicLawJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if j.Phase < SalicLawPhasePlaying || j.Phase > SalicLawPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > SalicLawTotalCards {
		return fmt.Errorf("stock too large: %d", len(j.Stock))
	}
	if j.OpenPiles < 0 || j.OpenPiles > SalicLawTableauCnt {
		return fmt.Errorf("invalid open pile count: %d", j.OpenPiles)
	}
	if len(j.ActionLog) > salicLawMaxSliceLen || len(j.History) > salicLawMaxSliceLen {
		return errors.New("saliclaw: input array exceeds maximum allowed size")
	}
	for i := range SalicLawFoundationCnt {
		if len(j.Foundation[i]) > SalicLawFoundationTarget {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range SalicLawTableauCnt {
		if len(j.Tableau[i]) > SalicLawTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		c.trumpCards = j.TrumpCards
	}
	c.tableau = j.Tableau
	c.foundation = j.Foundation
	c.stock = j.Stock
	c.openPiles = j.OpenPiles
	c.phase = j.Phase
	c.moveCount = j.MoveCount
	c.actionLog = j.ActionLog
	c.isStalemate = j.IsStalemate
	c.history = j.History
	return nil
}
