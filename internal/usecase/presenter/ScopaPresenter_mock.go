//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockScopaPresenter スコパプレゼンターモック。
type MockScopaPresenter struct {
	MockGamePresenter[interfaces.ScopaGame]
}

// HintOutput モック
func (_m *MockScopaPresenter) HintOutput(g interfaces.ScopaGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
