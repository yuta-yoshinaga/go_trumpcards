package usecase

type playableGame interface {
	gameEndChecker
	IsHumanTurn() bool
}

type gameEndChecker interface {
	GetGameEndFlag() bool
}

type outputPresenter[G any] interface {
	Output(game G, lastErr error) string
}

// guardNotPlayable returns early output if the game has ended or it is not the human's turn.
func guardNotPlayable[G playableGame](game G, p outputPresenter[G]) (string, bool) {
	if game.GetGameEndFlag() || !game.IsHumanTurn() {
		return p.Output(game, nil), true
	}
	return "", false
}

// guardGameEnd returns early output if the game has ended.
func guardGameEnd[G gameEndChecker](game G, p outputPresenter[G]) (string, bool) {
	if game.GetGameEndFlag() {
		return p.Output(game, nil), true
	}
	return "", false
}

// roundAdvancer is a game that can start its next round.
type roundAdvancer interface {
	gameEndChecker
	NextRound()
}

// advanceRound runs the shared "start the next round" step: bail out if the
// game is over, advance, let the CPUs play, then present.
//
// Consolidates 22 byte-identical NextRound implementations across the rummy and
// trick-taking interactors (Burraco, Canasta, Carioca, Chinchon, Conquian,
// ContractRummy, …). They differed only in receiver name and concrete type,
// which is why a name-based search never grouped them — see issue #5368.
//
// runCpu is a callback rather than a method on the interface: each interactor's
// runCpuTurns is unexported, so no interface can name it.
//
// The guard is the reason all 22 bodies were identical: without it a finished
// game deals another round. Order matters too — the CPUs must act on the new
// round, not the one that just ended. Both are pinned by tests.
func advanceRound[G roundAdvancer](game G, p outputPresenter[G], runCpu func()) string {
	if out, blocked := guardGameEnd(game, p); blocked {
		return out
	}
	game.NextRound()
	runCpu()
	return p.Output(game, nil)
}
