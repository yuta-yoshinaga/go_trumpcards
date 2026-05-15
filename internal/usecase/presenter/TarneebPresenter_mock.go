//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTarneebPresenter Tarneeb プレゼンターモック
type MockTarneebPresenter struct {
	MockGamePresenter[interfaces.TarneebGame]
}

// HintOutput モック
func (_m *MockTarneebPresenter) HintOutput(t interfaces.TarneebGame) string {
	ret := _m.Called(t)
	return ret.Get(0).(string)
}
