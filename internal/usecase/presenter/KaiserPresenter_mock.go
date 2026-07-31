//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKaiserPresenter カイザー (Kaiser) プレゼンターモック
type MockKaiserPresenter = MockGamePresenter[interfaces.KaiserGame]
