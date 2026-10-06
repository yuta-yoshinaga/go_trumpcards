package domain

// MaxActionLog bounds action history persisted with every Cloudflare Worker KV
// request. Workers save KV on every request and each move uses about 107 bytes.
// Capping only undo history changed the observed 1102 failure from move 117 to
// about 1,100; 200 action-log entries use about 21 KB, keeping persisted logs
// bounded without changing turn numbers below the cap (#10148).
const MaxActionLog = 200

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
	// dropped counts entries trimmed from the front. Only games whose undo
	// snapshots carry a mark persist it in KV: Spiderette, WillOTheWisp,
	// CribbageSquares, PokerSquares, MonteCarlo, and FourteenOut. It resets to 0
	// after restore for all other games.
	dropped int
}

// GetActionLog returns the entries recorded so far, oldest first.
func (b *actionLogBase) GetActionLog() []*ActionLogEntry { return b.actionLog }

// appendLogCode records one action with a locale-independent detail code.
func (b *actionLogBase) appendLogCode(playerIdx int, actionType, detailCode string, detailParams map[string]string, cards []*Card) {
	b.appendLogCodeAt(b.nextTurnNumber(), playerIdx, actionType, detailCode, detailParams, cards)
}

// nextTurnNumber preserves len+1 numbering below the cap. After the cap, or
// after an undo in a game that tracks dropped entries, it continues from the
// last TurnNumber+1. Even when dropped resets to 0, games that do not persist it
// continue from the last number, so numbering remains uninterrupted.
func (b *actionLogBase) nextTurnNumber() int {
	if len(b.actionLog) > 0 && (b.dropped > 0 || len(b.actionLog) >= MaxActionLog) {
		return b.actionLog[len(b.actionLog)-1].TurnNumber + 1
	}
	return len(b.actionLog) + 1
}

// actionLogMark records the absolute log position for an undo snapshot.
func (b *actionLogBase) actionLogMark() int { return b.dropped + len(b.actionLog) }

// truncateActionLog restores an undo snapshot's log position. The mark includes
// dropped entries, so undo removes the right number even when the cap trimmed
// entries from the front after the snapshot was taken. If mark is less than
// dropped (older than every retained entry), the log is cleared; only moves after
// that snapshot remain, so clearing them is intended.
func (b *actionLogBase) truncateActionLog(mark int) {
	keep := mark - b.dropped
	if keep < 0 {
		keep = 0
	}
	if keep > len(b.actionLog) {
		keep = len(b.actionLog)
	}
	b.actionLog = b.actionLog[:keep]
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
	if excess := len(b.actionLog) - MaxActionLog; excess > 0 {
		b.actionLog = b.actionLog[excess:]
		b.dropped += excess
	}
}
