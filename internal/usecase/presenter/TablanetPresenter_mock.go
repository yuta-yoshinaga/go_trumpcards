//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTablanetPresenter はタブラネットプレゼンターモック。
type MockTablanetPresenter struct {
	MockGamePresenter[interfaces.TablanetGame]
}

// HintOutput モック
func (_m *MockTablanetPresenter) HintOutput(g interfaces.TablanetGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
