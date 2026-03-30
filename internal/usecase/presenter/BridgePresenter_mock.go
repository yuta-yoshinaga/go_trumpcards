//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBridgePresenter ブリッジプレゼンターモック
type MockBridgePresenter struct {
	MockGamePresenter[interfaces.BridgeGame]
}

// HintOutput モック
func (_m *MockBridgePresenter) HintOutput(b interfaces.BridgeGame) string {
	ret := _m.Called(b)
	return ret.Get(0).(string)
}
