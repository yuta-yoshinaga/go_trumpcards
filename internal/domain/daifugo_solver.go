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
		revolution:     d.revolutionActive,
		elevenBack:     d.elevenBackActive,
		sequenceLocked: d.sequenceLocked,
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

// generateOpeningMoves generates all valid plays from a clear table
func (s *daifugoSolver) generateOpeningMoves(hand []*Card) []solverPlay {
	var moves []solverPlay

	// When sequenceLocked, only sequence plays are allowed
	if !s.sequenceLocked {
		groups := s.groupByValue(hand)
		jokers := s.getJokers(hand)

		// Sorted values for deterministic order (strongest first for better pruning)
		values := s.sortedValuesByStrength(groups)

		for _, v := range values {
			group := groups[v]
			// Generate plays of size 1..len(group)
			for size := 1; size <= len(group); size++ {
				play := make([]*Card, size)
				copy(play, group[:size])
				moves = append(moves, solverPlay{cards: play})
			}
			// Augment with jokers for larger groups
			if len(jokers) > 0 {
				maxAug := len(group) + len(jokers)
				for augSize := len(group) + 1; augSize <= maxAug; augSize++ {
					jokersNeeded := augSize - len(group)
					play := make([]*Card, augSize)
					copy(play, group)
					for ji := 0; ji < jokersNeeded; ji++ {
						play[len(group)+ji] = jokers[ji]
					}
					moves = append(moves, solverPlay{cards: play})
				}
			}
		}

		// Pure joker plays
		for i := 1; i <= len(jokers); i++ {
			play := make([]*Card, i)
			copy(play, jokers[:i])
			moves = append(moves, solverPlay{cards: play})
		}
	}

	// Sequence plays (階段)
	if s.config.SequenceEnabled {
		seqMoves := s.generateSequencePlays(hand)
		moves = append(moves, seqMoves...)
	}

	// Mark 8-cut plays
	if s.config.EightCutEnabled {
		for i := range moves {
			if s.containsNonJokerValue(moves[i].cards, 8) {
				moves[i].is8Cut = true
			}
		}
	}

	return moves
}

// generateResponseMoves generates valid plays responding to existing table cards
func (s *daifugoSolver) generateResponseMoves(hand []*Card, tableCards []*Card,
	tableIsSeq bool, suitLocked bool, lockedSuit int, numberLocked bool) []solverPlay {
	var moves []solverPlay
	needed := len(tableCards)

	// Sequence response: generate valid sequence plays that beat the table sequence
	if tableIsSeq {
		if s.config.SequenceEnabled {
			tableMinStr := s.sequenceMinStrength(tableCards)
			seqMoves := s.generateSequenceResponsePlays(hand, needed, tableMinStr)
			moves = append(moves, seqMoves...)
		}
		// Mark 8-cut plays in sequence responses
		if s.config.EightCutEnabled {
			for i := range moves {
				if s.containsNonJokerValue(moves[i].cards, 8) {
					moves[i].is8Cut = true
				}
			}
		}
		return moves
	}

	tableStr := s.playStrength(tableCards)
	groups := s.groupByValue(hand)
	jokers := s.getJokers(hand)

	// Sort by descending strength for determinism and better pruning
	values := s.sortedValuesByStrength(groups)

	for _, v := range values {
		group := groups[v]
		str := s.cardStrength(v)
		if str <= tableStr {
			continue
		}

		// Number lock check
		if numberLocked && s.config.NumberLockEnabled {
			tableBase := getBaseValue(tableCards)
			if tableBase > 0 {
				tableBaseStr := s.cardStrength(tableBase)
				if str-tableBaseStr != 1 {
					continue
				}
			}
		}

		if len(group) >= needed {
			play := s.selectCardsForSuitLock(group, needed, suitLocked, lockedSuit)
			if play != nil {
				moves = append(moves, solverPlay{cards: play})
			}
		}
		if len(group) < needed && len(group)+len(jokers) >= needed {
			play := s.selectCardsForSuitLock(group, len(group), suitLocked, lockedSuit)
			if play != nil {
				augmented := make([]*Card, needed)
				copy(augmented, play)
				jokersNeeded := needed - len(group)
				for ji := 0; ji < jokersNeeded; ji++ {
					augmented[len(group)+ji] = jokers[ji]
				}
				moves = append(moves, solverPlay{cards: augmented})
			}
		}
	}

	// Joker single
	if needed == 1 && len(jokers) > 0 && DaifugoJokerStrength > tableStr {
		moves = append(moves, solverPlay{cards: []*Card{jokers[0]}})
	}

	// Pure joker group
	if needed > 1 && len(jokers) >= needed && DaifugoJokerStrength > tableStr {
		play := make([]*Card, needed)
		copy(play, jokers[:needed])
		moves = append(moves, solverPlay{cards: play})
	}

	// Spade-3 counter: CPU gets another turn with spade-3 on table
	if s.config.SpadeThreeEnabled && needed == 1 && len(tableCards) == 1 && IsJoker(tableCards[0]) {
		for _, c := range hand {
			if !IsJoker(c) && c.GetDesign() == CardDesignSpade && c.GetValue() == 3 {
				moves = append(moves, solverPlay{cards: []*Card{c}, isSpadeThreeReturn: true})
				break
			}
		}
	}

	// Mark 8-cut plays
	if s.config.EightCutEnabled {
		for i := range moves {
			if s.containsNonJokerValue(moves[i].cards, 8) {
				moves[i].is8Cut = true
			}
		}
	}

	return moves
}

// selectCardsForSuitLock picks `needed` cards from group satisfying suit lock.
// For full lock, the first non-joker card must match the locked suit.
// Returns nil if suit lock cannot be satisfied.
func (s *daifugoSolver) selectCardsForSuitLock(group []*Card, needed int,
	suitLocked bool, lockedSuit int) []*Card {
	if !suitLocked || s.config.SuitLockMode == DaifugoSuitLockNone {
		play := make([]*Card, needed)
		copy(play, group[:needed])
		return play
	}

	if s.config.SuitLockMode == DaifugoSuitLockFull {
		// Full lock: first non-joker card's suit must match locked suit
		matchIdx := -1
		for i, c := range group {
			if !IsJoker(c) && c.GetDesign() == lockedSuit {
				matchIdx = i
				break
			}
		}
		if matchIdx < 0 {
			return nil
		}
		// guaranteed: len(group) >= needed from all callers
		play := make([]*Card, needed)
		play[0] = group[matchIdx]
		j := 1
		for i := 0; i < len(group) && j < needed; i++ {
			if i != matchIdx {
				play[j] = group[i]
				j++
			}
		}
		return play
	}

	// Partial lock: at least one card must match
	hasMatch := false
	for _, c := range group {
		if c.GetDesign() == lockedSuit {
			hasMatch = true
			break
		}
	}
	if !hasMatch {
		return nil
	}
	play := make([]*Card, needed)
	copy(play, group[:needed])
	return play
}

// isUnbeatablePlay checks if a solverPlay cannot be beaten by any single opponent.
// It dispatches to sequence or group unbeatability checks.
func (s *daifugoSolver) isUnbeatablePlay(move solverPlay) bool {
	if move.isSequence {
		return s.isUnbeatableSequence(move.cards)
	}
	return s.isUnbeatable(move.cards)
}

// isUnbeatableSequence checks if a sequence play cannot be beaten by any opponent.
// An opponent can beat a sequence by having a same-length sequence of the same suit
// (or different suit) with higher min strength. We conservatively check all suits.
func (s *daifugoSolver) isUnbeatableSequence(play []*Card) bool {
	needed := len(play)
	playMinStr := s.sequenceMinStrength(play)

	for _, oppHand := range s.oppHands {
		if s.canBeatSequence(oppHand, needed, playMinStr) {
			return false
		}
	}
	return true
}

// canBeatSequence checks if an opponent's hand can form a sequence of `needed` cards
// with a higher min strength than `playMinStr`.
func (s *daifugoSolver) canBeatSequence(oppHand []*Card, needed int, playMinStr int) bool {
	// Group non-joker cards by suit and sort by strength
	suitGroups := make(map[int][]int) // suit -> sorted strengths
	jokerCount := 0
	for _, c := range oppHand {
		if IsJoker(c) {
			jokerCount++
			continue
		}
		suit := c.GetDesign()
		str := s.cardStrength(c.GetValue())
		suitGroups[suit] = append(suitGroups[suit], str)
	}

	for _, strengths := range suitGroups {
		sort.Ints(strengths)
		// Try to build a sequence of `needed` starting from each position
		for si := 0; si < len(strengths); si++ {
			if s.canFormSequenceFromStrengths(strengths, si, jokerCount, needed, playMinStr) {
				return true
			}
		}
	}

	return false
}

// canFormSequenceFromStrengths checks if we can form a consecutive sequence of `needed` cards
// starting from strengths[si], filling gaps with jokers, such that the minimum strength > playMinStr.
func (s *daifugoSolver) canFormSequenceFromStrengths(strengths []int, si int, jokerCount int, needed int, playMinStr int) bool {
	count := 1
	lastStr := strengths[si]
	jokersUsed := 0
	sci := si + 1

	for count < needed {
		targetStr := lastStr + 1
		if sci < len(strengths) && strengths[sci] == targetStr {
			count++
			lastStr = targetStr
			sci++
		} else if sci < len(strengths) && strengths[sci] == lastStr {
			// Duplicate strength, skip
			sci++
			continue
		} else if jokersUsed < jokerCount {
			jokersUsed++
			count++
			lastStr = targetStr
		} else {
			return false
		}
	}

	// Check that the minimum strength of this sequence beats the play
	minStr := strengths[si]
	return minStr > playMinStr
}

// sequenceMinStrength returns the minimum card strength in a sequence play.
func (s *daifugoSolver) sequenceMinStrength(cards []*Card) int {
	minStr := DaifugoJokerStrength + 1
	for _, c := range cards {
		var str int
		if IsJoker(c) {
			str = DaifugoJokerStrength
		} else {
			str = s.cardStrength(c.GetValue())
		}
		if str < minStr {
			minStr = str
		}
	}
	return minStr
}

// generateSequencePlays generates all valid sequence plays (3+ cards, same suit, consecutive)
// from the given hand for opening (clear table).
func (s *daifugoSolver) generateSequencePlays(hand []*Card) []solverPlay {
	var moves []solverPlay

	// Group non-joker cards by suit
	suitCards := make(map[int][]*Card) // suit -> cards
	var jokers []*Card
	for _, c := range hand {
		if IsJoker(c) {
			jokers = append(jokers, c)
			continue
		}
		suit := c.GetDesign()
		suitCards[suit] = append(suitCards[suit], c)
	}

	for _, cards := range suitCards {
		// Sort by strength
		sort.Slice(cards, func(i, j int) bool {
			return s.cardStrength(cards[i].GetValue()) < s.cardStrength(cards[j].GetValue())
		})
		seqs := s.findAllSequences(cards, jokers, 3)
		for _, seq := range seqs {
			moves = append(moves, solverPlay{cards: seq, isSequence: true})
		}
	}

	return moves
}

// generateSequenceResponsePlays generates valid sequence plays that beat the table sequence.
func (s *daifugoSolver) generateSequenceResponsePlays(hand []*Card, needed int, tableMinStr int) []solverPlay {
	var moves []solverPlay

	// Group non-joker cards by suit
	suitCards := make(map[int][]*Card)
	var jokers []*Card
	for _, c := range hand {
		if IsJoker(c) {
			jokers = append(jokers, c)
			continue
		}
		suit := c.GetDesign()
		suitCards[suit] = append(suitCards[suit], c)
	}

	for _, cards := range suitCards {
		sort.Slice(cards, func(i, j int) bool {
			return s.cardStrength(cards[i].GetValue()) < s.cardStrength(cards[j].GetValue())
		})
		seqs := s.findAllSequences(cards, jokers, needed)
		for _, seq := range seqs {
			if len(seq) == needed && s.sequenceMinStrength(seq) > tableMinStr {
				moves = append(moves, solverPlay{cards: seq, isSequence: true})
			}
		}
	}

	return moves
}

// findAllSequences finds all valid sequences of exactly `minLen` or more from sorted same-suit cards,
// using jokers to fill gaps.
func (s *daifugoSolver) findAllSequences(sortedCards []*Card, jokers []*Card, minLen int) [][]*Card {
	var results [][]*Card

	n := len(sortedCards)
	for si := 0; si < n; si++ {
		// Try building sequences of different lengths starting from sortedCards[si]
		for targetLen := minLen; targetLen <= n+len(jokers); targetLen++ {
			seq := s.tryBuildSolverSequence(sortedCards, si, jokers, targetLen)
			if seq != nil {
				results = append(results, seq)
			} else {
				break // Can't build longer sequence if this fails
			}
		}
	}

	return results
}

// tryBuildSolverSequence attempts to build a sequence of `needed` cards starting from
// sortedCards[si], filling gaps with jokers.
func (s *daifugoSolver) tryBuildSolverSequence(sortedCards []*Card, si int, jokers []*Card, needed int) []*Card {
	seq := []*Card{sortedCards[si]}
	lastStr := s.cardStrength(sortedCards[si].GetValue())
	jokersUsed := 0
	sci := si + 1

	for len(seq) < needed {
		targetStr := lastStr + 1

		if sci < len(sortedCards) && s.cardStrength(sortedCards[sci].GetValue()) == targetStr {
			seq = append(seq, sortedCards[sci])
			lastStr = targetStr
			sci++
		} else if sci < len(sortedCards) && s.cardStrength(sortedCards[sci].GetValue()) == lastStr {
			// Duplicate strength, skip
			sci++
			continue
		} else if jokersUsed < len(jokers) {
			seq = append(seq, jokers[jokersUsed])
			jokersUsed++
			lastStr = targetStr
		} else {
			return nil
		}
	}

	return seq
}

// isUnbeatable checks if a play cannot be beaten by any single opponent
func (s *daifugoSolver) isUnbeatable(play []*Card) bool {
	for _, oppHand := range s.oppHands {
		if s.canBeat(oppHand, play) {
			return false
		}
	}
	return true
}

// canBeat checks if an opponent's hand contains a play that can beat the given play.
// Note: number lock and suit lock constraints on opponents are not checked here.
// This makes isUnbeatable conservative: it may return false even when opponents
// cannot legally play a beating card. Safe but may miss some winning sequences.
func (s *daifugoSolver) canBeat(oppHand []*Card, play []*Card) bool {
	count := len(play)
	playStr := s.playStrength(play)

	// Spade-3 counter: opponent can counter single joker with spade 3
	if s.config.SpadeThreeEnabled && count == 1 && IsJoker(play[0]) {
		for _, c := range oppHand {
			if !IsJoker(c) && c.GetDesign() == CardDesignSpade && c.GetValue() == 3 {
				return true
			}
		}
	}

	groups := s.groupByValue(oppHand)
	jokers := s.getJokers(oppHand)

	// Check value groups (with optional joker augmentation)
	for _, group := range groups {
		str := s.cardStrength(group[0].GetValue())
		if str > playStr {
			if len(group) >= count {
				return true
			}
			if len(group)+len(jokers) >= count {
				return true
			}
		}
	}

	// Single joker beats non-joker single
	if count == 1 && !IsJoker(play[0]) && len(jokers) > 0 && DaifugoJokerStrength > playStr {
		return true
	}

	// Pure joker group can beat any play with lower strength
	if count > 1 && len(jokers) >= count && DaifugoJokerStrength > playStr {
		return true
	}

	return false
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

// applyRevolution toggles revolution if the play has 4+ cards
func (s *daifugoSolver) applyRevolution(cards []*Card) {
	if len(cards) >= 4 {
		s.revolution = !s.revolution
	}
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
	result := solver.solve(cpuHand, d.tableCards, d.tableIsSequence,
		d.suitLocked, d.lockedSuit, d.numberLocked)
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
