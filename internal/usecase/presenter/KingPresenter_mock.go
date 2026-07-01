//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockKingPresenter はキングプレゼンターモック。
type MockKingPresenter struct {
	MockGamePresenter[interfaces.KingGame]
}

// HintOutput モック
func (_m *MockKingPresenter) HintOutput(g interfaces.KingGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
