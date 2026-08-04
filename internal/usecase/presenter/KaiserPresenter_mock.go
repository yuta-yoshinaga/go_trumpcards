//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKaiserPresenter カイザー (Kaiser) プレゼンターモック
type MockKaiserPresenter struct {
	MockGamePresenter[interfaces.KaiserGame]
}

// HintOutput モック
func (_m *MockKaiserPresenter) HintOutput(g interfaces.KaiserGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
