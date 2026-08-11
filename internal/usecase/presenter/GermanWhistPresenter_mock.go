//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockGermanWhistPresenter ジャーマンホイストプレゼンターモック
type MockGermanWhistPresenter struct {
	MockGamePresenter[interfaces.GermanWhistGame]
}

// HintOutput モック
func (_m *MockGermanWhistPresenter) HintOutput(g interfaces.GermanWhistGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
