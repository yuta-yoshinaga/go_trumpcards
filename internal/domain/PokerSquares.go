// Package domain provides core game domain models.
package domain

import (
	"encoding/json"
	"errors"
	"fmt"
)

// PokerSquaresPhase はポーカー・スクエアズのフェーズを表す。
type PokerSquaresPhase int

// PokerSquaresのフェーズ定数
const (
	// PokerSquaresPhasePlaying プレイ中
	PokerSquaresPhasePlaying PokerSquaresPhase = iota
	// PokerSquaresPhaseComplete 完了
	PokerSquaresPhaseComplete
)

// PokerSquaresGridSize はグリッドの一辺のサイズ。
const PokerSquaresGridSize = 5

// PokerSquaresTotalCells は総セル数 (5x5=25)。
const PokerSquaresTotalCells = PokerSquaresGridSize * PokerSquaresGridSize

// pokerSquaresScoreTable は American scoring system のハンドランク -> スコア のマップ。
var pokerSquaresScoreTable = map[int]int{
	PokerHandHighCard:      0,
	PokerHandOnePair:       2,
	PokerHandTwoPair:       5,
	PokerHandThreeOfAKind:  10,
	PokerHandStraight:      15,
	PokerHandFlush:         20,
	PokerHandFullHouse:     25,
	PokerHandFourOfAKind:   50,
	PokerHandStraightFlush: 75,
	PokerHandRoyalFlush:    100,
}

// PokerSquares はポーカー・スクエアズのゲーム状態を表す。
type PokerSquares struct {
	trumpCards  *TrumpCards
	board       [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	currentCard *Card
	placedCount int
	phase       PokerSquaresPhase
	actionLog   []*ActionLogEntry
	history     []*pokerSquaresSnapshot
}

// pokerSquaresSnapshot はアンドゥ用の状態スナップショット。
type pokerSquaresSnapshot struct {
	board       [PokerSquaresGridSize][PokerSquaresGridSize]*Card
	currentCard *Card
	placedCount int
	phase       PokerSquaresPhase
	deckDrawCnt int
	actionLogLn int
}

// NewPokerSquares はコンストラクタ。
func NewPokerSquares(tc *TrumpCards) *PokerSquares {
	return &PokerSquares{trumpCards: tc}
}

// Reset はゲームを初期化する。デッキをシャッフルし、最初のカードを引く。
func (p *PokerSquares) Reset() {
	p.trumpCards.Shuffle()
	p.board = [PokerSquaresGridSize][PokerSquaresGridSize]*Card{}
	p.placedCount = 0
	p.phase = PokerSquaresPhasePlaying
	p.actionLog = nil
	p.history = nil
	p.currentCard = p.trumpCards.DrawCard()
}

// Place は現在のカードを指定セルに置く。
func (p *PokerSquares) Place(row, col int) error {
	if p.phase != PokerSquaresPhasePlaying {
		return errors.New("game is not in playing phase")
	}
	if row < 0 || row >= PokerSquaresGridSize || col < 0 || col >= PokerSquaresGridSize {
		return errors.New("invalid cell position")
	}
	if p.board[row][col] != nil {
		return errors.New("cell is already occupied")
	}
	if p.currentCard == nil {
		return errors.New("no current card to place")
	}
	p.takeSnapshot()
	placed := p.currentCard
	p.board[row][col] = placed
	p.placedCount++
	p.appendLog("place", fmt.Sprintf("(%d,%d) に配置", row, col), []*Card{placed})
	if p.placedCount >= PokerSquaresTotalCells {
		p.currentCard = nil
		p.phase = PokerSquaresPhaseComplete
	} else {
		p.currentCard = p.trumpCards.DrawCard()
	}
	return nil
}

// Undo は直前の配置を取り消す。
func (p *PokerSquares) Undo() error {
	if len(p.history) == 0 {
		return errors.New("cannot undo: no history")
	}
	snap := p.history[len(p.history)-1]
	p.history = p.history[:len(p.history)-1]
	for i := snap.deckDrawCnt; i < p.trumpCards.deckDrawCnt; i++ {
		p.trumpCards.deck[i].SetDraw(false)
	}
	p.trumpCards.deckDrawCnt = snap.deckDrawCnt
	p.board = snap.board
	p.currentCard = snap.currentCard
	p.placedCount = snap.placedCount
	p.phase = snap.phase
	if len(p.actionLog) > snap.actionLogLn {
		p.actionLog = p.actionLog[:snap.actionLogLn]
	}
	return nil
}

// CanUndo はアンドゥ可能かを返す。
func (p *PokerSquares) CanUndo() bool {
	return len(p.history) > 0
}

// GiveUp はゲームを放棄する。
func (p *PokerSquares) GiveUp() {
	if p.phase == PokerSquaresPhasePlaying {
		p.phase = PokerSquaresPhaseComplete
		p.currentCard = nil
		p.appendLog("giveup", "ギブアップしました", nil)
	}
}

// GetPhase はフェーズを返す。
func (p *PokerSquares) GetPhase() PokerSquaresPhase { return p.phase }

// SetPhase はフェーズを設定する (テスト用)。
func (p *PokerSquares) SetPhase(phase PokerSquaresPhase) { p.phase = phase }

// GetBoard はボードを返す。
func (p *PokerSquares) GetBoard() [PokerSquaresGridSize][PokerSquaresGridSize]*Card {
	return p.board
}

// SetBoard はボードを設定する (テスト用)。
func (p *PokerSquares) SetBoard(b [PokerSquaresGridSize][PokerSquaresGridSize]*Card) {
	p.board = b
}

// GetCurrentCard は次に配置するカードを返す。
func (p *PokerSquares) GetCurrentCard() *Card { return p.currentCard }

// SetCurrentCard は次に配置するカードを設定する (テスト用)。
func (p *PokerSquares) SetCurrentCard(c *Card) { p.currentCard = c }

// GetPlacedCount は配置済みカード枚数を返す。
func (p *PokerSquares) GetPlacedCount() int { return p.placedCount }

// SetPlacedCount は配置済みカード枚数を設定する (テスト用)。
func (p *PokerSquares) SetPlacedCount(n int) { p.placedCount = n }

// GetActionLog は棋譜を返す。
func (p *PokerSquares) GetActionLog() []*ActionLogEntry { return p.actionLog }

// IsComplete はゲームが完了したかを返す。
func (p *PokerSquares) IsComplete() bool {
	return p.placedCount >= PokerSquaresTotalCells
}

// EvaluateRow は指定行を 5 枚ポーカーハンドとして評価し、ランク定数を返す。
// 行が埋まっていない場合は -1 を返す。
func (p *PokerSquares) EvaluateRow(r int) int {
	if r < 0 || r >= PokerSquaresGridSize {
		return -1
	}
	cards := make([]*Card, 0, PokerSquaresGridSize)
	for c := range PokerSquaresGridSize {
		if p.board[r][c] == nil {
			return -1
		}
		cards = append(cards, p.board[r][c])
	}
	return evalFiveCardHand(cards)
}

// EvaluateCol は指定列を 5 枚ポーカーハンドとして評価し、ランク定数を返す。
// 列が埋まっていない場合は -1 を返す。
func (p *PokerSquares) EvaluateCol(c int) int {
	if c < 0 || c >= PokerSquaresGridSize {
		return -1
	}
	cards := make([]*Card, 0, PokerSquaresGridSize)
	for r := range PokerSquaresGridSize {
		if p.board[r][c] == nil {
			return -1
		}
		cards = append(cards, p.board[r][c])
	}
	return evalFiveCardHand(cards)
}

// RowScore は指定行の得点 (American scoring) を返す。
func (p *PokerSquares) RowScore(r int) int {
	rank := p.EvaluateRow(r)
	if rank < 0 {
		return 0
	}
	return pokerSquaresRankToScore(rank)
}

// ColScore は指定列の得点 (American scoring) を返す。
func (p *PokerSquares) ColScore(c int) int {
	rank := p.EvaluateCol(c)
	if rank < 0 {
		return 0
	}
	return pokerSquaresRankToScore(rank)
}

// TotalScore は 10 ハンドの合計得点を返す。
func (p *PokerSquares) TotalScore() int {
	total := 0
	for i := range PokerSquaresGridSize {
		total += p.RowScore(i)
		total += p.ColScore(i)
	}
	return total
}

// pokerSquaresRankToScore はハンドランクを得点に変換する。
func pokerSquaresRankToScore(rank int) int {
	if s, ok := pokerSquaresScoreTable[rank]; ok {
		return s
	}
	// FiveOfAKind (ジョーカー未使用の 52 枚デッキでは発生しない) は 4K 相当にマップ
	if rank == PokerHandFiveOfAKind {
		return pokerSquaresScoreTable[PokerHandFourOfAKind]
	}
	return 0
}

// takeSnapshot は現在の状態をスナップショットとして保存する。
func (p *PokerSquares) takeSnapshot() {
	p.history = append(p.history, &pokerSquaresSnapshot{
		board:       p.board,
		currentCard: p.currentCard,
		placedCount: p.placedCount,
		phase:       p.phase,
		deckDrawCnt: p.trumpCards.deckDrawCnt,
		actionLogLn: len(p.actionLog),
	})
}

// appendLog は棋譜エントリを追加する。
func (p *PokerSquares) appendLog(actionType, detail string, cards []*Card) {
	p.actionLog = append(p.actionLog, &ActionLogEntry{
		TurnNumber: p.placedCount,
		PlayerIdx:  0,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// pokerSquaresJSON はシリアライズ用のワイヤーフォーマット。
type pokerSquaresJSON struct {
	TrumpCards  *TrumpCards                                       `json:"tc"`
	Board       [PokerSquaresGridSize][PokerSquaresGridSize]*Card `json:"bd"`
	CurrentCard *Card                                             `json:"cc"`
	PlacedCount int                                               `json:"pc"`
	Phase       PokerSquaresPhase                                 `json:"ps"`
	ActionLog   []*ActionLogEntry                                 `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (p *PokerSquares) MarshalJSON() ([]byte, error) {
	return json.Marshal(pokerSquaresJSON{
		TrumpCards:  p.trumpCards,
		Board:       p.board,
		CurrentCard: p.currentCard,
		PlacedCount: p.placedCount,
		Phase:       p.phase,
		ActionLog:   p.actionLog,
	})
}

// pokerSquaresMaxSliceLen はデシリアライズ時のスライスサイズ上限。
const pokerSquaresMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (p *PokerSquares) UnmarshalJSON(data []byte) error {
	var j pokerSquaresJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.ActionLog) > pokerSquaresMaxSliceLen {
		return fmt.Errorf("pokersquares: input array exceeds maximum allowed size")
	}
	p.trumpCards = j.TrumpCards
	if p.trumpCards == nil {
		p.trumpCards = NewTrumpCards(0)
	}
	p.board = j.Board
	p.currentCard = j.CurrentCard
	p.placedCount = j.PlacedCount
	p.phase = j.Phase
	p.actionLog = j.ActionLog
	if p.actionLog == nil {
		p.actionLog = make([]*ActionLogEntry, 0)
	}
	p.history = nil
	return nil
}
