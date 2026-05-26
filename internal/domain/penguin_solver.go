package domain

import "container/heap"

// PenguinSolverMaxIterations is the maximum number of states the solver explores.
const PenguinSolverMaxIterations = 100000

type penguinState struct {
	tableau    [PenguinTableauCnt][]*Card
	freeCells  [PenguinCellCnt]*Card
	foundation [PenguinFoundationCnt]int // count per suit
	g          int
	h          int
	index      int
}

type penguinPQ []*penguinState

func (pq penguinPQ) Len() int { return len(pq) }

func (pq penguinPQ) Less(i, j int) bool {
	return (pq[i].g + pq[i].h) < (pq[j].g + pq[j].h)
}

func (pq penguinPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *penguinPQ) Push(x any) {
	n := len(*pq)
	item := x.(*penguinState)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *penguinPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

type penguinSolver struct {
	visited       map[[52]uint16]struct{}
	iterations    int
	maxIterations int
	initialState  *penguinState
	baseRank      int
}

func newPenguinSolver(p *Penguin) *penguinSolver {
	s := &penguinSolver{
		visited:       make(map[[52]uint16]struct{}),
		maxIterations: PenguinSolverMaxIterations,
		baseRank:      p.baseRank,
	}
	state := &penguinState{}
	for i := range PenguinTableauCnt {
		state.tableau[i] = make([]*Card, len(p.tableau[i]))
		copy(state.tableau[i], p.tableau[i])
	}
	state.freeCells = p.freeCells
	for i := range PenguinFoundationCnt {
		state.foundation[i] = len(p.foundation[i])
	}
	state.g = 0
	state.h = penguinHeuristic(state)
	s.initialState = state
	return s
}

func penguinHeuristic(st *penguinState) int {
	total := 0
	for i := range PenguinFoundationCnt {
		total += st.foundation[i]
	}
	return 52 - total
}

func (s *penguinSolver) isSolvable() bool {
	return s.astar()
}

func (s *penguinSolver) astar() bool {
	if isSolvedPenguinState(s.initialState) {
		return true
	}

	pq := &penguinPQ{}
	heap.Init(pq)
	heap.Push(pq, s.initialState)

	key := penguinStateKey(s.initialState)
	s.visited[key] = struct{}{}

	for pq.Len() > 0 {
		s.iterations++
		if s.iterations > s.maxIterations {
			return true
		}

		current := heap.Pop(pq).(*penguinState)

		successors := s.generateSuccessors(current)
		for _, next := range successors {
			if isSolvedPenguinState(next) {
				return true
			}
			sk := penguinStateKey(next)
			if _, ok := s.visited[sk]; ok {
				continue
			}
			s.visited[sk] = struct{}{}
			heap.Push(pq, next)
		}
	}

	return false
}

func (s *penguinSolver) generateSuccessors(st *penguinState) []*penguinState {
	var successors []*penguinState

	// 1. Tableau -> Foundation
	for col := range PenguinTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < PenguinFoundationCnt && canPlaceOnPenguinFoundation(card, fIdx, st.foundation, s.baseRank) {
			next := copyPenguinState(st)
			next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = penguinHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 2. FreeCell -> Foundation
	for cell := range PenguinCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < PenguinFoundationCnt && canPlaceOnPenguinFoundation(card, fIdx, st.foundation, s.baseRank) {
			next := copyPenguinState(st)
			next.freeCells[cell] = nil
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = penguinHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	emptyColRank := penguinPrevRank(s.baseRank)
	for fromCol := range PenguinTableauCnt {
		if len(st.tableau[fromCol]) == 0 {
			continue
		}
		card := st.tableau[fromCol][len(st.tableau[fromCol])-1]
		for toCol := range PenguinTableauCnt {
			if toCol == fromCol {
				continue
			}
			if canPlaceOnPenguinTableau(card, toCol, st.tableau, emptyColRank) {
				next := copyPenguinState(st)
				next.tableau[fromCol] = next.tableau[fromCol][:len(next.tableau[fromCol])-1]
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = penguinHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 4. FreeCell -> Tableau
	for cell := range PenguinCellCnt {
		if st.freeCells[cell] == nil {
			continue
		}
		card := st.freeCells[cell]
		for toCol := range PenguinTableauCnt {
			if canPlaceOnPenguinTableau(card, toCol, st.tableau, emptyColRank) {
				next := copyPenguinState(st)
				next.freeCells[cell] = nil
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = penguinHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 5. Tableau -> FreeCell
	for col := range PenguinTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		for cell := range PenguinCellCnt {
			if st.freeCells[cell] == nil {
				next := copyPenguinState(st)
				next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
				next.freeCells[cell] = card
				next.g = st.g + 1
				next.h = penguinHeuristic(next)
				successors = append(successors, next)
				break
			}
		}
	}

	return successors
}

func copyPenguinState(st *penguinState) *penguinState {
	next := &penguinState{}
	for i := range PenguinTableauCnt {
		next.tableau[i] = make([]*Card, len(st.tableau[i]))
		copy(next.tableau[i], st.tableau[i])
	}
	next.freeCells = st.freeCells
	next.foundation = st.foundation
	return next
}

func isSolvedPenguinState(st *penguinState) bool {
	for i := range PenguinFoundationCnt {
		if st.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

// penguinPrevRank is the package-level version for use in the solver.
func penguinPrevRank(r int) int {
	return ((r + 11) % 13) + 1
}

// penguinNextRank is the package-level version for use in the solver.
func penguinNextRank(r int) int {
	return (r % 13) + 1
}

func canPlaceOnPenguinTableau(card *Card, col int, tableau [PenguinTableauCnt][]*Card, emptyColRank int) bool {
	colCards := tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == emptyColRank
	}
	topCard := colCards[len(colCards)-1]
	return card.GetDesign() == topCard.GetDesign() && card.GetValue() == penguinPrevRank(topCard.GetValue())
}

func canPlaceOnPenguinFoundation(card *Card, fIdx int, foundation [PenguinFoundationCnt]int, baseRank int) bool {
	count := foundation[fIdx]
	if count == 0 {
		return card.GetValue() == baseRank
	}
	expectedRank := baseRank
	for i := 0; i < count; i++ {
		expectedRank = penguinNextRank(expectedRank)
	}
	return card.GetValue() == expectedRank
}

func penguinStateKey(st *penguinState) [52]uint16 {
	var key [52]uint16
	for col := range PenguinTableauCnt {
		for pos, card := range st.tableau[col] {
			if card == nil || card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = uint16(col*64 + pos + 1)
			}
		}
	}
	for cell := range PenguinCellCnt {
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
