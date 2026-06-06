//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTressettePresenter トレセッテのプレゼンターモック
type MockTressettePresenter struct {
	MockGamePresenter[interfaces.TressetteGame]
}

// HintOutput モック
func (_m *MockTressettePresenter) HintOutput(g interfaces.TressetteGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
