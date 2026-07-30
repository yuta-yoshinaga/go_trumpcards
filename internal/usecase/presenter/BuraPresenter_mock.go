//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBuraPresenter ブラ プレゼンターモック
type MockBuraPresenter struct {
	MockGamePresenter[interfaces.BuraGame]
}

// HintOutput モック
func (_m *MockBuraPresenter) HintOutput(b interfaces.BuraGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
