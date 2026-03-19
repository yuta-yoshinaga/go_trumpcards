package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSpadesPresenter スペードプレゼンターモック
type MockSpadesPresenter = MockGamePresenter[interfaces.SpadesGame]
