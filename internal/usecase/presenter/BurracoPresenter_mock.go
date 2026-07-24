//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBurracoPresenter ブラーコプレゼンターモック
type MockBurracoPresenter struct {
	MockGamePresenter[interfaces.BurracoGame]
}

// HintOutput モック
func (_m *MockBurracoPresenter) HintOutput(g interfaces.BurracoGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
