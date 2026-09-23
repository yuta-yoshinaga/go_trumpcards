//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockIndianRummyPresenter インドラミープレゼンターモック
type MockIndianRummyPresenter struct {
	MockGamePresenter[interfaces.IndianRummyGame]
}

// HintOutput モック
func (_m *MockIndianRummyPresenter) HintOutput(g interfaces.IndianRummyGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
