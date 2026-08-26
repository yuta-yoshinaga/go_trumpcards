//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// NapoleonsSquarePhase ナポレオンズ・スクエアのゲームフェーズ
type NapoleonsSquarePhase int

// NapoleonsSquareのフェーズ定数
const (
	// NapoleonsSquarePhasePlaying プレイ中
	NapoleonsSquarePhasePlaying NapoleonsSquarePhase = iota
	// NapoleonsSquarePhaseGameClear ゲームクリア
	NapoleonsSquarePhaseGameClear
	// NapoleonsSquarePhaseGameOver ゲームオーバー
	NapoleonsSquarePhaseGameOver
)

// NapoleonsSquareTableauCnt タブローの列数。正方形の各辺に 3 列ずつ並べる。
const NapoleonsSquareTableauCnt = 12

// NapoleonsSquareColumnLen 配り時の 1 列あたりの枚数
const NapoleonsSquareColumnLen = 4

// NapoleonsSquareFoundationCnt 基礎札の数。2 デッキなのでスートごとに 2 つ。
const NapoleonsSquareFoundationCnt = 8

// napoleonsSquareSuitOrder 基礎札インデックスとスートの対応。0..3 が 1 組目、
// 4..7 が 2 組目。固定しておくと配り直しても UI の位置が動かない。
var napoleonsSquareSuitOrder = [NapoleonsSquareFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// NapoleonsSquareTableauCard タブロー上のカード。全札が表向きだが、他のソリティアと
// 同じ形にしておくとプレゼンターを使い回せる。
type NapoleonsSquareTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// NapoleonsSquareHint ナポレオンズ・スクエアのヒント
type NapoleonsSquareHint struct {
	// FromZone 移動元 "waste" / "tableau" / "stock"
	FromZone string
	// FromCol 移動元のタブロー列（FromZone が "tableau" のときのみ）
	FromCol int
	// CardIndex 移動元の列内インデックス。連番グループの先頭を指す。
	CardIndex int
	// ToZone 移動先 "foundation" / "tableau"
	ToZone string
	// ToCol 移動先のインデックス（基礎札またはタブロー列）
	ToCol int
}

// NapoleonsSquare ナポレオンズ・スクエアゲームクラス。
//
// 2 デッキ 104 枚の大型ソリティア。8 枚のエースを中央の基礎札として置き、残りを
// 12 列 4 枚のタブロー（正方形）と山札に分ける。基礎札は同スート昇順 (A→K)、
// タブローは同スート降順 (K→A) に積む。
//
// タブローでは「同スートで 1 ずつ下がる連番」をまとめて動かせる点が
// FortyThieves との最大の違いで、空き列には任意の札・任意のグループを置ける。
//
// issue #4391 は「8 枚のエースを置き、残り 96 枚を 12 列×4 枚のタブローに配る」と
// しているが 12×4 = 48 であって 96 ではない。列構成の方が実際のルールと一致する
// ため、48 枚をタブローに配り残り 48 枚を山札としている。
type NapoleonsSquare struct {
	trumpCards *TrumpCards
	tableau    [NapoleonsSquareTableauCnt][]*NapoleonsSquareTableauCard
	stock      []*Card
	waste      []*Card
	foundation [NapoleonsSquareFoundationCnt][]*Card
	phase      NapoleonsSquarePhase
	moveCount  int
	actionLogBase
	history     []*napoleonsSquareSnapshot
	isStalemate bool
}

// napoleonsSquareSnapshot アンドゥ用スナップショット
type napoleonsSquareSnapshot struct {
	tableau     [NapoleonsSquareTableauCnt][]*NapoleonsSquareTableauCard
	stock       []*Card
	waste       []*Card
	foundation  [NapoleonsSquareFoundationCnt][]*Card
	phase       NapoleonsSquarePhase
	moveCount   int
	isStalemate bool
}

// NewNapoleonsSquare コンストラクタ
func NewNapoleonsSquare(trumpCards *TrumpCards) *NapoleonsSquare {
	return &NapoleonsSquare{trumpCards: trumpCards}
}

// NewDefaultNapoleonsSquare returns NapoleonsSquare with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultNapoleonsSquare() *NapoleonsSquare {
	return NewNapoleonsSquare(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (ns *NapoleonsSquare) Reset() {
	ns.trumpCards.Shuffle()
	ns.phase = NapoleonsSquarePhasePlaying
	ns.moveCount = 0
	ns.actionLog = nil
	ns.history = nil
	ns.isStalemate = false
	ns.stock = nil
	ns.waste = nil

	for i := range NapoleonsSquareFoundationCnt {
		ns.foundation[i] = nil
	}
	for i := range NapoleonsSquareTableauCnt {
		ns.tableau[i] = nil
	}

	// エースを 8 枚抜いて基礎札に置き、残りを配る。スートごとに 2 つの基礎札が
	// あるので、同じスートの 2 枚目は 4..7 側に入る。
	remaining := make([]*Card, 0, CardCnt*2)
	for {
		card := ns.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if card.GetValue() == 1 {
			if fIdx := ns.emptyFoundationForSuit(card.GetDesign()); fIdx >= 0 {
				ns.foundation[fIdx] = []*Card{card}
				continue
			}
		}
		remaining = append(remaining, card)
	}

	for col := range NapoleonsSquareTableauCnt {
		for range NapoleonsSquareColumnLen {
			if len(remaining) == 0 {
				break
			}
			card := remaining[0]
			remaining = remaining[1:]
			ns.tableau[col] = append(ns.tableau[col],
				&NapoleonsSquareTableauCard{Card: card, FaceUp: true})
		}
	}
	ns.stock = remaining

	ns.checkStalemate()
}

// emptyFoundationForSuit 指定スートの空いている基礎札インデックスを返す（無ければ -1）
func (ns *NapoleonsSquare) emptyFoundationForSuit(design int) int {
	for i, d := range napoleonsSquareSuitOrder {
		if d == design && len(ns.foundation[i]) == 0 {
			return i
		}
	}
	return -1
}

// Draw 山札からウェイストへ 1 枚めくる（リサイクルなし、1 巡のみ）
func (ns *NapoleonsSquare) Draw() error {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(ns.stock) == 0 {
		return errors.New("no cards in stock")
	}
	ns.takeSnapshot()
	card := ns.stock[0]
	ns.stock = ns.stock[1:]
	ns.waste = append(ns.waste, card)
	ns.afterMove("draw", "山札→ウェイスト", card)
	return nil
}

// MoveWasteToTableau ウェイスト最上段をタブローへ移す
func (ns *NapoleonsSquare) MoveWasteToTableau(col int) error {
	if err := ns.requirePlaying(); err != nil {
		return err
	}
	if col < 0 || col >= NapoleonsSquareTableauCnt {
		return errors.New("invalid column")
	}
	if len(ns.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ns.waste[len(ns.waste)-1]
	if !ns.canPlaceOnTableau(card, col) {
		return errors.New("cannot place card on tableau")
	}
	ns.takeSnapshot()
	ns.waste = ns.waste[:len(ns.waste)-1]
	ns.tableau[col] = append(ns.tableau[col],
		&NapoleonsSquareTableauCard{Card: card, FaceUp: true})
	ns.afterMove("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), card)
	return nil
}

// MoveWasteToFoundation ウェイスト最上段を基礎札へ移す
func (ns *NapoleonsSquare) MoveWasteToFoundation() error {
	if err := ns.requirePlaying(); err != nil {
		return err
	}
	if len(ns.waste) == 0 {
		return errors.New("waste is empty")
	}
	card := ns.waste[len(ns.waste)-1]
	fIdx := ns.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ns.takeSnapshot()
	ns.waste = ns.waste[:len(ns.waste)-1]
	ns.foundation[fIdx] = append(ns.foundation[fIdx], card)
	ns.afterMove("move", fmt.Sprintf("ウェイスト→基礎札%d", fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー間でカードを移す。cardIndex 以降が同スート降順の
// 連番であれば、その塊ごとまとめて動かせる。
func (ns *NapoleonsSquare) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if err := ns.requirePlaying(); err != nil {
		return err
	}
	if fromCol < 0 || fromCol >= NapoleonsSquareTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= NapoleonsSquareTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := ns.tableau[fromCol]
	// -1 は「最上段 1 枚」を表す。BeleagueredCastle など既存のソリティアと同じ約束で、
	// 呼び出し側が列の長さを知らなくても最上段を動かせる。
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	group := fromCards[cardIndex:]
	if !napoleonsSquareIsRun(group) {
		return errors.New("cards do not form a same-suit descending run")
	}
	if !ns.canPlaceOnTableau(group[0].Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	ns.takeSnapshot()
	// group は fromCards の内部を指しているので、切り詰める前にコピーを取る。
	moved := append([]*NapoleonsSquareTableauCard(nil), group...)
	ns.tableau[fromCol] = fromCards[:cardIndex]
	ns.tableau[toCol] = append(ns.tableau[toCol], moved...)
	ns.afterMove("move",
		fmt.Sprintf("タブロー列%d→タブロー列%d(%d枚)", fromCol, toCol, len(moved)),
		moved[0].Card)
	return nil
}

// MoveTableauToFoundation タブロー最上段を基礎札へ移す
func (ns *NapoleonsSquare) MoveTableauToFoundation(col int) error {
	if err := ns.requirePlaying(); err != nil {
		return err
	}
	if col < 0 || col >= NapoleonsSquareTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := ns.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1].Card
	fIdx := ns.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	ns.takeSnapshot()
	ns.tableau[col] = fromCards[:len(fromCards)-1]
	ns.foundation[fIdx] = append(ns.foundation[fIdx], card)
	ns.afterMove("move", fmt.Sprintf("タブロー列%d→基礎札%d", col, fIdx), card)
	return nil
}

// GiveUp ギブアップ
func (ns *NapoleonsSquare) GiveUp() {
	if ns.phase == NapoleonsSquarePhasePlaying {
		ns.phase = NapoleonsSquarePhaseGameOver
		ns.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札へ送れる手を優先し、無ければタブローの手、
// それも無ければ山札をめくる手を返す。手詰まり判定もこの関数を使う。
func (ns *NapoleonsSquare) GetHint() *NapoleonsSquareHint {
	if h := ns.foundationHint(); h != nil {
		return h
	}
	if h := ns.tableauHint(); h != nil {
		return h
	}
	if ns.phase == NapoleonsSquarePhasePlaying && len(ns.stock) > 0 {
		return &NapoleonsSquareHint{FromZone: "stock", FromCol: -1, CardIndex: -1, ToZone: "waste", ToCol: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (ns *NapoleonsSquare) foundationHint() *NapoleonsSquareHint {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return nil
	}
	if len(ns.waste) > 0 {
		if fIdx := ns.findFoundation(ns.waste[len(ns.waste)-1]); fIdx >= 0 {
			return &NapoleonsSquareHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	for col := range NapoleonsSquareTableauCnt {
		pile := ns.tableau[col]
		if len(pile) == 0 {
			continue
		}
		if fIdx := ns.findFoundation(pile[len(pile)-1].Card); fIdx >= 0 {
			return &NapoleonsSquareHint{
				FromZone: "tableau", FromCol: col, CardIndex: len(pile) - 1,
				ToZone: "foundation", ToCol: fIdx,
			}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
//
// 空き列への移動は「同じ列から出して同じ列に戻す」堂々巡りになりやすいので、
// 提案するのは実際に盤面が進む手だけに絞る（空き列へは 1 列まるごとの移動を
// 提案しない）。
func (ns *NapoleonsSquare) tableauHint() *NapoleonsSquareHint {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return nil
	}
	if len(ns.waste) > 0 {
		for col := range NapoleonsSquareTableauCnt {
			if ns.canPlaceOnTableau(ns.waste[len(ns.waste)-1], col) {
				return &NapoleonsSquareHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: col}
			}
		}
	}
	for from := range NapoleonsSquareTableauCnt {
		pile := ns.tableau[from]
		for idx := range pile {
			if !napoleonsSquareIsRun(pile[idx:]) {
				continue
			}
			for to := range NapoleonsSquareTableauCnt {
				if to == from {
					continue
				}
				// 列ごと空き列へ動かしても盤面は進まない。
				if idx == 0 && len(ns.tableau[to]) == 0 {
					continue
				}
				if ns.canPlaceOnTableau(pile[idx].Card, to) {
					return &NapoleonsSquareHint{
						FromZone: "tableau", FromCol: from, CardIndex: idx,
						ToZone: "tableau", ToCol: to,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (ns *NapoleonsSquare) AutoComplete() error {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := ns.foundationHint()
		if h == nil {
			break
		}
		var err error
		if h.FromZone == "waste" {
			err = ns.MoveWasteToFoundation()
		} else {
			err = ns.MoveTableauToFoundation(h.FromCol)
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
func (ns *NapoleonsSquare) Undo() error {
	if len(ns.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := ns.history[len(ns.history)-1]
	ns.history = ns.history[:len(ns.history)-1]
	ns.tableau = snap.tableau
	ns.stock = snap.stock
	ns.waste = snap.waste
	ns.foundation = snap.foundation
	ns.phase = snap.phase
	ns.moveCount = snap.moveCount
	ns.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (ns *NapoleonsSquare) CanUndo() bool { return len(ns.history) > 0 }

// UndoN n 手戻す
func (ns *NapoleonsSquare) UndoN(n int) error {
	return undoNChecked(ns, n, len(ns.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (ns *NapoleonsSquare) UndoToEscape() int {
	return undoToEscape(ns.isStalemate, ns.history, func(s *napoleonsSquareSnapshot) bool { return s.isStalemate })
}

// AllFaceUp ナポレオンズ・スクエアは常に全札が表向き
func (ns *NapoleonsSquare) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (ns *NapoleonsSquare) GetPhase() NapoleonsSquarePhase { return ns.phase }

// GetMoveCount 手数取得
func (ns *NapoleonsSquare) GetMoveCount() int { return ns.moveCount }

// GetStockCount 山札の残り枚数
func (ns *NapoleonsSquare) GetStockCount() int { return len(ns.stock) }

// GetWaste ウェイストを取得
func (ns *NapoleonsSquare) GetWaste() []*Card { return ns.waste }

// GetTableau タブローを取得
func (ns *NapoleonsSquare) GetTableau() [NapoleonsSquareTableauCnt][]*NapoleonsSquareTableauCard {
	return ns.tableau
}

// GetFoundation 基礎札を取得
func (ns *NapoleonsSquare) GetFoundation() [NapoleonsSquareFoundationCnt][]*Card {
	return ns.foundation
}

// GetGameEndFlag ゲーム終了フラグ
func (ns *NapoleonsSquare) GetGameEndFlag() bool { return ns.phase != NapoleonsSquarePhasePlaying }

// IsStalemate 手詰まりか
func (ns *NapoleonsSquare) IsStalemate() bool { return ns.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (ns *NapoleonsSquare) requirePlaying() error {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// napoleonsSquareIsRun 同スートで 1 ずつ下がる連番か（1 枚は常に真）
func napoleonsSquareIsRun(cards []*NapoleonsSquareTableauCard) bool {
	for i := 1; i < len(cards); i++ {
		prev, cur := cards[i-1].Card, cards[i].Card
		if prev == nil || cur == nil {
			return false
		}
		if prev.GetDesign() != cur.GetDesign() || cur.GetValue() != prev.GetValue()-1 {
			return false
		}
	}
	return true
}

// canPlaceOnTableau タブローに置けるか（空き列は任意、以降は同スートで 1 つ下）
func (ns *NapoleonsSquare) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := ns.tableau[col]
	if len(pile) == 0 {
		return true
	}
	top := pile[len(pile)-1].Card
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()-1
}

// canPlaceOnFoundation 基礎札に置けるか（空はエースのみ、以降は同スートで 1 つ上）
func (ns *NapoleonsSquare) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := ns.foundation[fIdx]
	if len(pile) == 0 {
		// 基礎札はスートに固定されているので、空でも他スートのエースは置けない。
		// FortyThieves は固定していないのでこの条件が無く、そのまま写すと
		// ♠A が ♣ の枠を開けてしまう。
		return card.GetValue() == 1 && napoleonsSquareSuitOrder[fIdx] == card.GetDesign()
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1
}

// findFoundation 置ける基礎札のインデックスを探す（見つからなければ -1）
func (ns *NapoleonsSquare) findFoundation(card *Card) int {
	for i := range NapoleonsSquareFoundationCnt {
		if ns.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (ns *NapoleonsSquare) afterMove(actionType, detail string, card *Card) {
	afterMove(&ns.moveCount, ns, actionType, detail, card)
}

// checkGameClear 8 つの基礎札がすべて K まで積み上がったか
func (ns *NapoleonsSquare) checkGameClear() {
	for i := range NapoleonsSquareFoundationCnt {
		if len(ns.foundation[i]) != CardValueMax {
			return
		}
	}
	ns.phase = NapoleonsSquarePhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。二重に持つと片方だけ直したときに食い違う。
func (ns *NapoleonsSquare) checkStalemate() {
	if ns.phase != NapoleonsSquarePhasePlaying {
		return
	}
	ns.isStalemate = ns.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (ns *NapoleonsSquare) takeSnapshot() {
	snap := &napoleonsSquareSnapshot{
		phase:       ns.phase,
		moveCount:   ns.moveCount,
		isStalemate: ns.isStalemate,
		stock:       append([]*Card(nil), ns.stock...),
		waste:       append([]*Card(nil), ns.waste...),
	}
	for i := range NapoleonsSquareFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), ns.foundation[i]...)
	}
	for i := range NapoleonsSquareTableauCnt {
		snap.tableau[i] = append([]*NapoleonsSquareTableauCard(nil), ns.tableau[i]...)
	}
	ns.history = appendSnapshot(ns.history, snap)
}

// appendLog 棋譜エントリを追加
func (ns *NapoleonsSquare) appendLog(actionType, detail string, cards []*Card) {
	ns.appendLogAt(ns.moveCount, 0, actionType, detail, cards)
}

// napoleonsSquareMaxSliceLen caps slice sizes during deserialisation.
const napoleonsSquareMaxSliceLen = 1000

// napoleonsSquareSnapshotJSON is the wire format for a single undo snapshot.
// napoleonsSquareSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type napoleonsSquareSnapshotJSON struct {
	Tableau     [NapoleonsSquareTableauCnt][]*NapoleonsSquareTableauCard `json:"tb"`
	Stock       []*Card                                                  `json:"st"`
	Waste       []*Card                                                  `json:"wa"`
	Foundation  [NapoleonsSquareFoundationCnt][]*Card                    `json:"fd"`
	Phase       NapoleonsSquarePhase                                     `json:"ps"`
	MoveCount   int                                                      `json:"mc"`
	IsStalemate bool                                                     `json:"sl"`
}

// MarshalJSON implements json.Marshaler for napoleonsSquareSnapshot.
func (s *napoleonsSquareSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(napoleonsSquareSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Waste:       s.waste,
		Foundation:  s.foundation,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for napoleonsSquareSnapshot.
func (s *napoleonsSquareSnapshot) UnmarshalJSON(data []byte) error {
	var j napoleonsSquareSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > napoleonsSquareMaxSliceLen ||
		len(j.Waste) > napoleonsSquareMaxSliceLen {
		return errors.New("napoleonssquare: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > napoleonsSquareMaxSliceLen {
			return errors.New("napoleonssquare: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > napoleonsSquareMaxSliceLen {
			return errors.New("napoleonssquare: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	s.waste = j.Waste
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// napoleonsSquareJSON is the JSON wire format for NapoleonsSquare.
type napoleonsSquareJSON struct {
	TrumpCards  *TrumpCards                                              `json:"tc"`
	Tableau     [NapoleonsSquareTableauCnt][]*NapoleonsSquareTableauCard `json:"tb"`
	Stock       []*Card                                                  `json:"st"`
	Waste       []*Card                                                  `json:"ws"`
	Foundation  [NapoleonsSquareFoundationCnt][]*Card                    `json:"fd"`
	Phase       NapoleonsSquarePhase                                     `json:"ps"`
	MoveCount   int                                                      `json:"mc"`
	ActionLog   []*ActionLogEntry                                        `json:"al"`
	IsStalemate bool                                                     `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*napoleonsSquareSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (ns *NapoleonsSquare) MarshalJSON() ([]byte, error) {
	return json.Marshal(&napoleonsSquareJSON{
		TrumpCards:  ns.trumpCards,
		Tableau:     ns.tableau,
		Stock:       ns.stock,
		Waste:       ns.waste,
		Foundation:  ns.foundation,
		Phase:       ns.phase,
		MoveCount:   ns.moveCount,
		ActionLog:   ns.actionLog,
		IsStalemate: ns.isStalemate,
		History:     ns.history,
	})
}

// NapoleonsSquareTotalCards 2 デッキ分の総枚数。勝利は 8 つの組札にこの枚数を
// 積み切ることなので、進捗の分母そのもの (#5554)。
const NapoleonsSquareTotalCards = CardCnt * 2

// napoleonsSquareTotalCards は旧名。内部の境界チェックが使う。
const napoleonsSquareTotalCards = NapoleonsSquareTotalCards

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (ns *NapoleonsSquare) UnmarshalJSON(data []byte) error {
	var j napoleonsSquareJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > napoleonsSquareMaxSliceLen || len(j.History) > napoleonsSquareMaxSliceLen {
		return errors.New("napoleonssquare: input array exceeds maximum allowed size")
	}
	if j.Phase < NapoleonsSquarePhasePlaying || j.Phase > NapoleonsSquarePhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > napoleonsSquareTotalCards || len(j.Waste) > napoleonsSquareTotalCards {
		return fmt.Errorf("stock/waste too large: %d/%d", len(j.Stock), len(j.Waste))
	}
	for i := range NapoleonsSquareFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range NapoleonsSquareTableauCnt {
		if len(j.Tableau[i]) > napoleonsSquareTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		ns.trumpCards = j.TrumpCards
	}
	ns.tableau = j.Tableau
	ns.stock = j.Stock
	ns.waste = j.Waste
	ns.foundation = j.Foundation
	ns.phase = j.Phase
	ns.moveCount = j.MoveCount
	ns.actionLog = j.ActionLog
	ns.isStalemate = j.IsStalemate
	ns.history = j.History
	return nil
}
