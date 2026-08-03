//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLobaPresenter ロバ プレゼンターモック
type MockLobaPresenter struct {
	MockGamePresenter[interfaces.LobaGame]
}

// HintOutput モック
func (_m *MockLobaPresenter) HintOutput(c interfaces.LobaGame) string {
	ret := _m.Called(c)
	return ret.Get(0).(string)
}
