//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockWillOTheWispPresenter ウィル・オ・ザ・ウィスププレゼンターモック
type MockWillOTheWispPresenter struct {
	MockGamePresenter[interfaces.WillOTheWispGame]
}

// HintOutput モック
func (_m *MockWillOTheWispPresenter) HintOutput(s interfaces.WillOTheWispGame) string {
	ret := _m.Called(s)
	return ret.Get(0).(string)
}
