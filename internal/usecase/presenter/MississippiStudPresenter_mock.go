//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMississippiStudPresenter ミシシッピ・スタッドプレゼンターモック
type MockMississippiStudPresenter = MockGamePresenter[interfaces.MississippiStudGame]
