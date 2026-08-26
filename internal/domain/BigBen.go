//go:build !js || !wasm || extra3

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// BigBenPhase ビッグ・ベンのゲームフェーズ
type BigBenPhase int

// BigBenのフェーズ定数
const (
	// BigBenPhasePlaying プレイ中
	BigBenPhasePlaying BigBenPhase = iota
	// BigBenPhaseGameClear ゲームクリア
	BigBenPhaseGameClear
	// BigBenPhaseGameOver ゲームオーバー
	BigBenPhaseGameOver
)

// BigBenFoundationCnt 基礎札の数。時計の文字盤 12 個ぶん。
const BigBenFoundationCnt = 12

// BigBenTableauCnt タブローの列数
const BigBenTableauCnt = 8

// BigBenColumnLen 1 列あたりの配り枚数
const BigBenColumnLen = 5

// BigBenTotalCards 使用する総枚数（52 枚 2 組）
const BigBenTotalCards = CardCnt * 2

// BigBenDealMinimum 補充で各列を満たす最低枚数。**手が尽きたら、すべての列が
// この枚数になるまで山札から配る。**すでに全列がこれ以上なら 1 巡だけ配る。
const BigBenDealMinimum = 3

// bigBenStarter は文字盤 1 つぶんの初期カードと目標ランク。
type bigBenStarter struct {
	design int
	value  int
}

// bigBenStarters 文字盤に最初から置かれる 12 枚。
//
// インデックス i は時計盤の位置で、**9 時から時計回り**に並ぶ。9 時の ♣2 から
// 1 ランクずつ上がり、スートは ♣♥♠♦ を繰り返して 8 時の ♦K で一周する。
// 各文字盤の目標ランクは**その位置の時刻**（9 時なら 9、12 時なら Q）。
//
// **この並びが 104 枚にちょうど閉じる。**9〜12 時の 4 本が 8 枚、残り 8 本が
// 9 枚で 4×8 + 8×9 = 104。だから「全札が文字盤へ乗る」と「各文字盤が自分の
// 時刻を表示する」は同じことを言っている。並びを崩すと、勝てないか札が余る。
var bigBenStarters = [BigBenFoundationCnt]bigBenStarter{
	{CardDesignClover, 2},             // 9 時: ♣2 → 9 まで（8 枚）
	{CardDesignHeart, 3},              // 10 時: ♥3 → 10 まで（8 枚）
	{CardDesignSpade, 4},              // 11 時: ♠4 → J まで（8 枚）
	{CardDesignDiamond, 5},            // 12 時: ♦5 → Q まで（8 枚）
	{CardDesignClover, 6},             // 1 時: ♣6 → K で折り返して A まで（9 枚）
	{CardDesignHeart, 7},              // 2 時: ♥7 → 2 まで（9 枚）
	{CardDesignSpade, 8},              // 3 時: ♠8 → 3 まで（9 枚）
	{CardDesignDiamond, 9},            // 4 時: ♦9 → 4 まで（9 枚）
	{CardDesignClover, 10},            // 5 時: ♣10 → 5 まで（9 枚）
	{CardDesignHeart, 11},             // 6 時: ♥J → 6 まで（9 枚）
	{CardDesignSpade, 12},             // 7 時: ♠Q → 7 まで（9 枚）
	{CardDesignDiamond, CardValueMax}, // 8 時: ♦K → 8 まで（9 枚）
}

// bigBenHourOrder は文字盤インデックスから時刻（＝目標ランク）への対応。
// 9 時始まりなので、添字 0..11 が 9,10,11,12,1,2,...,8 になる。
var bigBenHourOrder = [BigBenFoundationCnt]int{9, 10, 11, 12, 1, 2, 3, 4, 5, 6, 7, 8}

// BigBenTargetRank 文字盤 idx（0 始まり、9 時から時計回り）の目標ランクを返す。
func BigBenTargetRank(idx int) int {
	if idx < 0 || idx >= BigBenFoundationCnt {
		return 0
	}
	return bigBenHourOrder[idx]
}

// BigBenTableauCard タブロー上のカード。全札が表向きだが、他の
// ソリティアと同じ形にしておくとプレゼンターを使い回せる。
type BigBenTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// BigBenHint ビッグ・ベンのヒント。
//
// **補充の勧めは移動ではない。**FromZone が "stock" のときは ToZone も "stock"
// で、列を指さない（-1）。移動の体裁に落とすと存在しない列 -1 が画面に漏れる。
type BigBenHint struct {
	// FromZone 移動元 "tableau"、または補充を勧める "stock"
	FromZone string
	// FromCol 移動元のタブロー列（補充のときは -1）
	FromCol int
	// ToZone 移動先 "foundation" / "tableau"、または補充を勧める "stock"
	ToZone string
	// ToIdx 移動先のインデックス（文字盤またはタブロー列）
	ToIdx int
}

// BigBen ビッグ・ベンゲームクラス。
//
// 52 枚 2 組（104 枚）。時計盤の 12 枚が文字盤として最初から置かれ、続く 40 枚が
// 8 列×5 枚のタブローへ、残り **52 枚が山札**になる。
//
// 文字盤は 9 時の ♣2 から時計回りに 1 ランクずつ上がり、8 時の ♦K で一周する。
// 同スートで昇順（K の次は A に折り返す）に積み、**その位置の時刻に対応する
// ランク**に達したら完成で、それ以上は受け付けない。全文字盤が完成したとき、
// 104 枚がちょうど乗り切っている。
//
// **タブローは同スートで降順**、上札のみ、**空き列は埋められない**。手が尽きたら
// `Deal` で山札から補充する ── 各列が 3 枚になるまで配り、すでに全列が 3 枚以上
// なら 1 巡だけ配る。**再配りは無い。**
//
// クローン元のグランドファーザーズ・クロックとは 5 点異なる: 1 組 52 枚で山札を
// 持たない / 文字盤の並びと目標ランクが違う（1 時＝A 始まり）/ タブローは
// スート不問 / 空き列に任意の札を置ける / 補充が無い。
type BigBen struct {
	trumpCards *TrumpCards
	foundation [BigBenFoundationCnt][]*Card
	tableau    [BigBenTableauCnt][]*BigBenTableauCard
	stock      []*Card
	phase      BigBenPhase
	moveCount  int
	actionLogBase
	history     []*bigBenSnapshot
	isStalemate bool
}

// bigBenSnapshot アンドゥ用スナップショット
type bigBenSnapshot struct {
	foundation  [BigBenFoundationCnt][]*Card
	tableau     [BigBenTableauCnt][]*BigBenTableauCard
	stock       []*Card
	phase       BigBenPhase
	moveCount   int
	isStalemate bool
}

// NewBigBen コンストラクタ
func NewBigBen(trumpCards *TrumpCards) *BigBen {
	return &BigBen{trumpCards: trumpCards}
}

// NewDefaultBigBen returns BigBen with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultBigBen() *BigBen {
	return NewBigBen(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (gc *BigBen) Reset() {
	gc.trumpCards.Shuffle()
	gc.phase = BigBenPhasePlaying
	gc.moveCount = 0
	gc.actionLog = nil
	gc.history = nil
	gc.isStalemate = false

	for i := range BigBenFoundationCnt {
		gc.foundation[i] = nil
	}
	for i := range BigBenTableauCnt {
		gc.tableau[i] = nil
	}

	// 文字盤の 12 枚は固定なので、山から抜き出して所定の位置に置く。残りが
	// ちょうど 40 枚になり、8 列×5 枚に収まる。
	gc.stock = nil
	remaining := make([]*Card, 0, BigBenTotalCards-BigBenFoundationCnt)
	for {
		card := gc.trumpCards.DrawCard()
		if card == nil {
			break
		}
		if idx := bigBenStarterIndex(card); idx >= 0 && len(gc.foundation[idx]) == 0 {
			gc.foundation[idx] = []*Card{card}
			continue
		}
		remaining = append(remaining, card)
	}

	for col := range BigBenTableauCnt {
		for range BigBenColumnLen {
			if len(remaining) == 0 {
				break
			}
			card := remaining[0]
			remaining = remaining[1:]
			gc.tableau[col] = append(gc.tableau[col],
				&BigBenTableauCard{Card: card, FaceUp: true})
		}
	}
	// **配り切らずに残す。**52 枚が山札になり、手が尽きたときの補充源になる。
	// クローン元は 1 組 52 枚がちょうど配り切れるので山札を持たない。
	gc.stock = remaining

	gc.checkStalemate()
}

// bigBenStarterIndex カードが文字盤の初期札ならその位置を返す（違えば -1）。
func bigBenStarterIndex(card *Card) int {
	if card == nil {
		return -1
	}
	for i, s := range bigBenStarters {
		if s.design == card.GetDesign() && s.value == card.GetValue() {
			return i
		}
	}
	return -1
}

// MoveTableauToFoundation タブロー最上段を文字盤へ移す
func (gc *BigBen) MoveTableauToFoundation(col, fIdx int) error {
	if err := gc.requirePlaying(); err != nil {
		return err
	}
	card, err := gc.topOf(col)
	if err != nil {
		return err
	}
	if fIdx < 0 || fIdx >= BigBenFoundationCnt {
		return errors.New("invalid foundation index")
	}
	if !gc.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on that clock face")
	}
	gc.takeSnapshot()
	gc.popTop(col)
	gc.foundation[fIdx] = append(gc.foundation[fIdx], card)
	gc.afterMove("move", fmt.Sprintf("タブロー列%d→文字盤%d", col, fIdx), card)
	return nil
}

// MoveTableauToTableau タブロー最上段を別の列へ移す（同スートで 1 つ下。空き列は埋められない）
func (gc *BigBen) MoveTableauToTableau(fromCol, toCol int) error {
	if err := gc.requirePlaying(); err != nil {
		return err
	}
	card, err := gc.topOf(fromCol)
	if err != nil {
		return err
	}
	if toCol < 0 || toCol >= BigBenTableauCnt {
		return errors.New("invalid destination column")
	}
	if fromCol == toCol {
		return errors.New("source and destination are the same column")
	}
	if !gc.canPlaceOnTableau(card, toCol) {
		return NewDomainErrorCode(ErrInvalidPlay, "bigben.errCannotPlaceTableau", nil)
	}
	gc.takeSnapshot()
	gc.popTop(fromCol)
	gc.tableau[toCol] = append(gc.tableau[toCol],
		&BigBenTableauCard{Card: card, FaceUp: true})
	gc.afterMove("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), card)
	return nil
}

// Deal 山札から補充する。
//
// **各列が 3 枚になるまで配る。**すでに全列が 3 枚以上なら 1 巡だけ配る ──
// そこで何もしないと、山札が残っているのに手が進まなくなる。再配りは無い。
func (gc *BigBen) Deal() error {
	if err := gc.requirePlaying(); err != nil {
		return err
	}
	if len(gc.stock) == 0 {
		return NewDomainErrorCode(ErrDeckExhausted, "bigben.errStockEmptyNoRedeal", nil)
	}
	gc.takeSnapshot()

	dealt := 0
	needsTopUp := false
	for col := range BigBenTableauCnt {
		if len(gc.tableau[col]) < BigBenDealMinimum {
			needsTopUp = true
			break
		}
	}
	for col := range BigBenTableauCnt {
		want := 1
		if needsTopUp {
			want = BigBenDealMinimum - len(gc.tableau[col])
		}
		for range want {
			if len(gc.stock) == 0 {
				break
			}
			card := gc.stock[0]
			gc.stock = gc.stock[1:]
			gc.tableau[col] = append(gc.tableau[col], &BigBenTableauCard{Card: card, FaceUp: true})
			dealt++
		}
	}
	gc.moveCount++
	gc.appendLog("deal", fmt.Sprintf("山札から%d枚補充した", dealt), nil)
	gc.checkGameClear()
	gc.checkStalemate()
	return nil
}

// GetStockCount 山札の残り枚数
func (gc *BigBen) GetStockCount() int { return len(gc.stock) }

// GiveUp ギブアップ
func (gc *BigBen) GiveUp() {
	if gc.phase == BigBenPhasePlaying {
		gc.phase = BigBenPhaseGameOver
		gc.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。文字盤へ送れる手を優先し、無ければタブローの手を返す。
// 手詰まり判定もこの関数を使う。
func (gc *BigBen) GetHint() *BigBenHint {
	// **終わった局面ではヒントを出さない。**クローン元は foundationHint と
	// tableauHint がそれぞれフェーズを見ていたので不要だったが、補充の枝は
	// それを通らない ── ギブアップ後も「配れ」と言い続けてしまう。
	if gc.phase != BigBenPhasePlaying {
		return nil
	}
	if h := gc.foundationHint(); h != nil {
		return h
	}
	if h := gc.tableauHint(); h != nil {
		return h
	}
	// **山札が残っていれば手詰まりではない。**補充がこのゲームの逃げ道なので、
	// ここを返さないと「配れば動くのに手詰まり」と言う盤ができる。行き先の列を
	// 持たないので、移動の体裁には落とさない。
	if len(gc.stock) > 0 {
		return &BigBenHint{FromZone: "stock", FromCol: -1, ToZone: "stock", ToIdx: -1}
	}
	return nil
}

// foundationHint 文字盤へ送れる手を 1 つ返す（オートコンプリート用）。
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 文字盤に上げるだけの操作である。
func (gc *BigBen) foundationHint() *BigBenHint {
	if gc.phase != BigBenPhasePlaying {
		return nil
	}
	for col := range BigBenTableauCnt {
		card := gc.topCard(col)
		if card == nil {
			continue
		}
		for fIdx := range BigBenFoundationCnt {
			if gc.canPlaceOnFoundation(card, fIdx) {
				return &BigBenHint{FromZone: "tableau", FromCol: col, ToZone: "foundation", ToIdx: fIdx}
			}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
//
// 空き列への移動は、その列に 1 枚しかない場合は「出して戻す」だけで盤面が進ま
// ないため提案しない。放置するとヒントが無限に同じ手を勧める。
func (gc *BigBen) tableauHint() *BigBenHint {
	if gc.phase != BigBenPhasePlaying {
		return nil
	}
	for from := range BigBenTableauCnt {
		card := gc.topCard(from)
		if card == nil {
			continue
		}
		for to := range BigBenTableauCnt {
			if to == from {
				continue
			}
			if gc.canPlaceOnTableau(card, to) {
				return &BigBenHint{FromZone: "tableau", FromCol: from, ToZone: "tableau", ToIdx: to}
			}
		}
	}
	return nil
}

// AutoComplete 文字盤へ送れる札がなくなるまで自動で送る
func (gc *BigBen) AutoComplete() error {
	if gc.phase != BigBenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := gc.foundationHint()
		if h == nil {
			break
		}
		if err := gc.MoveTableauToFoundation(h.FromCol, h.ToIdx); err != nil {
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
func (gc *BigBen) Undo() error {
	if len(gc.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := gc.history[len(gc.history)-1]
	gc.history = gc.history[:len(gc.history)-1]
	gc.foundation = snap.foundation
	gc.tableau = snap.tableau
	gc.stock = snap.stock
	gc.phase = snap.phase
	gc.moveCount = snap.moveCount
	gc.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (gc *BigBen) CanUndo() bool { return len(gc.history) > 0 }

// UndoN n 手戻す
func (gc *BigBen) UndoN(n int) error {
	return undoNChecked(gc, n, len(gc.history))
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (gc *BigBen) UndoToEscape() int {
	return undoToEscape(gc.isStalemate, gc.history, func(s *bigBenSnapshot) bool { return s.isStalemate })
}

// AllFaceUp 常に全札が表向き
func (gc *BigBen) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (gc *BigBen) GetPhase() BigBenPhase { return gc.phase }

// GetMoveCount 手数取得
func (gc *BigBen) GetMoveCount() int { return gc.moveCount }

// GetFoundation 文字盤を取得
func (gc *BigBen) GetFoundation() [BigBenFoundationCnt][]*Card {
	return gc.foundation
}

// GetTableau タブローを取得
func (gc *BigBen) GetTableau() [BigBenTableauCnt][]*BigBenTableauCard {
	return gc.tableau
}

// GetGameEndFlag ゲーム終了フラグ
func (gc *BigBen) GetGameEndFlag() bool {
	return gc.phase != BigBenPhasePlaying
}

// IsStalemate 手詰まりか
func (gc *BigBen) IsStalemate() bool { return gc.isStalemate }

// IsFoundationComplete 文字盤 fIdx が目標ランクに達しているか
func (gc *BigBen) IsFoundationComplete(fIdx int) bool {
	if fIdx < 0 || fIdx >= BigBenFoundationCnt {
		return false
	}
	pile := gc.foundation[fIdx]
	if len(pile) == 0 {
		return false
	}
	return pile[len(pile)-1].GetValue() == BigBenTargetRank(fIdx)
}

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (gc *BigBen) requirePlaying() error {
	if gc.phase != BigBenPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// topOf 指定列の最上段カードを返す（列が空・範囲外ならエラー）
func (gc *BigBen) topOf(col int) (*Card, error) {
	if col < 0 || col >= BigBenTableauCnt {
		return nil, errors.New("invalid column index")
	}
	pile := gc.tableau[col]
	if len(pile) == 0 {
		return nil, errors.New("column is empty")
	}
	return pile[len(pile)-1].Card, nil
}

// topCard 指定列の最上段カードを返す（空なら nil）。手の探索用。
func (gc *BigBen) topCard(col int) *Card {
	pile := gc.tableau[col]
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1].Card
}

// popTop 指定列の最上段を取り除く
func (gc *BigBen) popTop(col int) {
	gc.tableau[col] = gc.tableau[col][:len(gc.tableau[col])-1]
}

// canPlaceOnTableau タブローに置けるか（同スートで 1 つ下。空き列は埋められない）。
// 文字盤と違い、こちらは A の下へ K を置くような折り返しをしない。
func (gc *BigBen) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := gc.tableau[col]
	// **空き列は埋められない。**クローン元は任意の札を置けるので、その分岐を
	// 残すと、このゲームには無い逃げ場が生まれる。
	if len(pile) == 0 {
		return false
	}
	top := pile[len(pile)-1].Card
	// **同スートで 1 つ下。**クローン元はスートを見ない。
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()-1
}

// canPlaceOnFoundation 文字盤に置けるか。
// 同スートで 1 つ上、K の次は A に折り返す。目標ランクに達した文字盤は完成
// なので何も受け付けない — 折り返しを許すと 1 周して積み続けられてしまう。
func (gc *BigBen) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := gc.foundation[fIdx]
	if len(pile) == 0 {
		return false
	}
	if gc.IsFoundationComplete(fIdx) {
		return false
	}
	top := pile[len(pile)-1]
	if card.GetDesign() != top.GetDesign() {
		return false
	}
	return card.GetValue() == bigBenNextRank(top.GetValue())
}

// bigBenNextRank K の次を A に折り返す昇順の次ランク
func bigBenNextRank(v int) int {
	if v >= CardValueMax {
		return 1
	}
	return v + 1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (gc *BigBen) afterMove(actionType, detail string, card *Card) {
	gc.moveCount++
	gc.appendLog(actionType, detail, []*Card{card})
	gc.checkGameClear()
	gc.checkStalemate()
}

// checkGameClear 12 の文字盤すべてが目標ランクに達したか
func (gc *BigBen) checkGameClear() {
	for i := range BigBenFoundationCnt {
		if !gc.IsFoundationComplete(i) {
			return
		}
	}
	gc.phase = BigBenPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は文字盤とタブローの両方を見るので、「ヒントが無い」と「手詰まり」は
// 同じ条件になる。二重に持つと片方だけ直したときに静かに食い違う。
func (gc *BigBen) checkStalemate() {
	if gc.phase != BigBenPhasePlaying {
		return
	}
	gc.isStalemate = gc.GetHint() == nil
}

// takeSnapshot 現在の状態を保存する
func (gc *BigBen) takeSnapshot() {
	snap := &bigBenSnapshot{
		// **山札も巻き戻す。**載せ忘れると、補充を undo したときに札は列から
		// 消えるのに山札には戻らず、盤から枚数が失われる。
		stock:       append([]*Card(nil), gc.stock...),
		phase:       gc.phase,
		moveCount:   gc.moveCount,
		isStalemate: gc.isStalemate,
	}
	for i := range BigBenFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), gc.foundation[i]...)
	}
	for i := range BigBenTableauCnt {
		snap.tableau[i] = append([]*BigBenTableauCard(nil), gc.tableau[i]...)
	}
	gc.history = appendSnapshot(gc.history, snap)
}

// appendLog 棋譜エントリを追加
func (gc *BigBen) appendLog(actionType, detail string, cards []*Card) {
	gc.appendLogAt(gc.moveCount, 0, actionType, detail, cards)
}

// bigBenMaxSliceLen caps slice sizes during deserialisation.
const bigBenMaxSliceLen = 1000

// bigBenSnapshotJSON is the wire format for a single undo snapshot.
// bigBenSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type bigBenSnapshotJSON struct {
	Foundation [BigBenFoundationCnt][]*Card           `json:"fd"`
	Tableau    [BigBenTableauCnt][]*BigBenTableauCard `json:"tb"`
	// Stock も載せる。**メモリ上のスナップショットを直しただけでは足りない** ──
	// Worker は毎リクエスト KV から組み直すので、ここに無い項は往復のたびに
	// 消え、補充を undo した札が山札へ戻らなくなる。
	Stock       []*Card     `json:"st"`
	Phase       BigBenPhase `json:"ps"`
	MoveCount   int         `json:"mc"`
	IsStalemate bool        `json:"sl"`
}

// MarshalJSON implements json.Marshaler for bigBenSnapshot.
func (s *bigBenSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(bigBenSnapshotJSON{
		Foundation:  s.foundation,
		Tableau:     s.tableau,
		Stock:       s.stock,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for bigBenSnapshot.
func (s *bigBenSnapshot) UnmarshalJSON(data []byte) error {
	var j bigBenSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	for _, pile := range j.Foundation {
		if len(pile) > bigBenMaxSliceLen {
			return errors.New("bigben: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Tableau {
		if len(pile) > bigBenMaxSliceLen {
			return errors.New("bigben: snapshot pile exceeds maximum allowed size")
		}
	}
	if len(j.Stock) > bigBenMaxSliceLen {
		return errors.New("bigben: snapshot stock exceeds maximum allowed size")
	}
	s.foundation = j.Foundation
	s.tableau = j.Tableau
	s.stock = j.Stock
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// bigBenJSON is the JSON wire format for BigBen.
type bigBenJSON struct {
	TrumpCards *TrumpCards                            `json:"tc"`
	Foundation [BigBenFoundationCnt][]*Card           `json:"fd"`
	Tableau    [BigBenTableauCnt][]*BigBenTableauCard `json:"tb"`
	// Stock は必ず載せる。Worker はリクエストごとに KV から盤を組み直すので、
	// 載せ忘れると 1 回目の往復以降ずっと「山札が空」になり、補充が二度と
	// 効かない。History と同じ理由 (#4478)。
	Stock       []*Card           `json:"st"`
	Phase       BigBenPhase       `json:"ps"`
	MoveCount   int               `json:"mc"`
	ActionLog   []*ActionLogEntry `json:"al"`
	IsStalemate bool              `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*bigBenSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (gc *BigBen) MarshalJSON() ([]byte, error) {
	return json.Marshal(&bigBenJSON{
		TrumpCards:  gc.trumpCards,
		Foundation:  gc.foundation,
		Tableau:     gc.tableau,
		Stock:       gc.stock,
		Phase:       gc.phase,
		MoveCount:   gc.moveCount,
		ActionLog:   gc.actionLog,
		IsStalemate: gc.isStalemate,
		History:     gc.history,
	})
}

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (gc *BigBen) UnmarshalJSON(data []byte) error {
	var j bigBenJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > bigBenMaxSliceLen {
		return errors.New("bigben: stock exceeds maximum allowed size")
	}
	if len(j.ActionLog) > bigBenMaxSliceLen || len(j.History) > bigBenMaxSliceLen {
		return errors.New("bigben: input array exceeds maximum allowed size")
	}
	if j.Phase < BigBenPhasePlaying || j.Phase > BigBenPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	for i := range BigBenFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("clock face %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range BigBenTableauCnt {
		if len(j.Tableau[i]) > CardCnt {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		gc.trumpCards = j.TrumpCards
	}
	gc.foundation = j.Foundation
	gc.tableau = j.Tableau
	gc.stock = j.Stock
	gc.phase = j.Phase
	gc.moveCount = j.MoveCount
	gc.actionLog = j.ActionLog
	gc.isStalemate = j.IsStalemate
	gc.history = j.History
	return nil
}
