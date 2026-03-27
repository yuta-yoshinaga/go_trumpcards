//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDoubtPresenter ダウトプレゼンターモック
type MockDoubtPresenter = MockGamePresenter[interfaces.DoubtGame]
