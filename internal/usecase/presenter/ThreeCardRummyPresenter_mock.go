//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThreeCardRummyPresenter スリーカード・ラミープレゼンターモック
type MockThreeCardRummyPresenter struct {
	MockGamePresenter[interfaces.ThreeCardRummyGame]
}

// HintOutput モック
func (_m *MockThreeCardRummyPresenter) HintOutput(g interfaces.ThreeCardRummyGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
