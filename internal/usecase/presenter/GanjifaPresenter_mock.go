//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGanjifaPresenter ガンジファのプレゼンターモック
type MockGanjifaPresenter struct {
	MockGamePresenter[interfaces.GanjifaGame]
}

// HintOutput モック
func (_m *MockGanjifaPresenter) HintOutput(g interfaces.GanjifaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
