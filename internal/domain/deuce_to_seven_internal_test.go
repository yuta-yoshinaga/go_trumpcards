package domain

// Internal-package test helpers for DeuceToSeven that need access to private
// fields. Mirrors the shape of badugi_internal_test.go.

func (d *DeuceToSeven) setActedFlags(flags []bool)   { d.round.actedFlags = flags }
func (d *DeuceToSeven) setRaiseCount(count int)      { d.round.raiseCount = count }
func (d *DeuceToSeven) setStartingChips(chips []int) { d.round.startingChips = chips }
