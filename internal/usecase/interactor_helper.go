package usecase

// execAndPresent calls an error-returning domain action and presents the result.
func execAndPresent[G any](game G, p outputPresenter[G], action func() error) string {
	err := action()
	return p.Output(game, err)
}

// runAndPresent calls a void domain action and presents the result.
func runAndPresent[G any](game G, p outputPresenter[G], action func()) string {
	action()
	return p.Output(game, nil)
}
