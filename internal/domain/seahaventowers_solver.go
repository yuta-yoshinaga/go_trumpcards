//go:build !js || !wasm || solo

package domain

import "container/heap"

// SeahavenTowersSolverMaxIterations is the maximum number of states the solver explores.
// Keeps runtime under ~50ms and memory under ~10MB.
const SeahavenTowersSolverMaxIterations = 100000

// seahavenTowersState represents a board state in the A* search.
type seahavenTowersState struct {
	tableau    [SeahavenTowersTableauCnt][]*Card
	freeCells  [SeahavenTowersCellCnt]*Card
	foundation [SeahavenTowersFoundationCnt]int // count per suit
	g          int                              // moves made so far
	h          int                              // heuristic estimate of remaining moves
	index      int                              // index in heap (maintained by container/heap)
}

// seahavenTowersPQ implements heap.Interface for A* priority queue.
type seahavenTowersPQ []*seahavenTowersState

func (pq seahavenTowersPQ) Len() int { return len(pq) }

func (pq seahavenTowersPQ) Less(i, j int) bool {
	return (pq[i].g + pq[i].h) < (pq[j].g + pq[j].h)
}

func (pq seahavenTowersPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *seahavenTowersPQ) Push(x any) {
	n := len(*pq)
	item := x.(*seahavenTowersState)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *seahavenTowersPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// seahavenTowersSolver performs A* search with memoization to determine if a Seahaven Towers board is solvable.
type seahavenTowersSolver struct {
	visited       map[[52]uint16]struct{}
	iterations    int
	maxIterations int
	initialState  *seahavenTowersState
}

func newSeahavenTowersSolver(s *SeahavenTowers) *seahavenTowersSolver {
	solver := &seahavenTowersSolver{
		visited:       make(map[[52]uint16]struct{}),
		maxIterations: SeahavenTowersSolverMaxIterations,
	}
	state := &seahavenTowersState{}
	for i := range SeahavenTowersTableauCnt {
		state.tableau[i] = make([]*Card, len(s.tableau[i]))
		copy(state.tableau[i], s.tableau[i])
	}
	state.freeCells = s.freeCells
	for i := range SeahavenTowersFoundationCnt {
		state.foundation[i] = len(s.foundation[i])
	}
	state.g = 0
	state.h = seahavenTowersHeuristic(state)
	solver.initialState = state
	return solver
}

// seahavenTowersHeuristic returns an admissible estimate of remaining moves —
// the number of cards not yet on a foundation pile.
func seahavenTowersHeuristic(st *seahavenTowersState) int {
	total := 0
	for i := range SeahavenTowersFoundationCnt {
		total += st.foundation[i]
	}
	return 52 - total
}

// isSolvable returns true if the board can be solved, false if proven unsolvable.
// If the iteration limit is exceeded, returns true (unknown = not stalemate).
func (s *seahavenTowersSolver) isSolvable() bool {
	return s.astar()
}

func (s *seahavenTowersSolver) astar() bool {
	if isSeahavenTowersSolvedState(s.initialState) {
		return true
	}

	pq := &seahavenTowersPQ{}
	heap.Init(pq)
	heap.Push(pq, s.initialState)

	key := seahavenTowersStateKeyFromState(s.initialState)
	s.visited[key] = struct{}{}

	for pq.Len() > 0 {
		s.iterations++
		if s.iterations > s.maxIterations {
			return true // unknown = not stalemate
		}

		current := heap.Pop(pq).(*seahavenTowersState)

		successors := s.generateSuccessors(current)
		for _, next := range successors {
			if isSeahavenTowersSolvedState(next) {
				return true
			}
			sk := seahavenTowersStateKeyFromState(next)
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
func (s *seahavenTowersSolver) generateSuccessors(st *seahavenTowersState) []*seahavenTowersState {
	var successors []*seahavenTowersState

	// 1. Tableau -> Foundation
	for col := range SeahavenTowersTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < SeahavenTowersFoundationCnt && seahavenTowersCanPlaceOnFoundation(card, fIdx, st.foundation) {
			next := copySeahavenTowersState(st)
			next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = seahavenTowersHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 2. Reserved cell -> Foundation
	for cell := range SeahavenTowersCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < SeahavenTowersFoundationCnt && seahavenTowersCanPlaceOnFoundation(card, fIdx, st.foundation) {
			next := copySeahavenTowersState(st)
			next.freeCells[cell] = nil
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = seahavenTowersHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	for fromCol := range SeahavenTowersTableauCnt {
		if len(st.tableau[fromCol]) == 0 {
			continue
		}
		card := st.tableau[fromCol][len(st.tableau[fromCol])-1]
		for toCol := range SeahavenTowersTableauCnt {
			if toCol == fromCol {
				continue
			}
			if seahavenTowersCanPlaceOnTableau(card, toCol, st.tableau) {
				next := copySeahavenTowersState(st)
				next.tableau[fromCol] = next.tableau[fromCol][:len(next.tableau[fromCol])-1]
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = seahavenTowersHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 4. Reserved cell -> Tableau
	for cell := range SeahavenTowersCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		for toCol := range SeahavenTowersTableauCnt {
			if seahavenTowersCanPlaceOnTableau(card, toCol, st.tableau) {
				next := copySeahavenTowersState(st)
				next.freeCells[cell] = nil
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = seahavenTowersHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 5. Tableau -> Reserved cell
	for col := range SeahavenTowersTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		for cell := range SeahavenTowersCellCnt {
			if st.freeCells[cell] == nil {
				next := copySeahavenTowersState(st)
				next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
				next.freeCells[cell] = card
				next.g = st.g + 1
				next.h = seahavenTowersHeuristic(next)
				successors = append(successors, next)
				break // Only try one empty reserved cell (equivalent moves)
			}
		}
	}

	return successors
}

// copySeahavenTowersState creates a deep copy of a seahavenTowersState.
// NOTE: g and h are left at 0 in the copy; the caller must set them before pushing to the priority queue.
func copySeahavenTowersState(st *seahavenTowersState) *seahavenTowersState {
	next := &seahavenTowersState{}
	for i := range SeahavenTowersTableauCnt {
		next.tableau[i] = make([]*Card, len(st.tableau[i]))
		copy(next.tableau[i], st.tableau[i])
	}
	next.freeCells = st.freeCells
	next.foundation = st.foundation
	return next
}

func isSeahavenTowersSolvedState(st *seahavenTowersState) bool {
	for i := range SeahavenTowersFoundationCnt {
		if st.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

// seahavenTowersCanPlaceOnTableau enforces the Seahaven Towers stacking rules:
// same-suit descending on non-empty columns, and Kings-only on empty columns.
func seahavenTowersCanPlaceOnTableau(card *Card, col int, tableau [SeahavenTowersTableauCnt][]*Card) bool {
	colCards := tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == topCard.GetValue()-1
}

func seahavenTowersCanPlaceOnFoundation(card *Card, fIdx int, foundation [SeahavenTowersFoundationCnt]int) bool {
	count := foundation[fIdx]
	if count == 0 {
		return card.GetValue() == 1
	}
	return card.GetValue() == count+1
}

// stateKey returns the state key for the solver's initial state (used in tests).
func (s *seahavenTowersSolver) stateKey() [52]uint16 {
	return seahavenTowersStateKeyFromState(s.initialState)
}

// seahavenTowersStateKeyFromState encodes the board state into a compact key for memoization.
// Each card (identified by (design-1)*13+value-1) maps to a uint16 location.
// Tableau: col*64 + pos + 1 (supports up to 63 cards per column).
// Reserved: 640 + cell. Foundation: 0 (default).
//
// The reserved base is 640, not 512 (as FreeCell uses): with 10 columns, the
// tableau range reaches `9*64 + 51 + 1 = 628`, and 512+cell would collide
// with col=8/pos=0 (8*64+0+1 = 513 == 512+1), letting the solver mark
// distinct boards as already-visited and producing false stalemates.
func seahavenTowersStateKeyFromState(st *seahavenTowersState) [52]uint16 {
	var key [52]uint16
	for col := range SeahavenTowersTableauCnt {
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
	for cell := range SeahavenTowersCellCnt {
		if st.freeCells[cell] != nil {
			card := st.freeCells[cell]
			if card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = uint16(640 + cell)
			}
		}
	}
	return key
}
