//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTarocchiniPresenter タロッキーニのプレゼンターモック
type MockTarocchiniPresenter struct {
	MockGamePresenter[interfaces.TarocchiniGame]
}

// HintOutput モック
func (_m *MockTarocchiniPresenter) HintOutput(g interfaces.TarocchiniGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
