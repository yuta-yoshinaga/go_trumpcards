//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBauernschnapsenPresenter バウエルンシュナプセンプレゼンターモック
type MockBauernschnapsenPresenter struct {
	MockGamePresenter[interfaces.BauernschnapsenGame]
}

// HintOutput モック
func (_m *MockBauernschnapsenPresenter) HintOutput(g interfaces.BauernschnapsenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
