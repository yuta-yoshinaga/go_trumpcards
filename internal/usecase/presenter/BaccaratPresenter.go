package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// BaccaratPresenter バカラプレゼンターインタフェース
type BaccaratPresenter interface {
	Output(b interfaces.BaccaratGame, lastErr error) string
	ActionLogOutput(b interfaces.BaccaratGame) string
}
