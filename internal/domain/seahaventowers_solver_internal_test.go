//go:build test

package domain

import (
	"container/heap"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSeahavenTowersSolver_NearComplete(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	var tableau [SeahavenTowersTableauCnt][]*Card
	tableau[0] = []*Card{NewCard(CardDesignSpade, CardValueMax, false)}
	tableau[1] = []*Card{NewCard(CardDesignClover, CardValueMax, false)}
	tableau[2] = []*Card{NewCard(CardDesignHeart, CardValueMax, false)}
	tableau[3] = []*Card{NewCard(CardDesignDiamond, CardValueMax, false)}
	s.SetTableau(tableau)

	var foundation [SeahavenTowersFoundationCnt][]*Card
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		for v := 1; v < CardValueMax; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	s.SetFoundation(foundation)

	var cells [SeahavenTowersCellCnt]*Card
	s.SetFreeCells(cells)

	solver := newSeahavenTowersSolver(s)
	assert.True(t, solver.isSolvable())
}

func TestSeahavenTowersSolver_AlreadySolved(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	var tableau [SeahavenTowersTableauCnt][]*Card
	s.SetTableau(tableau)
	var cells [SeahavenTowersCellCnt]*Card
	s.SetFreeCells(cells)

	var foundation [SeahavenTowersFoundationCnt][]*Card
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		for v := 1; v <= CardValueMax; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	s.SetFoundation(foundation)

	solver := newSeahavenTowersSolver(s)
	assert.True(t, solver.isSolvable())
}

func TestSeahavenTowersSolver_Blocked(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	// Cells blocked with non-King cards that can't form same-suit-descending stacks.
	var cells [SeahavenTowersCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, 11, false)
	cells[1] = NewCard(CardDesignClover, 11, false)
	s.SetFreeCells(cells)

	// Tableau columns each hold a single non-King card; no two cards share a
	// suit with consecutive ranks, so no tableau-to-tableau move is legal.
	// No Aces, so no foundation move possible.
	var tableau [SeahavenTowersTableauCnt][]*Card
	tableau[0] = []*Card{NewCard(CardDesignSpade, 5, false)}
	tableau[1] = []*Card{NewCard(CardDesignClover, 5, false)}
	tableau[2] = []*Card{NewCard(CardDesignHeart, 5, false)}
	tableau[3] = []*Card{NewCard(CardDesignDiamond, 5, false)}
	tableau[4] = []*Card{NewCard(CardDesignSpade, 8, false)}
	tableau[5] = []*Card{NewCard(CardDesignClover, 8, false)}
	tableau[6] = []*Card{NewCard(CardDesignHeart, 8, false)}
	tableau[7] = []*Card{NewCard(CardDesignDiamond, 8, false)}
	tableau[8] = []*Card{NewCard(CardDesignHeart, 2, false)}
	tableau[9] = []*Card{NewCard(CardDesignDiamond, 2, false)}
	s.SetTableau(tableau)

	var foundation [SeahavenTowersFoundationCnt][]*Card
	s.SetFoundation(foundation)

	solver := newSeahavenTowersSolver(s)
	assert.False(t, solver.isSolvable())
	assert.Greater(t, len(solver.visited), 0)
}

func TestSeahavenTowersSolver_IterationLimitReturnsTrue(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	// Deterministic far-from-solved board: each of the first four columns
	// exposes an Ace (a guaranteed foundation move → ≥1 successor at the first
	// node), with the remaining 44 cards buried across the other columns. This
	// guarantees the search reaches a second loop iteration regardless of
	// shuffle, so the maxIterations=1 cap is hit (returns true, iterations > 1).
	// (A shuffled Reset board occasionally had no move or solved within one
	// iteration, which made this flake on CI.)
	var tableau [SeahavenTowersTableauCnt][]*Card
	for suit := 1; suit <= 4; suit++ {
		// King buried, Ace exposed on top of columns 0..3.
		tableau[suit-1] = []*Card{NewCard(suit, CardValueMax, false), NewCard(suit, 1, false)}
	}
	scatterCol := 4
	for suit := 1; suit <= 4; suit++ {
		for val := 2; val < CardValueMax; val++ {
			tableau[scatterCol] = append(tableau[scatterCol], NewCard(suit, val, false))
			scatterCol++
			if scatterCol >= SeahavenTowersTableauCnt {
				scatterCol = 4
			}
		}
	}
	s.SetTableau(tableau)
	s.SetFreeCells([SeahavenTowersCellCnt]*Card{})
	s.SetFoundation([SeahavenTowersFoundationCnt][]*Card{})

	solver := newSeahavenTowersSolver(s)
	solver.maxIterations = 1

	assert.True(t, solver.isSolvable())
	assert.Greater(t, solver.iterations, 1)
}

func TestSeahavenTowersSolver_FreeCellAceToFoundation(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	var tableau [SeahavenTowersTableauCnt][]*Card
	s.SetTableau(tableau)

	var cells [SeahavenTowersCellCnt]*Card
	cells[0] = NewCard(CardDesignSpade, CardValueMax, false) // King in reserved cell
	s.SetFreeCells(cells)

	var foundation [SeahavenTowersFoundationCnt][]*Card
	for i := 0; i < SeahavenTowersFoundationCnt; i++ {
		maxV := CardValueMax
		if i+1 == CardDesignSpade {
			maxV = CardValueMax - 1
		}
		for v := 1; v <= maxV; v++ {
			foundation[i] = append(foundation[i], NewCard(i+1, v, false))
		}
	}
	s.SetFoundation(foundation)

	solver := newSeahavenTowersSolver(s)
	assert.True(t, solver.isSolvable())
}

func TestSeahavenTowersSolver_StateKeyDifferentStates(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	var tableau1 [SeahavenTowersTableauCnt][]*Card
	tableau1[0] = []*Card{NewCard(CardDesignSpade, 1, false)}
	s.SetTableau(tableau1)
	var cells [SeahavenTowersCellCnt]*Card
	s.SetFreeCells(cells)
	var foundation [SeahavenTowersFoundationCnt][]*Card
	s.SetFoundation(foundation)

	solver1 := newSeahavenTowersSolver(s)
	key1 := solver1.stateKey()

	var tableau2 [SeahavenTowersTableauCnt][]*Card
	tableau2[1] = []*Card{NewCard(CardDesignSpade, 1, false)}
	s.SetTableau(tableau2)

	solver2 := newSeahavenTowersSolver(s)
	key2 := solver2.stateKey()

	assert.NotEqual(t, key1, key2)
}

func TestSeahavenTowersSolver_HeuristicAdmissible(t *testing.T) {
	st := &seahavenTowersState{}
	st.foundation[0] = 5
	st.foundation[1] = 3
	st.foundation[2] = 10
	st.foundation[3] = 2
	h := seahavenTowersHeuristic(st)
	assert.Equal(t, 52-(5+3+10+2), h)
}

func TestSeahavenTowersSolver_PriorityQueueOrder(t *testing.T) {
	pq := &seahavenTowersPQ{}
	heap.Init(pq)

	heap.Push(pq, &seahavenTowersState{g: 5, h: 10}) // f=15
	heap.Push(pq, &seahavenTowersState{g: 2, h: 3})  // f=5
	heap.Push(pq, &seahavenTowersState{g: 8, h: 1})  // f=9

	first := heap.Pop(pq).(*seahavenTowersState)
	assert.Equal(t, 5, first.g+first.h)
	second := heap.Pop(pq).(*seahavenTowersState)
	assert.Equal(t, 9, second.g+second.h)
	third := heap.Pop(pq).(*seahavenTowersState)
	assert.Equal(t, 15, third.g+third.h)
}

func TestSeahavenTowersCanPlaceOnTableau_SolverRules(t *testing.T) {
	var tableau [SeahavenTowersTableauCnt][]*Card

	t.Run("King on empty column", func(t *testing.T) {
		king := NewCard(CardDesignSpade, CardValueMax, false)
		assert.True(t, seahavenTowersCanPlaceOnTableau(king, 0, tableau))
	})

	t.Run("non-King on empty column rejected", func(t *testing.T) {
		five := NewCard(CardDesignSpade, 5, false)
		assert.False(t, seahavenTowersCanPlaceOnTableau(five, 0, tableau))
	})

	t.Run("same-suit descending accepted", func(t *testing.T) {
		tableau[1] = []*Card{NewCard(CardDesignSpade, 6, false)}
		five := NewCard(CardDesignSpade, 5, false)
		assert.True(t, seahavenTowersCanPlaceOnTableau(five, 1, tableau))
	})

	t.Run("different suit rejected", func(t *testing.T) {
		tableau[2] = []*Card{NewCard(CardDesignSpade, 6, false)}
		five := NewCard(CardDesignHeart, 5, false)
		assert.False(t, seahavenTowersCanPlaceOnTableau(five, 2, tableau))
	})

	t.Run("non-descending rejected", func(t *testing.T) {
		tableau[3] = []*Card{NewCard(CardDesignSpade, 6, false)}
		eight := NewCard(CardDesignSpade, 8, false)
		assert.False(t, seahavenTowersCanPlaceOnTableau(eight, 3, tableau))
	})
}

// TestSeahavenTowersSolver_NoTableauReservedKeyCollision is a regression guard
// for the encoding bug where col=8/pos=0 hashed to the same uint16 (513) as
// reserved cell 1 (512+1). With 10 tableau columns the tableau key range now
// reaches `9*64 + 51 + 1 = 628`, so the reserved base must live above that.
// Two boards differing only by swapping a card between tableau[8][0] and
// freeCells[1] must produce distinct state keys.
func TestSeahavenTowersSolver_NoTableauReservedKeyCollision(t *testing.T) {
	s := NewSeahavenTowers(NewTrumpCards(0))
	s.Reset()

	cardA := NewCard(CardDesignSpade, 5, false)
	cardB := NewCard(CardDesignHeart, 9, false)

	// Board 1: cardA on tableau[8][0], cardB on reserved cell 1.
	var tableau1 [SeahavenTowersTableauCnt][]*Card
	tableau1[8] = []*Card{cardA}
	s.SetTableau(tableau1)
	var cells1 [SeahavenTowersCellCnt]*Card
	cells1[1] = cardB
	s.SetFreeCells(cells1)
	var foundation [SeahavenTowersFoundationCnt][]*Card
	s.SetFoundation(foundation)

	solver1 := newSeahavenTowersSolver(s)
	key1 := solver1.stateKey()

	// Board 2: swap the two cards' locations.
	var tableau2 [SeahavenTowersTableauCnt][]*Card
	tableau2[8] = []*Card{cardB}
	s.SetTableau(tableau2)
	var cells2 [SeahavenTowersCellCnt]*Card
	cells2[1] = cardA
	s.SetFreeCells(cells2)

	solver2 := newSeahavenTowersSolver(s)
	key2 := solver2.stateKey()

	assert.NotEqual(t, key1, key2, "boards differing by tableau[8][0]/reserved[1] swap must hash differently")
}

func TestSeahavenTowersCanPlaceOnFoundation_SolverRules(t *testing.T) {
	var foundation [SeahavenTowersFoundationCnt]int

	t.Run("Ace on empty foundation", func(t *testing.T) {
		ace := NewCard(CardDesignSpade, 1, false)
		assert.True(t, seahavenTowersCanPlaceOnFoundation(ace, 0, foundation))
	})

	t.Run("non-Ace on empty foundation rejected", func(t *testing.T) {
		two := NewCard(CardDesignSpade, 2, false)
		assert.False(t, seahavenTowersCanPlaceOnFoundation(two, 0, foundation))
	})

	t.Run("next-rank accepted", func(t *testing.T) {
		foundation[0] = 5
		six := NewCard(CardDesignSpade, 6, false)
		assert.True(t, seahavenTowersCanPlaceOnFoundation(six, 0, foundation))
	})

	t.Run("wrong rank rejected", func(t *testing.T) {
		foundation[1] = 3
		eight := NewCard(CardDesignClover, 8, false)
		assert.False(t, seahavenTowersCanPlaceOnFoundation(eight, 1, foundation))
	})
}
