package domain

import "sort"

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
		// Build exact-length sequences directly (no need for findAllSequences)
		for si := range cards {
			seq := s.tryBuildSolverSequence(cards, si, jokers, needed)
			if seq != nil && s.sequenceMinStrength(seq) > tableMinStr {
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
