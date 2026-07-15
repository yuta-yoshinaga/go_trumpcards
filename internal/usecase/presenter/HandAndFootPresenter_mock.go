//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockHandAndFootPresenter ハンドアンドフットプレゼンターモック
type MockHandAndFootPresenter struct {
	MockGamePresenter[interfaces.HandAndFootGame]
}

// HintOutput モック
func (_m *MockHandAndFootPresenter) HintOutput(g interfaces.HandAndFootGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
