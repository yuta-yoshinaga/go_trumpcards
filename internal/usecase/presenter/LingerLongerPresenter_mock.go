//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLingerLongerPresenter リンガーロンガープレゼンターモック
type MockLingerLongerPresenter struct {
	MockGamePresenter[interfaces.LingerLongerGame]
}

// HintOutput モック
func (_m *MockLingerLongerPresenter) HintOutput(s interfaces.LingerLongerGame) string {
	return _m.Called(s).Get(0).(string)
}
