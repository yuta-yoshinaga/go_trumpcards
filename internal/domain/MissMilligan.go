//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// MissMilliganPhase ミス・ミリガンのゲームフェーズ
type MissMilliganPhase int

// MissMilliganのフェーズ定数
const (
	// MissMilliganPhasePlaying プレイ中
	MissMilliganPhasePlaying MissMilliganPhase = iota
	// MissMilliganPhaseGameClear ゲームクリア
	MissMilliganPhaseGameClear
	// MissMilliganPhaseGameOver ゲームオーバー
	MissMilliganPhaseGameOver
)

// MissMilliganTableauCnt タブローの列数
const MissMilliganTableauCnt = 8

// MissMilliganFoundationCnt 基礎札の数。2 デッキなのでスートごとに 2 つ。
const MissMilliganFoundationCnt = 8

// missMilliganSuitOrder 基礎札インデックスとスートの対応。0..3 が 1 組目、
// 4..7 が 2 組目。固定しておくと配り直しても UI の位置が動かない。
var missMilliganSuitOrder = [MissMilliganFoundationCnt]int{
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
	CardDesignSpade, CardDesignClover, CardDesignHeart, CardDesignDiamond,
}

// MissMilliganTableauCard タブロー上のカード。全札が表向きだが、他のソリティアと
// 同じ形にしておくとプレゼンターを使い回せる。
type MissMilliganTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// MissMilliganHint ミス・ミリガンのヒント
type MissMilliganHint struct {
	// FromZone 移動元 "tableau" / "waived" / "stock"
	FromZone string
	// FromCol 移動元のタブロー列（FromZone が "tableau" のときのみ）
	FromCol int
	// CardIndex 移動元の列内インデックス。連番グループの先頭を指す。
	CardIndex int
	// ToZone 移動先 "foundation" / "tableau"
	ToZone string
	// ToIdx 移動先のインデックス（基礎札またはタブロー列）
	ToIdx int
}

// MissMilligan ミス・ミリガンゲームクラス。
//
// 2 デッキ 104 枚。8 列のタブローに 1 枚ずつ配って始め、手が尽きたら山札から
// **8 枚を一度に**（各列へ 1 枚ずつ）配り足す。タブローは色違いの降順、基礎札は
// スートごとに A から昇順。**空き列にはキング（またはキングを先頭とする連番）
// しか置けない。**
//
// 「ウェイブ（waiving）」がこのゲームの救済ルール: 山札を使い切ったあとに限り、
// タブローの正しい連番を 1 組だけ脇に持ち上げて保持でき、下に埋まっていた札を
// 動かしてから戻せる。保持中は新たなウェイブができない。
//
// issue #4407 のルール 5 は「山札から 1 枚ずつめくり…捨て札へ送る」としているが、
// ミス・ミリガンに捨て札は無く、山札は 8 枚単位で配り足す。実際の規則を採った。
// また issue が触れていない「空き列はキングのみ」も本来の規則なので実装している。
type MissMilligan struct {
	trumpCards *TrumpCards
	tableau    [MissMilliganTableauCnt][]*MissMilliganTableauCard
	stock      []*Card
	foundation [MissMilliganFoundationCnt][]*Card
	waived     []*Card
	phase      MissMilliganPhase
	moveCount  int
	actionLogBase
	history     []*missMilliganSnapshot
	isStalemate bool
}

// missMilliganSnapshot アンドゥ用スナップショット
type missMilliganSnapshot struct {
	tableau     [MissMilliganTableauCnt][]*MissMilliganTableauCard
	stock       []*Card
	foundation  [MissMilliganFoundationCnt][]*Card
	waived      []*Card
	phase       MissMilliganPhase
	moveCount   int
	isStalemate bool
}

// NewMissMilligan コンストラクタ
func NewMissMilligan(trumpCards *TrumpCards) *MissMilligan {
	return &MissMilligan{trumpCards: trumpCards}
}

// NewDefaultMissMilligan returns MissMilligan with two combined 52-card decks.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultMissMilligan() *MissMilligan {
	return NewMissMilligan(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームリセット
func (mm *MissMilligan) Reset() {
	mm.trumpCards.Shuffle()
	mm.phase = MissMilliganPhasePlaying
	mm.moveCount = 0
	mm.actionLog = nil
	mm.history = nil
	mm.isStalemate = false
	mm.waived = nil
	mm.stock = nil

	for i := range MissMilliganFoundationCnt {
		mm.foundation[i] = nil
	}
	for i := range MissMilliganTableauCnt {
		mm.tableau[i] = nil
	}

	// 最初の配りは 8 枚（各列 1 枚）。残り 96 枚は山札。
	for col := range MissMilliganTableauCnt {
		card := mm.trumpCards.DrawCard()
		if card == nil {
			break
		}
		mm.tableau[col] = append(mm.tableau[col],
			&MissMilliganTableauCard{Card: card, FaceUp: true})
	}
	for {
		card := mm.trumpCards.DrawCard()
		if card == nil {
			break
		}
		mm.stock = append(mm.stock, card)
	}

	mm.checkStalemate()
}

// Deal 山札から各列へ 1 枚ずつ、8 枚を一度に配り足す。
// ミス・ミリガンに捨て札は無く、山札はこの形でしか減らない。
func (mm *MissMilligan) Deal() error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if len(mm.stock) == 0 {
		return errors.New("no cards in stock")
	}
	if len(mm.waived) > 0 {
		// 保持したままの配り足しは、戻す先を自分で潰しかねない。
		return errors.New("return the waived cards before dealing")
	}
	mm.takeSnapshot()
	dealt := 0
	for col := range MissMilliganTableauCnt {
		if len(mm.stock) == 0 {
			break
		}
		card := mm.stock[0]
		mm.stock = mm.stock[1:]
		mm.tableau[col] = append(mm.tableau[col],
			&MissMilliganTableauCard{Card: card, FaceUp: true})
		dealt++
	}
	mm.afterMove("deal", fmt.Sprintf("山札→各列に%d枚配布", dealt), nil)
	return nil
}

// MoveTableauToTableau タブロー間でカードを移す。cardIndex 以降が色違い降順の
// 連番であれば、その塊ごとまとめて動かせる。
func (mm *MissMilligan) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if fromCol < 0 || fromCol >= MissMilliganTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= MissMilliganTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := mm.tableau[fromCol]
	// -1 は「最上段 1 枚」。BeleagueredCastle など既存のソリティアと同じ約束。
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	group := fromCards[cardIndex:]
	if !missMilliganIsRun(group) {
		return errors.New("cards do not form an alternating-colour descending run")
	}
	if !mm.canPlaceOnTableau(group[0].Card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	mm.takeSnapshot()
	// group は fromCards の内部を指しているので、切り詰める前にコピーを取る。
	moved := append([]*MissMilliganTableauCard(nil), group...)
	mm.tableau[fromCol] = fromCards[:cardIndex]
	mm.tableau[toCol] = append(mm.tableau[toCol], moved...)
	mm.afterMove("move",
		fmt.Sprintf("タブロー列%d→タブロー列%d(%d枚)", fromCol, toCol, len(moved)),
		moved[0].Card)
	return nil
}

// MoveTableauToFoundation タブロー最上段を基礎札へ移す
func (mm *MissMilligan) MoveTableauToFoundation(col int) error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if col < 0 || col >= MissMilliganTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := mm.tableau[col]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	card := fromCards[len(fromCards)-1].Card
	fIdx := mm.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	mm.takeSnapshot()
	mm.tableau[col] = fromCards[:len(fromCards)-1]
	mm.foundation[fIdx] = append(mm.foundation[fIdx], card)
	mm.afterMove("move", fmt.Sprintf("タブロー列%d→基礎札%d", col, fIdx), card)
	return nil
}

// Waive タブローの連番を 1 組だけ脇へ持ち上げる（救済ルール）。
// 山札を使い切ったあとにしか使えず、保持できるのは 1 組だけ。
func (mm *MissMilligan) Waive(col, cardIndex int) error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if len(mm.stock) > 0 {
		return errors.New("waiving is only allowed once the stock is exhausted")
	}
	if len(mm.waived) > 0 {
		return errors.New("cards are already waived")
	}
	if col < 0 || col >= MissMilliganTableauCnt {
		return errors.New("invalid column")
	}
	fromCards := mm.tableau[col]
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex < 0 || cardIndex >= len(fromCards) {
		return errors.New("invalid card index")
	}
	group := fromCards[cardIndex:]
	if !missMilliganIsRun(group) {
		return errors.New("cards do not form an alternating-colour descending run")
	}
	mm.takeSnapshot()
	waived := make([]*Card, 0, len(group))
	for _, tc := range group {
		waived = append(waived, tc.Card)
	}
	mm.tableau[col] = fromCards[:cardIndex]
	mm.waived = waived
	mm.afterMove("waive", fmt.Sprintf("列%dから%d枚を保持", col, len(waived)), waived[0])
	return nil
}

// PlaceWaived 保持中の札をタブローへ戻す
func (mm *MissMilligan) PlaceWaived(toCol int) error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if len(mm.waived) == 0 {
		return errors.New("nothing is waived")
	}
	if toCol < 0 || toCol >= MissMilliganTableauCnt {
		return errors.New("invalid column")
	}
	if !mm.canPlaceOnTableau(mm.waived[0], toCol) {
		return errors.New("cannot place waived cards there")
	}
	mm.takeSnapshot()
	for _, c := range mm.waived {
		mm.tableau[toCol] = append(mm.tableau[toCol],
			&MissMilliganTableauCard{Card: c, FaceUp: true})
	}
	n := len(mm.waived)
	head := mm.waived[0]
	mm.waived = nil
	mm.afterMove("move", fmt.Sprintf("保持→タブロー列%d(%d枚)", toCol, n), head)
	return nil
}

// MoveWaivedToFoundation 保持中の 1 枚を基礎札へ送る。
// 連番を保持しているときは基礎札に積めないので 1 枚のときだけ許す。
func (mm *MissMilligan) MoveWaivedToFoundation() error {
	if err := mm.requirePlaying(); err != nil {
		return err
	}
	if len(mm.waived) == 0 {
		return errors.New("nothing is waived")
	}
	if len(mm.waived) > 1 {
		return errors.New("only a single waived card can go to a foundation")
	}
	card := mm.waived[0]
	fIdx := mm.findFoundation(card)
	if fIdx < 0 {
		return errors.New("cannot place card on foundation")
	}
	mm.takeSnapshot()
	mm.waived = nil
	mm.foundation[fIdx] = append(mm.foundation[fIdx], card)
	mm.afterMove("move", fmt.Sprintf("保持→基礎札%d", fIdx), card)
	return nil
}

// GiveUp ギブアップ
func (mm *MissMilligan) GiveUp() {
	if mm.phase == MissMilliganPhasePlaying {
		mm.phase = MissMilliganPhaseGameOver
		mm.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint 手を 1 つ提示する。基礎札 → タブロー → 山札の順。手詰まり判定も兼ねる。
func (mm *MissMilligan) GetHint() *MissMilliganHint {
	if h := mm.foundationHint(); h != nil {
		return h
	}
	if h := mm.tableauHint(); h != nil {
		return h
	}
	if mm.phase == MissMilliganPhasePlaying && len(mm.stock) > 0 && len(mm.waived) == 0 {
		return &MissMilliganHint{FromZone: "stock", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToIdx: -1}
	}
	return nil
}

// foundationHint 基礎札へ送れる手を 1 つ返す（オートコンプリート用）。
// タブローの手を混ぜてはいけない — オートコンプリートは盤面をかき混ぜず、
// 基礎札に上げるだけの操作である。
func (mm *MissMilligan) foundationHint() *MissMilliganHint {
	if mm.phase != MissMilliganPhasePlaying {
		return nil
	}
	if len(mm.waived) == 1 {
		if fIdx := mm.findFoundation(mm.waived[0]); fIdx >= 0 {
			return &MissMilliganHint{FromZone: "waived", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToIdx: fIdx}
		}
	}
	for col := range MissMilliganTableauCnt {
		card := mm.topCard(col)
		if card == nil {
			continue
		}
		if fIdx := mm.findFoundation(card); fIdx >= 0 {
			return &MissMilliganHint{
				FromZone: "tableau", FromCol: col, CardIndex: len(mm.tableau[col]) - 1,
				ToZone: "foundation", ToIdx: fIdx,
			}
		}
	}
	return nil
}

// tableauHint タブローへ置ける手を 1 つ返す。
//
// 保持中の札を戻す手があればそれを最優先する — 保持したままでは配り足しも
// 2 度目のウェイブもできないため。
//
// ただし戻せないときに探索を打ち切ってはいけない。**盤面をかき混ぜて戻す先を
// 作ること自体がウェイブの目的**であり、MoveTableauToTableau と
// MoveTableauToFoundation は保持中でも禁止していない。ここで nil を返すと、
// 手が残っているのに手詰まりと判定されてしまう（#4472 のレビュー指摘）。
func (mm *MissMilligan) tableauHint() *MissMilliganHint {
	if mm.phase != MissMilliganPhasePlaying {
		return nil
	}
	if len(mm.waived) > 0 {
		for col := range MissMilliganTableauCnt {
			if mm.canPlaceOnTableau(mm.waived[0], col) {
				return &MissMilliganHint{FromZone: "waived", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToIdx: col}
			}
		}
		// 戻せないので、戻す先を作る手を下の通常探索で探す。
	}
	for from := range MissMilliganTableauCnt {
		pile := mm.tableau[from]
		for idx := range pile {
			if !missMilliganIsRun(pile[idx:]) {
				continue
			}
			for to := range MissMilliganTableauCnt {
				if to == from {
					continue
				}
				// 列ごと空き列へ動かしても盤面は進まない。
				if idx == 0 && len(mm.tableau[to]) == 0 {
					continue
				}
				if mm.canPlaceOnTableau(pile[idx].Card, to) {
					return &MissMilliganHint{
						FromZone: "tableau", FromCol: from, CardIndex: idx,
						ToZone: "tableau", ToIdx: to,
					}
				}
			}
		}
	}
	return nil
}

// AutoComplete 基礎札へ送れる札がなくなるまで自動で送る
func (mm *MissMilligan) AutoComplete() error {
	if mm.phase != MissMilliganPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	moved := false
	for {
		h := mm.foundationHint()
		if h == nil {
			break
		}
		var err error
		if h.FromZone == "waived" {
			err = mm.MoveWaivedToFoundation()
		} else {
			err = mm.MoveTableauToFoundation(h.FromCol)
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
func (mm *MissMilligan) Undo() error {
	if len(mm.history) == 0 {
		return errors.New("nothing to undo")
	}
	snap := mm.history[len(mm.history)-1]
	mm.history = mm.history[:len(mm.history)-1]
	mm.tableau = snap.tableau
	mm.stock = snap.stock
	mm.foundation = snap.foundation
	mm.waived = snap.waived
	mm.phase = snap.phase
	mm.moveCount = snap.moveCount
	mm.isStalemate = snap.isStalemate
	return nil
}

// CanUndo アンドゥ可能か
func (mm *MissMilligan) CanUndo() bool { return len(mm.history) > 0 }

// UndoN n 手戻す
func (mm *MissMilligan) UndoN(n int) error {
	if n <= 0 {
		return errors.New("n must be positive")
	}
	if n > len(mm.history) {
		return errors.New("not enough history")
	}
	for range n {
		if err := mm.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// UndoToEscape 膠着状態から抜けるのに必要なアンドゥ回数（膠着でなければ 0、不可なら -1）
func (mm *MissMilligan) UndoToEscape() int {
	if !mm.isStalemate {
		return 0
	}
	for i := len(mm.history) - 1; i >= 0; i-- {
		if !mm.history[i].isStalemate {
			return len(mm.history) - i
		}
	}
	return -1
}

// AllFaceUp 常に全札が表向き
func (mm *MissMilligan) AllFaceUp() bool { return true }

// GetPhase フェーズ取得
func (mm *MissMilligan) GetPhase() MissMilliganPhase { return mm.phase }

// GetMoveCount 手数取得
func (mm *MissMilligan) GetMoveCount() int { return mm.moveCount }

// GetStockCount 山札の残り枚数
func (mm *MissMilligan) GetStockCount() int { return len(mm.stock) }

// GetWaived 保持中の札を取得
func (mm *MissMilligan) GetWaived() []*Card { return mm.waived }

// CanWaive ウェイブが可能か（山札を使い切り、まだ何も保持していない）
func (mm *MissMilligan) CanWaive() bool {
	return mm.phase == MissMilliganPhasePlaying && len(mm.stock) == 0 && len(mm.waived) == 0
}

// GetTableau タブローを取得
func (mm *MissMilligan) GetTableau() [MissMilliganTableauCnt][]*MissMilliganTableauCard {
	return mm.tableau
}

// GetFoundation 基礎札を取得
func (mm *MissMilligan) GetFoundation() [MissMilliganFoundationCnt][]*Card { return mm.foundation }

// GetGameEndFlag ゲーム終了フラグ
func (mm *MissMilligan) GetGameEndFlag() bool { return mm.phase != MissMilliganPhasePlaying }

// IsStalemate 手詰まりか
func (mm *MissMilligan) IsStalemate() bool { return mm.isStalemate }

// --- Private helpers ---

// requirePlaying プレイ中でなければエラーを返す
func (mm *MissMilligan) requirePlaying() error {
	if mm.phase != MissMilliganPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	return nil
}

// missMilliganIsRun 色違いで 1 つずつ下がる連番か（1 枚は常に真）
func missMilliganIsRun(cards []*MissMilliganTableauCard) bool {
	for i := 1; i < len(cards); i++ {
		prev, cur := cards[i-1].Card, cards[i].Card
		if prev == nil || cur == nil {
			return false
		}
		if cur.GetValue() != prev.GetValue()-1 {
			return false
		}
		if missMilliganIsRed(prev.GetDesign()) == missMilliganIsRed(cur.GetDesign()) {
			return false
		}
	}
	return true
}

// missMilliganIsRed 赤いスートか
func missMilliganIsRed(design int) bool {
	return design == CardDesignHeart || design == CardDesignDiamond
}

// topCard 指定列の最上段カードを返す（空なら nil）
func (mm *MissMilligan) topCard(col int) *Card {
	pile := mm.tableau[col]
	if len(pile) == 0 {
		return nil
	}
	return pile[len(pile)-1].Card
}

// canPlaceOnTableau タブローに置けるか。
// **空き列はキングのみ** — issue #4407 は触れていないが本来の規則で、これが無いと
// 空き列が万能の置き場になってゲームが崩れる。
func (mm *MissMilligan) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := mm.tableau[col]
	if len(pile) == 0 {
		return card.GetValue() == CardValueMax
	}
	top := pile[len(pile)-1].Card
	if card.GetValue() != top.GetValue()-1 {
		return false
	}
	return missMilliganIsRed(card.GetDesign()) != missMilliganIsRed(top.GetDesign())
}

// canPlaceOnFoundation 基礎札に置けるか（空はそのスートのエースのみ、以降は同スートで 1 つ上）
func (mm *MissMilligan) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := mm.foundation[fIdx]
	if len(pile) == 0 {
		// 基礎札はスートに固定されているので、空でも他スートのエースは置けない。
		return card.GetValue() == 1 && missMilliganSuitOrder[fIdx] == card.GetDesign()
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1
}

// findFoundation 置ける基礎札のインデックスを探す（見つからなければ -1）
func (mm *MissMilligan) findFoundation(card *Card) int {
	for i := range MissMilliganFoundationCnt {
		if mm.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// afterMove 手数・棋譜・終了判定をまとめて進める
func (mm *MissMilligan) afterMove(actionType, detail string, card *Card) {
	mm.moveCount++
	var cards []*Card
	if card != nil {
		cards = []*Card{card}
	}
	mm.appendLog(actionType, detail, cards)
	mm.checkGameClear()
	mm.checkStalemate()
}

// checkGameClear 8 つの基礎札すべてが K まで積み上がったか
func (mm *MissMilligan) checkGameClear() {
	for i := range MissMilliganFoundationCnt {
		if len(mm.foundation[i]) != CardValueMax {
			return
		}
	}
	mm.phase = MissMilliganPhaseGameClear
}

// checkStalemate 打つ手が一つも無い状態か。
// GetHint は基礎札・タブロー・山札のすべてを見るので、「ヒントが無い」と
// 「手詰まり」は同じ条件になる。二重に持つと片方だけ直したときに食い違う。
//
// ウェイブは救済手段なので、まだ使えるうちは手詰まりとしない。保持中の場合は
// GetHint が「戻す手」と「戻す先を作る手」の両方を見ているので、そちらで尽きる。
func (mm *MissMilligan) checkStalemate() {
	if mm.phase != MissMilliganPhasePlaying {
		return
	}
	if mm.GetHint() != nil {
		mm.isStalemate = false
		return
	}
	mm.isStalemate = !mm.canWaiveAnything()
}

// canWaiveAnything 脇へ持ち上げられる札が 1 枚でもあるか
func (mm *MissMilligan) canWaiveAnything() bool {
	if !mm.CanWaive() {
		return false
	}
	for col := range MissMilliganTableauCnt {
		if len(mm.tableau[col]) > 0 {
			return true
		}
	}
	return false
}

// takeSnapshot 現在の状態を保存する
func (mm *MissMilligan) takeSnapshot() {
	snap := &missMilliganSnapshot{
		phase:       mm.phase,
		moveCount:   mm.moveCount,
		isStalemate: mm.isStalemate,
		stock:       append([]*Card(nil), mm.stock...),
		waived:      append([]*Card(nil), mm.waived...),
	}
	for i := range MissMilliganFoundationCnt {
		snap.foundation[i] = append([]*Card(nil), mm.foundation[i]...)
	}
	for i := range MissMilliganTableauCnt {
		snap.tableau[i] = append([]*MissMilliganTableauCard(nil), mm.tableau[i]...)
	}
	mm.history = append(mm.history, snap)
}

// appendLog 棋譜エントリを追加
func (mm *MissMilligan) appendLog(actionType, detail string, cards []*Card) {
	mm.appendLogAt(mm.moveCount, 0, actionType, detail, cards)
}

// missMilliganMaxSliceLen caps slice sizes during deserialisation.
const missMilliganMaxSliceLen = 1000

// missMilliganSnapshotJSON is the wire format for a single undo snapshot.
// missMilliganSnapshot uses unexported fields, so marshalling it directly would emit
// `[{},{}]` -- the undo depth would survive but every snapshot would be blank,
// and Undo would wipe the board instead of rewinding it (#4478).
type missMilliganSnapshotJSON struct {
	Tableau     [MissMilliganTableauCnt][]*MissMilliganTableauCard `json:"tb"`
	Stock       []*Card                                            `json:"st"`
	Foundation  [MissMilliganFoundationCnt][]*Card                 `json:"fd"`
	Waived      []*Card                                            `json:"wv"`
	Phase       MissMilliganPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	IsStalemate bool                                               `json:"sl"`
}

// MarshalJSON implements json.Marshaler for missMilliganSnapshot.
func (s *missMilliganSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(missMilliganSnapshotJSON{
		Tableau:     s.tableau,
		Stock:       s.stock,
		Foundation:  s.foundation,
		Waived:      s.waived,
		Phase:       s.phase,
		MoveCount:   s.moveCount,
		IsStalemate: s.isStalemate,
	})
}

// UnmarshalJSON implements json.Unmarshaler for missMilliganSnapshot.
func (s *missMilliganSnapshot) UnmarshalJSON(data []byte) error {
	var j missMilliganSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > missMilliganMaxSliceLen ||
		len(j.Waived) > missMilliganMaxSliceLen {
		return errors.New("missmilligan: snapshot array exceeds maximum allowed size")
	}
	for _, pile := range j.Tableau {
		if len(pile) > missMilliganMaxSliceLen {
			return errors.New("missmilligan: snapshot pile exceeds maximum allowed size")
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > missMilliganMaxSliceLen {
			return errors.New("missmilligan: snapshot pile exceeds maximum allowed size")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	s.foundation = j.Foundation
	s.waived = j.Waived
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	s.isStalemate = j.IsStalemate
	return nil
}

// missMilliganJSON is the JSON wire format for MissMilligan.
type missMilliganJSON struct {
	TrumpCards  *TrumpCards                                        `json:"tc"`
	Tableau     [MissMilliganTableauCnt][]*MissMilliganTableauCard `json:"tb"`
	Stock       []*Card                                            `json:"st"`
	Foundation  [MissMilliganFoundationCnt][]*Card                 `json:"fd"`
	Waived      []*Card                                            `json:"wv"`
	Phase       MissMilliganPhase                                  `json:"ps"`
	MoveCount   int                                                `json:"mc"`
	ActionLog   []*ActionLogEntry                                  `json:"al"`
	IsStalemate bool                                               `json:"sl"`
	// History must round-trip: the Cloudflare Worker is stateless per request
	// and rebuilds the game from KV every call, so an unpersisted undo stack
	// means Undo/UndoN/UndoToEscape silently never work in production (#4478).
	History []*missMilliganSnapshot `json:"hi,omitempty"`
}

// MarshalJSON KV スナップショット用のシリアライズ
func (mm *MissMilligan) MarshalJSON() ([]byte, error) {
	return json.Marshal(&missMilliganJSON{
		TrumpCards:  mm.trumpCards,
		Tableau:     mm.tableau,
		Stock:       mm.stock,
		Foundation:  mm.foundation,
		Waived:      mm.waived,
		Phase:       mm.phase,
		MoveCount:   mm.moveCount,
		ActionLog:   mm.actionLog,
		IsStalemate: mm.isStalemate,
		History:     mm.history,
	})
}

// missMilliganTotalCards 2 デッキ分の総枚数
const missMilliganTotalCards = CardCnt * 2

// UnmarshalJSON KV スナップショットからの復元。KV には以前のバージョンが書いた任意の
// バイト列が入りうるので、壊れた状態でゲームを開始させないよう値域を検証する。
func (mm *MissMilligan) UnmarshalJSON(data []byte) error {
	var j missMilliganJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > missMilliganMaxSliceLen || len(j.History) > missMilliganMaxSliceLen {
		return errors.New("missmilligan: input array exceeds maximum allowed size")
	}
	if j.Phase < MissMilliganPhasePlaying || j.Phase > MissMilliganPhaseGameOver {
		return fmt.Errorf("invalid phase: %d", j.Phase)
	}
	if j.MoveCount < 0 {
		return fmt.Errorf("invalid move count: %d", j.MoveCount)
	}
	if len(j.Stock) > missMilliganTotalCards {
		return fmt.Errorf("stock too large: %d", len(j.Stock))
	}
	if len(j.Waived) > CardValueMax {
		return fmt.Errorf("waived too large: %d", len(j.Waived))
	}
	for i := range MissMilliganFoundationCnt {
		if len(j.Foundation[i]) > CardValueMax {
			return fmt.Errorf("foundation %d holds %d cards", i, len(j.Foundation[i]))
		}
	}
	for i := range MissMilliganTableauCnt {
		if len(j.Tableau[i]) > missMilliganTotalCards {
			return fmt.Errorf("tableau %d holds %d cards", i, len(j.Tableau[i]))
		}
	}
	if j.TrumpCards != nil {
		mm.trumpCards = j.TrumpCards
	}
	mm.tableau = j.Tableau
	mm.stock = j.Stock
	mm.foundation = j.Foundation
	mm.waived = j.Waived
	mm.phase = j.Phase
	mm.moveCount = j.MoveCount
	mm.actionLog = j.ActionLog
	mm.isStalemate = j.IsStalemate
	mm.history = j.History
	return nil
}
