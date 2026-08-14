//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGoofspielPresenter ゴフスピールプレゼンターモック
type MockGoofspielPresenter struct {
	MockGamePresenter[interfaces.GoofspielGame]
}

// HintOutput モック
func (_m *MockGoofspielPresenter) HintOutput(s interfaces.GoofspielGame) string {
	return _m.Called(s).Get(0).(string)
}
