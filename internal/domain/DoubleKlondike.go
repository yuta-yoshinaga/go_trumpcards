//go:build !js || !wasm || extra2

package domain

import (
	"errors"
	"fmt"
)

// Double Klondike (Gargantua) 盤面定数。
const (
	// DoubleKlondikeTableauCnt タブロー列数。
	DoubleKlondikeTableauCnt = 9
	// DoubleKlondikeFoundationCnt ファウンデーション本数 (2 デッキ = スート毎 2 本)。
	DoubleKlondikeFoundationCnt = 8
	// DoubleKlondikeDrawCount ストックから一度にめくる枚数。
	DoubleKlondikeDrawCount = 3
	// doubleKlondikeMaxSliceLen JSON 復元時のスライス長上限。
	doubleKlondikeMaxSliceLen = 10000
)

// DoubleKlondikePhase ゲームフェーズ。
type DoubleKlondikePhase int

// Double Klondike のフェーズ定数。
const (
	// DoubleKlondikePhasePlaying プレイ中。
	DoubleKlondikePhasePlaying DoubleKlondikePhase = iota
	// DoubleKlondikePhaseGameClear クリア (8 ファウンデーション完成)。
	DoubleKlondikePhaseGameClear
	// DoubleKlondikePhaseGameOver ギブアップ。
	DoubleKlondikePhaseGameOver
)

// DoubleKlondikeTableauCard タブロー上のカード。Klondike とは別バケット (classic worker)
// に属するため、DoubleKlondikeTableauCard を共有せず独自に定義する。
type DoubleKlondikeTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// DoubleKlondikeHint 推奨手。FromZone/ToZone は "tableau"/"waste"/"foundation"。
type DoubleKlondikeHint struct {
	FromZone  string
	FromCol   int
	CardIndex int
	ToZone    string
	ToCol     int
}

// DoubleKlondike Double Klondike (ガルガンチュア) 本体。状態のみを保持する。
type DoubleKlondike struct {
	trumpCards *TrumpCards
	tableau    [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard
	stock      []*Card
	waste      []*Card
	foundation [DoubleKlondikeFoundationCnt][]*Card
	phase      DoubleKlondikePhase
	moveCount  int
	actionLogBase
	history []*doubleKlondikeSnapshot
}

// doubleKlondikeSnapshot Undo 用スナップショット。
type doubleKlondikeSnapshot struct {
	tableau    [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard
	stock      []*Card
	waste      []*Card
	foundation [DoubleKlondikeFoundationCnt][]*Card
	phase      DoubleKlondikePhase
	moveCount  int
}

// NewDoubleKlondike コンストラクタ。
func NewDoubleKlondike(trumpCards *TrumpCards) *DoubleKlondike {
	return &DoubleKlondike{trumpCards: trumpCards}
}

// NewDefaultDoubleKlondike 2 デッキ (104 枚) で生成する。
func NewDefaultDoubleKlondike() *DoubleKlondike {
	return NewDoubleKlondike(NewTrumpCardsWithDecks(2, 0))
}

// Reset ゲームを初期化する。9 列に 1..9 枚ずつ配り、各列の最上位のみ表向き。
func (g *DoubleKlondike) Reset() {
	g.trumpCards.Shuffle()
	g.phase = DoubleKlondikePhasePlaying
	g.moveCount = 0
	g.actionLog = nil
	g.history = nil
	for i := 0; i < DoubleKlondikeTableauCnt; i++ {
		g.tableau[i] = make([]*DoubleKlondikeTableauCard, 0, i+1)
		for j := 0; j <= i; j++ {
			g.tableau[i] = append(g.tableau[i], &DoubleKlondikeTableauCard{Card: g.trumpCards.DrawCard(), FaceUp: j == i})
		}
	}
	for i := 0; i < DoubleKlondikeFoundationCnt; i++ {
		g.foundation[i] = nil
	}
	g.stock = nil
	g.waste = nil
	for g.trumpCards.GetRemainingCount() > 0 {
		g.stock = append(g.stock, g.trumpCards.DrawCard())
	}
	g.appendLog("deal", "新しいゲームを開始しました", nil)
}

// --- rules ---

// dkIsBlack 黒スート (スペード・クラブ) なら true。
func dkIsBlack(c *Card) bool {
	return c.GetDesign() == CardDesignSpade || c.GetDesign() == CardDesignClover
}

// canPlaceOnTableau タブロー col に card を置けるか。空列は K のみ、それ以外は
// 交互色かつ降順 (value = top-1)。
func (g *DoubleKlondike) canPlaceOnTableau(card *Card, col int) bool {
	if card == nil {
		return false
	}
	pile := g.tableau[col]
	if len(pile) == 0 {
		return card.GetValue() == CardValueMax
	}
	top := pile[len(pile)-1]
	if !top.FaceUp || top.Card == nil {
		return false
	}
	return dkIsBlack(card) != dkIsBlack(top.Card) && card.GetValue() == top.Card.GetValue()-1
}

// canPlaceOnFoundation ファウンデーション fIdx に card を置けるか。空本は A のみ、
// それ以外は同スートかつ昇順。
func (g *DoubleKlondike) canPlaceOnFoundation(card *Card, fIdx int) bool {
	if card == nil {
		return false
	}
	pile := g.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == 1
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == top.GetValue()+1
}

// findFoundation card を置ける最初のファウンデーション番号を返す (なければ -1)。
func (g *DoubleKlondike) findFoundation(card *Card) int {
	for i := 0; i < DoubleKlondikeFoundationCnt; i++ {
		if g.canPlaceOnFoundation(card, i) {
			return i
		}
	}
	return -1
}

// autoFlipTableau 列 col の最上位が裏向きなら表に返す。
func (g *DoubleKlondike) autoFlipTableau(col int) {
	cards := g.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

// --- public actions ---

// Draw ストックからウェイストへ DrawCount 枚めくる。空ならウェイストをリサイクル。
func (g *DoubleKlondike) Draw() error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if len(g.stock) == 0 {
		if len(g.waste) == 0 {
			return errors.New("doubleklondike: no cards in stock or waste")
		}
		g.takeSnapshot()
		for i := len(g.waste) - 1; i >= 0; i-- {
			g.stock = append(g.stock, g.waste[i])
		}
		g.waste = nil
		g.appendLog("recycle", "ウェイストをストックに戻しました", nil)
		return nil
	}
	g.takeSnapshot()
	count := DoubleKlondikeDrawCount
	if count > len(g.stock) {
		count = len(g.stock)
	}
	for i := 0; i < count; i++ {
		card := g.stock[len(g.stock)-1]
		g.stock = g.stock[:len(g.stock)-1]
		g.waste = append(g.waste, card)
	}
	g.moveCount++
	g.appendLog("draw", "ストックからカードを引きました", nil)
	return nil
}

// MoveWasteToTableau ウェイストのトップをタブロー col へ移す。
func (g *DoubleKlondike) MoveWasteToTableau(col int) error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if col < 0 || col >= DoubleKlondikeTableauCnt {
		return errors.New("doubleklondike: invalid column")
	}
	if len(g.waste) == 0 {
		return errors.New("doubleklondike: waste is empty")
	}
	card := g.waste[len(g.waste)-1]
	if !g.canPlaceOnTableau(card, col) {
		return errors.New("doubleklondike: cannot place card on tableau")
	}
	g.takeSnapshot()
	g.waste = g.waste[:len(g.waste)-1]
	g.tableau[col] = append(g.tableau[col], &DoubleKlondikeTableauCard{Card: card, FaceUp: true})
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("ウェイスト→タブロー列%d", col), []*Card{card})
	return nil
}

// MoveWasteToFoundation ウェイストのトップをファウンデーションへ移す。
func (g *DoubleKlondike) MoveWasteToFoundation() error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if len(g.waste) == 0 {
		return errors.New("doubleklondike: waste is empty")
	}
	card := g.waste[len(g.waste)-1]
	fIdx := g.findFoundation(card)
	if fIdx < 0 {
		return errors.New("doubleklondike: cannot place card on foundation")
	}
	g.takeSnapshot()
	g.waste = g.waste[:len(g.waste)-1]
	g.foundation[fIdx] = append(g.foundation[fIdx], card)
	g.moveCount++
	g.appendLog("move", "ウェイスト→ファンデーション", []*Card{card})
	g.checkGameClear()
	return nil
}

// MoveTableauToTableau 列 fromCol の cardIndex 以降を列 toCol へ移す。
func (g *DoubleKlondike) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= DoubleKlondikeTableauCnt || toCol < 0 || toCol >= DoubleKlondikeTableauCnt {
		return errors.New("doubleklondike: invalid column")
	}
	if fromCol == toCol {
		return errors.New("doubleklondike: from and to columns are the same")
	}
	from := g.tableau[fromCol]
	if cardIndex < 0 || cardIndex >= len(from) {
		return errors.New("doubleklondike: invalid card index")
	}
	if !from[cardIndex].FaceUp {
		return errors.New("doubleklondike: card is face down")
	}
	if !g.canPlaceOnTableau(from[cardIndex].Card, toCol) {
		return errors.New("doubleklondike: cannot place card on tableau")
	}
	g.takeSnapshot()
	moving := from[cardIndex:]
	moved := make([]*Card, len(moving))
	for i, mc := range moving {
		g.tableau[toCol] = append(g.tableau[toCol], mc)
		moved[i] = mc.Card
	}
	g.tableau[fromCol] = from[:cardIndex]
	g.autoFlipTableau(fromCol)
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), moved)
	return nil
}

// MoveTableauToFoundation 列 col の最上位をファウンデーションへ移す。
func (g *DoubleKlondike) MoveTableauToFoundation(col int) error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if col < 0 || col >= DoubleKlondikeTableauCnt {
		return errors.New("doubleklondike: invalid column")
	}
	from := g.tableau[col]
	if len(from) == 0 {
		return errors.New("doubleklondike: tableau column is empty")
	}
	card := from[len(from)-1].Card
	fIdx := g.findFoundation(card)
	if fIdx < 0 {
		return errors.New("doubleklondike: cannot place card on foundation")
	}
	g.takeSnapshot()
	g.tableau[col] = from[:len(from)-1]
	g.foundation[fIdx] = append(g.foundation[fIdx], card)
	g.autoFlipTableau(col)
	g.moveCount++
	g.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	g.checkGameClear()
	return nil
}

// GiveUp 投了する。
func (g *DoubleKlondike) GiveUp() {
	if g.phase == DoubleKlondikePhasePlaying {
		g.phase = DoubleKlondikePhaseGameOver
		g.appendLog("giveup", "ギブアップしました", nil)
	}
}

// checkGameClear 8 ファウンデーション全完成 (104 枚) でクリア。
func (g *DoubleKlondike) checkGameClear() {
	total := 0
	for _, pile := range g.foundation {
		total += len(pile)
	}
	if total == 104 {
		g.phase = DoubleKlondikePhaseGameClear
		g.appendLog("clear", "クリア！", nil)
	}
}

// AllFaceUp 全タブローカードが表向きかを返す。
func (g *DoubleKlondike) AllFaceUp() bool {
	for col := 0; col < DoubleKlondikeTableauCnt; col++ {
		for _, tc := range g.tableau[col] {
			if !tc.FaceUp {
				return false
			}
		}
	}
	return true
}

// IsStalemate ストック・ウェイストが尽き、合法手も無い手詰まりかを返す。
func (g *DoubleKlondike) IsStalemate() bool {
	if g.phase != DoubleKlondikePhasePlaying {
		return false
	}
	if len(g.stock) > 0 || len(g.waste) > 0 {
		return false
	}
	return g.GetHint() == nil
}

// AutoComplete 全カードが表向きのとき、出せるファウンデーション手を出し切る。
func (g *DoubleKlondike) AutoComplete() error {
	if g.phase != DoubleKlondikePhasePlaying {
		return errors.New("doubleklondike: game is not in playing phase")
	}
	if !g.AllFaceUp() {
		return errors.New("doubleklondike: not all cards are face up")
	}
	g.takeSnapshot()
	for {
		moved := false
		for len(g.waste) > 0 {
			fIdx := g.findFoundation(g.waste[len(g.waste)-1])
			if fIdx < 0 {
				break
			}
			card := g.waste[len(g.waste)-1]
			g.waste = g.waste[:len(g.waste)-1]
			g.foundation[fIdx] = append(g.foundation[fIdx], card)
			g.moveCount++
			moved = true
		}
		for col := 0; col < DoubleKlondikeTableauCnt; col++ {
			if len(g.tableau[col]) == 0 {
				continue
			}
			card := g.tableau[col][len(g.tableau[col])-1].Card
			fIdx := g.findFoundation(card)
			if fIdx < 0 {
				continue
			}
			g.tableau[col] = g.tableau[col][:len(g.tableau[col])-1]
			g.foundation[fIdx] = append(g.foundation[fIdx], card)
			g.moveCount++
			moved = true
		}
		if !moved {
			break
		}
	}
	g.checkGameClear()
	return nil
}

// GetHint 推奨手を 1 つ返す。タブロー/ウェイスト→ファウンデーションを優先する。
func (g *DoubleKlondike) GetHint() *DoubleKlondikeHint {
	if g.phase != DoubleKlondikePhasePlaying {
		return nil
	}
	for col := 0; col < DoubleKlondikeTableauCnt; col++ {
		if len(g.tableau[col]) == 0 {
			continue
		}
		card := g.tableau[col][len(g.tableau[col])-1].Card
		if fIdx := g.findFoundation(card); fIdx >= 0 {
			return &DoubleKlondikeHint{FromZone: "tableau", FromCol: col, CardIndex: len(g.tableau[col]) - 1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	if len(g.waste) > 0 {
		if fIdx := g.findFoundation(g.waste[len(g.waste)-1]); fIdx >= 0 {
			return &DoubleKlondikeHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	for fromCol := 0; fromCol < DoubleKlondikeTableauCnt; fromCol++ {
		from := g.tableau[fromCol]
		firstFaceUp := -1
		for i, tc := range from {
			if tc.FaceUp {
				firstFaceUp = i
				break
			}
		}
		if firstFaceUp <= 0 {
			continue // all face up (no card to expose) or empty
		}
		card := from[firstFaceUp].Card
		for toCol := 0; toCol < DoubleKlondikeTableauCnt; toCol++ {
			if toCol != fromCol && g.canPlaceOnTableau(card, toCol) {
				return &DoubleKlondikeHint{FromZone: "tableau", FromCol: fromCol, CardIndex: firstFaceUp, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	if len(g.waste) > 0 {
		card := g.waste[len(g.waste)-1]
		for toCol := 0; toCol < DoubleKlondikeTableauCnt; toCol++ {
			if g.canPlaceOnTableau(card, toCol) {
				return &DoubleKlondikeHint{FromZone: "waste", FromCol: -1, CardIndex: -1, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	return nil
}

func (g *DoubleKlondike) appendLog(action, detail string, cards []*Card) {
	g.appendLogAt(g.moveCount, 0, action, detail, cards)
}

// --- undo ---

func (g *DoubleKlondike) takeSnapshot() {
	snap := &doubleKlondikeSnapshot{phase: g.phase, moveCount: g.moveCount}
	for i := 0; i < DoubleKlondikeTableauCnt; i++ {
		snap.tableau[i] = make([]*DoubleKlondikeTableauCard, len(g.tableau[i]))
		for j, tc := range g.tableau[i] {
			if tc != nil {
				// Deep-copy: FaceUp is mutated in-place (autoFlipTableau), so a
				// shallow pointer copy would corrupt the snapshot on Undo.
				snap.tableau[i][j] = &DoubleKlondikeTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
			}
		}
	}
	snap.stock = make([]*Card, len(g.stock))
	copy(snap.stock, g.stock)
	snap.waste = make([]*Card, len(g.waste))
	copy(snap.waste, g.waste)
	for i := 0; i < DoubleKlondikeFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(g.foundation[i]))
		copy(snap.foundation[i], g.foundation[i])
	}
	g.history = append(g.history, snap)
}

func (g *DoubleKlondike) restoreSnapshot(snap *doubleKlondikeSnapshot) {
	g.tableau = snap.tableau
	g.stock = snap.stock
	g.waste = snap.waste
	g.foundation = snap.foundation
	g.phase = snap.phase
	g.moveCount = snap.moveCount
}

// Undo 直近の 1 手を取り消す。
func (g *DoubleKlondike) Undo() error {
	if len(g.history) == 0 {
		return errors.New("doubleklondike: nothing to undo")
	}
	snap := g.history[len(g.history)-1]
	g.history = g.history[:len(g.history)-1]
	g.restoreSnapshot(snap)
	return nil
}

// CanUndo Undo 可能か。
func (g *DoubleKlondike) CanUndo() bool { return len(g.history) > 0 }

// UndoN n 回 Undo する。
func (g *DoubleKlondike) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if len(g.history) == 0 {
			break
		}
		if err := g.Undo(); err != nil {
			return err
		}
	}
	return nil
}

// --- accessors ---

// GetPhase 現在のフェーズ。
func (g *DoubleKlondike) GetPhase() DoubleKlondikePhase { return g.phase }

// SetPhase フェーズを設定する (テスト用)。
func (g *DoubleKlondike) SetPhase(p DoubleKlondikePhase) { g.phase = p }

// GetGameEndFlag プレイ中でなければ true。
func (g *DoubleKlondike) GetGameEndFlag() bool { return g.phase != DoubleKlondikePhasePlaying }

// GetMoveCount 累計手数。
func (g *DoubleKlondike) GetMoveCount() int { return g.moveCount }

// GetStockCount ストック残枚数。
func (g *DoubleKlondike) GetStockCount() int { return len(g.stock) }

// GetWaste ウェイストのカード一覧。
func (g *DoubleKlondike) GetWaste() []*Card { return g.waste }

// GetTableau タブローを返す。
func (g *DoubleKlondike) GetTableau() [DoubleKlondikeTableauCnt][]*DoubleKlondikeTableauCard {
	return g.tableau
}

// GetFoundation ファウンデーションを返す。
func (g *DoubleKlondike) GetFoundation() [DoubleKlondikeFoundationCnt][]*Card { return g.foundation }

// GetActionLog アクションログを返す。
func (g *DoubleKlondike) GetActionLog() []*ActionLogEntry { return g.actionLog }
