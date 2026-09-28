package domain

// restoreGamePlayer returns the snapshot's GamePlayer, or a fresh non-human
// one when the snapshot had none (older sessions / hand-built fixtures).
func restoreGamePlayer(gp *GamePlayer) *GamePlayer {
	if gp != nil {
		return gp
	}
	return NewGamePlayer(false)
}

// assignIfSet copies *src into *dst when the snapshot carried the field and
// leaves *dst untouched otherwise.
func assignIfSet[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}
