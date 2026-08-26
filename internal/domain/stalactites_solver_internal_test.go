//go:build test

package domain

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStalactitesSolver_NearComplete(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Set up near-complete: all foundation at 12, one card left per suit on tableau
	var tableau [StalactitesTableauCnt][]*Card
	tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, false)}
	tableau[1] = []*Card{NewCard(CardDesignClover, CardValueMax, false)}
	tableau[2] = []*Card{NewCard(CardDesignHeart, CardValueMax, false)}
	tableau[3] = []*Card{NewCard(CardDesignDiamond, CardValueMax, false)}
	f.SetTableau(tableau)

	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		for v := 1; v < CardValueMax; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)

	var cells [StalactitesCellCnt]*Card
	f.SetCells(cells)

	solver := newStalactitesSolver(f)
	assert.True(t, solver.isSolvable())
}

func TestStalactitesSolver_Blocked(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Set up a blocked board: all free cells occupied, no valid moves
	var tableau [StalactitesTableauCnt][]*Card
	// Column 0: two cards in wrong order (same color, blocking)
	tableau[0] = []*Card{
		NewCard(CardDesignSpade, 3, false),
		NewCard(CardDesignSpade, 5, false),
	}
	// Column 1: two cards in wrong order
	tableau[1] = []*Card{
		NewCard(CardDesignClover, 3, false),
		NewCard(CardDesignClover, 5, false),
	}
	// Column 2: two cards in wrong order
	tableau[2] = []*Card{
		NewCard(CardDesignHeart, 3, false),
		NewCard(CardDesignHeart, 5, false),
	}
	// Column 3: two cards in wrong order
	tableau[3] = []*Card{
		NewCard(CardDesignDiamond, 3, false),
		NewCard(CardDesignDiamond, 5, false),
	}
	// Columns 4-7: cards blocking each other
	tableau[4] = []*Card{
		NewCard(CardDesignSpade, 7, false),
		NewCard(CardDesignSpade, 9, false),
	}
	tableau[5] = []*Card{
		NewCard(CardDesignClover, 7, false),
		NewCard(CardDesignClover, 9, false),
	}
	tableau[6] = []*Card{
		NewCard(CardDesignHeart, 7, false),
		NewCard(CardDesignHeart, 9, false),
	}
	tableau[7] = []*Card{
		NewCard(CardDesignDiamond, 7, false),
		NewCard(CardDesignDiamond, 9, false),
	}
	f.SetTableau(tableau)

	// All free cells occupied
	var cells [StalactitesCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, 11, false)
	cells[1] = NewCard(CardDesignClover, 11, false)
	cells[2] = NewCard(CardDesignHeart, 11, false)
	cells[3] = NewCard(CardDesignDiamond, 11, false)
	f.SetCells(cells)

	// Foundation has some cards
	var foundation [StalactitesFoundationCnt][]*Card
	foundation[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
	foundation[1] = []*Card{NewCard(CardDesignClover, 1, false)}
	foundation[2] = []*Card{NewCard(CardDesignHeart, 1, false)}
	foundation[3] = []*Card{NewCard(CardDesignDiamond, 1, false)}
	f.SetFoundation(foundation)

	solver := newStalactitesSolver(f)
	assert.False(t, solver.isSolvable())
}

func TestStalactitesSolver_AlreadySolved(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// All cards on foundation
	var tableau [StalactitesTableauCnt][]*Card
	f.SetTableau(tableau)
	var cells [StalactitesCellCnt]*Card
	f.SetCells(cells)

	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		for v := 1; v <= CardValueMax; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)

	solver := newStalactitesSolver(f)
	assert.True(t, solver.isSolvable())
}

func TestStalactitesSolver_IterationLimitReturnsTrue(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset() // Full random board — many states to explore

	solver := newStalactitesSolver(f)
	// Set a very low iteration limit to guarantee the limit is hit
	solver.maxIterations = 1

	// When the iteration limit is exceeded, solver returns true (unknown = not stalemate)
	assert.True(t, solver.isSolvable())
	// Verify we actually hit the limit (iterations > 1)
	assert.Greater(t, solver.iterations, 1)
}

func TestStalactitesSolver_StalactitesToFoundation(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Ace in free cell, empty tableau, empty foundation -> should be solvable step
	var tableau [StalactitesTableauCnt][]*Card
	// Put remaining cards in valid sequences so game is solvable
	// Simplify: near complete game with one card in free cell
	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		for v := 1; v <= CardValueMax; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	// Remove the last spade card from foundation, put it in free cell
	foundation[0] = foundation[0][:12]
	f.SetFoundation(foundation)
	f.SetTableau(tableau)

	var cells [StalactitesCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, CardValueMax, false)
	f.SetCells(cells)

	solver := newStalactitesSolver(f)
	assert.True(t, solver.isSolvable())
}

func TestStalactitesSolver_StalactitesKingToFoundation(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Card in free cell, move to tableau then to foundation
	// Heart Queen (12) in stalactites, Spade King (13) on tableau
	// Foundation: Spade 1-12, Heart 1-11, Clover 1-13, Diamond 1-13
	// Solution: move Heart Queen to foundation (needs 12, has 11), then Heart King on tableau... wait.
	// Simpler: just have a card in stalactites that can go directly to foundation
	var tableau [StalactitesTableauCnt][]*Card
	f.SetTableau(tableau)

	var cells [StalactitesCellCnt]*Card
	cells[0] = NewCard(CardDesignHeart, CardValueMax, false) // Heart King in free cell
	f.SetCells(cells)

	// Foundation: Heart has 12 cards, rest complete
	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		maxV := CardValueMax
		if i+1 == CardDesignHeart {
			maxV = 12 // Heart missing King
		}
		for v := 1; v <= maxV; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)

	solver := newStalactitesSolver(f)
	assert.True(t, solver.isSolvable())
}

func TestStalactitesSolver_StateKeyDifferentStates(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	var tableau1 [StalactitesTableauCnt][]*Card
	tableau1[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
	f.SetTableau(tableau1)
	var cells [StalactitesCellCnt]*Card
	f.SetCells(cells)
	var foundation [StalactitesFoundationCnt][]*Card
	f.SetFoundation(foundation)

	solver1 := newStalactitesSolver(f)
	key1 := solver1.stateKey()

	// Different state: card in different column
	var tableau2 [StalactitesTableauCnt][]*Card
	tableau2[1] = []*Card{NewCard(CardDesignSpade, 1, false)}
	f.SetTableau(tableau2)

	solver2 := newStalactitesSolver(f)
	key2 := solver2.stateKey()

	assert.NotEqual(t, key1, key2)
}

func TestStalactitesSolver_GameClearSkipsStalemate(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Set up: all cards on foundation except one King on tableau
	var tableau [StalactitesTableauCnt][]*Card
	tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, false)}
	f.SetTableau(tableau)
	var cells [StalactitesCellCnt]*Card
	f.SetCells(cells)

	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		maxV := CardValueMax
		if i+1 == CardDesignSpade {
			maxV = CardValueMax - 1 // Spade missing King
		}
		for v := 1; v <= maxV; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)
	f.SetPhase(StalactitesPhasePlaying)

	// Move the last card to foundation — triggers checkGameClear then checkStalemate
	err := f.MoveTableauToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
	assert.False(t, f.IsStalemate()) // checkStalemate should early-return on non-playing phase
}

func TestStalactitesSolver_GameClearViaStalactitesSkipsStalemate(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// All cards on foundation except one King in free cell
	var tableau [StalactitesTableauCnt][]*Card
	f.SetTableau(tableau)
	var cells [StalactitesCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, CardValueMax, false)
	f.SetCells(cells)

	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		maxV := CardValueMax
		if i+1 == CardDesignSpade {
			maxV = CardValueMax - 1
		}
		for v := 1; v <= maxV; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)
	f.SetPhase(StalactitesPhasePlaying)

	err := f.MoveStalactitesToFoundation(0)
	assert.NoError(t, err)
	assert.Equal(t, StalactitesPhaseGameClear, f.GetPhase())
	assert.False(t, f.IsStalemate())
}

func TestStalactitesSolver_MemoizationPreventsRevisit(t *testing.T) {
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	// Simple blocked state: one card, no valid moves, all free cells full
	var tableau [StalactitesTableauCnt][]*Card
	tableau[0] = []*Card{NewCard(CardDesignSpade, 5, false)}
	f.SetTableau(tableau)

	var cells [StalactitesCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, 8, false)
	cells[1] = NewCard(CardDesignClover, 8, false)
	cells[2] = NewCard(CardDesignHeart, 8, false)
	cells[3] = NewCard(CardDesignDiamond, 8, false)
	f.SetCells(cells)

	var foundation [StalactitesFoundationCnt][]*Card
	f.SetFoundation(foundation)

	solver := newStalactitesSolver(f)
	result := solver.isSolvable()
	assert.False(t, result)
	// Verify visited map was populated
	assert.Greater(t, len(solver.visited), 0)
}

func TestStalactitesSolver_AStarHeuristicAdmissible(t *testing.T) {
	// Verify stalactitesHeuristic returns correct value: 52 - sum(foundation counts)
	st := &stalactitesState{}
	st.foundation[0] = 5
	st.foundation[1] = 3
	st.foundation[2] = 10
	st.foundation[3] = 2
	h := stalactitesHeuristic(st)
	assert.Equal(t, 52-(5+3+10+2), h)
}

func TestStalactitesSolver_AStarSolvableBoard(t *testing.T) {
	// Set up a moderately complex solvable board:
	// Foundation has A-10 for all suits, remaining cards (J, Q, K) arranged
	// in solvable order on tableau (alternating color, descending)
	f := NewStalactites(NewTrumpCards(0))
	f.Reset()
	// These fixtures seed foundation piles as A..K, so pin the base rank to match.
	// Reset picks it from the deal, and the solver counts ranks FROM the base --
	// leaving it shuffled made these tests agree with the solver's old FreeCell
	// model instead of exercising the real rule.
	f.baseRank = 1

	var foundation [StalactitesFoundationCnt][]*Card
	for i := 0; i < StalactitesFoundationCnt; i++ {
		foundation[i] = make([]*Card, 0)
		for v := 1; v <= 10; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	f.SetFoundation(foundation)

	// Place J, Q, K for each suit on tableau in solvable arrangements
	// Column 0: Spade K (13), Heart Q (12), Clover J (11) — alternating color, descending
	// Column 1: Heart K (13), Spade Q (12), Diamond J (11) — alternating color, descending
	// Column 2: Clover K (13), Diamond Q (12), Spade J (11) — alternating color, descending
	// Column 3: Diamond K (13), Clover Q (12), Heart J (11) — alternating color, descending
	var tableau [StalactitesTableauCnt][]*Card
	tableau[0] = []*Card{
		NewCard(CardDesignSpade, 13, false),
		NewCard(CardDesignHeart, 12, false),
		NewCard(CardDesignClover, 11, false),
	}
	tableau[1] = []*Card{
		NewCard(CardDesignHeart, 13, false),
		NewCard(CardDesignSpade, 12, false),
		NewCard(CardDesignDiamond, 11, false),
	}
	tableau[2] = []*Card{
		NewCard(CardDesignClover, 13, false),
		NewCard(CardDesignDiamond, 12, false),
		NewCard(CardDesignSpade, 11, false),
	}
	tableau[3] = []*Card{
		NewCard(CardDesignDiamond, 13, false),
		NewCard(CardDesignClover, 12, false),
		NewCard(CardDesignHeart, 11, false),
	}
	f.SetTableau(tableau)

	var cells [StalactitesCellCnt]*Card
	f.SetCells(cells)

	solver := newStalactitesSolver(f)
	assert.True(t, solver.isSolvable())
	// A* with stalactitesHeuristic should solve this efficiently (far fewer iterations than 100k limit)
	assert.Less(t, solver.iterations, 1000)
}

func TestStalactitesSolver_AStarPriorityQueueOrder(t *testing.T) {
	// Verify the priority queue orders by f = g + h (lowest first)
	pq := &stalactitesPQ{}
	heap.Init(pq)

	s1 := &stalactitesState{g: 5, h: 10} // f=15
	s2 := &stalactitesState{g: 2, h: 3}  // f=5
	s3 := &stalactitesState{g: 8, h: 1}  // f=9

	heap.Push(pq, s1)
	heap.Push(pq, s2)
	heap.Push(pq, s3)

	// Should dequeue in order: s2 (f=5), s3 (f=9), s1 (f=15)
	first := heap.Pop(pq).(*stalactitesState)
	assert.Equal(t, 5, first.g+first.h)

	second := heap.Pop(pq).(*stalactitesState)
	assert.Equal(t, 9, second.g+second.h)

	third := heap.Pop(pq).(*stalactitesState)
	assert.Equal(t, 15, third.g+third.h)
}

func TestCanPlaceOnTableau_SolverAllowsNonKingOnEmptyStal(t *testing.T) {
	var tableau [StalactitesTableauCnt][]*Card
	// Zero-value solver models a standard Stalactites board (sameSuit == false).
	solver := &stalactitesSolver{}

	t.Run("non-King on empty column", func(t *testing.T) {
		// Stalactites rule: any card can go on an empty column
		heart5 := NewCard(CardDesignHeart, 5, false)
		assert.True(t, solver.canPlaceOnTableau(heart5, 0, tableau))
	})

	t.Run("King on empty column", func(t *testing.T) {
		king := NewCard(CardDesignSpade, CardValueMax, false)
		assert.True(t, solver.canPlaceOnTableau(king, 0, tableau))
	})

	t.Run("valid alternating-color descending on non-empty", func(t *testing.T) {
		tableau[1] = []*Card{NewCard(CardDesignSpade, 6, false)}
		heart5 := NewCard(CardDesignHeart, 5, false)
		assert.True(t, solver.canPlaceOnTableau(heart5, 1, tableau))
	})

	t.Run("invalid same-color on non-empty", func(t *testing.T) {
		tableau[2] = []*Card{NewCard(CardDesignSpade, 6, false)}
		clover5 := NewCard(CardDesignClover, 5, false)
		assert.False(t, solver.canPlaceOnTableau(clover5, 2, tableau))
	})

	// The Baker's Game case that lived here was dropped with the sameSuit flag:
	// FreeCell.go carries that variant, and Stalactites always builds by
	// alternating colour. It is still covered by freecell_solver_internal_test.go.
}
