//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHoneymoonBridgePresenter ハネムーンブリッジプレゼンターモック
type MockHoneymoonBridgePresenter struct {
	MockGamePresenter[interfaces.HoneymoonBridgeGame]
}

// HintOutput モック
func (_m *MockHoneymoonBridgePresenter) HintOutput(s interfaces.HoneymoonBridgeGame) string {
	return _m.Called(s).Get(0).(string)
}
