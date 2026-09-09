//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockQuinzePresenter カーンズ プレゼンターモック
type MockQuinzePresenter struct {
	MockGamePresenter[interfaces.QuinzeGame]
}
