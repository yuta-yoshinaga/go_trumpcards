package domain

// Internal-package test helpers for Badugi that need access to private
// fields. Mirrors the shape of poker_internal_test.go.

func (b *Badugi) setActedFlags(flags []bool)   { b.round.actedFlags = flags }
func (b *Badugi) setRaiseCount(count int)      { b.round.raiseCount = count }
func (b *Badugi) setStartingChips(chips []int) { b.round.startingChips = chips }
