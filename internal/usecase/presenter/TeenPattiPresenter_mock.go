//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockTeenPattiPresenter ティーン・パティのプレゼンターモック
type MockTeenPattiPresenter struct {
	MockGamePresenter[interfaces.TeenPattiGame]
}

// HintOutput モック
func (_m *MockTeenPattiPresenter) HintOutput(g interfaces.TeenPattiGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
