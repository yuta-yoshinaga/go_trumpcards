//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSutdaPresenter はソッタのプレゼンターモック。
type MockSutdaPresenter struct {
	MockGamePresenter[interfaces.SutdaGame]
}

// HintOutput モック
func (_m *MockSutdaPresenter) HintOutput(g interfaces.SutdaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
