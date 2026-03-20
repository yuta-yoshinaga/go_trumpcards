package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOmahaPresenter オマハホールデムプレゼンターモック
type MockOmahaPresenter = MockGamePresenter[interfaces.OmahaGame]
