//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockScoponePresenter スコポーネプレゼンターモック。
type MockScoponePresenter struct {
	MockGamePresenter[interfaces.ScoponeGame]
}

// HintOutput モック
func (_m *MockScoponePresenter) HintOutput(g interfaces.ScoponeGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
