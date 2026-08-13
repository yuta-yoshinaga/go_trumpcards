//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockZwanzigerrufenPresenter ツヴァンツィガールーフェンのプレゼンターモック。
type MockZwanzigerrufenPresenter struct {
	MockGamePresenter[interfaces.ZwanzigerrufenGame]
}

// HintOutput モック
func (_m *MockZwanzigerrufenPresenter) HintOutput(g interfaces.ZwanzigerrufenGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
