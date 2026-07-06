//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockRookPresenter ルーク(Rook)プレゼンターモック
type MockRookPresenter struct {
	MockGamePresenter[interfaces.RookGame]
}

// HintOutput モック
func (_m *MockRookPresenter) HintOutput(g interfaces.RookGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
