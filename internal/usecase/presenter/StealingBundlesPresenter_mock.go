//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockStealingBundlesPresenter スティーリングバンドルプレゼンターモック
type MockStealingBundlesPresenter struct {
	MockGamePresenter[interfaces.StealingBundlesGame]
}

// HintOutput モック
func (_m *MockStealingBundlesPresenter) HintOutput(s interfaces.StealingBundlesGame) string {
	return _m.Called(s).Get(0).(string)
}
