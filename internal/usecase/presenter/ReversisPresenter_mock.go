//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockReversisPresenter レヴェルシプレゼンターモック
type MockReversisPresenter struct {
	MockGamePresenter[interfaces.ReversisGame]
}

// HintOutput モック
func (_m *MockReversisPresenter) HintOutput(r interfaces.ReversisGame) string {
	return _m.Called(r).Get(0).(string)
}
