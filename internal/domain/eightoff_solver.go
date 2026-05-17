package domain

import "container/heap"

// EightOffSolverMaxIterations is the maximum number of states the solver explores.
// Keeps runtime under ~50ms and memory under ~10MB.
const EightOffSolverMaxIterations = 100000

// eightOffState represents a board state in the A* search.
type eightOffState struct {
	tableau    [EightOffTableauCnt][]*Card
	freeCells  [EightOffCellCnt]*Card
	foundation [EightOffFoundationCnt]int // count per suit
	g          int                        // moves made so far
	h          int                        // heuristic estimate of remaining moves
	index      int                        // index in heap (maintained by container/heap)
}

// eightOffPQ implements heap.Interface for A* priority queue.
type eightOffPQ []*eightOffState

func (pq eightOffPQ) Len() int { return len(pq) }

func (pq eightOffPQ) Less(i, j int) bool {
	return (pq[i].g + pq[i].h) < (pq[j].g + pq[j].h)
}

func (pq eightOffPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *eightOffPQ) Push(x any) {
	n := len(*pq)
	item := x.(*eightOffState)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *eightOffPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// eightOffSolver performs A* search with memoization to determine if an Eight Off
// board is solvable.
type eightOffSolver struct {
	visited       map[[52]uint16]struct{}
	iterations    int
	maxIterations int
	initialState  *eightOffState
}

func newEightOffSolver(e *EightOff) *eightOffSolver {
	s := &eightOffSolver{
		visited:       make(map[[52]uint16]struct{}),
		maxIterations: EightOffSolverMaxIterations,
	}
	state := &eightOffState{}
	for i := range EightOffTableauCnt {
		state.tableau[i] = make([]*Card, len(e.tableau[i]))
		copy(state.tableau[i], e.tableau[i])
	}
	state.freeCells = e.freeCells
	for i := range EightOffFoundationCnt {
		state.foundation[i] = len(e.foundation[i])
	}
	state.g = 0
	state.h = eightOffHeuristic(state)
	s.initialState = state
	return s
}

// eightOffHeuristic returns an admissible estimate of remaining moves.
func eightOffHeuristic(st *eightOffState) int {
	total := 0
	for i := range EightOffFoundationCnt {
		total += st.foundation[i]
	}
	return 52 - total
}

// isSolvable returns true if the board can be solved, false if proven unsolvable.
// If the iteration limit is exceeded, returns true (unknown = not stalemate).
func (s *eightOffSolver) isSolvable() bool {
	return s.astar()
}

func (s *eightOffSolver) astar() bool {
	if isSolvedEightOffState(s.initialState) {
		return true
	}

	pq := &eightOffPQ{}
	heap.Init(pq)
	heap.Push(pq, s.initialState)

	key := eightOffStateKey(s.initialState)
	s.visited[key] = struct{}{}

	for pq.Len() > 0 {
		s.iterations++
		if s.iterations > s.maxIterations {
			return true
		}

		current := heap.Pop(pq).(*eightOffState)

		successors := s.generateSuccessors(current)
		for _, next := range successors {
			if isSolvedEightOffState(next) {
				return true
			}
			sk := eightOffStateKey(next)
			if _, ok := s.visited[sk]; ok {
				continue
			}
			s.visited[sk] = struct{}{}
			heap.Push(pq, next)
		}
	}

	return false
}

// generateSuccessors generates all valid successor states from the current state.
func (s *eightOffSolver) generateSuccessors(st *eightOffState) []*eightOffState {
	var successors []*eightOffState

	// 1. Tableau -> Foundation
	for col := range EightOffTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < EightOffFoundationCnt && canPlaceOnEightOffFoundation(card, fIdx, st.foundation) {
			next := copyEightOffState(st)
			next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = eightOffHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 2. FreeCell -> Foundation
	for cell := range EightOffCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < EightOffFoundationCnt && canPlaceOnEightOffFoundation(card, fIdx, st.foundation) {
			next := copyEightOffState(st)
			next.freeCells[cell] = nil
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = eightOffHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	for fromCol := range EightOffTableauCnt {
		if len(st.tableau[fromCol]) == 0 {
			continue
		}
		card := st.tableau[fromCol][len(st.tableau[fromCol])-1]
		for toCol := range EightOffTableauCnt {
			if toCol == fromCol {
				continue
			}
			if canPlaceOnEightOffTableau(card, toCol, st.tableau) {
				next := copyEightOffState(st)
				next.tableau[fromCol] = next.tableau[fromCol][:len(next.tableau[fromCol])-1]
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = eightOffHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 4. FreeCell -> Tableau
	for cell := range EightOffCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		for toCol := range EightOffTableauCnt {
			if canPlaceOnEightOffTableau(card, toCol, st.tableau) {
				next := copyEightOffState(st)
				next.freeCells[cell] = nil
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = eightOffHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 5. Tableau -> FreeCell
	for col := range EightOffTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		for cell := range EightOffCellCnt {
			if st.freeCells[cell] == nil {
				next := copyEightOffState(st)
				next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
				next.freeCells[cell] = card
				next.g = st.g + 1
				next.h = eightOffHeuristic(next)
				successors = append(successors, next)
				break
			}
		}
	}

	return successors
}

// copyEightOffState creates a deep copy of an eightOffState.
// NOTE: g and h are left at 0 in the copy; the caller must set them before pushing.
func copyEightOffState(st *eightOffState) *eightOffState {
	next := &eightOffState{}
	for i := range EightOffTableauCnt {
		next.tableau[i] = make([]*Card, len(st.tableau[i]))
		copy(next.tableau[i], st.tableau[i])
	}
	next.freeCells = st.freeCells
	next.foundation = st.foundation
	return next
}

func isSolvedEightOffState(st *eightOffState) bool {
	for i := range EightOffFoundationCnt {
		if st.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

// canPlaceOnEightOffTableau implements Eight Off rules: empty column only takes
// a King, otherwise same suit descending by one.
func canPlaceOnEightOffTableau(card *Card, col int, tableau [EightOffTableauCnt][]*Card) bool {
	colCards := tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

func canPlaceOnEightOffFoundation(card *Card, fIdx int, foundation [EightOffFoundationCnt]int) bool {
	count := foundation[fIdx]
	if count == 0 {
		return card.GetValue() == 1
	}
	return card.GetValue() == count+1
}

// eightOffStateKey encodes the board state into a compact key for memoization.
// Each card (identified by (design-1)*13+value-1) maps to a uint16 location.
// Tableau: col*64 + pos + 1 (supports up to 63 cards per column).
// FreeCell: 512 + cell (cell ranges 0..7). Foundation: 0 (default).
func eightOffStateKey(st *eightOffState) [52]uint16 {
	var key [52]uint16
	for col := range EightOffTableauCnt {
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
	for cell := range EightOffCellCnt {
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
