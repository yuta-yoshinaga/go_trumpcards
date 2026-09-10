//go:build !js || !wasm || extra4

package domain

import (
	"encoding/json"
	"fmt"
)

const (
	// BassetPhaseBetting is the phase in which the player places a wager.
	BassetPhaseBetting = 1
	// BassetPhaseTurn is the phase in which the bank deals two cards.
	BassetPhaseTurn = 2
	// BassetPhaseDecision is the phase in which a winning player chooses paroli or cash-out.
	BassetPhaseDecision = 3
	// BassetPhaseRoundEnd is the phase after the deck has been exhausted.
	BassetPhaseRoundEnd = 4
	// BassetPhaseGameEnd is the terminal phase.
	BassetPhaseGameEnd = 5
)

const (
	BassetDeckSize     = 52
	BassetTurnCards    = 2
	BassetTurnsPerDeal = BassetDeckSize / BassetTurnCards
	BassetMinRank      = 1
	BassetMaxRank      = 13
)

// BassetPayoutMultipliers contains the fixed payout odds for each paroli stage.
var BassetPayoutMultipliers = [...]int{1, 7, 15, 33, 67}

// BassetBet is a wager on one rank. Stage is the current paroli stage.
type BassetBet struct {
	Amount int `json:"am"`
	Stage  int `json:"st"`
}

// BassetTurnResult describes the two cards dealt by the bank.
type BassetTurnResult struct {
	BankerCard *Card `json:"bc"`
	PlayerCard *Card `json:"pc"`
	Hit        bool  `json:"ht"`
	Net        int   `json:"nt"`
}

// Basset is the single-player Venetian banking game.
type Basset struct {
	config      BassetConfig
	trumpCards  *TrumpCards
	chips       ChipHolder
	bet         *BassetBet
	betRank     int
	turnsPlayed int
	lastTurn    *BassetTurnResult
	phase       int
	gameEndFlag bool
	totalPayout int
	actionLogBase
}

// NewBasset creates a Basset game.
func NewBasset(trumpCards *TrumpCards) *Basset {
	return NewBassetWithConfig(trumpCards, DefaultBassetConfig())
}

// NewBassetWithConfig creates a Basset game with configuration.
func NewBassetWithConfig(trumpCards *TrumpCards, config BassetConfig) *Basset {
	if err := config.Validate(); err != nil {
		config = DefaultBassetConfig()
	}
	if trumpCards == nil {
		trumpCards = NewTrumpCards(0)
	}
	b := &Basset{config: config, trumpCards: trumpCards, phase: BassetPhaseBetting}
	b.chips.SetChips(config.StartChips)
	return b
}

// NewDefaultBasset creates a default Basset game.
func NewDefaultBasset() *Basset { return NewBasset(NewTrumpCards(0)) }

// Reset starts a fresh deal.
func (b *Basset) Reset() {
	if b.chips.GetChips() < b.config.MinBet {
		b.chips.SetChips(b.config.StartChips)
	}
	b.trumpCards = NewTrumpCards(0)
	b.trumpCards.Shuffle()
	b.bet, b.betRank = nil, 0
	b.turnsPlayed, b.lastTurn, b.totalPayout = 0, nil, 0
	b.phase, b.gameEndFlag = BassetPhaseBetting, false
	b.actionLog = nil
}

// PlayerPlaceBet places or replaces the player's rank wager.
func (b *Basset) PlayerPlaceBet(rank, amount int) error {
	if b.phase != BassetPhaseBetting && b.phase != BassetPhaseTurn {
		return NewDomainError(ErrWrongPhase, "Betting is not available now.")
	}
	if rank < BassetMinRank || rank > BassetMaxRank {
		return NewDomainError(ErrInvalidPlay, "Bet rank must be between 1 and 13.")
	}
	if amount < b.config.MinBet || amount%b.config.MinBet != 0 || amount > b.config.MaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid bet amount.")
	}
	previous, stage := 0, 0
	if b.bet != nil {
		if b.betRank != rank {
			if b.bet.Stage > 0 {
				return NewDomainError(ErrInvalidPlay, "Cannot change bet rank during paroli.")
			}
			b.chips.AddChips(b.bet.Amount)
		} else {
			previous, stage = b.bet.Amount, b.bet.Stage
		}
	}
	delta := amount - previous
	if delta > 0 && !b.chips.SubtractChips(delta) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	if delta < 0 {
		b.chips.AddChips(-delta)
	}
	b.bet, b.betRank = &BassetBet{Amount: amount, Stage: stage}, rank
	b.appendLog(0, "bet", fmt.Sprintf("rank=%d amount=%d stage=%d", rank, amount, stage), nil)
	return nil
}

// PlayerDealTurn deals the banker card first and the player card second.
func (b *Basset) PlayerDealTurn() error {
	if b.phase != BassetPhaseBetting && b.phase != BassetPhaseTurn {
		return NewDomainError(ErrWrongPhase, "A turn is not available now.")
	}
	if b.bet == nil {
		return NewDomainError(ErrInvalidPlay, "Place a bet before dealing.")
	}
	if b.trumpCards.GetRemainingCount() < BassetTurnCards {
		return NewDomainError(ErrDeckExhausted, "The deck is exhausted.")
	}
	banker, player := b.trumpCards.DrawCard(), b.trumpCards.DrawCard()
	b.turnsPlayed++
	hit := banker.GetValue() == b.betRank || player.GetValue() == b.betRank
	b.lastTurn = &BassetTurnResult{BankerCard: banker, PlayerCard: player, Hit: hit}
	if banker.GetValue() == b.betRank {
		b.totalPayout -= b.bet.Amount
		b.bet, b.betRank = nil, 0
	} else if player.GetValue() == b.betRank {
		b.phase = BassetPhaseDecision
		b.appendLog(-1, "hit", "player card matched; choose paroli or cash out", []*Card{banker, player})
		return nil
	}
	b.phase = b.nextPhase()
	b.appendLog(-1, "turn", fmt.Sprintf("banker=%d player=%d hit=%t", banker.GetValue(), player.GetValue(), hit), []*Card{banker, player})
	return nil
}

// PlayerTakeWinnings cashes out a winning wager and resets its paroli stage.
func (b *Basset) PlayerTakeWinnings() error {
	if b.phase != BassetPhaseDecision || b.bet == nil {
		return NewDomainError(ErrWrongPhase, "There are no winnings to take.")
	}
	payout := b.bet.Amount * (1 + BassetPayoutMultipliers[b.bet.Stage])
	b.chips.AddChips(payout)
	b.totalPayout += b.bet.Amount * BassetPayoutMultipliers[b.bet.Stage]
	b.bet, b.betRank = nil, 0
	b.phase = b.nextPhase()
	b.appendLog(0, "takeWinnings", fmt.Sprintf("payout=%d", payout), nil)
	return nil
}

// PlayerPressParoli leaves the wager on the table and advances its payout stage.
func (b *Basset) PlayerPressParoli() error {
	if b.phase != BassetPhaseDecision || b.bet == nil {
		return NewDomainError(ErrWrongPhase, "Paroli is only available after a win.")
	}
	if b.bet.Stage >= len(BassetPayoutMultipliers)-1 {
		return b.PlayerTakeWinnings()
	}
	b.bet.Stage++
	b.phase = b.nextPhase()
	b.appendLog(0, "paroli", fmt.Sprintf("stage=%d multiplier=%d", b.bet.Stage, BassetPayoutMultipliers[b.bet.Stage]), nil)
	return nil
}

func (b *Basset) nextPhase() int {
	if b.trumpCards.GetRemainingCount() < BassetTurnCards {
		return BassetPhaseRoundEnd
	}
	return BassetPhaseTurn
}

// NextRound starts the next deal or ends the game when the bankroll is empty.
func (b *Basset) NextRound() {
	if b.chips.GetChips() < b.config.MinBet {
		b.gameEndFlag, b.phase = true, BassetPhaseGameEnd
		return
	}
	b.Reset()
}

// GetPhase returns the current phase.
func (b *Basset) GetPhase() int { return b.phase }

// GetGameEndFlag reports whether the game has ended.
func (b *Basset) GetGameEndFlag() bool { return b.gameEndFlag }

// GetChips returns the bankroll.
func (b *Basset) GetChips() int { return b.chips.GetChips() }

// GetTurnsPlayed returns the number of dealt turns.
func (b *Basset) GetTurnsPlayed() int { return b.turnsPlayed }

// GetTurnsTotal returns the maximum number of turns in a deck.
func (b *Basset) GetTurnsTotal() int { return BassetDeckSize / BassetTurnCards }

// GetRemainingCount returns the number of undealt cards.
func (b *Basset) GetRemainingCount() int { return b.trumpCards.GetRemainingCount() }

// GetRemainingByRank returns undealt card counts by rank.
func (b *Basset) GetRemainingByRank() [BassetMaxRank + 1]int {
	var out [BassetMaxRank + 1]int
	for _, c := range b.trumpCards.deck {
		if c != nil && !c.GetDraw() && c.GetValue() >= BassetMinRank && c.GetValue() <= BassetMaxRank {
			out[c.GetValue()]++
		}
	}
	return out
}

// GetBet returns the current wager and rank.
func (b *Basset) GetBet() (*BassetBet, int) { return b.bet, b.betRank }

// GetLastTurn returns the most recent turn.
func (b *Basset) GetLastTurn() *BassetTurnResult { return b.lastTurn }

// GetTotalPayout returns the current deal's net payout.
func (b *Basset) GetTotalPayout() int { return b.totalPayout }

// GetConfig returns the game configuration.
func (b *Basset) GetConfig() BassetConfig { return b.config }

// SetPhase is a test helper.
func (b *Basset) SetPhase(phase int) { b.phase = phase }

// SetChips is a test helper.
func (b *Basset) SetChips(chips int) { b.chips.SetChips(chips) }

// SetDeckForTest replaces the deck order for deterministic tests.
func (b *Basset) SetDeckForTest(cards []*Card) {
	b.trumpCards = NewTrumpCardsWithSuits(0, []int{})
	b.trumpCards.deck, b.trumpCards.deckCnt = cards, len(cards)
	b.trumpCards.deckInit()
}

type bassetJSON struct {
	Config      BassetConfig      `json:"cf"`
	TrumpCards  *TrumpCards       `json:"tc"`
	Chips       *ChipHolder       `json:"ch"`
	Bet         *BassetBet        `json:"be"`
	BetRank     int               `json:"br"`
	TurnsPlayed int               `json:"tn"`
	LastTurn    *BassetTurnResult `json:"lt"`
	Phase       int               `json:"ps"`
	GameEndFlag bool              `json:"ge"`
	TotalPayout int               `json:"tp"`
	ActionLog   []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (b *Basset) MarshalJSON() ([]byte, error) {
	return json.Marshal(bassetJSON{b.config, b.trumpCards, &b.chips, b.bet, b.betRank, b.turnsPlayed, b.lastTurn, b.phase, b.gameEndFlag, b.totalPayout, b.actionLog})
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Basset) UnmarshalJSON(data []byte) error {
	var j bassetJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if err := j.Config.Validate(); err != nil || j.TrumpCards == nil || j.Phase < BassetPhaseBetting || j.Phase > BassetPhaseGameEnd || j.BetRank < 0 || j.BetRank > BassetMaxRank || j.Chips == nil {
		return fmt.Errorf("basset: invalid serialized state")
	}
	b.config, b.trumpCards, b.bet, b.betRank = j.Config, j.TrumpCards, j.Bet, j.BetRank
	b.chips, b.turnsPlayed, b.lastTurn, b.phase = *j.Chips, j.TurnsPlayed, j.LastTurn, j.Phase
	b.gameEndFlag, b.totalPayout, b.actionLog = j.GameEndFlag, j.TotalPayout, j.ActionLog
	return nil
}
