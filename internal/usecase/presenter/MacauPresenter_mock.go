//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMacauPresenter マカオプレゼンターモック
type MockMacauPresenter struct {
	MockGamePresenter[interfaces.MacauGame]
}

// HintOutput モック
func (_m *MockMacauPresenter) HintOutput(g interfaces.MacauGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
