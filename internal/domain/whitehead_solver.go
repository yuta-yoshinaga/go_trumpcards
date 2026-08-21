//go:build !js || !wasm || solo

package domain

// checkWhiteheadStalemate checks for immediate deadlock in Whitehead.
// Since Whitehead has hidden face-down cards, true solvability cannot be determined.
// Stalemate is declared when: no hint available AND stock is exhausted/cycled without progress.
func (k *Whitehead) checkWhiteheadStalemate() {
	if k.phase != WhiteheadPhasePlaying {
		return
	}
	hint := k.GetHint()
	if hint != nil {
		k.isStalemate = false
		return
	}
	// No moves available
	if len(k.stock) == 0 && len(k.waste) == 0 {
		// Both stock and waste empty, no moves -> stalemate
		k.isStalemate = true
		return
	}
	if len(k.stock) == 0 && k.noProgressCycles >= 1 {
		// Stock exhausted and we've cycled through waste at least once without progress
		k.isStalemate = true
		return
	}
	k.isStalemate = false
}
