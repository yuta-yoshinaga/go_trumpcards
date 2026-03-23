package domain

import "container/heap"

// FreeCellSolverMaxIterations is the maximum number of states the solver explores.
// Keeps runtime under ~50ms and memory under ~10MB.
const FreeCellSolverMaxIterations = 100000

// freeCellState represents a board state in the A* search.
type freeCellState struct {
	tableau    [FreeCellTableauCnt][]*Card
	freeCells  [FreeCellCellCnt]*Card
	foundation [FreeCellFoundationCnt]int // count per suit
	g          int                        // moves made so far
	h          int                        // heuristic estimate of remaining moves
	index      int                        // index in heap (maintained by container/heap)
}

// freeCellPQ implements heap.Interface for A* priority queue.
type freeCellPQ []*freeCellState

func (pq freeCellPQ) Len() int { return len(pq) }

func (pq freeCellPQ) Less(i, j int) bool {
	return (pq[i].g + pq[i].h) < (pq[j].g + pq[j].h)
}

func (pq freeCellPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *freeCellPQ) Push(x any) {
	n := len(*pq)
	item := x.(*freeCellState)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *freeCellPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// freeCellSolver performs A* search with memoization to determine if a FreeCell board is solvable.
type freeCellSolver struct {
	visited       map[[52]uint16]struct{}
	iterations    int
	maxIterations int
	initialState  *freeCellState
}

func newFreeCellSolver(f *FreeCell) *freeCellSolver {
	s := &freeCellSolver{
		visited:       make(map[[52]uint16]struct{}),
		maxIterations: FreeCellSolverMaxIterations,
	}
	state := &freeCellState{}
	// Deep copy tableau
	for i := range FreeCellTableauCnt {
		state.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(state.tableau[i], f.tableau[i])
	}
	// Copy freeCells
	state.freeCells = f.freeCells
	// Foundation: only need count per suit
	for i := range FreeCellFoundationCnt {
		state.foundation[i] = len(f.foundation[i])
	}
	state.g = 0
	state.h = heuristic(state)
	s.initialState = state
	return s
}

// heuristic returns an admissible estimate of remaining moves.
// It counts cards not yet on foundation (each needs at least 1 move).
func heuristic(st *freeCellState) int {
	total := 0
	for i := range FreeCellFoundationCnt {
		total += st.foundation[i]
	}
	return 52 - total
}

// isSolvable returns true if the board can be solved, false if proven unsolvable.
// If the iteration limit is exceeded, returns true (unknown = not stalemate).
func (s *freeCellSolver) isSolvable() bool {
	return s.astar()
}

func (s *freeCellSolver) astar() bool {
	// Check if already solved
	if isSolvedState(s.initialState) {
		return true
	}

	pq := &freeCellPQ{}
	heap.Init(pq)
	heap.Push(pq, s.initialState)

	key := stateKeyFromState(s.initialState)
	s.visited[key] = struct{}{}

	for pq.Len() > 0 {
		s.iterations++
		if s.iterations > s.maxIterations {
			return true // unknown = not stalemate
		}

		current := heap.Pop(pq).(*freeCellState)

		// Generate all successor states
		successors := s.generateSuccessors(current)
		for _, next := range successors {
			if isSolvedState(next) {
				return true
			}
			sk := stateKeyFromState(next)
			if _, ok := s.visited[sk]; ok {
				continue
			}
			s.visited[sk] = struct{}{}
			heap.Push(pq, next)
		}
	}

	return false // exhausted all states, no solution
}

// generateSuccessors generates all valid successor states from the current state.
func (s *freeCellSolver) generateSuccessors(st *freeCellState) []*freeCellState {
	var successors []*freeCellState

	// 1. Tableau -> Foundation
	for col := range FreeCellTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && canPlaceOnFoundation(card, fIdx, st.foundation) {
			next := copyState(st)
			next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = heuristic(next)
			successors = append(successors, next)
		}
	}

	// 2. FreeCell -> Foundation
	for cell := range FreeCellCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && canPlaceOnFoundation(card, fIdx, st.foundation) {
			next := copyState(st)
			next.freeCells[cell] = nil
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = heuristic(next)
			successors = append(successors, next)
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	for fromCol := range FreeCellTableauCnt {
		if len(st.tableau[fromCol]) == 0 {
			continue
		}
		card := st.tableau[fromCol][len(st.tableau[fromCol])-1]
		for toCol := range FreeCellTableauCnt {
			if toCol == fromCol {
				continue
			}
			if canPlaceOnTableau(card, toCol, st.tableau) {
				next := copyState(st)
				next.tableau[fromCol] = next.tableau[fromCol][:len(next.tableau[fromCol])-1]
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = heuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 4. FreeCell -> Tableau
	for cell := range FreeCellCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		for toCol := range FreeCellTableauCnt {
			if canPlaceOnTableau(card, toCol, st.tableau) {
				next := copyState(st)
				next.freeCells[cell] = nil
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = heuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 5. Tableau -> FreeCell
	for col := range FreeCellTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		for cell := range FreeCellCellCnt {
			if st.freeCells[cell] == nil {
				next := copyState(st)
				next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
				next.freeCells[cell] = card
				next.g = st.g + 1
				next.h = heuristic(next)
				successors = append(successors, next)
				break // Only try one empty free cell (equivalent moves)
			}
		}
	}

	return successors
}

// copyState creates a deep copy of a freeCellState.
// NOTE: g and h are left at 0 in the copy; the caller must set them before pushing to the priority queue.
func copyState(st *freeCellState) *freeCellState {
	next := &freeCellState{}
	for i := range FreeCellTableauCnt {
		next.tableau[i] = make([]*Card, len(st.tableau[i]))
		copy(next.tableau[i], st.tableau[i])
	}
	next.freeCells = st.freeCells
	next.foundation = st.foundation
	return next
}

func isSolvedState(st *freeCellState) bool {
	for i := range FreeCellFoundationCnt {
		if st.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

func canPlaceOnTableau(card *Card, col int, tableau [FreeCellTableauCnt][]*Card) bool {
	colCards := tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return isAlternateColor(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

func canPlaceOnFoundation(card *Card, fIdx int, foundation [FreeCellFoundationCnt]int) bool {
	count := foundation[fIdx]
	if count == 0 {
		return card.GetValue() == 1
	}
	return card.GetValue() == count+1
}

func isAlternateColor(card1, card2 *Card) bool {
	return isBlack(card1) != isBlack(card2)
}

func isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// stateKey returns the state key for the solver's initial state (used in tests).
func (s *freeCellSolver) stateKey() [52]uint16 {
	return stateKeyFromState(s.initialState)
}

// stateKeyFromState encodes the board state into a compact key for memoization.
// Each card (identified by (design-1)*13+value-1) maps to a uint16 location.
// Tableau: col*64 + pos + 1 (supports up to 63 cards per column).
// FreeCell: 512 + cell. Foundation: 0 (default).
// Jokers (design=0) are skipped since FreeCell uses only 52 standard cards.
func stateKeyFromState(st *freeCellState) [52]uint16 {
	var key [52]uint16
	for col := range FreeCellTableauCnt {
		for pos, card := range st.tableau[col] {
			if card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = uint16(col*64 + pos + 1)
			}
		}
	}
	for cell := range FreeCellCellCnt {
		if st.freeCells[cell] != nil {
			card := st.freeCells[cell]
			if card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = uint16(512 + cell)
			}
		}
	}
	return key
}
