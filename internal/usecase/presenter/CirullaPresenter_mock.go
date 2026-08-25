//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockCirullaPresenter はチルッラのプレゼンターモック。
type MockCirullaPresenter struct {
	MockGamePresenter[interfaces.CirullaGame]
}

// HintOutput モック
func (_m *MockCirullaPresenter) HintOutput(g interfaces.CirullaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
