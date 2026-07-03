//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGutsPresenter はガッツ (Guts) プレゼンターモック。
type MockGutsPresenter struct {
	MockGamePresenter[interfaces.GutsGame]
}

// HintOutput モック
func (_m *MockGutsPresenter) HintOutput(g interfaces.GutsGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
