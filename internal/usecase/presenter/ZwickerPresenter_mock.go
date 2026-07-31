//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockZwickerPresenter ツヴィッカー プレゼンターモック
type MockZwickerPresenter struct {
	MockGamePresenter[interfaces.ZwickerGame]
}

// HintOutput モック
func (_m *MockZwickerPresenter) HintOutput(c interfaces.ZwickerGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
