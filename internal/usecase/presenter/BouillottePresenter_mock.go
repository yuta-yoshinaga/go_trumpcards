//go:build test

package presenter

import "github.com/yuta-yoshinaga/go_trumpcards/internal/domain/interfaces"

// MockBouillottePresenter はブイヨット (Bouillotte) プレゼンターモック。
type MockBouillottePresenter struct {
	MockGamePresenter[interfaces.BouillotteGame]
}

// HintOutput モック
func (_m *MockBouillottePresenter) HintOutput(g interfaces.BouillotteGame) string {
	ret := _m.Called(g)
	return ret.Get(0).(string)
}
