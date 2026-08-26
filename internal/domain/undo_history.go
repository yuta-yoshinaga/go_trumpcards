package domain

// MaxUndoHistory bounds how many undo snapshots a game keeps.
//
// The snapshots are not just an in-memory convenience: every game's
// MarshalJSON carries them, and the Cloudflare Workers build persists that
// JSON to KV on every request. An unbounded history therefore makes each
// request serialise a payload that grows with the length of the game --
// measured at about 1.8 KB per move for Klondike -- and the Worker is billed
// CPU for marshalling and unmarshalling all of it, twice, every time.
//
// That is not a theoretical ceiling. Played against the deployed solo Worker,
// klondike answered HTTP 503 with Cloudflare error 1102 ("Worker exceeded
// resource limits") on move 117, having reached roughly 217 KB of session
// state; the same session recovered the moment it was reset. 30 snapshots is
// about 59 KB, which leaves roughly four times the headroom of the payload
// that actually failed, and still gives a player more undo depth than any
// solitaire realistically asks for.
const MaxUndoHistory = 30

// appendSnapshot adds snap to a game's undo history, discarding the oldest
// entries once the history is longer than MaxUndoHistory.
//
// It trims down to the cap rather than dropping a single entry, so a session
// restored from a KV value written before the cap existed is brought back
// inside the budget by its next move instead of draining one snapshot at a
// time.
func appendSnapshot[T any](history []T, snap T) []T {
	history = append(history, snap)
	if len(history) > MaxUndoHistory {
		history = history[len(history)-MaxUndoHistory:]
	}
	return history
}
