//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBuraPresenter ブラ プレゼンターモック
type MockBuraPresenter struct {
	MockGamePresenter[interfaces.BuraGame]
}
