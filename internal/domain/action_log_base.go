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
type actionLogBase struct {
	actionLog []*ActionLogEntry
}

// GetActionLog returns the entries recorded so far, oldest first.
func (b *actionLogBase) GetActionLog() []*ActionLogEntry { return b.actionLog }

// appendLogCode records one action with a locale-independent detail code.
func (b *actionLogBase) appendLogCode(playerIdx int, actionType, detailCode string, detailParams map[string]string, cards []*Card) {
	b.appendLogCodeAt(len(b.actionLog)+1, playerIdx, actionType, detailCode, detailParams, cards)
}

// appendLogCodeAt records one action with a caller-supplied turn number and a locale-independent detail code.
func (b *actionLogBase) appendLogCodeAt(turnNumber, playerIdx int, actionType, detailCode string, detailParams map[string]string, cards []*Card) {
	b.actionLog = append(b.actionLog, &ActionLogEntry{
		TurnNumber:   turnNumber,
		PlayerIdx:    playerIdx,
		ActionType:   actionType,
		DetailCode:   detailCode,
		DetailParams: detailParams,
		Cards:        cards,
	})
}
