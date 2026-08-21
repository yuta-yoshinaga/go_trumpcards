//go:build !js || !wasm || solo

package domain

import "container/heap"

// StalactitesSolverMaxIterations is the maximum number of states the solver explores.
// Keeps runtime under ~50ms and memory under ~10MB.
const StalactitesSolverMaxIterations = 100000

// stalactitesState represents a board state in the A* search.
type stalactitesState struct {
	tableau    [StalactitesTableauCnt][]*Card
	cells      [StalactitesCellCnt]*Card
	foundation [StalactitesFoundationCnt]int // count per suit
	g          int                           // moves made so far
	h          int                           // stalactitesHeuristic estimate of remaining moves
	index      int                           // index in heap (maintained by container/heap)
}

// stalactitesPQ implements heap.Interface for A* priority queue.
type stalactitesPQ []*stalactitesState

func (pq stalactitesPQ) Len() int { return len(pq) }

func (pq stalactitesPQ) Less(i, j int) bool {
	return (pq[i].g + pq[i].h) < (pq[j].g + pq[j].h)
}

func (pq stalactitesPQ) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *stalactitesPQ) Push(x any) {
	n := len(*pq)
	item := x.(*stalactitesState)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *stalactitesPQ) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[:n-1]
	return item
}

// stalactitesSolver performs A* search with memoization to determine if a Stalactites board is solvable.
type stalactitesSolver struct {
	visited       map[[52]uint16]struct{}
	iterations    int
	maxIterations int
	initialState  *stalactitesState
	// baseRank はファンデーションの開始ランク。**これを持たないと、ソルバは
	// FreeCell の規則（スート別パイル・A 始まり）で解けるかを判定してしまう。**
	// baseRank が A になるのは 13 回に 1 回なので、大半の配りで手詰まり判定が
	// 別のゲームの規則で計算されることになる。
	baseRank int
}

func newStalactitesSolver(f *Stalactites) *stalactitesSolver {
	s := &stalactitesSolver{
		visited:       make(map[[52]uint16]struct{}),
		maxIterations: StalactitesSolverMaxIterations,
		baseRank:      f.baseRank,
	}
	state := &stalactitesState{}
	// Deep copy tableau
	for i := range StalactitesTableauCnt {
		state.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(state.tableau[i], f.tableau[i])
	}
	// Copy cells
	state.cells = f.cells
	// Foundation: only need count per suit
	for i := range StalactitesFoundationCnt {
		state.foundation[i] = len(f.foundation[i])
	}
	state.g = 0
	state.h = stalactitesHeuristic(state)
	s.initialState = state
	return s
}

// stalactitesHeuristic returns an admissible estimate of remaining moves.
// It counts cards not yet on foundation (each needs at least 1 move).
func stalactitesHeuristic(st *stalactitesState) int {
	total := 0
	for i := range StalactitesFoundationCnt {
		total += st.foundation[i]
	}
	return 52 - total
}

// isSolvable returns true if the board can be solved, false if proven unsolvable.
// If the iteration limit is exceeded, returns true (unknown = not stalemate).
func (s *stalactitesSolver) isSolvable() bool {
	return s.astar()
}

func (s *stalactitesSolver) astar() bool {
	// Check if already solved
	if stalactitesIsSolvedState(s.initialState) {
		return true
	}

	pq := &stalactitesPQ{}
	heap.Init(pq)
	heap.Push(pq, s.initialState)

	key := stalactitesStateKeyFromState(s.initialState)
	s.visited[key] = struct{}{}

	for pq.Len() > 0 {
		s.iterations++
		if s.iterations > s.maxIterations {
			return true // unknown = not stalemate
		}

		current := heap.Pop(pq).(*stalactitesState)

		// Generate all successor states
		successors := s.generateSuccessors(current)
		for _, next := range successors {
			if stalactitesIsSolvedState(next) {
				return true
			}
			sk := stalactitesStateKeyFromState(next)
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
func (s *stalactitesSolver) generateSuccessors(st *stalactitesState) []*stalactitesState {
	var successors []*stalactitesState

	// 1. Tableau -> Foundation
	for col := range StalactitesTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		fIdx := stalactitesFoundationIndexFor(card, st.foundation, s.baseRank)
		if fIdx >= 0 {
			next := stalactitesCopyState(st)
			next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = stalactitesHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 2. Stalactites -> Foundation
	for cell := range StalactitesCellCnt {
		if st.cells[cell] == nil {
			continue
		}
		card := st.cells[cell]
		fIdx := stalactitesFoundationIndexFor(card, st.foundation, s.baseRank)
		if fIdx >= 0 {
			next := stalactitesCopyState(st)
			next.cells[cell] = nil
			next.foundation[fIdx]++
			next.g = st.g + 1
			next.h = stalactitesHeuristic(next)
			successors = append(successors, next)
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	for fromCol := range StalactitesTableauCnt {
		if len(st.tableau[fromCol]) == 0 {
			continue
		}
		card := st.tableau[fromCol][len(st.tableau[fromCol])-1]
		for toCol := range StalactitesTableauCnt {
			if toCol == fromCol {
				continue
			}
			if s.canPlaceOnTableau(card, toCol, st.tableau) {
				next := stalactitesCopyState(st)
				next.tableau[fromCol] = next.tableau[fromCol][:len(next.tableau[fromCol])-1]
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = stalactitesHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 4. Stalactites -> Tableau
	for cell := range StalactitesCellCnt {
		if st.cells[cell] == nil {
			continue
		}
		card := st.cells[cell]
		for toCol := range StalactitesTableauCnt {
			if s.canPlaceOnTableau(card, toCol, st.tableau) {
				next := stalactitesCopyState(st)
				next.cells[cell] = nil
				next.tableau[toCol] = append(next.tableau[toCol], card)
				next.g = st.g + 1
				next.h = stalactitesHeuristic(next)
				successors = append(successors, next)
			}
		}
	}

	// 5. Tableau -> Stalactites
	for col := range StalactitesTableauCnt {
		if len(st.tableau[col]) == 0 {
			continue
		}
		card := st.tableau[col][len(st.tableau[col])-1]
		for cell := range StalactitesCellCnt {
			if st.cells[cell] == nil {
				next := stalactitesCopyState(st)
				next.tableau[col] = next.tableau[col][:len(next.tableau[col])-1]
				next.cells[cell] = card
				next.g = st.g + 1
				next.h = stalactitesHeuristic(next)
				successors = append(successors, next)
				break // Only try one empty free cell (equivalent moves)
			}
		}
	}

	return successors
}

// stalactitesCopyState creates a deep copy of a stalactitesState.
// NOTE: g and h are left at 0 in the copy; the caller must set them before pushing to the priority queue.
func stalactitesCopyState(st *stalactitesState) *stalactitesState {
	next := &stalactitesState{}
	for i := range StalactitesTableauCnt {
		next.tableau[i] = make([]*Card, len(st.tableau[i]))
		copy(next.tableau[i], st.tableau[i])
	}
	next.cells = st.cells
	next.foundation = st.foundation
	return next
}

func stalactitesIsSolvedState(st *stalactitesState) bool {
	for i := range StalactitesFoundationCnt {
		if st.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

// canPlaceOnTableau はバリアント（フリーセル / Baker's Game）に応じてタブロー
// 積み上げ条件を判定する。Baker's Game では同じスートの降順、通常のフリーセルでは
// 赤黒交互の降順を要求する。
func (s *stalactitesSolver) canPlaceOnTableau(card *Card, col int, tableau [StalactitesTableauCnt][]*Card) bool {
	colCards := tableau[col]
	if len(colCards) == 0 {
		// 空列には任意のカードを置ける
		return true
	}
	topCard := colCards[len(colCards)-1]
	if card.GetValue() != topCard.GetValue()-1 {
		return false
	}
	return isAlternateColor(card, topCard)
}

// stalactitesRankAtOffset は開始ランクから n 枚進んだランク。K の次は A。
func stalactitesRankAtOffset(baseRank, n int) int {
	return (baseRank-1+n)%CardValueMax + 1
}

// stalactitesCanPlaceOnFoundation は pile が次に受け取るランクかどうかを見る。
// **スートは見ない。**パイルは「何枚積んだか」だけを持ち、次に必要なランクは
// 開始ランクからの枚数で決まる（K の次は A に戻る）。
func stalactitesCanPlaceOnFoundation(card *Card, fIdx int, foundation [StalactitesFoundationCnt]int, baseRank int) bool {
	return card.GetValue() == stalactitesRankAtOffset(baseRank, foundation[fIdx])
}

// stalactitesFoundationIndexFor はその札を受け取れるパイルを返す（無ければ -1）。
// ドメインの Stalactites.foundationIndexFor と同じ順序: 継続できるパイルを先に
// 探し、無ければ空のパイルを使う。
func stalactitesFoundationIndexFor(card *Card, foundation [StalactitesFoundationCnt]int, baseRank int) int {
	for i := range StalactitesFoundationCnt {
		if foundation[i] > 0 && stalactitesCanPlaceOnFoundation(card, i, foundation, baseRank) {
			return i
		}
	}
	for i := range StalactitesFoundationCnt {
		if foundation[i] == 0 && stalactitesCanPlaceOnFoundation(card, i, foundation, baseRank) {
			return i
		}
	}
	return -1
}

// stateKey returns the state key for the solver's initial state (used in tests).
func (s *stalactitesSolver) stateKey() [52]uint16 {
	return stalactitesStateKeyFromState(s.initialState)
}

// stalactitesStateKeyFromState encodes the board state into a compact key for memoization.
// Each card (identified by (design-1)*13+value-1) maps to a uint16 location.
// Tableau: col*64 + pos + 1 (supports up to 63 cards per column).
// Stalactites: 512 + cell. Foundation: 0 (default).
// Jokers (design=0) are skipped since Stalactites uses only 52 standard cards.
func stalactitesStateKeyFromState(st *stalactitesState) [52]uint16 {
	var key [52]uint16
	for col := range StalactitesTableauCnt {
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
	for cell := range StalactitesCellCnt {
		if st.cells[cell] != nil {
			card := st.cells[cell]
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
