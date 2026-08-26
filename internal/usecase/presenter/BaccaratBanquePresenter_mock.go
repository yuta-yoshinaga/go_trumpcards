//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBaccaratBanquePresenter はバカラ・バンクのプレゼンターモック。
type MockBaccaratBanquePresenter struct {
	MockGamePresenter[interfaces.BaccaratBanqueGame]
}

// HintOutput モック
func (_m *MockBaccaratBanquePresenter) HintOutput(g interfaces.BaccaratBanqueGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
