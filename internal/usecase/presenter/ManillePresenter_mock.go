//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockManillePresenter マニーユのプレゼンターモック
type MockManillePresenter struct {
	MockGamePresenter[interfaces.ManilleGame]
}

// HintOutput モック
func (_m *MockManillePresenter) HintOutput(g interfaces.ManilleGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
