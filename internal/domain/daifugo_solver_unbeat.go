package domain

import "sort"

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
// Jokers may represent cards below the minimum real card, so the true minimum
// is computed by deducting leftover jokers (after filling interior gaps) from
// the minimum real card strength.
func (s *daifugoSolver) sequenceMinStrength(cards []*Card) int {
	var realStrengths []int
	jokerCount := 0
	for _, c := range cards {
		if IsJoker(c) {
			jokerCount++
		} else {
			realStrengths = append(realStrengths, s.cardStrength(c.GetValue()))
		}
	}
	if len(realStrengths) == 0 {
		return DaifugoJokerStrength
	}
	sort.Ints(realStrengths)
	minReal := realStrengths[0]
	maxReal := realStrengths[len(realStrengths)-1]
	// Interior gaps that jokers fill between real cards
	interiorGaps := (maxReal - minReal + 1) - len(realStrengths)
	if interiorGaps < 0 {
		interiorGaps = 0
	}
	// Remaining jokers extend the sequence below the minimum real card
	extendBelow := jokerCount - interiorGaps
	if extendBelow < 0 {
		extendBelow = 0
	}
	return minReal - extendBelow
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
