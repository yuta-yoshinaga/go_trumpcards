//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKnockoutWhistPresenter ノックアウト・ホイストのプレゼンターモック
type MockKnockoutWhistPresenter struct {
	MockGamePresenter[interfaces.KnockoutWhistGame]
}

// HintOutput モック
func (_m *MockKnockoutWhistPresenter) HintOutput(g interfaces.KnockoutWhistGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
