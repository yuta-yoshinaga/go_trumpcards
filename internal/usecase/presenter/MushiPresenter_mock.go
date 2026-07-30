//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockMushiPresenter 虫 プレゼンターモック
type MockMushiPresenter struct {
	MockGamePresenter[interfaces.MushiGame]
}

// HintOutput モック
func (_m *MockMushiPresenter) HintOutput(m interfaces.MushiGame) string {
	ret := _m.Called(m)
	return ret.Get(0).(string)
}
