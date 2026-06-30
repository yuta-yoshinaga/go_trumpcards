//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockJassPresenter ヤスプレゼンターモック
type MockJassPresenter struct {
	MockGamePresenter[interfaces.JassGame]
}

// HintOutput モック
func (_m *MockJassPresenter) HintOutput(g interfaces.JassGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
