//go:build !js || !wasm || extra2

package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

// TehonbikiPhase is the round state.
type TehonbikiPhase int

const (
	TehonbikiPhaseBet TehonbikiPhase = iota
	TehonbikiPhaseResult
	TehonbikiPhaseGameEnd
)

// TehonbikiResult is the wager outcome.
type TehonbikiResult int

const (
	TehonbikiResultNone TehonbikiResult = iota
	TehonbikiResultWin
	TehonbikiResultLose
)

// Tehonbiki is the number-guessing game using six stock cards. The payout is
// the house table adopted by this repository: fair odds less ten percent.
type Tehonbiki struct {
	deck        []*Card
	parentCard  int
	player      *TehonbikiPlayer
	config      TehonbikiConfig
	phase       TehonbikiPhase
	betType     TehonbikiBetType
	numbers     []int
	bet, payout int
	result      TehonbikiResult
	roundNo     int
	gameEndFlag bool
	actionLog   []*ActionLogEntry
	turnNumber  int
}

func NewTehonbiki(_ *TrumpCards, p *TehonbikiPlayer, c TehonbikiConfig) *Tehonbiki {
	return NewDefaultTehonbikiWithPlayer(p, c)
}
func NewDefaultTehonbiki() *Tehonbiki {
	c := DefaultTehonbikiConfig()
	return NewDefaultTehonbikiWithPlayer(NewTehonbikiPlayer(c.InitialChips), c)
}
func NewDefaultTehonbikiWithPlayer(p *TehonbikiPlayer, c TehonbikiConfig) *Tehonbiki {
	g := &Tehonbiki{player: p, config: c, roundNo: 1}
	g.Reset()
	return g
}
func buildTehonbikiDeck() []*Card {
	d := make([]*Card, 0, 6)
	for n := 1; n <= 6; n++ {
		d = append(d, NewCard(CardDesignSpade, n, false))
	}
	return d
}
func (g *Tehonbiki) Reset() {
	g.deck = buildTehonbikiDeck()
	g.parentCard = rand.Intn(6) + 1
	g.phase = TehonbikiPhaseBet
	g.betType = ""
	g.numbers = nil
	g.bet = 0
	g.payout = 0
	g.result = TehonbikiResultNone
	g.roundNo = 1
	g.gameEndFlag = false
	g.actionLog = nil
	g.turnNumber = 0
}

// SetParentCard fixes the face-down card, which also makes tests deterministic.
func (g *Tehonbiki) SetParentCard(n int) error {
	if n < 1 || n > 6 {
		return errors.New("tehonbiki: parent card out of range")
	}
	g.parentCard = n
	return nil
}
func (g *Tehonbiki) PlaceBet(ns []int, k TehonbikiBetType, bet int) error {
	if g.gameEndFlag {
		return errors.New("tehonbiki: game already finished")
	}
	if g.phase != TehonbikiPhaseBet {
		return errors.New("tehonbiki: not allowed in this phase")
	}
	p, ok := TehonbikiPayouts[k]
	if !ok {
		return errors.New("tehonbiki: invalid bet type")
	}
	want := map[TehonbikiBetType]int{TehonbikiBetSingle: 1, TehonbikiBetDouble: 2, TehonbikiBetTriple: 3, TehonbikiBetHalf: 3}[k]
	if len(ns) != want {
		return errors.New("tehonbiki: invalid number count")
	}
	seen := map[int]bool{}
	for _, n := range ns {
		if n < 1 || n > 6 || seen[n] {
			return errors.New("tehonbiki: invalid number")
		}
		seen[n] = true
	}
	if bet < TehonbikiMinBet || bet > TehonbikiMaxBet || !g.player.SubtractChips(bet) {
		return errors.New("tehonbiki: invalid or insufficient bet")
	}
	g.numbers = append([]int(nil), ns...)
	g.betType = k
	g.bet = bet
	g.result = TehonbikiResultLose
	for _, n := range ns {
		if n == g.parentCard {
			g.result = TehonbikiResultWin
			g.payout = bet * p.MultiplierNum / p.MultiplierDen
			g.player.AddChips(bet + g.payout)
			break
		}
	}
	g.phase = TehonbikiPhaseResult
	return nil
}
func (g *Tehonbiki) NextRound() error {
	if g.phase != TehonbikiPhaseResult {
		return errors.New("tehonbiki: not allowed in this phase")
	}
	if g.player.GetChips() < TehonbikiMinBet {
		g.gameEndFlag = true
		g.phase = TehonbikiPhaseGameEnd
		return nil
	}
	g.roundNo++
	g.deck = buildTehonbikiDeck()
	g.parentCard = rand.Intn(6) + 1
	g.phase = TehonbikiPhaseBet
	g.numbers = nil
	g.bet = 0
	g.payout = 0
	g.result = TehonbikiResultNone
	return nil
}
func (g *Tehonbiki) GetConfig() TehonbikiConfig      { return g.config }
func (g *Tehonbiki) SetConfig(c TehonbikiConfig)     { g.config = c }
func (g *Tehonbiki) GetPhase() TehonbikiPhase        { return g.phase }
func (g *Tehonbiki) GetGameEndFlag() bool            { return g.gameEndFlag }
func (g *Tehonbiki) GetParentCard() int              { return g.parentCard }
func (g *Tehonbiki) GetNumbers() []int               { return g.numbers }
func (g *Tehonbiki) GetBetType() TehonbikiBetType    { return g.betType }
func (g *Tehonbiki) GetBet() int                     { return g.bet }
func (g *Tehonbiki) GetPayout() int                  { return g.payout }
func (g *Tehonbiki) GetResult() TehonbikiResult      { return g.result }
func (g *Tehonbiki) GetChips() int                   { return g.player.GetChips() }
func (g *Tehonbiki) SetChips(n int)                  { g.player.SetChips(n) }
func (g *Tehonbiki) GetPlayer() *TehonbikiPlayer     { return g.player }
func (g *Tehonbiki) GetRoundNumber() int             { return g.roundNo }
func (g *Tehonbiki) GetRemainingCards() int          { return 6 }
func (g *Tehonbiki) GetActionLog() []*ActionLogEntry { return g.actionLog }

// TehonbikiHint is the neutral hint shown before a wager.
type TehonbikiHint struct {
	PickIdx int
	Reason  string
}

func (g *Tehonbiki) GetHint() *TehonbikiHint {
	if g.phase != TehonbikiPhaseBet || g.gameEndFlag {
		return nil
	}
	return &TehonbikiHint{Reason: "choose a number"}
}

type tehonbikiJSON struct {
	ParentCard int              `json:"pc"`
	Player     *TehonbikiPlayer `json:"pl"`
	Config     TehonbikiConfig  `json:"cf"`
	Phase      TehonbikiPhase   `json:"ph"`
	BetType    TehonbikiBetType `json:"btp"`
	Numbers    []int            `json:"nm"`
	Bet        int              `json:"bt"`
	Payout     int              `json:"po"`
	Result     TehonbikiResult  `json:"rs"`
	Round      int              `json:"rn"`
	End        bool             `json:"ge"`
}

func (g *Tehonbiki) MarshalJSON() ([]byte, error) {
	return json.Marshal(tehonbikiJSON{g.parentCard, g.player, g.config, g.phase, g.betType, g.numbers, g.bet, g.payout, g.result, g.roundNo, g.gameEndFlag})
}
func (g *Tehonbiki) UnmarshalJSON(b []byte) error {
	var j tehonbikiJSON
	if err := json.Unmarshal(b, &j); err != nil {
		return err
	}
	if j.Player == nil || j.ParentCard < 1 || j.ParentCard > 6 {
		return fmt.Errorf("tehonbiki: invalid saved state")
	}
	g.parentCard = j.ParentCard
	g.deck = buildTehonbikiDeck()
	g.player = j.Player
	g.config = j.Config
	g.phase = j.Phase
	g.betType = j.BetType
	g.numbers = j.Numbers
	g.bet = j.Bet
	g.payout = j.Payout
	g.result = j.Result
	g.roundNo = j.Round
	g.gameEndFlag = j.End
	return nil
}
