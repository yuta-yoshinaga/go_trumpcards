//go:build test

package domain

// stateKey returns the state key for the solver's initial state (used in tests).
func (s *freeCellSolver) stateKey() [52]uint16 {
	return stateKeyFromState(s.initialState)
}
