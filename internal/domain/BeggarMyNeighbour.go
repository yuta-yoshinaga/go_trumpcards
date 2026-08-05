package domain

import (
	"encoding/json"
	"fmt"
)

// BeggarMyNeighbourPlayerCnt Beggar-My-Neighbour ゲームのプレイヤー数 (人間 + CPU)
const BeggarMyNeighbourPlayerCnt = 2

// BeggarMyNeighbourPhase Beggar-My-Neighbour ゲームのフェーズ
type BeggarMyNeighbourPhase int

// Beggar-My-Neighbour ゲームのフェーズ定数
const (
	// BeggarMyNeighbourPhasePlay 通常プレイ: 現在のプレイヤーが山札のトップを場に出す
	BeggarMyNeighbourPhasePlay BeggarMyNeighbourPhase = 0
	// BeggarMyNeighbourPhasePayPenalty ペナルティ支払い中
	BeggarMyNeighbourPhasePayPenalty BeggarMyNeighbourPhase = 1
	// BeggarMyNeighbourPhaseCollect 場の山を回収する
	BeggarMyNeighbourPhaseCollect BeggarMyNeighbourPhase = 2
	// BeggarMyNeighbourPhaseGameEnd ゲーム終了
	BeggarMyNeighbourPhaseGameEnd BeggarMyNeighbourPhase = 3
)

// beggarMyNeighbourPenaltyValue ペナルティカードの支払い枚数を返す (非ペナルティは 0)
// J=1, Q=2, K=3, A=4
func beggarMyNeighbourPenaltyValue(c *Card) int {
	if c == nil {
		return 0
	}
	switch c.GetValue() {
	case 1:
		return 4
	case 11:
		return 1
	case 12:
		return 2
	case 13:
		return 3
	}
	return 0
}

// BeggarMyNeighbour Beggar-My-Neighbour ゲームクラス
type BeggarMyNeighbour struct {
	trumpCards       *TrumpCards
	players          [BeggarMyNeighbourPlayerCnt]*BeggarMyNeighbourPlayer
	config           BeggarMyNeighbourConfig
	phase            BeggarMyNeighbourPhase
	centralPile      []*Card
	currentPlayerIdx int
	penaltyOwnerIdx  int
	penaltyRemaining int
	lastCardPlayed   *Card
	gameEndFlag      bool
	winnerIdx        int
	roundsPlayed     int
	actionLog        []*ActionLogEntry
}

// NewBeggarMyNeighbour コンストラクタ
func NewBeggarMyNeighbour(trumpCards *TrumpCards, players []*BeggarMyNeighbourPlayer, config BeggarMyNeighbourConfig) *BeggarMyNeighbour {
	g := &BeggarMyNeighbour{
		trumpCards:      trumpCards,
		config:          config,
		penaltyOwnerIdx: -1,
		winnerIdx:       -1,
	}
	for i := 0; i < BeggarMyNeighbourPlayerCnt && i < len(players); i++ {
		g.players[i] = players[i]
	}
	return g
}

// NewDefaultBeggarMyNeighbour returns BeggarMyNeighbour with the standard 2-player setup (1 human, 1 CPU)
// and DefaultBeggarMyNeighbourConfig.
func NewDefaultBeggarMyNeighbour() *BeggarMyNeighbour {
	players := []*BeggarMyNeighbourPlayer{
		NewBeggarMyNeighbourPlayer(true),
		NewBeggarMyNeighbourPlayer(false),
	}
	return NewBeggarMyNeighbour(NewTrumpCards(0), players, DefaultBeggarMyNeighbourConfig())
}

// Reset ゲームをリセットして新しいゲームを開始する
func (g *BeggarMyNeighbour) Reset() {
	g.phase = BeggarMyNeighbourPhasePlay
	g.centralPile = nil
	g.currentPlayerIdx = 0
	g.penaltyOwnerIdx = -1
	g.penaltyRemaining = 0
	g.lastCardPlayed = nil
	g.gameEndFlag = false
	g.winnerIdx = -1
	g.roundsPlayed = 0
	g.actionLog = nil

	for _, p := range g.players {
		p.Reset()
		p.ResetPiles()
		p.SetIsFinished(false)
	}

	g.trumpCards.Shuffle()

	cards := make([]*Card, 0, CardCnt)
	for range CardCnt {
		c := g.trumpCards.DrawCard()
		if c != nil {
			cards = append(cards, c)
		}
	}

	// Deal an equal share to each player. CardCnt (52) divides evenly by the
	// 2-player count today; the divisor documents that any odd remainder is
	// intentionally left undealt.
	half := len(cards) / BeggarMyNeighbourPlayerCnt
	for pi := range BeggarMyNeighbourPlayerCnt {
		start := pi * half
		g.players[pi].AddToDrawPile(cards[start : start+half]...)
	}
}

// Step 状態機械を1ステップ進める
func (g *BeggarMyNeighbour) Step() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	switch g.phase {
	case BeggarMyNeighbourPhasePlay:
		return g.stepPlay()
	case BeggarMyNeighbourPhasePayPenalty:
		return g.stepPayPenalty()
	case BeggarMyNeighbourPhaseCollect:
		return g.stepCollect()
	}
	return ErrWrongPhase
}

// beggarMyNeighbourAutoPlayMaxSteps caps the AutoPlay loop.
// Declared as a var so tests can lower it to exercise the cap-hit branch.
var beggarMyNeighbourAutoPlayMaxSteps = 200000

// AutoPlay 自動プレイ（決着まで Step を繰り返す）
func (g *BeggarMyNeighbour) AutoPlay() error {
	if g.gameEndFlag {
		return ErrGameEnded
	}
	for i := range beggarMyNeighbourAutoPlayMaxSteps {
		if g.gameEndFlag {
			return nil
		}
		if err := g.Step(); err != nil {
			return fmt.Errorf("step %d: %w", i+1, err)
		}
	}
	return fmt.Errorf("auto-play reached maximum steps (%d) without finishing", beggarMyNeighbourAutoPlayMaxSteps)
}

// stepPlay 通常プレイフェーズ: 現在プレイヤーがカードを1枚場に出す
func (g *BeggarMyNeighbour) stepPlay() error {
	c := g.players[g.currentPlayerIdx].DrawOne()
	if c == nil {
		// Current player cannot turn up a card → they lose; opponent wins.
		g.finishWithWinner(1 - g.currentPlayerIdx)
		return nil
	}
	g.centralPile = append(g.centralPile, c)
	g.lastCardPlayed = c
	g.appendLog(g.currentPlayerIdx, "play", "play card", []*Card{c})

	if pv := beggarMyNeighbourPenaltyValue(c); pv > 0 {
		g.penaltyOwnerIdx = g.currentPlayerIdx
		g.penaltyRemaining = pv
		g.currentPlayerIdx = 1 - g.currentPlayerIdx
		g.phase = BeggarMyNeighbourPhasePayPenalty
	} else {
		g.currentPlayerIdx = 1 - g.currentPlayerIdx
		// stay in Play phase
		if !g.players[0].HasCards() || !g.players[1].HasCards() {
			g.finishByTotal()
		}
	}
	return nil
}

// stepPayPenalty ペナルティ支払いフェーズ: 現在プレイヤーが1枚払う
func (g *BeggarMyNeighbour) stepPayPenalty() error {
	c := g.players[g.currentPlayerIdx].DrawOne()
	if c == nil {
		// Payer cannot complete the penalty → the penalty owner wins the game.
		g.finishWithWinner(g.penaltyOwnerIdx)
		return nil
	}
	g.centralPile = append(g.centralPile, c)
	g.lastCardPlayed = c
	g.penaltyRemaining--
	g.appendLog(g.currentPlayerIdx, "pay", fmt.Sprintf("pay penalty (%d remaining)", g.penaltyRemaining), []*Card{c})

	if pv := beggarMyNeighbourPenaltyValue(c); pv > 0 {
		// New penalty card: flip obligation to original payer
		g.penaltyOwnerIdx = g.currentPlayerIdx
		g.penaltyRemaining = pv
		g.currentPlayerIdx = 1 - g.currentPlayerIdx
		// stay in PayPenalty phase
	} else if g.penaltyRemaining == 0 {
		// All paid: penaltyOwner collects
		g.phase = BeggarMyNeighbourPhaseCollect
	} else {
		// continue paying (no penalty card, still remaining); if the payer is
		// now out of cards they cannot finish paying → the owner wins.
		if !g.players[g.currentPlayerIdx].HasCards() {
			g.finishWithWinner(g.penaltyOwnerIdx)
		}
	}
	return nil
}

// stepCollect 回収フェーズ: penaltyOwner が場の山を全部もらう
func (g *BeggarMyNeighbour) stepCollect() error {
	collector := g.penaltyOwnerIdx

	g.players[collector].AddToDiscardPile(g.centralPile...)
	g.appendLog(collector, "collect", fmt.Sprintf("+%d cards", len(g.centralPile)), nil)

	g.centralPile = nil
	g.penaltyOwnerIdx = -1
	g.penaltyRemaining = 0
	// The collector (the player who laid the unanswered penalty card) leads the
	// next round by turning up the first card.
	g.currentPlayerIdx = collector
	g.roundsPlayed++
	g.phase = BeggarMyNeighbourPhasePlay

	if !g.players[0].HasCards() || !g.players[1].HasCards() {
		g.finishByTotal()
	} else if g.roundsPlayed >= g.config.MaxRounds {
		g.finishByTotal()
	}
	return nil
}

// finishWithWinner ends the game with an explicit winner (the loser ran out of
// cards mid-play or mid-payment). winnerIdx must be a valid player index.
func (g *BeggarMyNeighbour) finishWithWinner(winnerIdx int) {
	g.gameEndFlag = true
	g.phase = BeggarMyNeighbourPhaseGameEnd
	g.winnerIdx = winnerIdx
	if winnerIdx >= 0 && winnerIdx < BeggarMyNeighbourPlayerCnt {
		g.players[winnerIdx].SetIsFinished(true)
	}
}

// finishByTotal 保有枚数の多い方を勝者としてゲームを終了する。
// 同数 (場の山が回収されないまま上限ラウンドに達した場合など) は引き分けとし、
// winnerIdx を -1 のままにする。
func (g *BeggarMyNeighbour) finishByTotal() {
	g.gameEndFlag = true
	g.phase = BeggarMyNeighbourPhaseGameEnd
	t0 := g.players[0].TotalCards()
	t1 := g.players[1].TotalCards()
	switch {
	case t0 > t1:
		g.winnerIdx = 0
	case t1 > t0:
		g.winnerIdx = 1
	default:
		g.winnerIdx = -1 // genuine draw; presenters handle winnerIdx == -1
	}
	if g.winnerIdx >= 0 {
		g.players[g.winnerIdx].SetIsFinished(true)
	}
}

// appendLog 棋譜にエントリを追加する
func (g *BeggarMyNeighbour) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	g.actionLog = append(g.actionLog, &ActionLogEntry{
		TurnNumber: len(g.actionLog) + 1,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}

// --- Getters ---

// GetPhase フェーズ取得
func (g *BeggarMyNeighbour) GetPhase() BeggarMyNeighbourPhase { return g.phase }

// GetGameEndFlag ゲーム終了フラグ取得
func (g *BeggarMyNeighbour) GetGameEndFlag() bool { return g.gameEndFlag }

// GetWinnerIdx 勝者インデックス取得
func (g *BeggarMyNeighbour) GetWinnerIdx() int { return g.winnerIdx }

// GetPlayerCnt プレイヤー数取得
func (g *BeggarMyNeighbour) GetPlayerCnt() int { return BeggarMyNeighbourPlayerCnt }

// GetPlayer プレイヤー取得
func (g *BeggarMyNeighbour) GetPlayer(i int) *BeggarMyNeighbourPlayer {
	if i < 0 || i >= BeggarMyNeighbourPlayerCnt {
		return nil
	}
	return g.players[i]
}

// GetCurrentPlayerIdx 現在のプレイヤーインデックス取得
func (g *BeggarMyNeighbour) GetCurrentPlayerIdx() int { return g.currentPlayerIdx }

// GetPenaltyOwnerIdx ペナルティ所有者インデックス取得 (-1 = なし)
func (g *BeggarMyNeighbour) GetPenaltyOwnerIdx() int { return g.penaltyOwnerIdx }

// GetPenaltyRemaining 残りペナルティ枚数取得
func (g *BeggarMyNeighbour) GetPenaltyRemaining() int { return g.penaltyRemaining }

// GetCentralPileSize 場の山の枚数取得
func (g *BeggarMyNeighbour) GetCentralPileSize() int { return len(g.centralPile) }

// GetLastCardPlayed 最後に出されたカード取得
func (g *BeggarMyNeighbour) GetLastCardPlayed() *Card { return g.lastCardPlayed }

// GetRoundsPlayed 消化ラウンド数取得
func (g *BeggarMyNeighbour) GetRoundsPlayed() int { return g.roundsPlayed }

// GetConfig 設定取得
func (g *BeggarMyNeighbour) GetConfig() BeggarMyNeighbourConfig { return g.config }

// SetConfig 設定更新
func (g *BeggarMyNeighbour) SetConfig(cfg BeggarMyNeighbourConfig) { g.config = cfg }

// SetRoundsPlayedForTest はテスト用に消化ラウンド数を設定する。
func (g *BeggarMyNeighbour) SetRoundsPlayedForTest(n int) { g.roundsPlayed = n }

// GetActionLog 棋譜取得
func (g *BeggarMyNeighbour) GetActionLog() []*ActionLogEntry { return g.actionLog }

// IsHumanTurn 常に人間入力待ち
func (g *BeggarMyNeighbour) IsHumanTurn() bool { return !g.gameEndFlag }

// --- JSON ---

// beggarMyNeighbourJSON is the JSON wire format for BeggarMyNeighbour.
type beggarMyNeighbourJSON struct {
	TrumpCards       *TrumpCards                                          `json:"tc"`
	Players          [BeggarMyNeighbourPlayerCnt]*BeggarMyNeighbourPlayer `json:"ps"`
	Config           BeggarMyNeighbourConfig                              `json:"cf"`
	Phase            BeggarMyNeighbourPhase                               `json:"ph"`
	CentralPile      []*Card                                              `json:"cp"`
	CurrentPlayerIdx int                                                  `json:"cu"`
	PenaltyOwnerIdx  int                                                  `json:"po"`
	PenaltyRemaining int                                                  `json:"pr"`
	LastCardPlayed   *Card                                                `json:"lc"`
	GameEndFlag      bool                                                 `json:"gf"`
	WinnerIdx        int                                                  `json:"wi"`
	RoundsPlayed     int                                                  `json:"rp"`
	ActionLog        []*ActionLogEntry                                    `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (g *BeggarMyNeighbour) MarshalJSON() ([]byte, error) {
	return json.Marshal(beggarMyNeighbourJSON{
		TrumpCards:       g.trumpCards,
		Players:          g.players,
		Config:           g.config,
		Phase:            g.phase,
		CentralPile:      g.centralPile,
		CurrentPlayerIdx: g.currentPlayerIdx,
		PenaltyOwnerIdx:  g.penaltyOwnerIdx,
		PenaltyRemaining: g.penaltyRemaining,
		LastCardPlayed:   g.lastCardPlayed,
		GameEndFlag:      g.gameEndFlag,
		WinnerIdx:        g.winnerIdx,
		RoundsPlayed:     g.roundsPlayed,
		ActionLog:        g.actionLog,
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (g *BeggarMyNeighbour) UnmarshalJSON(data []byte) error {
	var j beggarMyNeighbourJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	g.trumpCards = j.TrumpCards
	g.players = j.Players
	g.config = j.Config
	g.phase = j.Phase
	g.centralPile = j.CentralPile
	g.currentPlayerIdx = j.CurrentPlayerIdx
	g.penaltyOwnerIdx = j.PenaltyOwnerIdx
	g.penaltyRemaining = j.PenaltyRemaining
	g.lastCardPlayed = j.LastCardPlayed
	g.gameEndFlag = j.GameEndFlag
	g.winnerIdx = j.WinnerIdx
	g.roundsPlayed = j.RoundsPlayed
	g.actionLog = j.ActionLog
	return nil
}
