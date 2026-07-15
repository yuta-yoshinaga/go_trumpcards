//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSevenBridgePresenter セブンブリッジプレゼンターモック
type MockSevenBridgePresenter struct {
	MockGamePresenter[interfaces.SevenBridgeGame]
}

// HintOutput モック
func (_m *MockSevenBridgePresenter) HintOutput(g interfaces.SevenBridgeGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
