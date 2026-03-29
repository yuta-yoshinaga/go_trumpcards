//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockOhHellPresenter オー・ヘルプレゼンターモック
type MockOhHellPresenter struct {
	MockGamePresenter[interfaces.OhHellGame]
}

// HintOutput モック
func (_m *MockOhHellPresenter) HintOutput(o interfaces.OhHellGame) string {
	ret := _m.Called(o)
	return ret.Get(0).(string)
}
