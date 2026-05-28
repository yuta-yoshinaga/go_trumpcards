//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockDoudizhuPresenter 斗地主プレゼンターモック
type MockDoudizhuPresenter = MockGamePresenter[interfaces.DoudizhuGame]
