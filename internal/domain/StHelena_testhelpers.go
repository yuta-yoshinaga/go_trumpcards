//go:build test

package domain

// StHelenaFoundationSuit returns the suit assigned to a foundation.
func StHelenaFoundationSuit(fIdx int) int { return stHelenaFoundationSuit(fIdx) }

// StHelenaIsAscendingFoundation reports whether a foundation is ascending.
func StHelenaIsAscendingFoundation(fIdx int) bool { return fIdx < StHelenaAscendingFoundationCnt }
