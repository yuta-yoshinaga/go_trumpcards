//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHachiHachiPresenter は八八プレゼンターモック。
type MockHachiHachiPresenter struct {
	MockGamePresenter[interfaces.HachiHachiGame]
}

// HintOutput モック
func (_m *MockHachiHachiPresenter) HintOutput(g interfaces.HachiHachiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
