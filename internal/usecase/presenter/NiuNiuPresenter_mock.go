//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockNiuNiuPresenter 闘牛 プレゼンターモック
type MockNiuNiuPresenter struct {
	MockGamePresenter[interfaces.NiuNiuGame]
}
