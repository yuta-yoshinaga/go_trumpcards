//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// DuchessPhase ダッチェスのゲームフェーズ
type DuchessPhase int

// Duchessのフェーズ定数
const (
	// DuchessPhasePlaying プレイ中
	DuchessPhasePlaying DuchessPhase = iota
	// DuchessPhaseGameClear ゲームクリア
	DuchessPhaseGameClear
	// DuchessPhaseGameOver ゲームオーバー
	DuchessPhaseGameOver
)

// DuchessReserveCnt リザーブ扇の数
const DuchessReserveCnt = 4

// DuchessReserveFanSize 扇 1 つあたりの枚数
const DuchessReserveFanSize = 3

// DuchessTableauCnt タブローの列数
const DuchessTableauCnt = 4

// DuchessFoundationCnt 基礎札の数（スートごとに 1 つ）
const DuchessFoundationCnt = 4

// duchessSuitOrder 基礎札インデックスとスートの対応。固定しておくと配り直しても
// UI の位置が動かない。
var duchessSuitOrder = [DuchessFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// DuchessSuitIndex スートに対応する基礎札インデックスを返す（不明なスートは -1）。
func DuchessSuitIndex(design int) int {
	for i, d := range duchessSuitOrder {
		if d == design {
			return i
		}
	}
	return -1
}

// DuchessTableauCard タブロー上のカード。全札が表向きだが、他のソリティアと
// 同じ形にしておくとプレゼンターを使い回せる。
type DuchessTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// DuchessHint ダッチェスのヒント
type DuchessHint struct {
	// FromZone 移動元 "reserve" / "waste" / "tableau" / "stock"
	FromZone string
	// FromIdx 移動元のインデックス（リザーブ扇またはタブロー列。それ以外は -1）
	FromIdx int
	// CardIndex 移動元の列内インデックス。連番グループの先頭を指す（-1 は非該当）。
	CardIndex int
	// ToZone 移動先 "foundation" / "tableau" / "waste"
	ToZone string
	// ToIdx 移動先のインデックス（-1 は特定の列を指さない）
	ToIdx int
}

// Duchess ダッチェス（別名グレンウッド）ゲームクラス。
//
// カンフィールド系。**4 つのリザーブ扇に 3 枚ずつ（計 12 枚）**、タブロー 4 列に
// 1 枚ずつ、残り 36 枚が山札。捨て札は 1 枚ずつめくり、**やり直しは無い**。
//
// 最大の特徴は基礎札の開始ランクをプレイヤーが選ぶこと。ゲーム開始時、4 つの扇の
// 一番上から 1 枚を選び、そのランクが 4 つの基礎札すべての開始ランクになる。以降は
// 同スートで昇順、K の次は A に折り返し、開始ランクの 1 つ手前まで一周させる。
//
// タブローは色違いの降順（同じく K→A に折り返す）。**リザーブが残っているうちは、
// 空いた列はリザーブからしか埋められない** — カンフィールド系の要で、これが無いと
// リザーブを消化する動機が消える。
//
// issue #4408 の仕様案は扇を「それぞれ 4 枚ずつ計 16 枚」としているが、ダッチェス
// （グレンウッド）は 3 枚ずつ計 12 枚。issue の数字は 52 枚の内訳としては破綻して
// いないものの実際の規則と違うため、実ゲームに合わせた。issue が触れていない
// 「空き列はリザーブ優先」も本来の規則なので実装している。
type Duchess struct {
	trumpCards *TrumpCards
	reserve    [DuchessReserveCnt][]*Card
	tableau    [DuchessTableauCnt][]*DuchessTableauCard
	foundation [DuchessFoundationCnt][]*Card
	stock      []*Card
	waste      []*Card
	baseRank   int
	phase      DuchessPhase
	moveCount  int
	actionLogBase
	history     []*duchessSnapshot
	isStalemate bool
}

// duchessSnapshot アンドゥ用スナップショット
type duchessSnapshot struct {
	reserve     [DuchessReserveCnt][]*Card
	tableau     [DuchessTableauCnt][]*DuchessTableauCard
	foundation  [DuchessFoundationCnt][]*Card
	stock       []*Card
	waste       []*Card
	baseRank    int
	phase       DuchessPhase
	moveCount   int
	isStalemate bool
}

// NewDuchess コンストラクタ
func NewDuchess(trumpCards *TrumpCards) *Duchess {
	return &Duchess{trumpCards: trumpCards}
}

// NewDefaultDuchess returns Duchess with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultDuchess() *Duchess {
	return NewDuchess(NewTrumpCards(0))
}

// Reset ゲームリセット
func (d *Duchess) Reset() {
	d.trumpCards.Shuffle()
	d.phase = DuchessPhasePlaying
	d.moveCount = 0
	d.actionLog = nil
	d.history = nil
	d.isStalemate = false
	d.baseRank = 0
	d.stock = nil
	d.waste = nil

	for i := range DuchessFoundationCnt {
		d.foundation[i] = nil
	}
	for i := range DuchessReserveCnt {
		d.reserve[i] = nil
		for range DuchessReserveFanSize {
			if card := d.trumpCards.DrawCard(); card != nil {
				d.reserve[i] = append(d.reserve[i], card)
			}
		}
	}
	for i := range DuchessTableauCnt {
		d.tableau[i] = nil
		if card := d.trumpCards.DrawCard(); card != nil {
			d.tableau[i] = append(d.tableau[i], &DuchessTableauCard{Card: card, FaceUp: true})
		}
	}
	for {
		card := d.trumpCards.DrawCard()
		if card == nil {
			break
		}
		d.stock = append(d.stock, card)
	}

	d.checkStalemate()
}

// IsAwaitingBaseRank 開始ランクがまだ選ばれていないか。
//
// フェーズを増やさず専用フラグにしているのは、他のソリティアが Playing/GameClear/
// GameOver の 3 値で揃っており、ここだけ 4 値にすると Web/フロントの phase 番号が
// 全ゲームでずれるため。
func (d *Duchess) IsAwaitingBaseRank() bool {
	return d.phase == DuchessPhasePlaying && d.baseRank == 0
}

// GetBaseRank 基礎札の開始ランク（未選択なら 0）
func (d *Duchess) GetBaseRank() int { return d.baseRank }

// ChooseBaseRank リザーブ扇 idx の一番上を最初の基礎札に据え、そのランクを
// 4 つの基礎札すべての開始ランクにする。
func (d *Duchess) ChooseBaseRank(fanIdx int) error {
	if err := d.requirePlaying(); err != nil {
		return err
	}
	if d.baseRank != 0 {
		return errors.New("the base rank is already chosen")
	}
	card, err := d.reserveTop(fanIdx)
	if err != nil {
		return err
	}
	fIdx := DuchessSuitIndex(card.GetDesign())
	if fIdx < 0 {
		return errors.New("card has no foundation")
	}
	d.takeSnapshot()
	d.popReserve(fanIdx)
	d.baseRank = card.GetValue()
	d.foundation[fIdx] = []*Card{card}
	d.afterMove("base", fmt.Sprintf("開始ランクを%dに決定（リザーブ%d）", d.baseRank, fanIdx), card)
	return nil
}

// Draw 山札からウェイストへ 1 枚めくる（やり直しなし、1 巡のみ）
func (d *Duchess) Draw() error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	if len(d.stock) == 0 {
		return errors.New("no cards in stock")
	}
	d.takeSnapshot()
	card := d.stock[0]
	d.stock = d.stock[1:]
	d.waste = append(d.waste, card)
	d.afterMove("draw", "山札→ウェイスト", card)
	return nil
}

// MoveReserveToFoundation リザーブ扇の一番上を基礎札へ移す
func (d *Duchess) MoveReserveToFoundation(fanIdx int) error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	card, err := d.reserveTop(fanIdx)
	if err != nil {
		return err
	}
	fIdx := d.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	d.takeSnapshot()
	d.popReserve(fanIdx)
	d.foundation[fIdx] = append(d.foundation[fIdx], card)
	d.afterMove("move", fmt.Sprintf("リザーブ%d→基礎札%d", fanIdx, fIdx), card)
	return nil
}

// MoveReserveToTableau リザーブ扇の一番上をタブローへ移す
func (d *Duchess) MoveReserveToTableau(fanIdx, col int) error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	card, err := d.reserveTop(fanIdx)
	if err != nil {
		return err
	}
	if col < 0 || col >= DuchessTableauCnt {
		return errors.New("invalid column")
	}
	if !d.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	d.takeSnapshot()
	d.popReserve(fanIdx)
	d.tableau[col] = append(d.tableau[col], &DuchessTableauCard{Card: card, FaceUp: true})
	d.afterMove("move", fmt.Sprintf("リザーブ%d→タブロー列%d", fanIdx, col), card)
	return nil
}

// MoveWasteToFoundation ウェイスト最上段を基礎札へ移す
func (d *Duchess) MoveWasteToFoundation() error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	if len(d.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := d.waste[len(d.waste)-1]
	fIdx := d.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	d.takeSnapshot()
	d.waste = d.waste[:len(d.waste)-1]
	d.foundation[fIdx] = append(d.foundation[fIdx], card)
	d.afterMove("move", fmt.Sprintf("ウェイスト→基礎札%d", fIdx), card)
	return nil
}

// MoveWasteToTableau ウェイスト最上段をタブローへ移す
func (d *Duchess) MoveWasteToTableau(col int) error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	if len(d.waste) == 0 {
		return errors.New("waste is empty")
	}
	if col < 0 || col >= DuchessTableauCnt {
		return errors.New("invalid column")
	}
	card := d.waste[len(d.waste)-1]
	if !d.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	if d.emptyColumnIsReservedForReserve(col) {
		return errors.New("an empty column must be filled from the reserve first")
	}
	d.takeSnapshot()
	d.waste = d.waste[:len(d.waste)-1]
	d.tableau[col] = append(d.tableau[col], &DuchessTableauCard{Card: card, FaceUp: true})
	d.afterMove("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), card)
	return nil
}

// MoveTableauToFoundation タブロー最上段を基礎札へ移す
func (d *Duchess) MoveTableauToFoundation(col int) error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	if col < 0 || col >= DuchessTableauCnt {
		return errors.New("invalid column")
	}
	pile := d.tableau[col]
	if len(pile) == 0 {
		return errors.New("tableau column is empty")
	}
	card := pile[len(pile)-1].Card
	fIdx := d.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	d.takeSnapshot()
	d.tableau[col] = pile[:len(pile)-1]
	d.foundation[fIdx] = append(d.foundation[fIdx], card)
	d.afterMove("move", fmt.Sprintf("タブロー列%d→基礎札%d", col, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間でカードを移す。cardIndex 以降が色違い降順の
// 連番であれば、その塊ごとまとめて動かせる。
func (d *Duchess) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if err := d.requireBaseChosen(); err != nil {
		return err
	}
	if fromCol < 0 || fromCol >= DuchessTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= DuchessTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := d.tableau[fromCol]
	// -1 は「最上段 1 枚」。BeleagueredCastle など既存のソリティアと同じ約束。
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	group := fromCards[cardIndex:]
	if !duchessIsRun(group) {
		return errors.New("cards do not form an alternating-colour descending run")
	}
	if !d.canPlaceOnTableau(group[0].Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	if d.emptyColumnIsReservedForReserve(toCol) {
		return errors.New("an empty column must be filled from the reserve first")
	}
	d.takeSnapshot()
	// group は fromCards の内部を指しているので、切り詰める前にコピーを取る。
	moved := append([]*DuchessTableauCard(nil), group...)
	d.tableau[fromCol] = fromCards[:cardIndex]
	d.tableau[toCol] = append(d.tableau[toCol], moved...)
	d.afterMove("move",
		fmt.Sprintf("タブロー列%d→タブロー列%d(%d枚)", fromCol, toCol, len(moved)),
		moved[0].Card)
	return nil
}

// GiveUp ギブアップ
func (d *Duchess) GiveUp() {
	if d.phase == DuchessPhasePlaying {
		d.phase = DuchessPhaseGameOver
		d.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (d *Duchess) GetHint() *DuchessHint {
	if d.phase != DuchessPhasePlaying {
		return nil
	}
	// 開始ランクが未選択なら、まずそれを選ぶ以外にできることはない。
	if d.baseRank == 0 {
		for fan := range DuchessReserveCnt {
			if d.reserveTopCard(fan) != nil {
				return &DuchessHint{FromZone: "reserve", FromIdx: fan, CardIndex: -1, ToZone: "foundation", ToIdx: -1}
			}
		}
		return nil
	}
	if h := d.foundationHint(); h != nil {
		return h
	}
	if h := d.tableauHint(); h != nil {
		return h
	}
	if len(d.stock) > 0 {
		return &DuchessHint{FromZone: "stock", FromIdx: -1, CardIndex: -1, ToZone: "waste", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (d *Duchess) foundationHint() *DuchessHint {
	if d.phase != DuchessPhasePlaying || d.baseRank == 0 {
		return nil
	}
	for fan := range DuchessReserveCnt {
		card := d.reserveTopCard(fan)
		if card != nil && d.findFoundation(card) >= 0 {
			return &DuchessHint{FromZone: "reserve", FromIdx: fan, CardIndex: -1, ToZone: "foundation", ToIdx: d.findFoundation(card)}
		}
	}
	if len(d.waste) > 0 {
		if fIdx := d.findFoundation(d.waste[len(d.waste)-1]); fIdx >= 0 {
			return &DuchessHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for col := range DuchessTableauCnt {
		card := d.tableauTop(col)
		if card == nil {
			continue
		}
		if fIdx := d.findFoundation(card); fIdx >= 0 {
			return &DuchessHint{
				FromZone: "tableau", FromIdx: col, CardIndex: len(d.tableau[col]) - 1,
				ToZone: "foundation", ToIdx: fIdx,
			}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
//
// リザーブを空き列へ入れる手を最優先する。リザーブが残っているうちは空き列を
// そこからしか埋められず、リザーブを削ることがこのゲームの主目的だから。
func (d *Duchess) tableauHint() *DuchessHint {
	if d.phase != DuchessPhasePlaying || d.baseRank == 0 {
		return nil
	}
	for fan := range DuchessReserveCnt {
		card := d.reserveTopCard(fan)
		if card == nil {
			continue
		}
		for col := range DuchessTableauCnt {
			if d.canPlaceOnTableau(card, col) {
				return &DuchessHint{FromZone: "reserve", FromIdx: fan, CardIndex: -1, ToZone: "tableau", ToIdx: col}
			}
		}
	}
	if len(d.waste) > 0 {
		card := d.waste[len(d.waste)-1]
		for col := range DuchessTableauCnt {
			if d.canPlaceOnTableau(card, col) && !d.emptyColumnIsReservedForReserve(col) {
				return &DuchessHint{FromZone: "waste", FromIdx: -1, CardIndex: -1, ToZone: "tableau", ToIdx: col}
			}
		}
	}
	for from := range DuchessTableauCnt {
		pile := d.tableau[from]
		for idx := range pile {
			if !duchessIsRun(pile[idx:]) {
				continue
			}
			for to := range DuchessTableauCnt {
				if to == from {
					continue
				}
				// 列ごと空き列へ動かしても盤面は進まない。
				if idx == 0 && len(d.tableau[to]) == 0 {
					continue
				}
				if d.canPlaceOnTableau(pile[idx].Card, to) && !d.emptyColumnIsReservedForReserve(to) {
					return &DuchessHint{
						FromZone: "tableau", FromIdx: from, CardIndex: idx,
						ToZone: "tableau", ToIdx: to,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (d *Duchess) AutoComplete() error {
	if d.phase != DuchessPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if d.baseRank == 0 {
		return errors.New("choose the base rank first")
	}
	moved := false
	for {
		h := d.foundationHint()
		if h == nil {
			break
		}
		var err error
		switch h.FromZone {
		case "reserve":
			err = d.MoveReserveToFoundation(h.FromIdx)
		case "waste":
			err = d.MoveWasteToFoundation()
		default:
			err = d.MoveTableauToFoundation(h.FromIdx)
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
func (d *Duchess) Undo() error {
	if len(d.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := d.history[len(d.history)-1]
	d.history = d.history[:len(d.history)-1]
	d.reserve = snap.reserve
	d.tableau = snap.tableau
	d.foundation = snap.foundation
	d.stock = snap.stock
	d.waste = snap.waste
	d.baseRank = snap.baseRank
	d.phase = snap.phase
	d.moveCount = snap.moveCount
	d.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (d *Duchess) CanUndo() bool { return len(d.history) > 0 }

// UndoN n 手戻す
func (d *Duchess) UndoN(n int) error {
	return undoNChecked(d, n, len(d.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (d *Duchess) UndoToEscape() int {
	return undoToEscape(d.isStalemate, d.history, func(s *duchessSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (d *Duchess) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (d *Duchess) GetPhase() DuchessPhase { return d.phase }

// GetMoveCount 手数取得
func (d *Duchess) GetMoveCount() int { return d.moveCount }

// GetStockCount 山札の残り枚数
func (d *Duchess) GetStockCount() int { return len(d.stock) }

// GetWaste ウェイストを取得
func (d *Duchess) GetWaste() []*Card { return d.waste }

// GetReserve リザーブ扇を取得
func (d *Duchess) GetReserve() [DuchessReserveCnt][]*Card { return d.reserve }

// GetTableau タブローを取得
func (d *Duchess) GetTableau() [DuchessTableauCnt][]*DuchessTableauCard { return d.tableau }

// GetFoundation 基礎札を取得
func (d *Duchess) GetFoundation() [DuchessFoundationCnt][]*Card { return d.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (d *Duchess) GetGameEndFlag() bool { return d.phase != DuchessPhasePlaying }

// IsStalemate 手詰まりか
func (d *Duchess) IsStalemate() bool { return d.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (d *Duchess) requirePlaying() error {
	if d.phase != DuchessPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// requireBaseChosen プレイ中かつ開始ランクが選ばれているか
func (d *Duchess) requireBaseChosen() error {
	if err := d.requirePlaying(); err != nil {
		return err
	}
	if d.baseRank == 0 {
		return errors.New("choose the base rank first")
	}
	return nil
}

// duchessIsRun 色違いで 1 つずつ下がる連番か（1 枚は常に真）。
// タブローも K→A に折り返す。
func duchessIsRun(cards []*DuchessTableauCard) bool {
	for i := 1; i < len(cards); i++ {
		prev, cur := cards[i-1].Card, cards[i].Card
		if prev == nil || cur == nil {
			return false
		}
		if cur.GetValue() != duchessPrevRank(prev.GetValue()) {
			return false
		}
		if duchessIsRed(prev.GetDesign()) == duchessIsRed(cur.GetDesign()) {
			return false
		}
	}
	return true
}

// duchessIsRed 赤いスートか
func duchessIsRed(design int) bool {
	return design == CardDesignHeart || design == CardDesignDiamond
}

// duchessNextRank K の次を A に折り返す昇順の次ランク
func duchessNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// duchessPrevRank A の前を K に折り返す降順の次ランク
func duchessPrevRank(v int) int {
	if v <= 1 {
		return CardValueMax
	}
	return v - 1
}

// reserveTop 指定扇の一番上を返す（範囲外・空ならエラー）
func (d *Duchess) reserveTop(fanIdx int) (*Card, error) {
	if fanIdx < 0 || fanIdx >= DuchessReserveCnt {
		return nil, errors.New("invalid reserve index")
	}
	if len(d.reserve[fanIdx]) == 0 {
		return nil, errors.New("reserve fan is empty")
	}
	return d.reserve[fanIdx][len(d.reserve[fanIdx])-1], nil
}

// reserveTopCard 指定扇の一番上を返す（空なら nil）。手の探索用。
func (d *Duchess) reserveTopCard(fanIdx int) *Card {
	if len(d.reserve[fanIdx]) == 0 {
		return nil
	}
	return d.reserve[fanIdx][len(d.reserve[fanIdx])-1]
}

// popReserve 指定扇の一番上を取り除く
func (d *Duchess) popReserve(fanIdx int) {
	d.reserve[fanIdx] = d.reserve[fanIdx][:len(d.reserve[fanIdx])-1]
}

// tableauTop 指定列の最上段を返す（空なら nil）
func (d *Duchess) tableauTop(col int) *Card {
	if len(d.tableau[col]) == 0 {
		return nil
	}
	return d.tableau[col][len(d.tableau[col])-1].Card
}

// reserveRemaining リザーブに残っている総枚数
func (d *Duchess) reserveRemaining() int {
	n := 0
	for i := range DuchessReserveCnt {
		n += len(d.reserve[i])
	}
	return n
}

// emptyColumnIsReservedForReserve 空き列がリザーブ専用か。
// カンフィールド系の要: リザーブが残っているうちは空き列をリザーブからしか
// 埋められない。これが無いとリザーブを消化する動機が消える。
func (d *Duchess) emptyColumnIsReservedForReserve(col int) bool {
	return len(d.tableau[col]) == 0 && d.reserveRemaining() > 0
}

// canPlaceOnTableau タブローに置けるか（空き列は任意、以降は色違いで 1 つ下）。
// リザーブ優先の制限は呼び出し側で見る — こちらは純粋な重ね方の判定。
func (d *Duchess) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := d.tableau[col]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1].Card
	if card.GetValue() != duchessPrevRank(top.GetValue()) {
		return false
	}
	return duchessIsRed(card.GetDesign()) != duchessIsRed(top.GetDesign())
}

// canPlaceOnFoundation 基礎札に置けるか。
// 空なら開始ランクかつそのスート、以降は同スートで 1 つ上（K の次は A）。
// 開始ランクの 1 つ手前まで積んだら完成で、それ以上は受け付けない。
func (d *Duchess) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil || d.baseRank == 0 {
		return false
	}
	if duchessSuitOrder[fIdx] != card.GetDesign() {
		return false
	}
	pile := d.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == d.baseRank
	}
	if len(pile) >= CardValueMax {
		return false
	}
	return card.GetValue() == duchessNextRank(pile[len(pile)-1].GetValue())
}

// findFoundation 置ける基礎札のインデックスを探す（見つからなければ -1）
func (d *Duchess) findFoundation(card *Card) int {
	for i := range DuchessFoundationCnt {
		if d.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (d *Duchess) afterMove(actionType, detail string, card *Card) {
	d.moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	d.appendLog(actionType, detail, cards)
	d.checkGameClear()
	d.checkStalemate()
}

// checkGameClear 4 つの基礎札がすべて 13 枚（開始ランクから一周）になったか
func (d *Duchess) checkGameClear() {
	for i := range DuchessFoundationCnt {
		if len(d.foundation[i]) != CardValueMax {
			return
		}
	}
	d.phase = DuchessPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は開始ランク選択・基礎札・タブロー・山札のすべてを見るので、
// 「ヒントが無い」と「手詰まり」は同じ条件になる。
func (d *Duchess) checkStalemate() {
	if d.phase != DuchessPhasePlaying {
		return
	}
	d.isStalemate = d.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (d *Duchess) takeSnapshot() {
	snap := &duchessSnapshot{
		phase:       d.phase,
		moveCount:   d.moveCount,
		isStalemate: d.isStalemate,
		baseRank:    d.baseRank,
		stock:       append([]*Card(nil), d.stock...),
		waste:       append([]*Card(nil), d.waste...),
	}
	for i := range DuchessReserveCnt {
		snap.reserve[i] = append([]*Card(nil), d.reserve[i]...)
	}
	for i := range DuchessFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), d.foundation[i]...)
	}
	for i := range DuchessTableauCnt {
		snap.tableau[i] = append([]*DuchessTableauCard(nil), d.tableau[i]...)
	}
	d.history = append(d.history, snap)
}

// appendLog 棋譜エントリを追加
func (d *Duchess) appendLog(actionType, detail string, cards []*Card) {
	d.appendLogAt(d.moveCount, 0, actionType, detail, cards)
}

// duchessMaxSliceLen caps slice sizes during deserialisation.
const duchessMaxSliceLen = 1000

// duchessSnapshotJSON is the wire format for a single undo snapshot.
// duchessSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type duchessSnapshotJSON struct {
	Reserve     [DuchessReserveCnt][]*Card               `json:"rs"`
	Tableau     [DuchessTableauCnt][]*DuchessTableauCard `json:"tb"`
	Foundation  [DuchessFoundationCnt][]*Card            `json:"fd"`
	Stock       []*Card                                  `json:"st"`
	Waste       []*Card                                  `json:"wa"`
	BaseRank    int                                      `json:"br"`
	Phase       DuchessPhase                             `json:"ps"`
	MoveCount   int                                      `json:"mc"`
	IsStalemate bool                                     `json:"sl"`
}

// MarshalJSON implements json.Marshaler for duchessSnapshot.
func (s *duchessSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(duchessSnapshotJSON{
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

// UnmarshalJSON implements json.Unmarshaler for duchessSnapshot.
func (s *duchessSnapshot) UnmarshalJSON(data []byte) error {
	var j duchessSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > duchessMaxSliceLen ||
		len(j.Waste) > duchessMaxSliceLen {
		return errors.New("duchess: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Reserve {
		if len(pile) > duchessMaxSliceLen {
			return errors.New("duchess: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Tableau {
		if len(pile) > duchessMaxSliceLen {
			return errors.New("duchess: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > duchessMaxSliceLen {
			return errors.New("duchess: snapshot pile exceeds maximum allowed size")
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

// duchessJSON is the JSON wire format for Duchess.
type duchessJSON struct {
	TrumpCards  *TrumpCards                              `json:"tc"`
	Reserve     [DuchessReserveCnt][]*Card               `json:"rs"`
	Tableau     [DuchessTableauCnt][]*DuchessTableauCard `json:"tb"`
	Foundation  [DuchessFoundationCnt][]*Card            `json:"fd"`
	Stock       []*Card                                  `json:"st"`
	Waste       []*Card                                  `json:"ws"`
	BaseRank    int                                      `json:"br"`
	Phase       DuchessPhase                             `json:"ps"`
	MoveCount   int                                      `json:"mc"`
	ActionLog   []*ActionLogEntry                        `json:"al"`
	IsStalemate bool                                     `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*duchessSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (d *Duchess) MarshalJSON() ([]byte, error) {
	return json.Marshal(&duchessJSON{
		TrumpCards:  d.trumpCards,
		Reserve:     d.reserve,
		Tableau:     d.tableau,
		Foundation:  d.foundation,
		Stock:       d.stock,
		Waste:       d.waste,
		BaseRank:    d.baseRank,
		Phase:       d.phase,
		MoveCount:   d.moveCount,
		ActionLog:   d.actionLog,
		IsStalemate: d.isStalemate,
		History:     d.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (d *Duchess) UnmarshalJSON(data []byte) error {
	var j duchessJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > duchessMaxSliceLen || len(j.History) > duchessMaxSliceLen {
		return errors.New("duchess: input array exceeds maximum allowed size")
	}
	if j.Phase < DuchessPhasePlaying || j.Phase > DuchessPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	// 0 は「未選択」なので許すが、それ以外は実在するランクでなければならない。
	if j.BaseRank < 0 || j.BaseRank > CardValueMax {
		return fmt.Errorf("invalid base rank: %d", j.BaseRank)
	}
	if len(j.Stock) > CardCnt || len(j.Waste) > CardCnt {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	for i := range DuchessReserveCnt {
		if len(j.Reserve[i]) > CardCnt {
			return fmt.Errorf("reserve %d holds %d cards", i, len(j.Reserve[i]))
		}
	}
	for i := range DuchessFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range DuchessTableauCnt {
		if len(j.Tableau[i]) > CardCnt {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		d.trumpCards = j.TrumpCards
	}
	d.reserve = j.Reserve
	d.tableau = j.Tableau
	d.foundation = j.Foundation
	d.stock = j.Stock
	d.waste = j.Waste
	d.baseRank = j.BaseRank
	d.phase = j.Phase
	d.moveCount = j.MoveCount
	d.actionLog = j.ActionLog
	d.isStalemate = j.IsStalemate
	d.history = j.History
	return nil
}
