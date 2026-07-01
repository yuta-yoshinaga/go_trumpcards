//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockLooPresenter はループレゼンターモック。
type MockLooPresenter struct {
	MockGamePresenter[interfaces.LooGame]
}

// HintOutput モック
func (_m *MockLooPresenter) HintOutput(g interfaces.LooGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
