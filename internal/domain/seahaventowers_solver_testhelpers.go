//go:build test

package domain

// stateKey returns the state key for the solver's initial state (used in tests).
func (s *seahavenTowersSolver) stateKey() [52]uint16 {
	return seahavenTowersStateKeyFromState(s.initialState)
}
