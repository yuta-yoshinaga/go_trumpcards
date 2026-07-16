//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRummy500Presenter Rummy 500プレゼンターモック
type MockRummy500Presenter struct {
	MockGamePresenter[interfaces.Rummy500Game]
}

// HintOutput モック
func (_m *MockRummy500Presenter) HintOutput(g interfaces.Rummy500Game) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
