//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSlobberhannesPresenter スロバーハンネスプレゼンターモック
type MockSlobberhannesPresenter struct {
	MockGamePresenter[interfaces.SlobberhannesGame]
}

// HintOutput モック
func (_m *MockSlobberhannesPresenter) HintOutput(s interfaces.SlobberhannesGame) string {
	return _m.Called(s).Get(0).(string)
}
