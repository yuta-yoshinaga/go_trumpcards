//go:build !js || !wasm || extra

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// AgnesPhase アグネス・ソレルゲームフェーズ
type AgnesPhase int

// Agnesのフェーズ定数
const (
	// AgnesPhasePlaying プレイ中
	AgnesPhasePlaying AgnesPhase = iota
	// AgnesPhaseGameClear ゲームクリア
	AgnesPhaseGameClear
	// AgnesPhaseGameOver ゲームオーバー
	AgnesPhaseGameOver
)

// AgnesTableauCnt タブローの列数
const AgnesTableauCnt = 7

// AgnesFoundationCnt ファンデーションの数
const AgnesFoundationCnt = 4

// AgnesTableauCard タブロー上のカード (Klondike流に表裏を持つ)
type AgnesTableauCard struct {
	Card   *Card `json:"c"`
	FaceUp bool  `json:"f"`
}

// AgnesHint ヒント
type AgnesHint struct {
	FromZone  string // "tableau"
	FromCol   int
	CardIndex int
	ToZone    string // "tableau" / "foundation"
	ToCol     int
}

// Agnes アグネス・ソレルゲームクラス
type Agnes struct {
	trumpCards *TrumpCards
	tableau    [AgnesTableauCnt][]*AgnesTableauCard
	stock      []*Card
	foundation [AgnesFoundationCnt][]*Card
	baseRank   int
	phase      AgnesPhase
	moveCount  int
	actionLog  []*ActionLogEntry
	history    []*agnesSnapshot
}

// agnesSnapshot アンドゥ用スナップショット
type agnesSnapshot struct {
	tableau    [AgnesTableauCnt][]*AgnesTableauCard
	stock      []*Card
	foundation [AgnesFoundationCnt][]*Card
	phase      AgnesPhase
	moveCount  int
}

// NewAgnes コンストラクタ
func NewAgnes(trumpCards *TrumpCards) *Agnes {
	return &Agnes{trumpCards: trumpCards}
}

// NewDefaultAgnes returns Agnes with a standard single 52-card deck.
// Used as the single source of truth for CUI, Web, and Worker construction sites.
func NewDefaultAgnes() *Agnes {
	return NewAgnes(NewTrumpCards(0))
}

// Reset ゲームリセット
func (a *Agnes) Reset() {
	a.trumpCards.Shuffle()
	a.phase = AgnesPhasePlaying
	a.moveCount = 0
	a.actionLog = nil
	a.history = nil

	// タブローに7列、Klondikeの階段配り: 列iにはi+1枚、最後(底)だけ表
	for i := 0; i < AgnesTableauCnt; i++ {
		a.tableau[i] = make([]*AgnesTableauCard, 0, i+1)
		for j := 0; j <= i; j++ {
			card := a.trumpCards.DrawCard()
			tc := &AgnesTableauCard{
				Card:   card,
				FaceUp: j == i,
			}
			a.tableau[i] = append(a.tableau[i], tc)
		}
	}

	// ファンデーション初期化
	for i := 0; i < AgnesFoundationCnt; i++ {
		a.foundation[i] = nil
	}

	// ベースランク決定: 28枚配った後の次の1枚をファンデーションの1つに置く (Canfield流)
	base := a.trumpCards.DrawCard()
	a.baseRank = base.GetValue()
	fIdx := base.GetDesign() - 1
	if fIdx >= 0 && fIdx < AgnesFoundationCnt {
		a.foundation[fIdx] = []*Card{base}
	}

	// 残りの23枚をストックへ
	a.stock = nil
	for a.trumpCards.GetRemainingCount() > 0 {
		a.stock = append(a.stock, a.trumpCards.DrawCard())
	}
}

// DealStock ストックから各列に1枚ずつ表向きに配る (最後の配りは2枚)
func (a *Agnes) DealStock() error {
	if a.phase != AgnesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if len(a.stock) == 0 {
		return errors.New("stock is empty")
	}
	a.takeSnapshot()
	dealt := make([]*Card, 0, AgnesTableauCnt)
	for col := 0; col < AgnesTableauCnt && len(a.stock) > 0; col++ {
		card := a.stock[len(a.stock)-1]
		a.stock = a.stock[:len(a.stock)-1]
		a.tableau[col] = append(a.tableau[col], &AgnesTableauCard{Card: card, FaceUp: true})
		dealt = append(dealt, card)
	}
	a.moveCount++
	a.appendLog("deal", "ストックから各列に1枚ずつ配りました", dealt)
	return nil
}

// MoveTableauToTableau タブロー間で移動 (1枚のみ、同色降順、ラップなし)
func (a *Agnes) MoveTableauToTableau(fromCol, cardIndex, toCol int) error {
	if a.phase != AgnesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if fromCol < 0 || fromCol >= AgnesTableauCnt {
		return errors.New("invalid from column")
	}
	if toCol < 0 || toCol >= AgnesTableauCnt {
		return errors.New("invalid to column")
	}
	if fromCol == toCol {
		return errors.New("from and to columns are the same")
	}
	fromCards := a.tableau[fromCol]
	if len(fromCards) == 0 {
		return errors.New("tableau column is empty")
	}
	// 末尾(底、表向き)カードのみ移動可能。cardIndex==-1 は末尾を指す。
	if cardIndex == -1 {
		cardIndex = len(fromCards) - 1
	}
	if cardIndex != len(fromCards)-1 {
		return errors.New("only the end card can be moved")
	}
	tc := fromCards[cardIndex]
	if !tc.FaceUp {
		return errors.New("card is face down")
	}
	card := tc.Card
	if !a.canPlaceOnTableau(card, toCol) {
		return errors.New("cannot place card on tableau")
	}
	a.takeSnapshot()
	a.tableau[toCol] = append(a.tableau[toCol], tc)
	a.tableau[fromCol] = fromCards[:cardIndex]
	a.autoFlipTableau(fromCol)
	a.moveCount++
	a.appendLog("move", fmt.Sprintf("タブロー列%d→タブロー列%d", fromCol, toCol), []*Card{card})
	return nil
}

// MoveTableauToFoundation タブローからファンデーションに移動
func (a *Agnes) MoveTableauToFoundation(col int) error {
	if a.phase != AgnesPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if col < 0 || col >= AgnesTableauCnt {
		return errors.New("invalid column")
	}
	from := a.tableau[col]
	if len(from) == 0 {
		return errors.New("tableau column is empty")
	}
	tc := from[len(from)-1]
	if !tc.FaceUp {
		return errors.New("card is face down")
	}
	card := tc.Card
	fIdx := card.GetDesign() - 1
	if fIdx < 0 || fIdx >= AgnesFoundationCnt {
		return errors.New("invalid card for foundation")
	}
	if !a.canPlaceOnFoundation(card, fIdx) {
		return errors.New("cannot place card on foundation")
	}
	a.takeSnapshot()
	a.tableau[col] = from[:len(from)-1]
	a.foundation[fIdx] = append(a.foundation[fIdx], card)
	a.autoFlipTableau(col)
	a.moveCount++
	a.appendLog("move", fmt.Sprintf("タブロー列%d→ファンデーション", col), []*Card{card})
	a.checkGameClear()
	return nil
}

// GiveUp ギブアップ
func (a *Agnes) GiveUp() {
	if a.phase == AgnesPhasePlaying {
		a.phase = AgnesPhaseGameOver
		a.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetHint ヒントを取得
func (a *Agnes) GetHint() *AgnesHint {
	if a.phase != AgnesPhasePlaying {
		return nil
	}
	// 優先度1: タブローからファンデーションへ
	for col := 0; col < AgnesTableauCnt; col++ {
		if len(a.tableau[col]) == 0 {
			continue
		}
		tc := a.tableau[col][len(a.tableau[col])-1]
		if !tc.FaceUp {
			continue
		}
		card := tc.Card
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < AgnesFoundationCnt && a.canPlaceOnFoundation(card, fIdx) {
			return &AgnesHint{FromZone: "tableau", FromCol: col, CardIndex: len(a.tableau[col]) - 1, ToZone: "foundation", ToCol: fIdx}
		}
	}
	// 優先度2: タブローからタブローへ
	for fromCol := 0; fromCol < AgnesTableauCnt; fromCol++ {
		if len(a.tableau[fromCol]) == 0 {
			continue
		}
		tc := a.tableau[fromCol][len(a.tableau[fromCol])-1]
		if !tc.FaceUp {
			continue
		}
		card := tc.Card
		for toCol := 0; toCol < AgnesTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			if a.canPlaceOnTableau(card, toCol) {
				return &AgnesHint{FromZone: "tableau", FromCol: fromCol, CardIndex: len(a.tableau[fromCol]) - 1, ToZone: "tableau", ToCol: toCol}
			}
		}
	}
	return nil
}

// IsStalemate は合法手が 1 つも無いかを返す。ストックが残っていれば false。
//
// **判定は GetHint をそのまま使う。**ヒントが探す手（タブロー→ファンデーション、
// タブロー→タブロー）が合法手の全てなので、別のスキャンを書くと「手詰まり」と
// 言いながらヒントが手を返す状態が作れる (#4830)。
func (a *Agnes) IsStalemate() bool {
	if a.phase != AgnesPhasePlaying {
		return false
	}
	if len(a.stock) > 0 {
		return false
	}
	return a.GetHint() == nil
}

// --- Getters / Setters ---

// GetPhase フェーズ取得
func (a *Agnes) GetPhase() AgnesPhase { return a.phase }

// SetPhase フェーズ設定
func (a *Agnes) SetPhase(p AgnesPhase) { a.phase = p }

// GetMoveCount 移動回数取得
func (a *Agnes) GetMoveCount() int { return a.moveCount }

// GetStockCount ストック残枚数
func (a *Agnes) GetStockCount() int { return len(a.stock) }

// GetTableau タブロー取得
func (a *Agnes) GetTableau() [AgnesTableauCnt][]*AgnesTableauCard { return a.tableau }

// GetFoundation ファンデーション取得
func (a *Agnes) GetFoundation() [AgnesFoundationCnt][]*Card { return a.foundation }

// GetActionLog 棋譜取得
func (a *Agnes) GetActionLog() []*ActionLogEntry { return a.actionLog }

// GetGameEndFlag returns true once the game has left the playing phase.
func (a *Agnes) GetGameEndFlag() bool { return a.phase != AgnesPhasePlaying }

// GetBaseRank ベースランク取得
func (a *Agnes) GetBaseRank() int { return a.baseRank }

// SetBaseRank ベースランク設定 (テスト用)
func (a *Agnes) SetBaseRank(r int) { a.baseRank = r }

// SetTableau タブロー設定 (テスト用)
func (a *Agnes) SetTableau(t [AgnesTableauCnt][]*AgnesTableauCard) { a.tableau = t }

// SetStock ストック設定 (テスト用)
func (a *Agnes) SetStock(s []*Card) { a.stock = s }

// SetFoundation ファンデーション設定 (テスト用)
func (a *Agnes) SetFoundation(f [AgnesFoundationCnt][]*Card) { a.foundation = f }

// Undo 直前の操作を取り消す
func (a *Agnes) Undo() error {
	if a.phase != AgnesPhasePlaying {
		return errors.New("cannot undo: game is not in playing phase")
	}
	if len(a.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := a.history[len(a.history)-1]
	a.history = a.history[:len(a.history)-1]
	a.restoreSnapshot(snap)
	return nil
}

// CanUndo アンドゥ可能か
func (a *Agnes) CanUndo() bool {
	return len(a.history) > 0 && a.phase == AgnesPhasePlaying
}

// UndoN n回連続アンドゥ
func (a *Agnes) UndoN(n int) error {
	for i := 0; i < n; i++ {
		if err := a.Undo(); err != nil {
			return fmt.Errorf("undo step %d failed: %w", i+1, err)
		}
	}
	return nil
}

// --- Private helpers ---

// canPlaceOnTableau 同色降順(ラップなし)、空列は手動配置不可
func (a *Agnes) canPlaceOnTableau(card *Card, col int) bool {
	colCards := a.tableau[col]
	if len(colCards) == 0 {
		// 空列はストックの配りでしか埋まらない (手動配置不可)
		return false
	}
	top := colCards[len(colCards)-1].Card
	return a.isSameColor(card, top) && card.GetValue() == top.GetValue()-1
}

// canPlaceOnFoundation ベースランクから上昇(K→Aラップ)、同スート
func (a *Agnes) canPlaceOnFoundation(card *Card, fIdx int) bool {
	pile := a.foundation[fIdx]
	if len(pile) == 0 {
		return card.GetValue() == a.baseRank
	}
	top := pile[len(pile)-1]
	return card.GetDesign() == top.GetDesign() && card.GetValue() == a.nextRank(top.GetValue())
}

func (a *Agnes) nextRank(r int) int {
	return (r % 13) + 1
}

func (a *Agnes) isSameColor(x, y *Card) bool {
	return a.isBlack(x) == a.isBlack(y)
}

func (a *Agnes) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// autoFlipTableau タブローの末尾の裏カードを自動フリップ
func (a *Agnes) autoFlipTableau(col int) {
	cards := a.tableau[col]
	if len(cards) > 0 && !cards[len(cards)-1].FaceUp {
		cards[len(cards)-1].FaceUp = true
	}
}

func (a *Agnes) checkGameClear() {
	for i := 0; i < AgnesFoundationCnt; i++ {
		if len(a.foundation[i]) != CardValueMax {
			return
		}
	}
	a.phase = AgnesPhaseGameClear
}

func (a *Agnes) takeSnapshot() {
	snap := &agnesSnapshot{
		phase:     a.phase,
		moveCount: a.moveCount,
	}
	for i := 0; i < AgnesTableauCnt; i++ {
		snap.tableau[i] = make([]*AgnesTableauCard, len(a.tableau[i]))
		for j, tc := range a.tableau[i] {
			snap.tableau[i][j] = &AgnesTableauCard{Card: tc.Card, FaceUp: tc.FaceUp}
		}
	}
	snap.stock = make([]*Card, len(a.stock))
	copy(snap.stock, a.stock)
	for i := 0; i < AgnesFoundationCnt; i++ {
		snap.foundation[i] = make([]*Card, len(a.foundation[i]))
		copy(snap.foundation[i], a.foundation[i])
	}
	a.history = append(a.history, snap)
}

func (a *Agnes) restoreSnapshot(snap *agnesSnapshot) {
	a.tableau = snap.tableau
	a.stock = snap.stock
	a.foundation = snap.foundation
	a.phase = snap.phase
	a.moveCount = snap.moveCount
}

func (a *Agnes) appendLog(actionType, detail string, cards []*Card) {
	a.actionLog = append(a.actionLog, &ActionLogEntry{
		TurnNumber: a.moveCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// agnesJSON is the JSON wire format for Agnes.
type agnesJSON struct {
	TrumpCards *TrumpCards                          `json:"tc"`
	Tableau    [AgnesTableauCnt][]*AgnesTableauCard `json:"tb"`
	Stock      []*Card                              `json:"st"`
	Foundation [AgnesFoundationCnt][]*Card          `json:"fd"`
	BaseRank   int                                  `json:"br"`
	Phase      AgnesPhase                           `json:"ps"`
	MoveCount  int                                  `json:"mc"`
	ActionLog  []*ActionLogEntry                    `json:"al"`
	History    []*agnesSnapshot                     `json:"hi,omitempty"`
}

// agnesSnapshotJSON is the wire format for a single undo snapshot.
type agnesSnapshotJSON struct {
	Tableau    [AgnesTableauCnt][]*AgnesTableauCard `json:"tb"`
	Stock      []*Card                              `json:"st"`
	Foundation [AgnesFoundationCnt][]*Card          `json:"fd"`
	Phase      AgnesPhase                           `json:"ps"`
	MoveCount  int                                  `json:"mc"`
}

// MarshalJSON implements json.Marshaler for agnesSnapshot.
func (s *agnesSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(agnesSnapshotJSON{
		Tableau:    s.tableau,
		Stock:      s.stock,
		Foundation: s.foundation,
		Phase:      s.phase,
		MoveCount:  s.moveCount,
	})
}

// agnesMaxSliceLen caps slice sizes during deserialisation.
const agnesMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler for agnesSnapshot.
func (s *agnesSnapshot) UnmarshalJSON(data []byte) error {
	var j agnesSnapshotJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > agnesMaxSliceLen {
		return fmt.Errorf("agnes: snapshot array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > agnesMaxSliceLen {
			return fmt.Errorf("agnes: snapshot tableau column exceeds maximum allowed size")
		}
		for _, tc := range col {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("agnes: snapshot tableau contains nil element")
			}
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > agnesMaxSliceLen {
			return fmt.Errorf("agnes: snapshot foundation pile exceeds maximum allowed size")
		}
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("agnes: snapshot foundation contains nil element")
			}
		}
	}
	for _, c := range j.Stock {
		if c == nil {
			return fmt.Errorf("agnes: snapshot stock contains nil element")
		}
	}
	s.tableau = j.Tableau
	s.stock = j.Stock
	if s.stock == nil {
		s.stock = make([]*Card, 0)
	}
	s.foundation = j.Foundation
	s.phase = j.Phase
	s.moveCount = j.MoveCount
	return nil
}

// MarshalJSON implements json.Marshaler.
func (a *Agnes) MarshalJSON() ([]byte, error) {
	return json.Marshal(agnesJSON{
		TrumpCards: a.trumpCards,
		Tableau:    a.tableau,
		Stock:      a.stock,
		Foundation: a.foundation,
		BaseRank:   a.baseRank,
		Phase:      a.phase,
		MoveCount:  a.moveCount,
		ActionLog:  a.actionLog,
		History:    a.history,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (a *Agnes) UnmarshalJSON(data []byte) error {
	var j agnesJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.Stock) > agnesMaxSliceLen || len(j.ActionLog) > agnesMaxSliceLen ||
		len(j.History) > agnesMaxSliceLen {
		return fmt.Errorf("agnes: input array exceeds maximum allowed size")
	}
	for _, col := range j.Tableau {
		if len(col) > agnesMaxSliceLen {
			return fmt.Errorf("agnes: tableau column exceeds maximum allowed size")
		}
		for _, tc := range col {
			if tc == nil || tc.Card == nil {
				return fmt.Errorf("agnes: tableau contains nil element")
			}
		}
	}
	for _, pile := range j.Foundation {
		if len(pile) > agnesMaxSliceLen {
			return fmt.Errorf("agnes: foundation pile exceeds maximum allowed size")
		}
		for _, c := range pile {
			if c == nil {
				return fmt.Errorf("agnes: foundation contains nil element")
			}
		}
	}
	for _, c := range j.Stock {
		if c == nil {
			return fmt.Errorf("agnes: stock contains nil element")
		}
	}
	a.trumpCards = j.TrumpCards
	if a.trumpCards == nil {
		a.trumpCards = NewTrumpCards(0)
	}
	a.tableau = j.Tableau
	a.stock = j.Stock
	if a.stock == nil {
		a.stock = make([]*Card, 0)
	}
	a.foundation = j.Foundation
	a.baseRank = j.BaseRank
	a.phase = j.Phase
	a.moveCount = j.MoveCount
	a.actionLog = j.ActionLog
	if a.actionLog == nil {
		a.actionLog = make([]*ActionLogEntry, 0)
	}
	a.history = j.History
	if a.history == nil {
		a.history = make([]*agnesSnapshot, 0)
	}
	return nil
}
