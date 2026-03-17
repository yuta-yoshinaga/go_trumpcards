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
