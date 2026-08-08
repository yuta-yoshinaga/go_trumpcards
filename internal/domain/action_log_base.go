package domain

// actionLogBase holds a game's action log and the two operations every game
// performs on it. It is embedded rather than duplicated: 223 games had written
// their own `appendLog`, and 244 their own `GetActionLog`, from the same handful
// of bodies. See issue #5185.
//
// Embedding keeps `g.actionLog` working verbatim at every existing call site —
// promoted fields are readable and assignable — so `Reset` implementations that
// do `g.actionLog = nil`, and the MarshalJSON/UnmarshalJSON pairs that map the
// field to an exported DTO, need no changes. That matters: the action log is
// persisted to KV for the Cloudflare Workers, and a codec change there would
// silently drop history rather than fail loudly.
//
// This base carries the numbered variant, where TurnNumber counts entries.
// Solitaire games number their entries by move count instead and keep their own
// appendLog for now; collapsing those needs a different base, not this one.
type actionLogBase struct {
	actionLog []*ActionLogEntry
}

// GetActionLog returns the entries recorded so far, oldest first.
func (b *actionLogBase) GetActionLog() []*ActionLogEntry { return b.actionLog }

// appendLog records one action. TurnNumber is assigned from the number of
// entries already present, so the first entry is turn 1.
func (b *actionLogBase) appendLog(playerIdx int, actionType, detail string, cards []*Card) {
	b.appendLogAt(len(b.actionLog)+1, playerIdx, actionType, detail, cards)
}

// appendLogAt records one action under a caller-supplied turn number, for games
// that number entries by something other than the entry count. The solitaires
// number by move count, which lives on the game struct and so is not visible
// from here — they pass it in rather than each keeping a copy of this body.
func (b *actionLogBase) appendLogAt(turnNumber, playerIdx int, actionType, detail string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber: turnNumber,
		PlayerIdx:  playerIdx,
		ActionType: actionType,
		Detail:     detail,
		Cards:      cards,
	})
}
