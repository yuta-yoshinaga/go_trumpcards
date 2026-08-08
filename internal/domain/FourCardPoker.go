//go:build !js || !wasm || casino

package domain

import (
	"encoding/json"
	"fmt"
)

// FourCardPoker phase constants.
const (
	FourCardPokerPhaseBet    = 1 // Bet phase
	FourCardPokerPhaseAction = 2 // Action phase (Play/Fold after seeing dealer's upcard)
	FourCardPokerPhaseEnd    = 3 // End phase
)

// FourCardPoker default values.
const (
	FourCardPokerDefaultChips = 1000  // Default chip stack
	FourCardPokerMinBet       = 10    // Minimum ante (and minimum Aces Up)
	FourCardPokerMaxBet       = 10000 // Maximum ante (and maximum Aces Up)
	FourCardPokerPlayerCards  = 5     // Cards dealt to the player
	FourCardPokerDealerCards  = 6     // Cards dealt to the dealer (1 face-up)
	FourCardPokerHandSize     = 4     // Final hand size used for evaluation
	FourCardPokerMinPlayMul   = 1     // Minimum Play bet multiplier of ante
	FourCardPokerMaxPlayMul   = 3     // Maximum Play bet multiplier of ante
)

// Auto Ante Bonus payout multipliers (paid in addition to ante on Play; refunded on Fold? — no,
// Ante Bonus is forfeited on Fold; only credited when player plays and the hand qualifies).
const (
	FourCardPokerAnteBonusThreeOfAKind  = 2  // 3 of a Kind 2:1
	FourCardPokerAnteBonusStraightFlush = 20 // Straight Flush 20:1
	FourCardPokerAnteBonusFourOfAKind   = 25 // 4 of a Kind 25:1
)

// Aces Up sidebet payout multipliers (standard Shuffle Master paytable).
// Aces Up qualifies only on Pair of Aces or better. All payouts are pure bonuses
// (the original Aces Up wager is returned in addition to the listed odds).
const (
	FourCardPokerAcesUpPairOfAces    = 1  // Pair of Aces 1:1
	FourCardPokerAcesUpTwoPair       = 2  // Two Pair 2:1
	FourCardPokerAcesUpStraight      = 4  // Straight 4:1
	FourCardPokerAcesUpFlush         = 5  // Flush 5:1
	FourCardPokerAcesUpThreeOfAKind  = 7  // 3 of a Kind 7:1
	FourCardPokerAcesUpStraightFlush = 40 // Straight Flush 40:1
	FourCardPokerAcesUpFourOfAKind   = 50 // 4 of a Kind 50:1
)

// FourCardPoker is the Four Card Poker domain entity.
type FourCardPoker struct {
	trumpCards      *TrumpCards
	playerHand      []*Card // 5 cards dealt to the player
	dealerHand      []*Card // 6 cards dealt to the dealer
	playerBest      []*Card // best 4-card subset of playerHand
	dealerBest      []*Card // best 4-card subset of dealerHand (revealed on resolve)
	chips           ChipHolder
	anteBet         int
	acesUpBet       int
	playBet         int
	playMultiplier  int // 1, 2, or 3
	phase           int
	gameEndFlag     bool
	result          GameResult
	antePayout      int // returned ante + 1:1 win (or 0 on lose)
	playPayout      int // returned play bet + 1:1 win (or 0 on lose)
	anteBonusPayout int // pure bonus, no original wager (independent of dealer)
	acesUpPayout    int // returned acesUp wager + bonus
	playerHandRank  int
	dealerHandRank  int
	actionLogBase
}

// NewFourCardPoker constructs an empty game bound to a deck.
func NewFourCardPoker(trumpCards *TrumpCards) *FourCardPoker {
	trumpCards.Shuffle()
	return &FourCardPoker{
		trumpCards: trumpCards,
		phase:      FourCardPokerPhaseBet,
	}
}

// NewDefaultFourCardPoker returns a fully-initialised game with default chips.
func NewDefaultFourCardPoker() *FourCardPoker {
	fcp := NewFourCardPoker(NewTrumpCards(0))
	fcp.chips.SetChips(FourCardPokerDefaultChips)
	return fcp
}

// Reset clears per-round state and reshuffles. Chip stack is preserved unless
// it has fallen below the minimum bet, in which case it is restored to the default.
func (fcp *FourCardPoker) Reset() {
	fcp.gameEndFlag = false
	fcp.phase = FourCardPokerPhaseBet
	fcp.playerHand = nil
	fcp.dealerHand = nil
	fcp.playerBest = nil
	fcp.dealerBest = nil
	fcp.anteBet = 0
	fcp.acesUpBet = 0
	fcp.playBet = 0
	fcp.playMultiplier = 0
	fcp.result = 0
	fcp.antePayout = 0
	fcp.playPayout = 0
	fcp.anteBonusPayout = 0
	fcp.acesUpPayout = 0
	fcp.playerHandRank = 0
	fcp.dealerHandRank = 0
	fcp.actionLog = nil
	if fcp.chips.GetChips() < FourCardPokerMinBet {
		fcp.chips.SetChips(FourCardPokerDefaultChips)
	}
	fcp.trumpCards = NewTrumpCards(0)
	for range 10 {
		fcp.trumpCards.Shuffle()
	}
}

// Bet places the ante (mandatory) and optional Aces Up sidebet, then deals.
func (fcp *FourCardPoker) Bet(ante, acesUp int) error {
	if fcp.phase != FourCardPokerPhaseBet {
		return NewDomainError(ErrWrongPhase, "Bet is only allowed during the bet phase.")
	}
	if ante < FourCardPokerMinBet || ante%FourCardPokerMinBet != 0 || ante > FourCardPokerMaxBet {
		return NewDomainError(ErrInvalidAmount, "Invalid ante amount.")
	}
	if acesUp < 0 {
		return NewDomainError(ErrInvalidAmount, "Aces Up bet must not be negative.")
	}
	if acesUp > 0 && (acesUp < FourCardPokerMinBet || acesUp%FourCardPokerMinBet != 0 || acesUp > FourCardPokerMaxBet) {
		return NewDomainError(ErrInvalidAmount, "Invalid Aces Up bet amount.")
	}
	totalCost := ante + acesUp
	if !fcp.chips.SubtractChips(totalCost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips.")
	}
	fcp.anteBet = ante
	fcp.acesUpBet = acesUp
	fcp.appendLog(0, "bet", fmt.Sprintf("ante=%d acesup=%d", ante, acesUp), nil)

	fcp.deal()
	fcp.phase = FourCardPokerPhaseAction
	return nil
}

// Play places the Play bet at the requested ante multiplier (1, 2, or 3) and resolves.
func (fcp *FourCardPoker) Play(multiplier int) error {
	if fcp.phase != FourCardPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Play is only allowed during the action phase.")
	}
	if multiplier < FourCardPokerMinPlayMul || multiplier > FourCardPokerMaxPlayMul {
		return NewDomainError(ErrInvalidAmount, "Play multiplier must be 1, 2, or 3.")
	}
	cost := fcp.anteBet * multiplier
	if !fcp.chips.SubtractChips(cost) {
		return NewDomainError(ErrInsufficientChips, "Insufficient chips for play bet.")
	}
	fcp.playBet = cost
	fcp.playMultiplier = multiplier
	fcp.appendLog(0, "play", fmt.Sprintf("play bet=%d (x%d ante)", cost, multiplier), nil)

	fcp.resolve()
	return nil
}

// Fold forfeits the ante (and Play bet, which is never placed). Aces Up is still evaluated.
func (fcp *FourCardPoker) Fold() error {
	if fcp.phase != FourCardPokerPhaseAction {
		return NewDomainError(ErrWrongPhase, "Fold is only allowed during the action phase.")
	}
	fcp.appendLog(0, "fold", "player folds", nil)

	fcp.updateBestHands()
	fcp.result = GameResultLose

	// Aces Up still pays
	fcp.evaluateAcesUp()
	if fcp.acesUpPayout > 0 {
		fcp.chips.AddChips(fcp.acesUpPayout)
	}

	fcp.gameEndFlag = true
	fcp.phase = FourCardPokerPhaseEnd
	fcp.appendLog(-1, "result", "player folded", nil)
	return nil
}

// deal distributes 5 cards to the player and 6 to the dealer.
func (fcp *FourCardPoker) deal() {
	fcp.playerHand = make([]*Card, 0, FourCardPokerPlayerCards)
	fcp.dealerHand = make([]*Card, 0, FourCardPokerDealerCards)
	for range FourCardPokerPlayerCards {
		fcp.playerHand = append(fcp.playerHand, fcp.trumpCards.DrawCard())
	}
	for range FourCardPokerDealerCards {
		fcp.dealerHand = append(fcp.dealerHand, fcp.trumpCards.DrawCard())
	}
	fcp.appendLog(-1, "deal", "dealt 5 to player and 6 to dealer", nil)
}

// updateBestHands picks the strongest 4-card subset for both player and dealer
// and caches the resulting ranks. Shared by Fold and resolve.
func (fcp *FourCardPoker) updateBestHands() {
	fcp.playerBest = pickBestFour(fcp.playerHand)
	fcp.dealerBest = pickBestFour(fcp.dealerHand)
	fcp.playerHandRank = evalFourCardHand(fcp.playerBest)
	fcp.dealerHandRank = evalFourCardHand(fcp.dealerBest)
}

// resolve picks best hands, compares, and credits all payouts.
func (fcp *FourCardPoker) resolve() {
	fcp.updateBestHands()

	cmp := compareFourCardHands(fcp.playerBest, fcp.dealerBest)
	if cmp > 0 {
		fcp.result = GameResultWin
	} else if cmp < 0 {
		fcp.result = GameResultLose
	} else {
		fcp.result = GameResultDraw
	}

	fcp.calculatePayouts()
	fcp.evaluateAnteBonus()
	fcp.evaluateAcesUp()

	totalPayout := fcp.antePayout + fcp.playPayout + fcp.anteBonusPayout + fcp.acesUpPayout
	if totalPayout > 0 {
		fcp.chips.AddChips(totalPayout)
	}

	fcp.gameEndFlag = true
	fcp.phase = FourCardPokerPhaseEnd

	var resultStr string
	switch fcp.result {
	case GameResultWin:
		resultStr = "player wins"
	case GameResultDraw:
		resultStr = "push"
	default:
		resultStr = "dealer wins"
	}
	fcp.appendLog(-1, "result", resultStr, nil)
}

// calculatePayouts settles Ante and Play bets. Four Card Poker has no dealer qualify.
func (fcp *FourCardPoker) calculatePayouts() {
	switch fcp.result {
	case GameResultWin:
		fcp.antePayout = fcp.anteBet * 2 // returned + 1:1
		fcp.playPayout = fcp.playBet * 2 // returned + 1:1
	case GameResultDraw:
		fcp.antePayout = fcp.anteBet
		fcp.playPayout = fcp.playBet
	case GameResultLose:
		fcp.antePayout = 0
		fcp.playPayout = 0
	}
}

// evaluateAnteBonus credits the automatic ante bonus when the player plays a qualifying hand.
// Paid as a pure bonus on top of the ante payout (no original wager returned with the bonus).
func (fcp *FourCardPoker) evaluateAnteBonus() {
	switch fcp.playerHandRank {
	case FourCardHandThreeOfAKind:
		fcp.anteBonusPayout = fcp.anteBet * FourCardPokerAnteBonusThreeOfAKind
	case FourCardHandStraightFlush:
		fcp.anteBonusPayout = fcp.anteBet * FourCardPokerAnteBonusStraightFlush
	case FourCardHandFourOfAKind:
		fcp.anteBonusPayout = fcp.anteBet * FourCardPokerAnteBonusFourOfAKind
	}
}

// evaluateAcesUp credits the Aces Up sidebet. Returned wager is included in the payout.
// Qualifies on Pair of Aces or better — lesser pairs lose the sidebet.
func (fcp *FourCardPoker) evaluateAcesUp() {
	if fcp.acesUpBet <= 0 {
		return
	}
	mul := 0
	switch fcp.playerHandRank {
	case FourCardHandFourOfAKind:
		mul = FourCardPokerAcesUpFourOfAKind
	case FourCardHandStraightFlush:
		mul = FourCardPokerAcesUpStraightFlush
	case FourCardHandThreeOfAKind:
		mul = FourCardPokerAcesUpThreeOfAKind
	case FourCardHandFlush:
		mul = FourCardPokerAcesUpFlush
	case FourCardHandStraight:
		mul = FourCardPokerAcesUpStraight
	case FourCardHandTwoPair:
		mul = FourCardPokerAcesUpTwoPair
	case FourCardHandPair:
		if fcp.pairIsAces(fcp.playerBest) {
			mul = FourCardPokerAcesUpPairOfAces
		}
	}
	if mul > 0 {
		fcp.acesUpPayout = fcp.acesUpBet + fcp.acesUpBet*mul
	}
}

// pairIsAces reports whether a Pair hand's pair value is Aces.
func (fcp *FourCardPoker) pairIsAces(cards []*Card) bool {
	if len(cards) != FourCardPokerHandSize {
		return false
	}
	counts := make(map[int]int)
	for _, c := range cards {
		counts[c.GetValue()]++
	}
	return counts[1] == 2
}

// --- Getters ---

// GetPlayerHand returns the 5-card player hand.
func (fcp *FourCardPoker) GetPlayerHand() []*Card { return fcp.playerHand }

// GetDealerHand returns the 6-card dealer hand.
func (fcp *FourCardPoker) GetDealerHand() []*Card { return fcp.dealerHand }

// GetPlayerBest returns the best 4-card subset of the player's hand.
func (fcp *FourCardPoker) GetPlayerBest() []*Card { return fcp.playerBest }

// GetDealerBest returns the best 4-card subset of the dealer's hand.
func (fcp *FourCardPoker) GetDealerBest() []*Card { return fcp.dealerBest }

// GetDealerUpCard returns the dealer's face-up card (first card), or nil before deal.
func (fcp *FourCardPoker) GetDealerUpCard() *Card {
	if len(fcp.dealerHand) == 0 {
		return nil
	}
	return fcp.dealerHand[0]
}

// GetPhase returns the current phase.
func (fcp *FourCardPoker) GetPhase() int { return fcp.phase }

// GetGameEndFlag reports whether the round has ended.
func (fcp *FourCardPoker) GetGameEndFlag() bool { return fcp.gameEndFlag }

// GetAnteBet returns the ante bet.
func (fcp *FourCardPoker) GetAnteBet() int { return fcp.anteBet }

// GetAcesUpBet returns the Aces Up sidebet.
func (fcp *FourCardPoker) GetAcesUpBet() int { return fcp.acesUpBet }

// GetPlayBet returns the Play bet.
func (fcp *FourCardPoker) GetPlayBet() int { return fcp.playBet }

// GetPlayMultiplier returns the chosen Play multiplier.
func (fcp *FourCardPoker) GetPlayMultiplier() int { return fcp.playMultiplier }

// GetResult returns the round result.
func (fcp *FourCardPoker) GetResult() GameResult { return fcp.result }

// GetAntePayout returns the ante payout (including returned wager on win/push).
func (fcp *FourCardPoker) GetAntePayout() int { return fcp.antePayout }

// GetPlayPayout returns the play payout (including returned wager on win/push).
func (fcp *FourCardPoker) GetPlayPayout() int { return fcp.playPayout }

// GetAnteBonusPayout returns the automatic ante bonus (pure bonus, no wager returned with it).
func (fcp *FourCardPoker) GetAnteBonusPayout() int { return fcp.anteBonusPayout }

// GetAcesUpPayout returns the Aces Up payout (including returned wager).
func (fcp *FourCardPoker) GetAcesUpPayout() int { return fcp.acesUpPayout }

// GetTotalPayout returns the sum of all payouts credited this round.
func (fcp *FourCardPoker) GetTotalPayout() int {
	return fcp.antePayout + fcp.playPayout + fcp.anteBonusPayout + fcp.acesUpPayout
}

// GetPlayerHandRank returns the player's hand rank.
func (fcp *FourCardPoker) GetPlayerHandRank() int { return fcp.playerHandRank }

// GetDealerHandRank returns the dealer's hand rank.
func (fcp *FourCardPoker) GetDealerHandRank() int { return fcp.dealerHandRank }

// GetChips returns the current chip stack.
func (fcp *FourCardPoker) GetChips() int { return fcp.chips.GetChips() }

// --- Test helpers ---

// SetPhase sets the phase (test only).
func (fcp *FourCardPoker) SetPhase(phase int) { fcp.phase = phase }

// SetPlayerHand sets the player's 5-card hand (test only).
func (fcp *FourCardPoker) SetPlayerHand(cards []*Card) { fcp.playerHand = cards }

// SetDealerHand sets the dealer's 6-card hand (test only).
func (fcp *FourCardPoker) SetDealerHand(cards []*Card) { fcp.dealerHand = cards }

// SetAnteBet sets the ante (test only).
func (fcp *FourCardPoker) SetAnteBet(amount int) { fcp.anteBet = amount }

// SetAcesUpBet sets the Aces Up sidebet (test only).
func (fcp *FourCardPoker) SetAcesUpBet(amount int) { fcp.acesUpBet = amount }

// SetPlayBet sets the Play bet (test only).
func (fcp *FourCardPoker) SetPlayBet(amount int) { fcp.playBet = amount }

// SetResult sets the round result (test only).
func (fcp *FourCardPoker) SetResult(result GameResult) { fcp.result = result }

// SetGameEndFlag sets the end flag (test only).
func (fcp *FourCardPoker) SetGameEndFlag(flag bool) { fcp.gameEndFlag = flag }

// SetChips sets the chip stack (test only).
func (fcp *FourCardPoker) SetChips(chips int) { fcp.chips.SetChips(chips) }

// SetPlayerHandRank sets the player rank (test only).
func (fcp *FourCardPoker) SetPlayerHandRank(rank int) { fcp.playerHandRank = rank }

// SetDealerHandRank sets the dealer rank (test only).
func (fcp *FourCardPoker) SetDealerHandRank(rank int) { fcp.dealerHandRank = rank }

// SetAntePayout sets the ante payout (test only).
func (fcp *FourCardPoker) SetAntePayout(p int) { fcp.antePayout = p }

// SetPlayPayout sets the play payout (test only).
func (fcp *FourCardPoker) SetPlayPayout(p int) { fcp.playPayout = p }

// SetAnteBonusPayout sets the bonus (test only).
func (fcp *FourCardPoker) SetAnteBonusPayout(p int) { fcp.anteBonusPayout = p }

// SetAcesUpPayout sets the aces-up payout (test only).
func (fcp *FourCardPoker) SetAcesUpPayout(p int) { fcp.acesUpPayout = p }

// fourCardPokerJSON is the JSON wire format.
type fourCardPokerJSON struct {
	TrumpCards      *TrumpCards       `json:"tc"`
	PlayerHand      []*Card           `json:"ph"`
	DealerHand      []*Card           `json:"dh"`
	PlayerBest      []*Card           `json:"pb4"`
	DealerBest      []*Card           `json:"db4"`
	Chips           *ChipHolder       `json:"ch"`
	AnteBet         int               `json:"ab"`
	AcesUpBet       int               `json:"au"`
	PlayBet         int               `json:"pb"`
	PlayMultiplier  int               `json:"pm"`
	Phase           int               `json:"ps"`
	GameEndFlag     bool              `json:"ge"`
	Result          GameResult        `json:"rs"`
	AntePayout      int               `json:"ap"`
	PlayPayout      int               `json:"plp"`
	AnteBonusPayout int               `json:"abp"`
	AcesUpPayout    int               `json:"aup"`
	PlayerHandRank  int               `json:"pr"`
	DealerHandRank  int               `json:"dr"`
	ActionLog       []*ActionLogEntry `json:"al"`
}

// MarshalJSON implements json.Marshaler.
func (fcp *FourCardPoker) MarshalJSON() ([]byte, error) {
	return json.Marshal(fourCardPokerJSON{
		TrumpCards:      fcp.trumpCards,
		PlayerHand:      fcp.playerHand,
		DealerHand:      fcp.dealerHand,
		PlayerBest:      fcp.playerBest,
		DealerBest:      fcp.dealerBest,
		Chips:           &fcp.chips,
		AnteBet:         fcp.anteBet,
		AcesUpBet:       fcp.acesUpBet,
		PlayBet:         fcp.playBet,
		PlayMultiplier:  fcp.playMultiplier,
		Phase:           fcp.phase,
		GameEndFlag:     fcp.gameEndFlag,
		Result:          fcp.result,
		AntePayout:      fcp.antePayout,
		PlayPayout:      fcp.playPayout,
		AnteBonusPayout: fcp.anteBonusPayout,
		AcesUpPayout:    fcp.acesUpPayout,
		PlayerHandRank:  fcp.playerHandRank,
		DealerHandRank:  fcp.dealerHandRank,
		ActionLog:       fcp.actionLog,
	})
}

const fourCardPokerMaxSliceLen = 1000

// UnmarshalJSON implements json.Unmarshaler.
func (fcp *FourCardPoker) UnmarshalJSON(data []byte) error {
	var j fourCardPokerJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	if len(j.PlayerHand) > fourCardPokerMaxSliceLen || len(j.DealerHand) > fourCardPokerMaxSliceLen ||
		len(j.PlayerBest) > fourCardPokerMaxSliceLen || len(j.DealerBest) > fourCardPokerMaxSliceLen ||
		len(j.ActionLog) > fourCardPokerMaxSliceLen {
		return fmt.Errorf("fourcardpoker: input array exceeds maximum allowed size")
	}
	fcp.trumpCards = j.TrumpCards
	if fcp.trumpCards == nil {
		fcp.trumpCards = NewTrumpCards(0)
	}
	fcp.playerHand = j.PlayerHand
	if fcp.playerHand == nil {
		fcp.playerHand = make([]*Card, 0)
	}
	fcp.dealerHand = j.DealerHand
	if fcp.dealerHand == nil {
		fcp.dealerHand = make([]*Card, 0)
	}
	fcp.playerBest = j.PlayerBest
	fcp.dealerBest = j.DealerBest
	if j.Chips != nil {
		fcp.chips = *j.Chips
	}
	fcp.anteBet = j.AnteBet
	fcp.acesUpBet = j.AcesUpBet
	fcp.playBet = j.PlayBet
	fcp.playMultiplier = j.PlayMultiplier
	fcp.phase = j.Phase
	fcp.gameEndFlag = j.GameEndFlag
	fcp.result = j.Result
	fcp.antePayout = j.AntePayout
	fcp.playPayout = j.PlayPayout
	fcp.anteBonusPayout = j.AnteBonusPayout
	fcp.acesUpPayout = j.AcesUpPayout
	fcp.playerHandRank = j.PlayerHandRank
	fcp.dealerHandRank = j.DealerHandRank
	fcp.actionLog = j.ActionLog
	if fcp.actionLog == nil {
		fcp.actionLog = make([]*ActionLogEntry, 0)
	}
	return nil
}
