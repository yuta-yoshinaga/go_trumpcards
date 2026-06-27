//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockThreeCardPresenter スリーカードポーカープレゼンターモック
type MockThreeCardPresenter struct {
	MockGamePresenter[interfaces.ThreeCardGame]
}

// HintOutput モック
func (_m *MockThreeCardPresenter) HintOutput(g interfaces.ThreeCardGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
