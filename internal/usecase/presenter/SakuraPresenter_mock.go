//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockSakuraPresenter はさくらプレゼンターモック。
type MockSakuraPresenter struct {
	MockGamePresenter[interfaces.SakuraGame]
}

// HintOutput モック
func (_m *MockSakuraPresenter) HintOutput(g interfaces.SakuraGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
