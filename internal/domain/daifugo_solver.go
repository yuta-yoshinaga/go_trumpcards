package domain

import "sort"

// DaifugoSolverMaxCards is the maximum hand size for triggering the endgame solver.
// When the CPU (Hard difficulty) has at most this many cards, the solver attempts
// to find a guaranteed winning play sequence.
const DaifugoSolverMaxCards = 8

// daifugoSolver performs endgame "perfect read" (完全読み) analysis for Hard AI.
// It searches for a guaranteed winning sequence of plays, where each intermediate
// play is either an 8-cut or unbeatable by all opponents.
type daifugoSolver struct {
	oppHands       [][]*Card     // Each active opponent's hand
	revolution     bool          // Current revolution state
	elevenBack     bool          // Current eleven back state
	sequenceLocked bool          // Only sequence plays allowed
	config         DaifugoConfig // Game config (for local rules)
}

func newDaifugoSolver(d *Daifugo, oppHands [][]*Card) *daifugoSolver {
	return &daifugoSolver{
		oppHands:       oppHands,
		revolution:     d.round.revolutionActive,
		elevenBack:     d.round.elevenBackActive,
		sequenceLocked: d.round.sequenceLocked,
		config:         d.config,
	}
}

func (s *daifugoSolver) cardStrength(v int) int {
	reversed := s.revolution != s.elevenBack
	if reversed {
		return DaifugoCardStrengthRevolution(v)
	}
	return DaifugoCardStrength(v)
}

// solverPlay represents a possible play (set of cards) from the hand
type solverPlay struct {
	cards              []*Card
	is8Cut             bool
	isSpadeThreeReturn bool // spade-3 counter: CPU gets another turn with this card on table
	isSequence         bool // true when this play is a sequence (階段)
}

// solve finds a guaranteed winning first play from the CPU's hand.
// tableCards nil means the table is clear and CPU has the lead.
// Returns the first play's cards or nil if no guaranteed win exists.
func (s *daifugoSolver) solve(hand []*Card, tableCards []*Card, tableIsSeq bool,
	suitLocked bool, lockedSuit int, numberLocked bool) []*Card {
	if len(hand) == 0 {
		return nil
	}

	var moves []solverPlay
	if tableCards == nil {
		moves = s.generateOpeningMoves(hand)
	} else {
		moves = s.generateResponseMoves(hand, tableCards, tableIsSeq,
			suitLocked, lockedSuit, numberLocked)
	}

	for _, move := range moves {
		remaining := s.removeCards(hand, move.cards)

		// Last play: empties hand
		if len(remaining) == 0 {
			if s.wouldBeIllegalFinish(move.cards) {
				continue
			}
			return move.cards
		}

		// Save state for backtracking
		savedRev, savedEB := s.revolution, s.elevenBack

		// 8-cut: table clears regardless of opponents
		if move.is8Cut {
			if s.trySolveWithClearTable(remaining, move.cards, savedRev, savedEB) {
				return move.cards
			}
			continue
		}

		// Spade-3 counter: CPU gets another turn with spade-3 on table
		if move.isSpadeThreeReturn {
			// Table shows spade-3 (strength 3), CPU's turn, no constraints
			if s.solve(remaining, move.cards, false, false, 0, false) != nil {
				return move.cards
			}
			continue
		}

		// Non-8-cut: check if unbeatable by all opponents
		if s.isUnbeatablePlay(move) {
			if s.trySolveWithClearTable(remaining, move.cards, savedRev, savedEB) {
				return move.cards
			}
		}
	}

	return nil
}

// trySolveWithClearTable attempts to solve after a table-clearing play (8-cut or unbeatable).
// It applies revolution, resets eleven back, and handles state restore for backtracking.
func (s *daifugoSolver) trySolveWithClearTable(remaining []*Card, moveCards []*Card,
	savedRev, savedEB bool) bool {
	s.applyRevolution(moveCards)
	s.elevenBack = false // table clear resets eleven back
	if s.solve(remaining, nil, false, false, 0, false) != nil {
		s.revolution, s.elevenBack = savedRev, savedEB
		return true
	}
	s.revolution, s.elevenBack = savedRev, savedEB
	return false
}

// applyRevolution toggles revolution if the play has 4+ cards
func (s *daifugoSolver) applyRevolution(cards []*Card) {
	if len(cards) >= 4 {
		s.revolution = !s.revolution
	}
}

// playStrength returns the strength of a play
func (s *daifugoSolver) playStrength(play []*Card) int {
	base := getBaseValue(play)
	if base < 0 {
		return DaifugoJokerStrength
	}
	return s.cardStrength(base)
}

// wouldBeIllegalFinish checks if finishing with these cards would be an illegal finish
func (s *daifugoSolver) wouldBeIllegalFinish(cards []*Card) bool {
	if !s.config.IllegalFinishEnabled {
		return false
	}
	// 8-cut finish
	if s.config.EightCutEnabled {
		for _, c := range cards {
			if !IsJoker(c) && c.GetValue() == 8 {
				return true
			}
		}
	}
	// Joker finish
	for _, c := range cards {
		if IsJoker(c) {
			return true
		}
	}
	// Revolution finish (4+ cards)
	if len(cards) >= 4 {
		return true
	}
	return false
}

// Helper methods

func (s *daifugoSolver) groupByValue(cards []*Card) map[int][]*Card {
	groups := make(map[int][]*Card)
	for _, c := range cards {
		if IsJoker(c) {
			continue
		}
		groups[c.GetValue()] = append(groups[c.GetValue()], c)
	}
	return groups
}

func (s *daifugoSolver) getJokers(cards []*Card) []*Card {
	var jokers []*Card
	for _, c := range cards {
		if IsJoker(c) {
			jokers = append(jokers, c)
		}
	}
	return jokers
}

// sortedValuesByStrength returns values sorted by descending card strength
func (s *daifugoSolver) sortedValuesByStrength(groups map[int][]*Card) []int {
	values := make([]int, 0, len(groups))
	for v := range groups {
		values = append(values, v)
	}
	sort.Slice(values, func(i, j int) bool {
		return s.cardStrength(values[i]) > s.cardStrength(values[j])
	})
	return values
}

func (s *daifugoSolver) containsNonJokerValue(cards []*Card, value int) bool {
	for _, c := range cards {
		if !IsJoker(c) && c.GetValue() == value {
			return true
		}
	}
	return false
}

func (s *daifugoSolver) removeCards(hand []*Card, toRemove []*Card) []*Card {
	removeSet := make(map[*Card]bool)
	for _, c := range toRemove {
		removeSet[c] = true
	}
	remaining := make([]*Card, 0, len(hand)-len(toRemove))
	for _, c := range hand {
		if !removeSet[c] {
			remaining = append(remaining, c)
		}
	}
	return remaining
}

// trySolveEndgame attempts to find a guaranteed winning play sequence for the Hard AI.
// Returns the first play as hand indices, or nil if no guaranteed win exists.
func (d *Daifugo) trySolveEndgame(player *DaifugoPlayer) []int {
	if player.GetCardsSize() > DaifugoSolverMaxCards || player.GetCardsSize() == 0 {
		return nil
	}

	// Collect CPU hand
	cpuHand := make([]*Card, player.GetCardsSize())
	for i := 0; i < player.GetCardsSize(); i++ {
		cpuHand[i] = player.GetCard(i)
	}

	// Collect active opponent hands
	var oppHands [][]*Card
	for _, p := range d.players {
		if p == player || p.GetIsFinished() {
			continue
		}
		hand := make([]*Card, p.GetCardsSize())
		for i := 0; i < p.GetCardsSize(); i++ {
			hand[i] = p.GetCard(i)
		}
		if len(hand) > 0 {
			oppHands = append(oppHands, hand)
		}
	}

	solver := newDaifugoSolver(d, oppHands)
	result := solver.solve(cpuHand, d.round.tableCards, d.round.tableIsSequence,
		d.round.suitLocked, d.round.lockedSuit, d.round.numberLocked)
	if result == nil {
		return nil
	}

	return d.mapSolverCardsToIndices(player, result)
}

// mapSolverCardsToIndices converts solver card pointers back to player hand indices
func (d *Daifugo) mapSolverCardsToIndices(player *DaifugoPlayer, cards []*Card) []int {
	indices := make([]int, 0, len(cards))
	used := make(map[int]bool)
	for _, c := range cards {
		for i := 0; i < player.GetCardsSize(); i++ {
			if !used[i] && player.GetCard(i) == c {
				indices = append(indices, i)
				used[i] = true
				break
			}
		}
	}
	sort.Ints(indices)
	return indices
}
