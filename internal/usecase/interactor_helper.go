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

// resetWithValidatedConfig はConfigのバリデーション→設定→リセット→出力の共通パターンを提供する。
// バリデーション失敗時はエラー出力を返し、成功時はsetConfigを呼んだ後resetAndPresentを実行する。
func resetWithValidatedConfig[G any, C interface{ Validate() error }](
	game G,
	presenter outputPresenter[G],
	cfg C,
	setConfig func(C),
	resetAndPresent func() string,
) string {
	if err := cfg.Validate(); err != nil {
		return presenter.Output(game, err)
	}
	setConfig(cfg)
	return resetAndPresent()
}
