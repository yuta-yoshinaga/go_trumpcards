package domain

// FreeCellSolverMaxIterations is the maximum number of states the solver explores.
// Keeps runtime under ~50ms and memory under ~10MB.
const FreeCellSolverMaxIterations = 100000

// freeCellSolver performs DFS with memoization to determine if a FreeCell board is solvable.
type freeCellSolver struct {
	tableau    [FreeCellTableauCnt][]*Card
	freeCells  [FreeCellCellCnt]*Card
	foundation [FreeCellFoundationCnt]int // track only count per suit (cards are ordered)
	visited    map[[52]byte]struct{}
	iterations int
}

func newFreeCellSolver(f *FreeCell) *freeCellSolver {
	s := &freeCellSolver{
		visited: make(map[[52]byte]struct{}),
	}
	// Deep copy tableau
	for i := 0; i < FreeCellTableauCnt; i++ {
		s.tableau[i] = make([]*Card, len(f.tableau[i]))
		copy(s.tableau[i], f.tableau[i])
	}
	// Copy freeCells
	s.freeCells = f.freeCells
	// Foundation: only need count per suit
	for i := 0; i < FreeCellFoundationCnt; i++ {
		s.foundation[i] = len(f.foundation[i])
	}
	return s
}

// isSolvable returns true if the board can be solved, false if proven unsolvable.
// If the iteration limit is exceeded, returns true (unknown = not stalemate).
func (s *freeCellSolver) isSolvable() bool {
	return s.dfs()
}

func (s *freeCellSolver) dfs() bool {
	// Check if solved
	if s.isSolved() {
		return true
	}

	// Iteration limit
	s.iterations++
	if s.iterations > FreeCellSolverMaxIterations {
		return true // unknown = not stalemate
	}

	// Memoization
	key := s.stateKey()
	if _, ok := s.visited[key]; ok {
		return false
	}
	s.visited[key] = struct{}{}

	// Try all moves in priority order

	// 1. Tableau -> Foundation
	for col := 0; col < FreeCellTableauCnt; col++ {
		if len(s.tableau[col]) == 0 {
			continue
		}
		card := s.tableau[col][len(s.tableau[col])-1]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && s.canPlaceOnFoundation(card, fIdx) {
			s.tableau[col] = s.tableau[col][:len(s.tableau[col])-1]
			s.foundation[fIdx]++
			if s.dfs() {
				s.foundation[fIdx]--
				s.tableau[col] = append(s.tableau[col], card)
				return true
			}
			s.foundation[fIdx]--
			s.tableau[col] = append(s.tableau[col], card)
		}
	}

	// 2. FreeCell -> Foundation
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		if s.freeCells[cell] == nil {
			continue
		}
		card := s.freeCells[cell]
		fIdx := card.GetDesign() - 1
		if fIdx >= 0 && fIdx < FreeCellFoundationCnt && s.canPlaceOnFoundation(card, fIdx) {
			s.freeCells[cell] = nil
			s.foundation[fIdx]++
			if s.dfs() {
				s.foundation[fIdx]--
				s.freeCells[cell] = card
				return true
			}
			s.foundation[fIdx]--
			s.freeCells[cell] = card
		}
	}

	// 3. Tableau -> Tableau (single card only for solver efficiency)
	for fromCol := 0; fromCol < FreeCellTableauCnt; fromCol++ {
		if len(s.tableau[fromCol]) == 0 {
			continue
		}
		card := s.tableau[fromCol][len(s.tableau[fromCol])-1]
		for toCol := 0; toCol < FreeCellTableauCnt; toCol++ {
			if toCol == fromCol {
				continue
			}
			// Skip moving to empty column if card is not King (usually not useful)
			if len(s.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if s.canPlaceOnTableau(card, toCol) {
				s.tableau[fromCol] = s.tableau[fromCol][:len(s.tableau[fromCol])-1]
				s.tableau[toCol] = append(s.tableau[toCol], card)
				if s.dfs() {
					s.tableau[toCol] = s.tableau[toCol][:len(s.tableau[toCol])-1]
					s.tableau[fromCol] = append(s.tableau[fromCol], card)
					return true
				}
				s.tableau[toCol] = s.tableau[toCol][:len(s.tableau[toCol])-1]
				s.tableau[fromCol] = append(s.tableau[fromCol], card)
			}
		}
	}

	// 4. FreeCell -> Tableau
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		if s.freeCells[cell] == nil {
			continue
		}
		card := s.freeCells[cell]
		for toCol := 0; toCol < FreeCellTableauCnt; toCol++ {
			if len(s.tableau[toCol]) == 0 && card.GetValue() != CardValueMax {
				continue
			}
			if s.canPlaceOnTableau(card, toCol) {
				s.freeCells[cell] = nil
				s.tableau[toCol] = append(s.tableau[toCol], card)
				if s.dfs() {
					s.tableau[toCol] = s.tableau[toCol][:len(s.tableau[toCol])-1]
					s.freeCells[cell] = card
					return true
				}
				s.tableau[toCol] = s.tableau[toCol][:len(s.tableau[toCol])-1]
				s.freeCells[cell] = card
			}
		}
	}

	// 5. Tableau -> FreeCell
	for col := 0; col < FreeCellTableauCnt; col++ {
		if len(s.tableau[col]) == 0 {
			continue
		}
		card := s.tableau[col][len(s.tableau[col])-1]
		for cell := 0; cell < FreeCellCellCnt; cell++ {
			if s.freeCells[cell] == nil {
				s.tableau[col] = s.tableau[col][:len(s.tableau[col])-1]
				s.freeCells[cell] = card
				if s.dfs() {
					s.freeCells[cell] = nil
					s.tableau[col] = append(s.tableau[col], card)
					return true
				}
				s.freeCells[cell] = nil
				s.tableau[col] = append(s.tableau[col], card)
				break // Only try one empty free cell (equivalent moves)
			}
		}
	}

	return false
}

func (s *freeCellSolver) isSolved() bool {
	for i := 0; i < FreeCellFoundationCnt; i++ {
		if s.foundation[i] != CardValueMax {
			return false
		}
	}
	return true
}

func (s *freeCellSolver) canPlaceOnTableau(card *Card, col int) bool {
	colCards := s.tableau[col]
	if len(colCards) == 0 {
		return card.GetValue() == CardValueMax
	}
	topCard := colCards[len(colCards)-1]
	return s.isAlternateColor(card, topCard) && card.GetValue() == topCard.GetValue()-1
}

func (s *freeCellSolver) canPlaceOnFoundation(card *Card, fIdx int) bool {
	count := s.foundation[fIdx]
	if count == 0 {
		return card.GetValue() == 1
	}
	// The top card value on foundation[fIdx] is count (since cards go 1,2,3...)
	return card.GetValue() == count+1
}

func (s *freeCellSolver) isAlternateColor(card1, card2 *Card) bool {
	return s.isBlack(card1) != s.isBlack(card2)
}

func (s *freeCellSolver) isBlack(card *Card) bool {
	return card.GetDesign() == CardDesignSpade || card.GetDesign() == CardDesignClover
}

// stateKey encodes the board state into a compact key for memoization.
// Each card (identified by (design-1)*13+value-1) maps to a location byte.
// Jokers (design=0) are skipped since FreeCell normally uses only 52 standard cards.
func (s *freeCellSolver) stateKey() [52]byte {
	var key [52]byte
	// Tableau cards: location = col*16 + position + 1
	for col := 0; col < FreeCellTableauCnt; col++ {
		for pos, card := range s.tableau[col] {
			if card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = byte(col*16 + pos + 1)
			}
		}
	}
	// FreeCell cards: location = 128 + cell
	for cell := 0; cell < FreeCellCellCnt; cell++ {
		if s.freeCells[cell] != nil {
			card := s.freeCells[cell]
			if card.GetDesign() < 1 || card.GetValue() < 1 {
				continue
			}
			idx := (card.GetDesign()-1)*CardValueMax + card.GetValue() - 1
			if idx >= 0 && idx < 52 {
				key[idx] = byte(128 + cell)
			}
		}
	}
	// Foundation cards have location 0 (default)
	return key
}
