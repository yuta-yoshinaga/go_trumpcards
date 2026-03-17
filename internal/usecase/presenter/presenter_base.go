package presenter

// GamePresenter ゲームプレゼンター汎用インタフェース
type GamePresenter[G any] interface {
	Output(game G, lastErr error) string
	ActionLogOutput(game G) string
}
