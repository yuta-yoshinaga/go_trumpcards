package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBaccaratPresenter バカラプレゼンターモック
type MockBaccaratPresenter = MockGamePresenter[interfaces.BaccaratGame]
