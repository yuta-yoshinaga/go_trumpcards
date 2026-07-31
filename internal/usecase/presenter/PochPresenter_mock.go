//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockPochPresenter ポッホ プレゼンターモック
type MockPochPresenter struct {
	MockGamePresenter[interfaces.PochGame]
}

// HintOutput モック
func (_m *MockPochPresenter) HintOutput(c interfaces.PochGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
