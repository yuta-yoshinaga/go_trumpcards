//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSuecaPresenter スエカのプレゼンターモック
type MockSuecaPresenter struct {
	MockGamePresenter[interfaces.SuecaGame]
}

// HintOutput モック
func (_m *MockSuecaPresenter) HintOutput(g interfaces.SuecaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
