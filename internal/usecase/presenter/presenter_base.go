package presenter

// GamePresenter ゲームプレゼンター汎用インタフェース
type GamePresenter[G any] interface {
	// Output ゲーム状態を出力する
	Output(game G, lastErr error) string
	// ActionLogOutput 棋譜を出力する
	ActionLogOutput(game G) string
}
