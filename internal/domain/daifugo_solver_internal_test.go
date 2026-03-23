package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func solverConfig() DaifugoConfig {
	return DaifugoConfig{
		EightCutEnabled:      true,
		IllegalFinishEnabled: true,
	}
}

func TestDaifugoSolver_solve(t *testing.T) {
	t.Run("single strongest card wins from clear table", func(t *testing.T) {
		// CPU has: 2♠ (strength 15)
		// Opponent has: A♠ (strength 14)
		cpuHand := []*Card{NewCard(CardDesignSpade, 2, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, 2, result[0].GetValue())
	})

	t.Run("single card always wins as last play", func(t *testing.T) {
		// CPU has only K♠ → play it as last play → finish (regardless of opponent cards)
		cpuHand := []*Card{NewCard(CardDesignSpade, 13, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("cannot win when intermediate plays are all beatable", func(t *testing.T) {
		// CPU has: K♠, 5♠ (strength 13, 5) - both beatable by opponent
		// Opponent has: 2♠ (strength 15)
		// Neither K nor 5 is unbeatable as intermediate play
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignSpade, 5, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("cannot win when intermediate beatable by joker", func(t *testing.T) {
		// CPU has: 2♠, A♠ (strength 15, 14)
		// Opponent has: Joker (strength 16) - can beat both
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignSpade, 1, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignJoker, 0, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("wins with multiple unbeatable plays", func(t *testing.T) {
		// CPU has: 2♠, A♠ (strength 15, 14)
		// Opponent has: K♠, Q♠ (strength 13, 12)
		// CPU plays 2 first (unbeatable), then A (unbeatable), win!
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 13, false), NewCard(CardDesignSpade, 12, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("wins with 8-cut then unbeatable last play", func(t *testing.T) {
		// CPU has: 8♠, 2♠
		// Opponent has: A♠
		// CPU plays 8 → 8-cut → plays 2 (strength 15 > 14), last play → win
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("8-cut finish is illegal", func(t *testing.T) {
		// CPU has only: 8♠
		// 8-cut finish is illegal → no guaranteed win
		cpuHand := []*Card{NewCard(CardDesignSpade, 8, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("joker finish is illegal", func(t *testing.T) {
		// CPU has only: Joker
		cpuHand := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("illegal finish disabled allows 8-cut finish", func(t *testing.T) {
		cpuHand := []*Card{NewCard(CardDesignSpade, 8, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		cfg := DaifugoConfig{EightCutEnabled: true, IllegalFinishEnabled: false}
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("pair play wins as unbeatable singles", func(t *testing.T) {
		// CPU has: 2♠, 2♥ (both strength 15, unbeatable as singles)
		// Opponent has: A♠, K♠ (strength 14, 13)
		// CPU plays 2♠ first (unbeatable), then 2♥ (last play) → win
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignSpade, 13, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		// Solver plays one 2 as single (unbeatable)
		assert.Equal(t, 1, len(result))
	})

	t.Run("pair beatable as intermediate prevents win", func(t *testing.T) {
		// CPU has: K♠, K♥, 5♠ (3 cards)
		// Opponent has: A♠, A♥ (pair of A beats pair of K)
		// Singles: K♠ beatable by A♠, K♥ beatable by A♥, 5♠ beatable
		// Pair KK: beatable by AA (intermediate)
		// No unbeatable intermediate play → no guaranteed win
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 13, false),
			NewCard(CardDesignHeart, 13, false),
			NewCard(CardDesignSpade, 5, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false), NewCard(CardDesignHeart, 1, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("pair beatable by opponent joker augmentation prevents win", func(t *testing.T) {
		// CPU has: A♠, A♥, 5♠ (3 cards)
		// Opponent has: 2♠, Joker (can form pair of 2+joker, strength 15 > 14)
		// Singles: A♠ beatable by joker, A♥ beatable by joker
		// Pair AA: beatable by 2+joker (intermediate)
		// No unbeatable intermediate → no guaranteed win
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
			NewCard(CardDesignSpade, 5, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignJoker, 0, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("opponents checked individually not pooled", func(t *testing.T) {
		// CPU has: A♠, A♥ (pair of A, strength 14)
		// Opponent1 has: 2♠ (can't form pair alone)
		// Opponent2 has: 2♥ (can't form pair alone)
		// Pooled they have pair of 2, but individually neither can beat
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)},
			{NewCard(CardDesignHeart, 2, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("empty hand returns nil", func(t *testing.T) {
		solver := &daifugoSolver{config: solverConfig()}
		result := solver.solve(nil, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("no opponents means all plays win", func(t *testing.T) {
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
		}
		solver := &daifugoSolver{oppHands: nil, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("spade-3 counter and illegal finish rule prevent win", func(t *testing.T) {
		// CPU has: Joker, 2♠. SpadeThreeEnabled.
		// Joker as intermediate: beatable by spade-3 → skip
		// 2♠ as intermediate: unbeatable (strength 15 > 3) → play 2♠ first
		// But remaining Joker: illegal finish! So 2♠ first doesn't work either.
		// No winning sequence.
		cpuHand := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		cfg := solverConfig()
		cfg.SpadeThreeEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("spade-3 avoided when illegal finish disabled", func(t *testing.T) {
		// Same setup but IllegalFinishEnabled = false
		// CPU plays 2♠ (unbeatable) → Joker (last play, allowed) → win
		cpuHand := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		cfg := DaifugoConfig{SpadeThreeEnabled: true, EightCutEnabled: true}
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 2, result[0].GetValue())
	})

	t.Run("revolution changes strength order", func(t *testing.T) {
		// Revolution active: 3 is strongest (strength 16-3=15... wait)
		// DaifugoCardStrengthRevolution(v) = 18 - DaifugoCardStrength(v)
		// 3: 18 - 3 = 15, 2: 18 - 15 = 3, A: 18 - 14 = 4
		// So in revolution, 3 is strongest
		cpuHand := []*Card{NewCard(CardDesignSpade, 3, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)}, // strength 3 in revolution
		}
		solver := &daifugoSolver{
			oppHands:   oppHands,
			revolution: true,
			config:     solverConfig(),
		}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		// 3 has strength 15, 2 has strength 3 → 3 is unbeatable
		assert.NotNil(t, result)
	})

	t.Run("revolution toggled by 4-card play affects subsequent plays", func(t *testing.T) {
		// CPU has: 5♠,5♥,5♦,5♣, 3♠ (quad of 5 + single 3)
		// Opponent has: 6♠, 6♥ (pair of 6 blocks pair of 5, single 6 blocks single 5)
		// Normal: 5 (str 5) beatable by 6 (str 6). Pair of 5 beatable by pair of 6.
		// Triple of 5: unbeatable (opponent can't form triple), but remaining [5♣,3♠]
		//   → 5♣ beatable by 6, 3♠ beatable by 6 → no win from triple path
		// Quad of 5 → revolution! Now 3♠ has strength 15, 6 has strength 12
		// CPU plays 3♠ as last play → win
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 5, false),
			NewCard(CardDesignSpade, 3, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 6, false), NewCard(CardDesignHeart, 6, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		// First play should be quad of 5 (triggers revolution)
		assert.Equal(t, 4, len(result))
	})

	t.Run("response to table cards", func(t *testing.T) {
		// Table has: K♠ (single, strength 13)
		// CPU has: A♠, 2♠ (can beat K with A or 2, then play remaining)
		// Opponent has: Q♠
		tableCards := []*Card{NewCard(CardDesignSpade, 13, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 12, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, tableCards, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		// Either A or 2 is valid (both beat K and are unbeatable by Q)
		v := result[0].GetValue()
		assert.True(t, v == 1 || v == 2, "expected A or 2, got %d", v)
	})

	t.Run("response with 8-cut clears table", func(t *testing.T) {
		// Table has: 7♠ (strength 7)
		// CPU has: 8♠, 3♠
		// Opponent has: 2♠
		// CPU plays 8 to beat 7 → 8-cut → table clears → plays 3 (last play)
		tableCards := []*Card{NewCard(CardDesignSpade, 7, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignSpade, 3, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, tableCards, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 8, result[0].GetValue())
	})

	t.Run("sequence response beats table sequence", func(t *testing.T) {
		// Table: 3♠-4♠-5♠ sequence (min strength 3)
		// CPU has: 6♠-7♠-8♠ (min strength 6, beats 3)
		// No opponents → should win
		tableCards := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 8, false),
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		solver := &daifugoSolver{oppHands: nil, config: cfg}
		result := solver.solve(cpuHand, tableCards, true, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("last play can be beatable", func(t *testing.T) {
		// CPU has: 2♠, 5♠
		// Opponent has: A♠
		// CPU plays 2 first (unbeatable: 15 > 14), then 5 as last play (beatable but OK)
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: solverConfig()}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("pure joker group is unbeatable", func(t *testing.T) {
		// CPU has: Joker1, Joker2, 3♠
		// Opponent has: 2♠, 2♥
		// CPU plays pair of jokers (strength 16, unbeatable as pair)
		// Then plays 3♠ as last play
		cpuHand := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignSpade, 3, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignHeart, 2, false)},
		}
		cfg := solverConfig()
		cfg.IllegalFinishEnabled = false // allow joker finish if needed
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
	})

	t.Run("number lock constrains response", func(t *testing.T) {
		// Table: 10♠ (strength 10), number lock active
		// CPU has: J♠ (strength 11, diff=1 OK), 2♠
		// Opponent has: 5♠
		tableCards := []*Card{NewCard(CardDesignSpade, 10, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 5, false)},
		}
		cfg := solverConfig()
		cfg.NumberLockEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, tableCards, false, false, 0, true)
		assert.NotNil(t, result)
		assert.Equal(t, 11, result[0].GetValue())
	})

	t.Run("number lock blocks non-consecutive response", func(t *testing.T) {
		// Table: 10♠ (strength 10), number lock active
		// CPU has: 2♠ (strength 15, diff=5, not 1) → blocked
		tableCards := []*Card{NewCard(CardDesignSpade, 10, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 5, false)},
		}
		cfg := solverConfig()
		cfg.NumberLockEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, tableCards, false, false, 0, true)
		assert.Nil(t, result)
	})

	t.Run("spade-3 counter fails to guarantee win continues to next move", func(t *testing.T) {
		// Revolution active: 3♠ has strength 15 (strongest), 5♠ has strength 13.
		// Table: Joker. CPU has: 3♠, 5♠. SpadeThreeEnabled.
		// Spade-3 counter: table becomes [3♠] (str 15 in revolution).
		// CPU remaining: 5♠ (str 13) can't beat 3♠ (str 15) → no response → fail.
		// No other move beats Joker → result nil.
		tableCards := []*Card{NewCard(CardDesignJoker, 0, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 5, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)},
		}
		cfg := solverConfig()
		cfg.SpadeThreeEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, revolution: true, config: cfg}
		result := solver.solve(cpuHand, tableCards, false, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("spade-3 counter response to joker on table", func(t *testing.T) {
		// Table: Joker. CPU has: 3♠, 2♠. SpadeThreeEnabled.
		// CPU can play spade-3 counter, then play 2 from clear table
		tableCards := []*Card{NewCard(CardDesignJoker, 0, false)}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 5, false)},
		}
		cfg := solverConfig()
		cfg.SpadeThreeEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, tableCards, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, CardDesignSpade, result[0].GetDesign())
		assert.Equal(t, 3, result[0].GetValue())
	})

	t.Run("sequence play as opening wins", func(t *testing.T) {
		// CPU has: A♠, 2♠, 3♠ — forms a sequence (strengths 14,15,3 — no, wait)
		// In normal order: 3→3, 4→4, ..., A→14, 2→15
		// So 3♠-4♠-5♠ would be a valid sequence with strengths 3,4,5
		// CPU has: Q♠, K♠, A♠ (strengths 12,13,14) — valid sequence
		// Opponent has: 5♠ (strength 5) — cannot form a 3-card sequence
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 12, false), // Q
			NewCard(CardDesignSpade, 13, false), // K
			NewCard(CardDesignSpade, 1, false),  // A
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 5, false)},
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("sequence with joker gap-filling wins", func(t *testing.T) {
		// CPU has: 4♠, Joker, 6♠ — sequence 4-5(joker)-6 (strengths 4,5,6)
		// Opponent has: 3♠ — cannot beat this sequence
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignSpade, 6, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignHeart, 3, false)},
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		cfg.IllegalFinishEnabled = false // joker in sequence is not a finish
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("sequence is only winning path", func(t *testing.T) {
		// CPU has: 10♠, J♠, Q♠ (strengths 10,11,12) — valid sequence
		// Opponent has: 2♠, A♠ (strengths 15,14) — beats any single card
		// Singles: 10 beatable by A, J beatable by A, Q beatable by A
		// Pairs: can't form pairs
		// Sequence: 10-J-Q unbeatable (opponent has 2♠+A♠, different values, same suit
		//   but A-2 are not consecutive in strength: 14,15 — only 2 cards, need 3)
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignSpade, 11, false),
			NewCard(CardDesignSpade, 12, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false), NewCard(CardDesignSpade, 1, false)},
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("sequence response with remaining cards wins", func(t *testing.T) {
		// Table: 3♥-4♥-5♥ sequence (min strength 3)
		// CPU has: 6♥-7♥-8♥ (beats table) + 2♠ (unbeatable single)
		// Opponent has: 9♠ (can't form 3-card sequence, single 9 strength 9)
		// CPU plays 6-7-8 (sequence response, 8-cut triggers since 8♥ present)
		// After 8-cut, table clears → play 2♠ as last card → win
		tableCards := []*Card{
			NewCard(CardDesignHeart, 3, false),
			NewCard(CardDesignHeart, 4, false),
			NewCard(CardDesignHeart, 5, false),
		}
		cpuHand := []*Card{
			NewCard(CardDesignHeart, 6, false),
			NewCard(CardDesignHeart, 7, false),
			NewCard(CardDesignHeart, 8, false),
			NewCard(CardDesignSpade, 2, false),
		}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 9, false)},
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, tableCards, true, false, 0, false)
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("sequence disabled returns nil for sequence table", func(t *testing.T) {
		// SequenceEnabled = false → no sequence moves generated
		tableCards := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignSpade, 7, false),
			NewCard(CardDesignSpade, 8, false),
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = false
		solver := &daifugoSolver{oppHands: nil, config: cfg}
		result := solver.solve(cpuHand, tableCards, true, false, 0, false)
		assert.Nil(t, result)
	})

	t.Run("sequence beatable by opponent sequence prevents win", func(t *testing.T) {
		// CPU has: 5♠-6♠-7♠ (strengths 5,6,7, min=5) — only option is sequence
		// Opponent has: 8♥-9♥-10♥ (strengths 8,9,10, min=8 > 5) — can beat
		// Singles: 5,6,7 all beatable by 8,9,10
		// No winning path
		cpuHand := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 6, false),
			NewCard(CardDesignSpade, 7, false),
		}
		oppHands := [][]*Card{
			{
				NewCard(CardDesignHeart, 8, false),
				NewCard(CardDesignHeart, 9, false),
				NewCard(CardDesignHeart, 10, false),
			},
		}
		cfg := solverConfig()
		cfg.SequenceEnabled = true
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		result := solver.solve(cpuHand, nil, false, false, 0, false)
		assert.Nil(t, result)
	})
}

func TestDaifugoSolver_isUnbeatable(t *testing.T) {
	t.Run("strongest single is unbeatable", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 2, false)} // strength 15
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 1, false)}, // strength 14
		}
		solver := &daifugoSolver{oppHands: oppHands, config: DaifugoConfig{}}
		assert.True(t, solver.isUnbeatable(play))
	})

	t.Run("not unbeatable when opponent has stronger", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 1, false)} // strength 14
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 2, false)}, // strength 15
		}
		solver := &daifugoSolver{oppHands: oppHands, config: DaifugoConfig{}}
		assert.False(t, solver.isUnbeatable(play))
	})

	t.Run("joker single is unbeatable without spade-3", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		solver := &daifugoSolver{oppHands: oppHands, config: DaifugoConfig{}}
		assert.True(t, solver.isUnbeatable(play))
	})

	t.Run("joker single beatable with spade-3", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHands := [][]*Card{
			{NewCard(CardDesignSpade, 3, false)},
		}
		cfg := DaifugoConfig{SpadeThreeEnabled: true}
		solver := &daifugoSolver{oppHands: oppHands, config: cfg}
		assert.False(t, solver.isUnbeatable(play))
	})

	t.Run("no opponents means unbeatable", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 3, false)}
		solver := &daifugoSolver{oppHands: nil, config: DaifugoConfig{}}
		assert.True(t, solver.isUnbeatable(play))
	})
}

func TestDaifugoSolver_canBeat(t *testing.T) {
	t.Run("stronger single beats weaker", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 13, false)}   // K, strength 13
		oppHand := []*Card{NewCard(CardDesignSpade, 1, false)} // A, strength 14
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.True(t, solver.canBeat(oppHand, play))
	})

	t.Run("weaker single cannot beat stronger", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 1, false)}     // A, strength 14
		oppHand := []*Card{NewCard(CardDesignSpade, 13, false)} // K, strength 13
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.False(t, solver.canBeat(oppHand, play))
	})

	t.Run("joker beats non-joker single", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignSpade, 2, false)}    // strength 15
		oppHand := []*Card{NewCard(CardDesignJoker, 0, false)} // strength 16
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.True(t, solver.canBeat(oppHand, play))
	})

	t.Run("cannot beat joker single without spade-3", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHand := []*Card{NewCard(CardDesignSpade, 2, false)} // 2 has strength 15 < 16
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.False(t, solver.canBeat(oppHand, play))
	})

	t.Run("spade-3 beats joker", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHand := []*Card{NewCard(CardDesignSpade, 3, false)}
		cfg := DaifugoConfig{SpadeThreeEnabled: true}
		solver := &daifugoSolver{config: cfg}
		assert.True(t, solver.canBeat(oppHand, play))
	})

	t.Run("non-spade 3 cannot beat joker", func(t *testing.T) {
		play := []*Card{NewCard(CardDesignJoker, 0, false)}
		oppHand := []*Card{NewCard(CardDesignHeart, 3, false)}
		cfg := DaifugoConfig{SpadeThreeEnabled: true}
		solver := &daifugoSolver{config: cfg}
		assert.False(t, solver.canBeat(oppHand, play))
	})

	t.Run("pair with joker augmentation beats pair", func(t *testing.T) {
		play := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		} // pair of A, strength 14
		oppHand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignJoker, 0, false),
		} // 2 + joker can form pair of 2, strength 15
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.True(t, solver.canBeat(oppHand, play))
	})

	t.Run("pure joker pair beats non-joker pair", func(t *testing.T) {
		play := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 2, false),
		}
		oppHand := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignJoker, 0, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.True(t, solver.canBeat(oppHand, play))
	})

	t.Run("cannot beat pure joker pair", func(t *testing.T) {
		play := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignJoker, 0, false),
		}
		oppHand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 2, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.False(t, solver.canBeat(oppHand, play))
	})
}

func TestDaifugoSolver_wouldBeIllegalFinish(t *testing.T) {
	t.Run("8-cut finish is illegal", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: true, EightCutEnabled: true}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{NewCard(CardDesignSpade, 8, false)}
		assert.True(t, solver.wouldBeIllegalFinish(cards))
	})

	t.Run("joker finish is illegal", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: true}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{NewCard(CardDesignJoker, 0, false)}
		assert.True(t, solver.wouldBeIllegalFinish(cards))
	})

	t.Run("quad finish is illegal (revolution)", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: true}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 5, false),
		}
		assert.True(t, solver.wouldBeIllegalFinish(cards))
	})

	t.Run("normal single finish is legal", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: true}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{NewCard(CardDesignSpade, 2, false)}
		assert.False(t, solver.wouldBeIllegalFinish(cards))
	})

	t.Run("illegal finish disabled allows everything", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: false, EightCutEnabled: true}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{NewCard(CardDesignSpade, 8, false)}
		assert.False(t, solver.wouldBeIllegalFinish(cards))
	})

	t.Run("8 finish legal when 8-cut disabled", func(t *testing.T) {
		cfg := DaifugoConfig{IllegalFinishEnabled: true, EightCutEnabled: false}
		solver := &daifugoSolver{config: cfg}
		cards := []*Card{NewCard(CardDesignSpade, 8, false)}
		assert.False(t, solver.wouldBeIllegalFinish(cards))
	})
}

func TestDaifugoSolver_generateOpeningMoves(t *testing.T) {
	t.Run("single card generates one move", func(t *testing.T) {
		hand := []*Card{NewCard(CardDesignSpade, 2, false)}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateOpeningMoves(hand)
		assert.Equal(t, 1, len(moves))
		assert.Equal(t, 1, len(moves[0].cards))
	})

	t.Run("pair generates single and pair moves", func(t *testing.T) {
		hand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignHeart, 2, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateOpeningMoves(hand)
		// single(2♠) + pair(2♠,2♥)
		assert.Equal(t, 2, len(moves))
	})

	t.Run("joker generates additional augmented and pure moves", func(t *testing.T) {
		hand := []*Card{
			NewCard(CardDesignSpade, 2, false),
			NewCard(CardDesignJoker, 0, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateOpeningMoves(hand)
		// single(2♠) + augmented pair(2♠+Joker) + pure joker single
		assert.Equal(t, 3, len(moves))
	})

	t.Run("8-cut marked correctly", func(t *testing.T) {
		hand := []*Card{
			NewCard(CardDesignSpade, 8, false),
			NewCard(CardDesignSpade, 2, false),
		}
		cfg := DaifugoConfig{EightCutEnabled: true}
		solver := &daifugoSolver{config: cfg}
		moves := solver.generateOpeningMoves(hand)
		eightCutCount := 0
		for _, m := range moves {
			if m.is8Cut {
				eightCutCount++
			}
		}
		assert.Greater(t, eightCutCount, 0)
	})
}

func TestDaifugoSolver_generateResponseMoves(t *testing.T) {
	t.Run("generates valid response beating table", func(t *testing.T) {
		tableCards := []*Card{NewCard(CardDesignSpade, 10, false)}
		hand := []*Card{
			NewCard(CardDesignSpade, 5, false),  // weaker
			NewCard(CardDesignSpade, 13, false), // stronger
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateResponseMoves(hand, tableCards, false, false, 0, false)
		// Only K (strength 13 > 10) should be a valid response
		assert.Equal(t, 1, len(moves))
		assert.Equal(t, 13, moves[0].cards[0].GetValue())
	})

	t.Run("no response when all cards weaker", func(t *testing.T) {
		tableCards := []*Card{NewCard(CardDesignSpade, 2, false)} // strength 15
		hand := []*Card{
			NewCard(CardDesignSpade, 1, false), // strength 14
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateResponseMoves(hand, tableCards, false, false, 0, false)
		assert.Equal(t, 0, len(moves))
	})

	t.Run("joker single response to non-joker table", func(t *testing.T) {
		tableCards := []*Card{NewCard(CardDesignSpade, 2, false)} // strength 15
		hand := []*Card{NewCard(CardDesignJoker, 0, false)}       // strength 16
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateResponseMoves(hand, tableCards, false, false, 0, false)
		assert.Equal(t, 1, len(moves))
		assert.True(t, IsJoker(moves[0].cards[0]))
	})

	t.Run("pure joker group response to pair on table", func(t *testing.T) {
		// Table has pair of A (strength 14)
		// CPU has only 2 jokers → pure joker pair (strength 16 > 14) responds
		tableCards := []*Card{
			NewCard(CardDesignSpade, 1, false),
			NewCard(CardDesignHeart, 1, false),
		}
		hand := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignJoker, 0, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		moves := solver.generateResponseMoves(hand, tableCards, false, false, 0, false)
		assert.Equal(t, 1, len(moves))
		assert.True(t, IsJoker(moves[0].cards[0]))
		assert.True(t, IsJoker(moves[0].cards[1]))
	})
}

func TestDaifugoSolver_selectCardsForSuitLock(t *testing.T) {
	t.Run("no suit lock returns first N cards", func(t *testing.T) {
		group := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		result := solver.selectCardsForSuitLock(group, 1, false, 0)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
	})

	t.Run("full lock reorders to match suit first", func(t *testing.T) {
		group := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{SuitLockMode: DaifugoSuitLockFull}}
		result := solver.selectCardsForSuitLock(group, 2, true, CardDesignHeart)
		assert.NotNil(t, result)
		assert.Equal(t, CardDesignHeart, result[0].GetDesign())
	})

	t.Run("full lock returns nil when no match", func(t *testing.T) {
		group := []*Card{
			NewCard(CardDesignSpade, 5, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{SuitLockMode: DaifugoSuitLockFull}}
		result := solver.selectCardsForSuitLock(group, 1, true, CardDesignHeart)
		assert.Nil(t, result)
	})

	t.Run("partial lock requires at least one match", func(t *testing.T) {
		group := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{SuitLockMode: DaifugoSuitLockPartial}}
		result := solver.selectCardsForSuitLock(group, 2, true, CardDesignHeart)
		assert.NotNil(t, result)
	})

	t.Run("partial lock returns nil when no match", func(t *testing.T) {
		group := []*Card{
			NewCard(CardDesignSpade, 5, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{SuitLockMode: DaifugoSuitLockPartial}}
		result := solver.selectCardsForSuitLock(group, 1, true, CardDesignHeart)
		assert.Nil(t, result)
	})
}

func TestDaifugoSolver_helpers(t *testing.T) {
	t.Run("groupByValue groups correctly", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignSpade, 10, false),
			NewCard(CardDesignJoker, 0, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		groups := solver.groupByValue(cards)
		assert.Equal(t, 2, len(groups))
		assert.Equal(t, 2, len(groups[5]))
		assert.Equal(t, 1, len(groups[10]))
	})

	t.Run("getJokers returns jokers only", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignJoker, 0, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		jokers := solver.getJokers(cards)
		assert.Equal(t, 2, len(jokers))
	})

	t.Run("removeCards removes by pointer identity", func(t *testing.T) {
		c1 := NewCard(CardDesignSpade, 5, false)
		c2 := NewCard(CardDesignHeart, 5, false)
		c3 := NewCard(CardDesignSpade, 10, false)
		hand := []*Card{c1, c2, c3}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		remaining := solver.removeCards(hand, []*Card{c2})
		assert.Equal(t, 2, len(remaining))
		assert.Same(t, c1, remaining[0])
		assert.Same(t, c3, remaining[1])
	})

	t.Run("sortedValuesByStrength returns descending", func(t *testing.T) {
		groups := map[int][]*Card{
			3: {NewCard(CardDesignSpade, 3, false)},
			1: {NewCard(CardDesignSpade, 1, false)},
			2: {NewCard(CardDesignSpade, 2, false)},
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		values := solver.sortedValuesByStrength(groups)
		// strength: 2→15, 1→14, 3→3
		assert.Equal(t, 2, values[0])
		assert.Equal(t, 1, values[1])
		assert.Equal(t, 3, values[2])
	})

	t.Run("containsNonJokerValue", func(t *testing.T) {
		cards := []*Card{
			NewCard(CardDesignJoker, 0, false),
			NewCard(CardDesignSpade, 8, false),
		}
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.True(t, solver.containsNonJokerValue(cards, 8))
		assert.False(t, solver.containsNonJokerValue(cards, 5))
	})

	t.Run("cardStrength normal", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.Equal(t, 15, solver.cardStrength(2))
		assert.Equal(t, 14, solver.cardStrength(1))
		assert.Equal(t, 13, solver.cardStrength(13))
		assert.Equal(t, 3, solver.cardStrength(3))
	})

	t.Run("cardStrength revolution", func(t *testing.T) {
		solver := &daifugoSolver{revolution: true, config: DaifugoConfig{}}
		assert.Equal(t, 3, solver.cardStrength(2))
		assert.Equal(t, 15, solver.cardStrength(3))
	})

	t.Run("cardStrength eleven back", func(t *testing.T) {
		solver := &daifugoSolver{elevenBack: true, config: DaifugoConfig{}}
		assert.Equal(t, 3, solver.cardStrength(2))
		assert.Equal(t, 15, solver.cardStrength(3))
	})

	t.Run("cardStrength revolution and eleven back cancel", func(t *testing.T) {
		solver := &daifugoSolver{revolution: true, elevenBack: true, config: DaifugoConfig{}}
		assert.Equal(t, 15, solver.cardStrength(2))
		assert.Equal(t, 3, solver.cardStrength(3))
	})

	t.Run("playStrength", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		assert.Equal(t, 15, solver.playStrength([]*Card{NewCard(CardDesignSpade, 2, false)}))
		assert.Equal(t, DaifugoJokerStrength, solver.playStrength([]*Card{NewCard(CardDesignJoker, 0, false)}))
	})

	t.Run("applyRevolution toggles for 4+ cards", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
			NewCard(CardDesignClover, 5, false),
		}
		solver.applyRevolution(cards)
		assert.True(t, solver.revolution)
		solver.applyRevolution(cards)
		assert.False(t, solver.revolution)
	})

	t.Run("trySolveWithClearTable succeeds", func(t *testing.T) {
		remaining := []*Card{NewCard(CardDesignSpade, 3, false)} // last play
		moveCards := []*Card{NewCard(CardDesignSpade, 2, false)} // single card, no revolution
		solver := &daifugoSolver{oppHands: nil, config: DaifugoConfig{}}
		result := solver.trySolveWithClearTable(remaining, moveCards, false, false)
		assert.True(t, result)
	})

	t.Run("trySolveWithClearTable fails and restores state", func(t *testing.T) {
		// CPU remaining: 5♠, 3♠. Opponent has 2♠ (strongest normal).
		// After quad revolution: 5→str 13, 3→str 15, 2→str 3
		// But 3♠ as intermediate (str 15) unbeatable → 5♠ last play → would win.
		// So use non-quad move (no revolution): 5♠ beatable (str 5 < 15), 3♠ beatable.
		remaining := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignSpade, 3, false),
		}
		moveCards := []*Card{NewCard(CardDesignSpade, 10, false)} // single, no revolution
		oppHands := [][]*Card{{NewCard(CardDesignSpade, 2, false)}}
		solver := &daifugoSolver{oppHands: oppHands, revolution: false, elevenBack: true, config: DaifugoConfig{}}
		result := solver.trySolveWithClearTable(remaining, moveCards, false, true)
		assert.False(t, result)
		// State should be restored
		assert.False(t, solver.revolution)
		assert.True(t, solver.elevenBack)
	})

	t.Run("applyRevolution no-op for fewer than 4 cards", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false),
			NewCard(CardDesignHeart, 5, false),
			NewCard(CardDesignDiamond, 5, false),
		}
		solver.applyRevolution(cards)
		assert.False(t, solver.revolution)
	})

	t.Run("sequenceMinStrength returns minimum card strength", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false), // strength 5
			NewCard(CardDesignSpade, 6, false), // strength 6
			NewCard(CardDesignSpade, 7, false), // strength 7
		}
		assert.Equal(t, 5, solver.sequenceMinStrength(cards))
	})

	t.Run("sequenceMinStrength with joker", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		cards := []*Card{
			NewCard(CardDesignSpade, 5, false), // strength 5
			NewCard(CardDesignJoker, 0, false), // strength 16
			NewCard(CardDesignSpade, 7, false), // strength 7
		}
		assert.Equal(t, 5, solver.sequenceMinStrength(cards))
	})

	t.Run("generateSequencePlays finds valid sequences", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{SequenceEnabled: true}}
		hand := []*Card{
			NewCard(CardDesignSpade, 3, false), // strength 3
			NewCard(CardDesignSpade, 4, false), // strength 4
			NewCard(CardDesignSpade, 5, false), // strength 5
		}
		moves := solver.generateSequencePlays(hand)
		assert.Greater(t, len(moves), 0)
		assert.True(t, moves[0].isSequence)
		assert.Equal(t, 3, len(moves[0].cards))
	})

	t.Run("generateSequencePlays with mixed suits", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{SequenceEnabled: true}}
		hand := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignHeart, 4, false), // different suit
			NewCard(CardDesignSpade, 5, false),
		}
		// No valid 3-card sequence of same suit
		moves := solver.generateSequencePlays(hand)
		assert.Equal(t, 0, len(moves))
	})

	t.Run("generateSequenceResponsePlays beats table", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{SequenceEnabled: true}}
		hand := []*Card{
			NewCard(CardDesignSpade, 6, false), // strength 6
			NewCard(CardDesignSpade, 7, false), // strength 7
			NewCard(CardDesignSpade, 8, false), // strength 8
		}
		moves := solver.generateSequenceResponsePlays(hand, 3, 3) // table min strength 3
		assert.Greater(t, len(moves), 0)
		assert.True(t, moves[0].isSequence)
	})

	t.Run("generateSequenceResponsePlays rejects weaker sequence", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{SequenceEnabled: true}}
		hand := []*Card{
			NewCard(CardDesignSpade, 3, false), // strength 3
			NewCard(CardDesignSpade, 4, false), // strength 4
			NewCard(CardDesignSpade, 5, false), // strength 5
		}
		moves := solver.generateSequenceResponsePlays(hand, 3, 6) // table min strength 6
		assert.Equal(t, 0, len(moves))
	})

	t.Run("canBeatSequence with stronger sequence", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		oppHand := []*Card{
			NewCard(CardDesignSpade, 8, false),  // strength 8
			NewCard(CardDesignSpade, 9, false),  // strength 9
			NewCard(CardDesignSpade, 10, false), // strength 10
		}
		// Can form 8-9-10 sequence (min strength 8 > 5)
		assert.True(t, solver.canBeatSequence(oppHand, 3, 5))
	})

	t.Run("canBeatSequence fails with weaker cards", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		oppHand := []*Card{
			NewCard(CardDesignSpade, 3, false), // strength 3
			NewCard(CardDesignSpade, 4, false), // strength 4
			NewCard(CardDesignSpade, 5, false), // strength 5
		}
		// Cannot beat sequence with min strength 8
		assert.False(t, solver.canBeatSequence(oppHand, 3, 8))
	})

	t.Run("canBeatSequence with joker fills gap", func(t *testing.T) {
		solver := &daifugoSolver{config: DaifugoConfig{}}
		oppHand := []*Card{
			NewCard(CardDesignSpade, 8, false),  // strength 8
			NewCard(CardDesignJoker, 0, false),  // fills 9
			NewCard(CardDesignSpade, 10, false), // strength 10
		}
		// Can form 8-Joker-10 sequence (min strength 8 > 5)
		assert.True(t, solver.canBeatSequence(oppHand, 3, 5))
	})

	t.Run("isUnbeatableSequence no opponents", func(t *testing.T) {
		solver := &daifugoSolver{oppHands: nil, config: DaifugoConfig{}}
		play := []*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		}
		assert.True(t, solver.isUnbeatableSequence(play))
	})

	t.Run("isUnbeatablePlay dispatches to sequence check", func(t *testing.T) {
		solver := &daifugoSolver{oppHands: nil, config: DaifugoConfig{}}
		move := solverPlay{
			cards:      []*Card{NewCard(CardDesignSpade, 3, false)},
			isSequence: true,
		}
		assert.True(t, solver.isUnbeatablePlay(move))
		move.isSequence = false
		assert.True(t, solver.isUnbeatablePlay(move))
	})
}

func TestDaifugo_trySolveEndgame(t *testing.T) {
	t.Run("returns nil when hand exceeds max", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{CpuDifficulty: DaifugoDifficultyHard}
		d := NewDaifugo(tc, players, cfg)
		// Give CPU 9 cards (exceeds max of 8)
		for i := 1; i <= 9; i++ {
			v := i
			if v > 13 {
				v = 13
			}
			players[1].AddCard(NewCard(CardDesignSpade, v, false))
		}
		result := d.trySolveEndgame(players[1])
		assert.Nil(t, result)
	})

	t.Run("returns nil when hand empty", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{CpuDifficulty: DaifugoDifficultyHard}
		d := NewDaifugo(tc, players, cfg)
		result := d.trySolveEndgame(players[1])
		assert.Nil(t, result)
	})

	t.Run("returns indices for guaranteed win", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{CpuDifficulty: DaifugoDifficultyHard}
		d := NewDaifugo(tc, players, cfg)
		// CPU (player 1): 2♠ (unbeatable)
		players[1].AddCard(NewCard(CardDesignSpade, 2, false))
		// Opponents: weaker cards
		players[0].AddCard(NewCard(CardDesignSpade, 5, false))
		players[2].AddCard(NewCard(CardDesignHeart, 5, false))
		players[3].AddCard(NewCard(CardDesignDiamond, 5, false))
		result := d.trySolveEndgame(players[1])
		assert.NotNil(t, result)
		assert.Equal(t, []int{0}, result)
	})

	t.Run("skips finished opponents", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{CpuDifficulty: DaifugoDifficultyHard}
		d := NewDaifugo(tc, players, cfg)
		// CPU (player 1): 5♠ (weak)
		players[1].AddCard(NewCard(CardDesignSpade, 5, false))
		// Player 2 has 2♠ but is finished → should be ignored
		players[2].AddCard(NewCard(CardDesignSpade, 2, false))
		players[2].SetIsFinished(true)
		// Player 0 and 3 have weaker cards
		players[0].AddCard(NewCard(CardDesignSpade, 3, false))
		players[3].AddCard(NewCard(CardDesignDiamond, 3, false))
		result := d.trySolveEndgame(players[1])
		assert.NotNil(t, result)
	})

	t.Run("solves when table is sequence with SequenceEnabled", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{
			CpuDifficulty:   DaifugoDifficultyHard,
			SequenceEnabled: true,
		}
		d := NewDaifugo(tc, players, cfg)
		// CPU (player 1): 8♠, 9♠, 10♠ — can beat the table sequence 3♠-4♠-5♠
		players[1].AddCard(NewCard(CardDesignSpade, 8, false))
		players[1].AddCard(NewCard(CardDesignSpade, 9, false))
		players[1].AddCard(NewCard(CardDesignSpade, 10, false))
		// Opponents: weaker cards
		players[0].AddCard(NewCard(CardDesignHeart, 3, false))
		players[2].AddCard(NewCard(CardDesignDiamond, 3, false))
		players[3].AddCard(NewCard(CardDesignClover, 3, false))
		// Set table as sequence 3♠-4♠-5♠
		d.SetTableCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		})
		d.SetTableIsSequence(true)
		result := d.trySolveEndgame(players[1])
		assert.NotNil(t, result)
		assert.Equal(t, 3, len(result))
	})

	t.Run("returns nil when table is sequence but SequenceEnabled is false", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{
			CpuDifficulty:   DaifugoDifficultyHard,
			SequenceEnabled: false,
		}
		d := NewDaifugo(tc, players, cfg)
		players[1].AddCard(NewCard(CardDesignSpade, 8, false))
		players[1].AddCard(NewCard(CardDesignSpade, 9, false))
		players[1].AddCard(NewCard(CardDesignSpade, 10, false))
		players[0].AddCard(NewCard(CardDesignHeart, 3, false))
		players[2].AddCard(NewCard(CardDesignDiamond, 3, false))
		players[3].AddCard(NewCard(CardDesignClover, 3, false))
		d.SetTableCards([]*Card{
			NewCard(CardDesignSpade, 3, false),
			NewCard(CardDesignSpade, 4, false),
			NewCard(CardDesignSpade, 5, false),
		})
		d.SetTableIsSequence(true)
		result := d.trySolveEndgame(players[1])
		assert.Nil(t, result)
	})

	t.Run("maps cards back to correct indices", func(t *testing.T) {
		tc := NewTrumpCards(0)
		players := []*DaifugoPlayer{
			NewDaifugoPlayer(true),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
			NewDaifugoPlayer(false),
		}
		cfg := DaifugoConfig{CpuDifficulty: DaifugoDifficultyHard}
		d := NewDaifugo(tc, players, cfg)
		// CPU: 5♠ (index 0), 2♠ (index 1)
		players[1].AddCard(NewCard(CardDesignSpade, 5, false))
		players[1].AddCard(NewCard(CardDesignSpade, 2, false))
		// Opponent: A♠
		players[0].AddCard(NewCard(CardDesignSpade, 1, false))
		players[2].AddCard(NewCard(CardDesignHeart, 3, false))
		players[3].AddCard(NewCard(CardDesignDiamond, 3, false))
		result := d.trySolveEndgame(players[1])
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result))
		// Should play 2♠ first (index 1, strongest card, unbeatable)
		assert.Equal(t, 1, result[0])
	})
}
